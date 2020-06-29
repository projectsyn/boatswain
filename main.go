package main

import (
	"fmt"
	"os"

	"github.com/projectsyn/boatswain/pkg/aws"
	"github.com/projectsyn/boatswain/pkg/k8sclient"

	corev1 "k8s.io/api/core/v1"
)

func replaceAsgNode(awsClient *aws.AwsClient, k8sClient *k8sclient.K8sClient,
	asg aws.AutoScalingGroup, instance aws.Instance, node *corev1.Node) error {
	// procedure:
	// 1. cordon node
	fmt.Println("Cordon node")
	if err := k8sClient.CordonNode(node); err != nil {
		return err
	}
	// 2. replace node in ASG
	fmt.Println("Replace ASG instance", instance.InstanceId)
	var newInstance *aws.Instance
	if i, err := awsClient.ReplaceNodeInASG(asg, instance); err == nil {
		newInstance = i
	} else {
		return err
	}
	// 3. Wait for new node ready
	fmt.Println("Wait for new node ready")
	if err := k8sClient.WaitUntilNodeReady(newInstance.InstancePrivateDnsName); err != nil {
		return err
	}
	// 4. Drain old node
	fmt.Println("Drain old node")
	if err := k8sClient.DrainNode(node); err != nil {
		return err
	}
	// 5. wait until no pods pending
	fmt.Println("Wait until no pods pending")
	if err := k8sClient.WaitUntilNoPodsPending(); err != nil {
		return err
	}
	// 6. Delete old K8s node object
	fmt.Println("Delete old node object")
	if err := k8sClient.DeleteNode(instance.InstancePrivateDnsName); err != nil {
		return err
	}
	// 7. Terminate old instance
	fmt.Println("Terminate old instance")
	return awsClient.TerminateInstance(instance.InstanceId)
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
