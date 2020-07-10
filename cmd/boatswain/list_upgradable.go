package boatswain

import (
	"github.com/projectsyn/boatswain/pkg/aws"
)

func (b *Boatswain) ListUpgradable() ([]*aws.Instance, error) {
	nodes := b.K8sClient.GetNodes()

	asgs, err := b.AwsClient.GetAutoScalingGroups()
	if err != nil {
		return nil, err
	}

	instances := []*aws.Instance{}

	for _, asg := range asgs.Groups {
		if asg.DesiredCapacity > 0 {
			for _, i := range asg.Instances {
				if i.LaunchTemplateVersion < asg.LaunchTemplateVersion {
					if _, ok := nodes[i.InstancePrivateDnsName]; ok {
						instances = append(instances, i)
					}
				}
			}
		}
	}
	return instances, nil
}
