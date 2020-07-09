package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
)

func (c *AwsClient) GetLatestEKSAmi(eksMajor string, eksMinor string) (string, error) {
	param := fmt.Sprintf("/aws/service/eks/optimized-ami/%v.%v/amazon-linux-2/recommended/image_id",
		eksMajor, eksMinor)
	result, err := c.SSM.GetParameter(&ssm.GetParameterInput{
		Name: aws.String(param),
	})
	if err != nil {
		return "", err
	}
	return *result.Parameter.Value, nil
}
