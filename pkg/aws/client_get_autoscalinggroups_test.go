package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
)

func TestGetAutoScalingGroups(t *testing.T) {
	ltId := "lt-0136aedc6db8deb50"
	cases := []struct {
		Instances        []*ec2.DescribeInstancesOutput
		InstanceStatuses map[string]*ec2.DescribeInstanceStatusOutput
		AsgResp          *autoscaling.DescribeAutoScalingGroupsOutput
		LtVersionResp    *ec2.DescribeLaunchTemplateVersionsOutput
		ExpectedCount    int64
	}{
		{
			Instances:        makeAutoScalingGroupInstances(),
			InstanceStatuses: makeAutoScalingGroupInstancesStatus(),
			AsgResp:          makeDescribeAutoScalingGroupsOutput(ltId, 1, 5),
			LtVersionResp:    makeDescribeLaunchTemplateVersionsOutput(ltId, 6),
			ExpectedCount:    3,
		},
	}

	for _, c := range cases {
		client := AwsClient{
			EC2: &mockEC2Client{LTVersionsResp: c.LtVersionResp, InstancesResp: c.Instances,
				InstanceStatusResp: c.InstanceStatuses},
			AutoScaling: &mockAutoScalingClient{Resp: c.AsgResp},
		}

		groups, err := client.GetAutoScalingGroups()
		assert.NoError(t, err)
		fmt.Println(groups)
	}
}
