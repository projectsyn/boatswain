package aws

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ssm"
)

var (
	awsEnvVars = []string{
		"AWS_REGION",
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
	}
)

func verifyAwsEnvVars() error {
	for _, e := range awsEnvVars {
		_, ok := os.LookupEnv(e)
		if !ok {
			return fmt.Errorf("environment variable %s is mandatory", e)
		}
	}
	return nil
}

func NewAwsClient(roleArn string) (*AwsClient, error) {
	if err := verifyAwsEnvVars(); err != nil {
		return nil, err
	}

	c := &AwsClient{}

	c.Session = session.New()

	if roleArn != "" {
		creds := stscreds.NewCredentials(c.Session, roleArn)
		c.Config = &aws.Config{Credentials: creds}
	}

	c.AutoScaling = autoscaling.New(c.Session, c.Config)
	c.EC2 = ec2.New(c.Session, c.Config)
	c.SSM = ssm.New(c.Session, c.Config)

	return c, nil
}
