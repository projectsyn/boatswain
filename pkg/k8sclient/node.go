package k8sclient

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/drain"
)

func (c *K8sClient) GetNodes() NodeMap {
	nodes, err := c.Client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err.Error())
	}
	nodesByName := NodeMap{}
	for _, node := range nodes.Items {
		// create copy of node object which we can take address of
		tmpNode := node
		nodeName := node.ObjectMeta.Name
		nodesByName[nodeName] = &tmpNode
	}
	return nodesByName
}

func (c *K8sClient) CordonNode(node *corev1.Node) error {
	fmt.Println("Cordoning node", node.ObjectMeta.Name)
	return drain.RunCordonOrUncordon(c.Drainer, node, true)
}

var TransientDrainError = errors.New("Server unable to handle drain request")

func (c *K8sClient) DrainNode(node *corev1.Node) error {
	fmt.Println("Draining node", node.ObjectMeta.Name)
	if err := c.CordonNode(node); err != nil {
		return err
	}
	err := drain.RunNodeDrain(c.Drainer, node.ObjectMeta.Name)
	if err != nil {
		errStr := err.Error()
		if strings.HasPrefix(errStr, "[error when evicting pod") {
			realErrStr := strings.SplitN(errStr, ": ", 2)[1]
			if strings.HasPrefix(realErrStr, "the server is currently unable to handle the request") {
				return TransientDrainError
			}
		}
	}
	return err
}

func (c *K8sClient) DeleteNode(nodeName string) error {
	return c.Client.CoreV1().Nodes().Delete(context.TODO(), nodeName, metav1.DeleteOptions{})
}

func isNodeReady(node *corev1.Node) bool {
	for _, cond := range node.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			if cond.Status == corev1.ConditionTrue {
				return true
			} else {
				return false
			}
		}
	}
	return false
}

func (c *K8sClient) WaitUntilNodeReady(nodeName string) (*corev1.Node, error) {
	nodeReady := false
	for !nodeReady {
		nodes := c.GetNodes()
		if newNode, ok := nodes[nodeName]; ok {
			fmt.Println("New node object exists")
			nodeReady = isNodeReady(newNode)
			fmt.Printf("Kubelet posting ready? %v\n", nodeReady)
			if nodeReady {
				return newNode, nil
			}
		}
		time.Sleep(15 * time.Second)
	}
	return nil, errors.New("should be unreachable")
}

func (c *K8sClient) SetNodeRoles(node *corev1.Node) error {
	nodeName := node.ObjectMeta.Name
	nodeRoles := []string{}
	for key, value := range node.ObjectMeta.Labels {
		if strings.HasPrefix(key, "node.kubernetes.io/") && value == "true" {
			role := strings.SplitN(key, "/", 2)[1]
			fmt.Printf("Identified role %v for node %v\n", role, nodeName)
			nodeRoles = append(nodeRoles, role)
		}
	}
	for _, role := range nodeRoles {
		noderole := fmt.Sprintf("node-role.kubernetes.io/%v", role)
		fmt.Println(noderole)
		node.ObjectMeta.Labels[noderole] = "true"
	}
	_, err := c.Client.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
	return err
}
