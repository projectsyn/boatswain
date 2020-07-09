package boatswain

import (
	"fmt"
	"strings"
	"unicode"
)

func (b *Boatswain) CheckAmi() error {
	eksVersion, err := b.K8sClient.GetServerVersion()
	if err != nil {
		return err
	}
	eksMinor := strings.TrimFunc(eksVersion.Minor, func(r rune) bool {
		return !unicode.IsNumber(r)
	})
	fmt.Printf("Current EKS Control Plane Version: v%v.%v\n", eksVersion.Major, eksMinor)
	latestAmi, err := b.AwsClient.GetLatestEKSAmi(eksVersion.Major, eksMinor)
	if err != nil {
		return err
	}

	asgs, err := b.AwsClient.GetAutoScalingGroups()
	if err != nil {
		return err
	}

	// TODO: show AMI name instead of ID
	newAmiAvail := false
	for _, asg := range asgs.Groups {
		if asg.CurrentAmi != latestAmi {
			fmt.Printf("ASG %v has outdated AMI %v (current AMI: %v)\n",
				asg.AutoScalingGroupName, asg.CurrentAmi, latestAmi)
			newAmiAvail = true
		}
	}
	if !newAmiAvail {
		fmt.Println("No ASGs with outdated AMI found")
	}
	return nil
}
