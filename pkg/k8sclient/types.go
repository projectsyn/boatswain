package k8sclient

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/drain"
)

// Kubernetes client
type K8sClient struct {
	Client  *kubernetes.Clientset
	config  *rest.Config
	Drainer *drain.Helper
}

// map of NodeName->Node
type NodeMap map[string]*corev1.Node
