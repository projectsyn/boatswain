package k8sclient

import (
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
)

func (c *K8sClient) GetServerVersion() (*version.Info, error) {
	d := discovery.NewDiscoveryClientForConfigOrDie(c.config)
	return d.ServerVersion()
}
