package aws

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
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
			instance := Instance{
				InstanceId:             *i.InstanceId,
				InstancePrivateDnsName: *instanceDns,
				LaunchTemplateVersion:  iltver,
			}
			agroup.Instances = append(agroup.Instances, instance)
		}
		groups.Groups = append(groups.Groups, agroup)
	}
	return &groups, nil
}
