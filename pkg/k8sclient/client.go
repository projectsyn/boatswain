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
	"k8s.io/kubectl/pkg/drain"
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
	d := &drain.Helper{
		Client:              clientset,
		GracePeriodSeconds:  30,
		Force:               true,
		IgnoreAllDaemonSets: true,
		DeleteLocalData:     true,
		Timeout:             time.Duration(0),
		Out:                 os.Stdout,
		ErrOut:              os.Stderr,
	}
	return &K8sClient{Client: clientset, Drainer: d}
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

func (c *K8sClient) CordonNode(node *corev1.Node) error {
	fmt.Println("Cordoning node", node.ObjectMeta.Name)
	return drain.RunCordonOrUncordon(c.Drainer, node, true)
}

func (c *K8sClient) DrainNode(node *corev1.Node) error {
	fmt.Println("Draining node", node.ObjectMeta.Name)
	if err := c.CordonNode(node); err != nil {
		return err
	}
	return drain.RunNodeDrain(c.Drainer, node.ObjectMeta.Name)
}
