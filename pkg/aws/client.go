package aws

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func NewAwsClient(roleArn string) *AwsClient {
	c := &AwsClient{}

	c.Session = session.New()
	creds := stscreds.NewCredentials(c.Session, roleArn)
	c.Config = &aws.Config{Credentials: creds}

	c.AutoScaling = autoscaling.New(c.Session, c.Config)
	c.EC2 = ec2.New(c.Session, c.Config)

	return c
}

func (c *AwsClient) getInstancePrivateDnsName(instanceId *string) (*string, error) {
	svc := c.EC2
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			instanceId,
		},
	}

	result, err := svc.DescribeInstances(input)
	if err != nil {
		return nil, err
	}
	return result.Reservations[0].Instances[0].PrivateDnsName, nil
}

func (c *AwsClient) getInstanceAvailabilityZone(instanceId *string) (*string, error) {
	svc := c.EC2
	input := &ec2.DescribeInstanceStatusInput{
		InstanceIds: []*string{
			instanceId,
		},
	}

	result, err := svc.DescribeInstanceStatus(input)
	if err != nil {
		return nil, err
	}
	return result.InstanceStatuses[0].AvailabilityZone, nil
}

func (c *AwsClient) GetAutoScalingGroups() (*AutoScalingGroups, error) {
	asg := c.AutoScaling
	ec2client := c.EC2

	result, err := asg.DescribeAutoScalingGroups(
		&autoscaling.DescribeAutoScalingGroupsInput{})
	if err != nil {
		return nil, err
	}
	groups := AutoScalingGroups{}
	for _, a := range result.AutoScalingGroups {
		if a.LaunchTemplate == nil {
			fmt.Println("Ignoring ASG without launch template")
			continue
		}
		agroup := AutoScalingGroup{
			AutoScalingGroupName: *a.AutoScalingGroupName,
			DesiredCapacity:      *a.DesiredCapacity,
			LaunchTemplateId:     *a.LaunchTemplate.LaunchTemplateId,
			MaxSize:              *a.MaxSize,
		}
		ltVersions, err := ec2client.DescribeLaunchTemplateVersions(
			&ec2.DescribeLaunchTemplateVersionsInput{
				LaunchTemplateId: a.LaunchTemplate.LaunchTemplateId,
			})
		if err != nil {
			fmt.Printf("Error fetching LT versions: %v", err)
			continue
		}
		agroup.LaunchTemplateVersion = *ltVersions.LaunchTemplateVersions[0].VersionNumber
		for _, i := range a.Instances {
			iltver, err := strconv.ParseInt(*i.LaunchTemplate.Version, 10, 64)
			if err != nil {
				// TODO: error handling
			}
			instanceDns, err := c.getInstancePrivateDnsName(i.InstanceId)
			if err != nil {
				fmt.Printf("Error getting Instance DNS name: %v", err)
				// TODO: error handling
			}
			instanceAz, err := c.getInstanceAvailabilityZone(i.InstanceId)
			if err != nil {
				fmt.Printf("Error getting Instance availability zone: %v", err)
				// TODO: error handling
			}
			instance := Instance{
				InstanceId:             *i.InstanceId,
				InstancePrivateDnsName: *instanceDns,
				LaunchTemplateVersion:  iltver,
				AvailabilityZone:       *instanceAz,
			}
			agroup.Instances = append(agroup.Instances, instance)
		}
		groups.Groups = append(groups.Groups, agroup)
	}
	return &groups, nil
}

// wait until ASG is scaled to desired capacity and all instances are in state "InService"
// some logic copied from
// https://github.com/aws/aws-sdk-go/blob/662385f2878df266eb80077fd5554c17534b3864/service/autoscaling/waiters.go#L79-L112
func (c *AwsClient) waitForAllNodesInService(asgName string) error {
	ctx := aws.BackgroundContext()
	w := request.Waiter{
		Name:        "WaitUntilAllInstancesInService",
		MaxAttempts: 40,
		Delay:       request.ConstantWaiterDelay(15 * time.Second),
		Acceptors: []request.WaiterAcceptor{
			{
				State:   request.SuccessWaiterState,
				Matcher: request.PathWaiterMatch, Argument: "contains(AutoScalingGroups[].[length(Instances[?LifecycleState=='InService']) == DesiredCapacity][], `true`)",
				Expected: false,
			},
			{
				State:   request.RetryWaiterState,
				Matcher: request.PathWaiterMatch, Argument: "contains(AutoScalingGroups[].[length(Instances[?LifecycleState=='InService']) == DesiredCapacity][], `false`)",
				Expected: true,
			},
		},
		NewRequest: func(opts []request.Option) (*request.Request, error) {
			input := &autoscaling.DescribeAutoScalingGroupsInput{
				AutoScalingGroupNames: []*string{
					aws.String(asgName),
				},
			}
			req, _ := c.AutoScaling.DescribeAutoScalingGroupsRequest(input)
			req.SetContext(ctx)
			return req, nil
		},
	}
	return w.WaitWithContext(ctx)
}

