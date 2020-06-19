package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
)

func TestGetInstancePrivateDnsName(t *testing.T) {
	cases := []struct {
		Instance string
		Resp     []*ec2.DescribeInstancesOutput
		Expected string
	}{
		{
			Instance: "i-0757e06c74906c484",
			Resp: []*ec2.DescribeInstancesOutput{
				makeDescribeInstancesOutput("i-0757e06c74906c484", "ip-10-200-105-130.eu-central-1.compute.internal"),
			},
			Expected: "ip-10-200-105-130.eu-central-1.compute.internal",
		},
	}

	for _, c := range cases {
		client := AwsClient{
			EC2: &mockEC2Client{InstancesResp: c.Resp},
		}
		privateDns, err := client.getInstancePrivateDnsName(&c.Instance)
		if err != nil {
			t.Fatalf("%d, unexpected error", err)
		}
		assert.Equal(t, c.Expected, *privateDns, "Error parsing private DNS name of instance")
	}

}
