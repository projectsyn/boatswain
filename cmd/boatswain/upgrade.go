package boatswain

import (
	"fmt"
	"log"

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
		return fmt.Errorf("getting ASGs: %w", err)
	}

	for _, asg := range asgs.Groups {
		fmt.Printf("ASG %v %v/%v instances, latest LT version %v\n",
			asg.AutoScalingGroupName, asg.DesiredCapacity, asg.MaxSize, asg.LaunchTemplateVersion)
		if asg.DesiredCapacity > 0 {
			for _, i := range asg.Instances {
				node, ok := nodes[i.InstancePrivateDnsName]
				if !ok {
					fmt.Printf("Skipping instance %v, no matching K8s node\n", i.InstancePrivateDnsName)
					continue
				}
				if singleNode != "" {
					if i.InstancePrivateDnsName == singleNode {
						fmt.Printf("found node %v (%v) -- replacing it\n",
							i.InstancePrivateDnsName, i.InstanceId)
						if err := boatswainpkg.ReplaceAsgNode(b.AwsClient,
							b.K8sClient,
							asg,
							i,
							node); err != nil {
							log.Fatal(err.Error())
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
							node); err != nil {
							log.Fatal(err.Error())
						}
					}
				}
			}
		}
	}
	return nil
}
