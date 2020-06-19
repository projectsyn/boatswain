package main

import (
	"fmt"
	"os"

	"github.com/projectsyn/boatswain/pkg/aws"
	"github.com/projectsyn/boatswain/pkg/k8sclient"

	corev1 "k8s.io/api/core/v1"
)

func replaceAsgNode(awsClient *aws.AwsClient, k8sClient *k8sclient.K8sClient,
	asgName string, instance aws.Instance, node *corev1.Node) error {
	// procedure:
	// 1. drain node
	if err := k8sClient.DrainNode(node); err != nil {
		return err
	}
	// 2. replace node in ASG
	// TBD: Is the replacement node
	// created in the same AZ as the
	// removed node when a specific node
	// is removed from ASG?
	fmt.Println("Replacing Node", instance.InstanceId)
	if err := awsClient.ReplaceNodeInASG(asgName, instance); err != nil {
		return err
	}
	// 3. Wait for no pods pending
	fmt.Println("Wait until no pods pending")
	// 4. Delete old K8s node object
	fmt.Println("Delete old node object")
	if err := k8sClient.DeleteNode(instance.InstancePrivateDnsName); err != nil {
		return err
	}
	return k8sClient.WaitUntilNoPodsPending()
}

func main() {
	awsClient := aws.NewAwsClient(os.Getenv("AWS_ASSUME_ROLE_ARN"))
	k8sClient := k8sclient.NewK8sClient()

	theNode := os.Getenv("NODE")
	if theNode != "" {
		fmt.Println("only draining", theNode)
	}

	nodes := k8sClient.GetNodes()

	asgs, err := awsClient.GetAutoScalingGroups()
	if err != nil {
		panic(err.Error())
	}

	for _, asg := range asgs.Groups {
		fmt.Printf("ASG %v %v/%v instances, latest LT version %v\n",
			asg.AutoScalingGroupName, asg.DesiredCapacity, asg.MaxSize, asg.LaunchTemplateVersion)
		if asg.DesiredCapacity > 0 {
			for _, i := range asg.Instances {
				if i.LaunchTemplateVersion < asg.LaunchTemplateVersion {
					if theNode != "" && i.InstancePrivateDnsName != theNode {
						continue
					}
					fmt.Printf("Instance %v (%v) uses old LaunchTemplateVersion %v\n",
						i.InstanceId, i.InstancePrivateDnsName, i.LaunchTemplateVersion)
					if err := replaceAsgNode(awsClient,
						k8sClient,
						asg.AutoScalingGroupName,
						i,
						nodes[i.InstancePrivateDnsName]); err != nil {
						panic(err.Error())
					}
				}
			}
		}
	}
}
