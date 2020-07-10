package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ssm"
)

func NewAwsClient(roleArn string) *AwsClient {
	c := &AwsClient{}

	c.Session = session.New()
	creds := stscreds.NewCredentials(c.Session, roleArn)
	c.Config = &aws.Config{Credentials: creds}

	c.AutoScaling = autoscaling.New(c.Session, c.Config)
	c.EC2 = ec2.New(c.Session, c.Config)
	c.SSM = ssm.New(c.Session, c.Config)

	return c
}
