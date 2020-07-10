package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

type AwsClient struct {
	Session     *session.Session
	Config      *aws.Config
	EC2         ec2iface.EC2API
	AutoScaling autoscalingiface.AutoScalingAPI
	SSM         ssmiface.SSMAPI
}

type AutoScalingGroups struct {
	Groups []*AutoScalingGroup
}

type AutoScalingGroup struct {
	AutoScalingGroupName  string
	DesiredCapacity       int64
	LaunchTemplateId      string
	LaunchTemplateVersion int64
	MaxSize               int64
	CurrentAmi            string
	Instances             []*Instance
	NewInstances          []*Instance
}

type Instance struct {
	InstanceId             string
	InstancePrivateDnsName string
	LaunchTemplateVersion  int64
	AvailabilityZone       string
}
