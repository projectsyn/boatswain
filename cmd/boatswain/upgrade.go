package boatswain

import (
	"fmt"

	boatswainpkg "github.com/projectsyn/boatswain/pkg/boatswain"
)

func (b *Boatswain) Upgrade(singleNode string, forceReplace bool) error {
	if forceReplace {
		// Force replacing all nodes overrides single node operation
		singleNode = ""
	}

	nodes := b.K8sClient.GetNodes()

	asgs, err := b.AwsClient.GetAutoScalingGroups()
	if err != nil {
		return err
	}

	for _, asg := range asgs.Groups {
		fmt.Printf("ASG %v %v/%v instances, latest LT version %v\n",
			asg.AutoScalingGroupName, asg.DesiredCapacity, asg.MaxSize, asg.LaunchTemplateVersion)
		if asg.DesiredCapacity > 0 {
			for _, i := range asg.Instances {
				if singleNode != "" {
					if i.InstancePrivateDnsName == singleNode {
						fmt.Printf("found node %v (%v) -- replacing it\n",
							i.InstancePrivateDnsName, i.InstanceId)
						if err := boatswainpkg.ReplaceAsgNode(b.AwsClient,
							b.K8sClient,
							asg,
							i,
							nodes[i.InstancePrivateDnsName]); err != nil {
							panic(err.Error())
						}
					}
				} else {
					if forceReplace || i.LaunchTemplateVersion < asg.LaunchTemplateVersion {
						fmt.Printf("Instance %v (%v) uses old LaunchTemplateVersion %v\n",
							i.InstanceId, i.InstancePrivateDnsName, i.LaunchTemplateVersion)
						if err := boatswainpkg.ReplaceAsgNode(b.AwsClient,
							b.K8sClient,
							asg,
							i,
							nodes[i.InstancePrivateDnsName]); err != nil {
							panic(err.Error())
						}
					}
				}
			}
		}
	}
	return nil
}