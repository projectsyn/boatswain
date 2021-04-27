package boatswain

import (
	"fmt"
	"strings"
	"unicode"
)

func (b *Boatswain) CheckAmi(eksVersionOverride string) error {
	eksVersion, err := b.K8sClient.GetServerVersion()
	if err != nil {
		return fmt.Errorf("getting K8s server version: %w", err)
	}
	eksMajor := eksVersion.Major
	eksMinor := strings.TrimFunc(eksVersion.Minor, func(r rune) bool {
		return !unicode.IsNumber(r)
	})
	fmt.Printf("Current EKS Control Plane Version: v%v.%v\n", eksMajor, eksMinor)
	if eksVersionOverride != "" {
		overrideParts := strings.Split(eksVersionOverride, ".")
		eksMajor = strings.Trim(overrideParts[0], "v")
		if eksMajor < eksVersion.Major || (eksMajor == eksVersion.Major && eksMinor > overrideParts[1]) {
			fmt.Println("Check for older EKS version than current cluster version requested, results may be unreliable")
		}
		eksMinor = overrideParts[1]
		fmt.Printf("User requested check for EKS version v%v.%v\n", eksMajor, eksMinor)
	}
	latestAmi, err := b.AwsClient.GetLatestEKSAmi(eksMajor, eksMinor)
	if err != nil {
		return fmt.Errorf("getting latest EKS AMI: %w", err)
	}

	asgs, err := b.AwsClient.GetAutoScalingGroups()
	if err != nil {
		return fmt.Errorf("getting ASGs: %w", err)
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
