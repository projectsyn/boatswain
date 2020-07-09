package main

import (
	"fmt"

	"github.com/projectsyn/boatswain/cmd/boatswain"

	"github.com/alecthomas/kong"
)

// TODO: consider routing more AWS/K8s credentials through CLI. Right now we
// rely on the AWS and k8s client libraries to extract the credentials from
// environment variables or default configs
type ApiCredentials struct {
	AwsAssumeRoleArn string `help:"ARN for role to assume on AWS" env:"AWS_ASSUME_ROLE_ARN" default:""`
}

type CheckAmiCmd struct {
	ApiCredentials
}

func (c *CheckAmiCmd) Run(ctx *kong.Context) error {
	b := boatswain.New(c.AwsAssumeRoleArn)
	return b.CheckAmi()
}

type ListUpgradableCmd struct {
	ApiCredentials
}

func (c *ListUpgradableCmd) Run(ctx *kong.Context) error {
	b := boatswain.New(c.AwsAssumeRoleArn)
	instances, err := b.ListUpgradable()
	if err != nil {
		return err
	}
	if len(instances) == 0 {
		fmt.Println("Everything up-to-date!")
		return nil
	}
	fmt.Println("Instances which are outdated")
	for _, i := range instances {
		fmt.Println("  ", i.InstancePrivateDnsName)
	}
	return nil
}

type UpgradeCmd struct {
	ApiCredentials
	SingleNode   string `help:"Provide K8s name of a single node to upgrade"`
	ForceReplace bool   `help:"Force replace all current nodes. Overrides replacing a single node"`
}

func (c *UpgradeCmd) Run(ctx *kong.Context) error {
	b := boatswain.New(c.AwsAssumeRoleArn)
	return b.Upgrade(c.SingleNode, c.ForceReplace)
}
