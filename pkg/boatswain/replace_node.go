package boatswain

import (
	"fmt"
	"time"

	"github.com/projectsyn/boatswain/pkg/aws"
	"github.com/projectsyn/boatswain/pkg/k8sclient"

	corev1 "k8s.io/api/core/v1"
)

func ReplaceAsgNode(awsClient *aws.AwsClient, k8sClient *k8sclient.K8sClient,
	asg *aws.AutoScalingGroup, instance *aws.Instance, node *corev1.Node) error {
	// procedure:
	// 1. cordon node
	fmt.Println("Cordon node")
	if err := k8sClient.CordonNode(node); err != nil {
		return err
	}
	// 2. replace node in ASG
	fmt.Println("Replace ASG instance", instance.InstanceId)
	newInstance, err := awsClient.ReplaceNodeInASG(asg, instance)
	if err != nil {
		return err
	}
	// 3. Wait for new node ready
	fmt.Printf("Wait for new node %v ready\n", newInstance.InstancePrivateDnsName)
	newNode, err := k8sClient.WaitUntilNodeReady(newInstance.InstancePrivateDnsName)
	if err != nil {
		return err
	}
	// 4. Ensure node-role labels exist on new node
	fmt.Println("Set node-role.kubernetes.io labels on new node")
	if err := k8sClient.SetNodeRoles(newNode); err != nil {
		fmt.Printf("Failed to set node roles for %v, continuing...\n",
			newNode.ObjectMeta.Name)
	}
	fmt.Printf("New node %v ready\n", newNode.ObjectMeta.Name)
	// 5. Drain old node
	fmt.Println("Drain old node")
	retryDrain := true
	for retryDrain {
		if err := k8sClient.DrainNode(node); err != nil {
			fmt.Println("Drain error: ", err)
			if err == k8sclient.ErrTransientDrain {
				time.Sleep(5 * time.Second)
				continue
			}
			fmt.Println("Ignoring error...")
		}
		retryDrain = false
	}
	// 6. wait until no pods pending
	fmt.Println("Wait until no pods pending")
	if err := k8sClient.WaitUntilNoPodsPending(); err != nil {
		return err
	}
	// 7. Delete old K8s node object
	fmt.Println("Delete old node object")
	if err := k8sClient.DeleteNode(instance.InstancePrivateDnsName); err != nil {
		return err
	}
	// 8. Terminate old instance
	fmt.Println("Terminate old instance")
	return awsClient.TerminateInstance(instance.InstanceId)
}
