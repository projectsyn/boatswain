package boatswain

import (
	"fmt"

	"github.com/projectsyn/boatswain/pkg/aws"
	"github.com/projectsyn/boatswain/pkg/k8sclient"
)

func New(awsAssumeRoleArn string) (*Boatswain, error) {
	aws, err := aws.NewAwsClient(awsAssumeRoleArn)
	if err != nil {
		return nil, fmt.Errorf("initializing AWS client: %w", err)
	}
	k8s, err := k8sclient.NewK8sClient()
	if err != nil {
		return nil, fmt.Errorf("initializing K8s client: %w", err)
	}
	return &Boatswain{
		AwsClient: aws,
		K8sClient: k8s,
	}, nil
}
