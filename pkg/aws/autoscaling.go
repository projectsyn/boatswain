package aws

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

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
		latestLt := ltVersions.LaunchTemplateVersions[0]
		agroup.LaunchTemplateVersion = *latestLt.VersionNumber
		agroup.CurrentAmi = *latestLt.LaunchTemplateData.ImageId
		for _, i := range a.Instances {
			instance, err := c.makeInstance(i)
			if err != nil {
				fmt.Printf("Error creating Instance struct for %v, continuing...\n", *i.InstanceId)
				continue
			}
			agroup.Instances = append(agroup.Instances, instance)
		}
		agroup.NewInstances = []*Instance{}
		groups.Groups = append(groups.Groups, &agroup)
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

func (c *AwsClient) DetachNodeFromASG(asg *AutoScalingGroup, instance *Instance) (*autoscaling.Activity, error) {
	asgSvc := c.AutoScaling
	detachInput := &autoscaling.DetachInstancesInput{
		AutoScalingGroupName: aws.String(asg.AutoScalingGroupName),
		InstanceIds:          []*string{aws.String(instance.InstanceId)},
		// don't decrement desired capacity, to immediately create
		// replacement instance
		ShouldDecrementDesiredCapacity: aws.Bool(false),
	}
	detach, err := asgSvc.DetachInstances(detachInput)
	if err != nil {
		return nil, err
	}
	return detach.Activities[0], nil
}

func (c *AwsClient) WaitForASGScaleUp(asg *AutoScalingGroup, activity *autoscaling.Activity) (*Instance, error) {
	asgSvc := c.AutoScaling
	describeActivitiesInput := &autoscaling.DescribeScalingActivitiesInput{
		AutoScalingGroupName: aws.String(asg.AutoScalingGroupName),
		MaxRecords:           aws.Int64(2),
	}
	newInstanceBooting := false
	newInstanceId := ""
	for !newInstanceBooting {
		activities, err := asgSvc.DescribeScalingActivities(describeActivitiesInput)
		if err != nil {
			return nil, err
		}
		for idx, a := range activities.Activities {
			if *a.ActivityId == *activity.ActivityId {
				msg := ""
				if a.StatusMessage != nil {
					msg = *a.StatusMessage
				}
				fmt.Printf("[%d] Detaching activity status: %v (%v)\n", idx, *a.StatusCode, msg)
				if idx != 1 {
					fmt.Println("New instance not created yet")
				} else {
					fmt.Println(*activities.Activities[0].StatusCode)
					if *activities.Activities[0].StatusCode == autoscaling.ScalingActivityStatusCodeSuccessful {
						newInstanceBooting = true
						descParts := strings.Split(*activities.Activities[0].Description, ": ")
						if len(descParts) > 1 {
							newInstanceId = descParts[1]
						}
					}
				}
				break
			}
		}
		if !newInstanceBooting {
			time.Sleep(15 * time.Second)
		}
	}
	if newInstanceId != "" {
		return c.makeInstanceFromId(newInstanceId)
	}
	return nil, nil
}

func (c *AwsClient) ReplaceNodeInASG(asg *AutoScalingGroup, instance *Instance) (*Instance, error) {
	fmt.Println("Detaching node")
	activity, err := c.DetachNodeFromASG(asg, instance)
	if err != nil {
		return nil, err
	}
	fmt.Println(*activity.ActivityId)

	fmt.Println("Waiting for scale-up")
	newInstance, err := c.WaitForASGScaleUp(asg, activity)
	if err != nil {
		return nil, err
	}

	if newInstance == nil {
		fmt.Println("WaitForASGScaleUp did not return Instance, identifying now")
		newInstance, err = c.IdentifyNewInstance(asg)
		if err != nil {
			return nil, err
		}
	}

	// remember new instance
	asg.NewInstances = append(asg.NewInstances, newInstance)

	return newInstance, nil
}
