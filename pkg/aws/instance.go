package aws

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

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
	if len(result.Reservations) == 0 {
		return nil, fmt.Errorf("No reservation found for id %v", instanceId)
	}
	if len(result.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("No instance found for id %v", instanceId)
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
	if len(result.InstanceStatuses) == 0 {
		return nil, fmt.Errorf("No InstanceStatus available for id %v", instanceId)
	}
	return result.InstanceStatuses[0].AvailabilityZone, nil
}

func (c *AwsClient) mkInstance(ltVersion string, instanceId *string) (*Instance, error) {
	iltver, err := strconv.ParseInt(ltVersion, 10, 64)
	if err != nil {
		return nil, err
	}
	instanceDns, err := c.getInstancePrivateDnsName(instanceId)
	if err != nil {
		fmt.Printf("Error getting Instance DNS name: %v\n", err)
		return nil, err
	}
	instanceAz, err := c.getInstanceAvailabilityZone(instanceId)
	if err != nil {
		fmt.Printf("Error getting Instance availability zone: %v, leaving empty\n", err)
		instanceAz = aws.String("")
	}
	return &Instance{
		InstanceId:             *instanceId,
		InstancePrivateDnsName: *instanceDns,
		LaunchTemplateVersion:  iltver,
		AvailabilityZone:       *instanceAz,
	}, nil
}

func (c *AwsClient) makeInstance(i *autoscaling.Instance) (*Instance, error) {
	return c.mkInstance(*i.LaunchTemplate.Version, i.InstanceId)
}

func (c *AwsClient) makeInstanceFromDetails(i *autoscaling.InstanceDetails) (*Instance, error) {
	return c.mkInstance(*i.LaunchTemplate.Version, i.InstanceId)
}

func (c *AwsClient) makeInstanceFromId(instanceId string) (*Instance, error) {
	result, err := c.AutoScaling.DescribeAutoScalingInstances(&autoscaling.DescribeAutoScalingInstancesInput{
		InstanceIds: []*string{aws.String(instanceId)},
	})
	if err != nil {
		return nil, err
	}
	if len(result.AutoScalingInstances) == 0 {
		return nil, fmt.Errorf("Unable to find AutoScaling instance with id %v", instanceId)
	}

	return c.makeInstanceFromDetails(result.AutoScalingInstances[0])
}

func (c *AwsClient) IdentifyNewInstance(asg *AutoScalingGroup) (*Instance, error) {
	asgSvc := c.AutoScaling
	instances, err := asgSvc.DescribeAutoScalingInstances(&autoscaling.DescribeAutoScalingInstancesInput{})
	if err != nil {
		return nil, err
	}
	prevInstanceIds := map[string]bool{}
	for _, i := range asg.Instances {
		prevInstanceIds[i.InstanceId] = true
	}
	for _, i := range asg.NewInstances {
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
		newInstance, err = c.makeInstanceFromDetails(i)
		if err != nil {
			return nil, err
		}
	}

	if err := c.waitForAllNodesInService(asg.AutoScalingGroupName); err != nil {
		return newInstance, err
	}

	return newInstance, nil
}

func (c *AwsClient) TerminateInstance(instanceId string) error {
	ec2svc := c.EC2

	input := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{aws.String(instanceId)},
	}
	_, err := ec2svc.TerminateInstances(input)
	return err
}
