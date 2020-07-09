package boatswain

import (
	"github.com/projectsyn/boatswain/pkg/aws"
	"github.com/projectsyn/boatswain/pkg/k8sclient"
)

func New(awsAssumeRoleArn string) *Boatswain {
	aws := aws.NewAwsClient(awsAssumeRoleArn)
	k8s := k8sclient.NewK8sClient()
	return &Boatswain{
		AwsClient: aws,
		K8sClient: k8s,
	}
}
