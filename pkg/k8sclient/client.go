package k8sclient

import (
	"context"
	"fmt"
	"os"
	"time"

	//"k8s.io/apimachinery/pkg/api/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func NewK8sClient() *K8sClient {
	// if you want to change the loading rules (which files in which order), you can do so here
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	// if you want to change override values or bind them to flags, there are methods to help you
	configOverrides := &clientcmd.ConfigOverrides{}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return &K8sClient{Client: clientset}
}

func (c *K8sClient) GetNodes() NodeMap {
	nodes, err := c.Client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
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

func (c *K8sClient) AnnotateNode(node *corev1.Node, annotations map[string]string) {
	for k, v := range annotations {
		node.ObjectMeta.Annotations[k] = v
	}
	_, err := c.Client.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
	if err != nil {
		panic(err.Error())
	}
}
