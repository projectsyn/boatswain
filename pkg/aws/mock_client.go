package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type mockEC2Client struct {
	ec2iface.EC2API
	InstancesResp      []*ec2.DescribeInstancesOutput
	InstanceStatusResp map[string]*ec2.DescribeInstanceStatusOutput
	LTVersionsResp     *ec2.DescribeLaunchTemplateVersionsOutput
}

type mockAutoScalingClient struct {
	autoscalingiface.AutoScalingAPI
	Resp *autoscaling.DescribeAutoScalingGroupsOutput
}

func (c *mockEC2Client) DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	for _, i := range c.InstancesResp {
		if *i.Reservations[0].Instances[0].InstanceId == *input.InstanceIds[0] {
			return i, nil
		}
	}
	return nil, awserr.New("404", "Instance not found", nil)
}

func makeDescribeInstancesOutput(instanceId string, privateDnsName string) *ec2.DescribeInstancesOutput {
	resp := &ec2.DescribeInstancesOutput{}
	reservation := ec2.Reservation{}
	instance := ec2.Instance{
		InstanceId:     aws.String(instanceId),
		PrivateDnsName: aws.String(privateDnsName),
	}

	reservation.Instances = append(reservation.Instances, &instance)
	resp.Reservations = append(resp.Reservations, &reservation)

	return resp
}

func (c *mockEC2Client) DescribeLaunchTemplateVersions(input *ec2.DescribeLaunchTemplateVersionsInput) (*ec2.DescribeLaunchTemplateVersionsOutput, error) {
	return c.LTVersionsResp, nil
}

func makeDescribeLaunchTemplateVersionsOutput(launchTemplateId string, latest int64, imageId string) *ec2.DescribeLaunchTemplateVersionsOutput {
	latestVer := ec2.LaunchTemplateVersion{
		LaunchTemplateId: aws.String(launchTemplateId),
		VersionNumber:    aws.Int64(latest),
		LaunchTemplateData: &ec2.ResponseLaunchTemplateData{
			ImageId: aws.String(imageId),
		},
	}
	return &ec2.DescribeLaunchTemplateVersionsOutput{
		LaunchTemplateVersions: []*ec2.LaunchTemplateVersion{
			&latestVer,
		},
	}
}

func (c *mockEC2Client) DescribeInstanceStatus(input *ec2.DescribeInstanceStatusInput) (*ec2.DescribeInstanceStatusOutput, error) {
	for instanceId, i := range c.InstanceStatusResp {
		if instanceId == *input.InstanceIds[0] {
			return i, nil
		}
	}
	return nil, awserr.New("404", "Instance status not found", nil)
}

func makeDescribeInstanceStatusOutput(az string) *ec2.DescribeInstanceStatusOutput {
	resp := &ec2.DescribeInstanceStatusOutput{}
	status := &ec2.InstanceStatus{
		AvailabilityZone: aws.String(az),
	}
	resp.InstanceStatuses = append(resp.InstanceStatuses, status)
	return resp
}

func (c *mockAutoScalingClient) DescribeAutoScalingGroups(input *autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	return c.Resp, nil
}

func makeDescribeAutoScalingGroupsOutput(launchTemplateId string, ltVersionA int64, ltVersionB int64) *autoscaling.DescribeAutoScalingGroupsOutput {
	instances := []*autoscaling.Instance{
		{
			AvailabilityZone: aws.String("eu-central-1c"),
			HealthStatus:     aws.String("Healthy"),
			InstanceId:       aws.String("i-020969a9bc57c5bb1"),
			InstanceType:     aws.String("m5a.xlarge"),
			LaunchTemplate: &autoscaling.LaunchTemplateSpecification{
				LaunchTemplateId:   aws.String(launchTemplateId),
				LaunchTemplateName: aws.String("amazeeio-test1-workers-120200319135046852500000003"),
				Version:            aws.String(fmt.Sprintf("%v", ltVersionA)),
			},
			LifecycleState:       aws.String("InService"),
			ProtectedFromScaleIn: aws.Bool(false),
		},
		{
			AvailabilityZone: aws.String("eu-central-1a"),
			HealthStatus:     aws.String("Healthy"),
			InstanceId:       aws.String("i-0757e06c74906c484"),
			InstanceType:     aws.String("m5a.large"),
			LaunchTemplate: &autoscaling.LaunchTemplateSpecification{
				LaunchTemplateId:   aws.String(launchTemplateId),
				LaunchTemplateName: aws.String("amazeeio-test1-workers-120200319135046852500000003"),
				Version:            aws.String(fmt.Sprintf("%v", ltVersionA)),
			},
			LifecycleState:       aws.String("InService"),
			ProtectedFromScaleIn: aws.Bool(false),
		},
		{
			AvailabilityZone: aws.String("eu-central-1b"),
			HealthStatus:     aws.String("Healthy"),
			InstanceId:       aws.String("i-07d0d125bef27b567"),
			InstanceType:     aws.String("m5a.xlarge"),
			LaunchTemplate: &autoscaling.LaunchTemplateSpecification{
				LaunchTemplateId:   aws.String(launchTemplateId),
				LaunchTemplateName: aws.String("amazeeio-test1-workers-120200319135046852500000003"),
				Version:            aws.String(fmt.Sprintf("%v", ltVersionB)),
			},
			LifecycleState:       aws.String("InService"),
			ProtectedFromScaleIn: aws.Bool(false),
		},
	}
	lt := autoscaling.LaunchTemplateSpecification{
		LaunchTemplateId:   aws.String(launchTemplateId),
		LaunchTemplateName: aws.String("amazeeio-test1-workers-120200319135046852500000003"),
		Version:            aws.String("$Latest"),
	}
	group := autoscaling.Group{
		AutoScalingGroupName: aws.String("amazeeio-test1-workers-120200319135047187100000005"),
		DesiredCapacity:      aws.Int64(3),
		LaunchTemplate:       &lt,
		MaxSize:              aws.Int64(6),
		MinSize:              aws.Int64(3),
		Instances:            instances,
	}
	asgout := autoscaling.DescribeAutoScalingGroupsOutput{}
	asgout.AutoScalingGroups = append(asgout.AutoScalingGroups, &group)

	return &asgout
}

func makeAutoScalingGroupInstances() []*ec2.DescribeInstancesOutput {
	return []*ec2.DescribeInstancesOutput{
		makeDescribeInstancesOutput("i-020969a9bc57c5bb1", "ip-10-200-164-185.eu-central-1.compute.internal"),
		makeDescribeInstancesOutput("i-07d0d125bef27b567", "ip-10-200-105-130.eu-central-1.compute.internal"),
		makeDescribeInstancesOutput("i-0757e06c74906c484", "ip-10-200-62-98.eu-central-1.compute.internal"),
	}

}

func makeAutoScalingGroupInstancesStatus() map[string]*ec2.DescribeInstanceStatusOutput {
	return map[string]*ec2.DescribeInstanceStatusOutput{
		"i-020969a9bc57c5bb1": makeDescribeInstanceStatusOutput("eu-central-1a"),
		"i-07d0d125bef27b567": makeDescribeInstanceStatusOutput("eu-central-1b"),
		"i-0757e06c74906c484": makeDescribeInstanceStatusOutput("eu-central-1c"),
	}
}
