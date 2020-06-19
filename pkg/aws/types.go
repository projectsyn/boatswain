package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type AwsClient struct {
	Session     *session.Session
	Config      *aws.Config
	EC2         ec2iface.EC2API
	AutoScaling autoscalingiface.AutoScalingAPI
}

type AutoScalingGroups struct {
	Groups []AutoScalingGroup
}

type AutoScalingGroup struct {
	AutoScalingGroupName  string
	DesiredCapacity       int64
	LaunchTemplateId      string
	LaunchTemplateVersion int64
	MaxSize               int64
	Instances             []Instance
}

type Instance struct {
	InstanceId             string
	InstancePrivateDnsName string
	LaunchTemplateVersion  int64
}
