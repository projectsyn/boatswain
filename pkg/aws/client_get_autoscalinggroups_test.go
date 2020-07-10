package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
)

func TestGetAutoScalingGroups(t *testing.T) {
	ltId := "lt-0136aedc6db8deb50"
	cases := []struct {
		Instances              []*ec2.DescribeInstancesOutput
		InstanceStatuses       map[string]*ec2.DescribeInstanceStatusOutput
		AsgResp                *autoscaling.DescribeAutoScalingGroupsOutput
		LtVersionResp          *ec2.DescribeLaunchTemplateVersionsOutput
		ExpectedInstanceCount  int
		ExpectedInstanceLtVerA int64
		ExpectedInstanceLtVerB int64
		LatestLtVer            int64
		ExpectedAsgCount       int
	}{
		{
			Instances:              makeAutoScalingGroupInstances(),
			InstanceStatuses:       makeAutoScalingGroupInstancesStatus(),
			AsgResp:                makeDescribeAutoScalingGroupsOutput(ltId, 1, 5),
			LtVersionResp:          makeDescribeLaunchTemplateVersionsOutput(ltId, 6, "ami-123456789abcdef"),
			ExpectedInstanceCount:  3,
			ExpectedInstanceLtVerA: 1,
			ExpectedInstanceLtVerB: 5,
			LatestLtVer:            6,
			ExpectedAsgCount:       1,
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
		assert.Equal(t, c.ExpectedAsgCount, len(groups.Groups), "Expected count of ASGs")
		if len(groups.Groups) > 0 {
			g := groups.Groups[0]
			assert.Equal(t, c.ExpectedInstanceCount, len(g.Instances), "Expected amount of instances in group")
			assert.Equal(t, "ami-123456789abcdef", g.CurrentAmi, "Current AMI correct")
			assert.Equal(t, c.LatestLtVer, g.LaunchTemplateVersion, "Correct latest LT version")
			for idx, i := range g.Instances {
				ltVer := c.ExpectedInstanceLtVerA
				if idx == len(g.Instances)-1 {
					ltVer = c.ExpectedInstanceLtVerB
				}
				assert.Equal(t, ltVer, i.LaunchTemplateVersion, "Correct instance LT version")
			}
		}
	}
}
