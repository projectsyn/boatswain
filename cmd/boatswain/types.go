package boatswain

import (
	"github.com/projectsyn/boatswain/pkg/aws"
	"github.com/projectsyn/boatswain/pkg/k8sclient"
)

type Boatswain struct {
	AwsClient *aws.AwsClient
	K8sClient *k8sclient.K8sClient
}