func (c *AwsClient) ReplaceNodeInASG(asg AutoScalingGroup, replacedInstance Instance) (*Instance, error) {
	fmt.Println("Replacing instance", replacedInstance.InstanceId)
	fmt.Println("AZ", replacedInstance.AvailabilityZone)

	asgSvc := c.AutoScaling
	detachInput := &autoscaling.DetachInstancesInput{
		AutoScalingGroupName: aws.String(asg.AutoScalingGroupName),
		InstanceIds:          []*string{aws.String(replacedInstance.InstanceId)},
		// don't decrement desired capacity, to immediately create
		// replacement instance
		ShouldDecrementDesiredCapacity: aws.Bool(false),
	}
	detach, err := asgSvc.DetachInstances(detachInput)
	if err != nil {
		return nil, err
	}
	fmt.Println(detach)

	// TODO: Wait for new instance provisioning started
	describeActivitiesInput := &autoscaling.DescribeScalingActivitiesInput{
		AutoScalingGroupName: aws.String(asg.AutoScalingGroupName),
		MaxRecords:           aws.Int64(2),
	}
	newInstanceCreating := false
	for !newInstanceCreating {
		activities, err := asgSvc.DescribeScalingActivities(describeActivitiesInput)
		if err != nil {
			return nil, err
		}
		for idx, activity := range activities.Activities {
			if *activity.ActivityId == *detach.Activities[0].ActivityId {
				fmt.Printf("[%d] Detaching activity status: %v\n", idx, *activity.StatusMessage)
				if idx != 1 {
					fmt.Println("New instance not created yet")
				} else {
					newInstanceCreating = true
				}
				break
			}
		}
		if newInstanceCreating {
			fmt.Println(*activities.Activities[1].StatusMessage)
		}
		time.Sleep(15 * time.Second)
	}

	instances, err := asgSvc.DescribeAutoScalingInstances(&autoscaling.DescribeAutoScalingInstancesInput{})
	if err != nil {
		return nil, err
	}
	prevInstanceIds := map[string]bool{}
	for _, i := range asg.Instances {
		prevInstanceIds[i.InstanceId] = true
	}
	fmt.Println(prevInstanceIds)
	var newInstance *Instance
	newInstance = nil
	for _, i := range instances.AutoScalingInstances {
		fmt.Printf("Checking instance %v\n", *i.InstanceId)
		if *i.AutoScalingGroupName != asg.AutoScalingGroupName {
			fmt.Printf("%v not considered due to ASG name %v\n", *i.InstanceId, *i.AutoScalingGroupName)
			continue
		}
		if _, ok := prevInstanceIds[*i.InstanceId]; ok {
			fmt.Printf("%v not considered due to being present in previous set of instances\n", *i.InstanceId)
			continue
		}
		iltver, err := strconv.ParseInt(*i.LaunchTemplate.Version, 10, 64)
		if err != nil {
			// TODO: error handling
		}
		instanceDns, err := c.getInstancePrivateDnsName(i.InstanceId)
		if err != nil {
			fmt.Printf("Error getting Instance DNS name: %v", err)
			// TODO: error handling
		}
		instanceAz, err := c.getInstanceAvailabilityZone(i.InstanceId)
		if err != nil {
			fmt.Printf("Error getting Instance availability zone: %v", err)
			// TODO: error handling
		}
		newInstance = &Instance{
			InstanceId:             *i.InstanceId,
			InstancePrivateDnsName: *instanceDns,
			LaunchTemplateVersion:  iltver,
			AvailabilityZone:       *instanceAz,
		}
	}

	if err := c.waitForAllNodesInService(asg.AutoScalingGroupName); err != nil {
		return newInstance, err
	}

	return newInstance, nil
}

func (c *AwsClient) TerminateInstance(instanceId string) error {
	return errors.New("Not yet implemented")
}
