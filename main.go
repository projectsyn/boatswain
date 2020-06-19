package main

import (
	"fmt"
	"os"

	"github.com/projectsyn/boatswain/pkg/aws"
	"github.com/projectsyn/boatswain/pkg/k8sclient"
)

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
			fmt.Println("Instances")
			for _, i := range asg.Instances {
				var marker string
				if i.LaunchTemplateVersion < asg.LaunchTemplateVersion {
					marker = "!"
					k8sClient.AnnotateNode(nodes[i.InstancePrivateDnsName], map[string]string{
						"launchTemplateVersion": fmt.Sprintf("%v", i.LaunchTemplateVersion)})
				} else {
					marker = " "
				}
				fmt.Printf("%v %v - %v - LT version %v\n",
					marker, i.InstanceId, i.InstancePrivateDnsName, i.LaunchTemplateVersion)
			}
		}
	}
}
