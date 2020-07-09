package main

import (
	"fmt"

	//	"github.com/projectsyn/boatswain/pkg/aws"
	//	"github.com/projectsyn/boatswain/pkg/k8sclient"
	"github.com/alecthomas/kong"
)

var (
	Version   = "undefined"
	BuildDate = "now"
)

type HelpCmd struct {
	Command []string `arg optional help:"Show help on command"`
}

// Run shows help.
func (c *HelpCmd) Run(realCtx *kong.Context) error {
	ctx, err := kong.Trace(realCtx.Kong, c.Command)
	if err != nil {
		return err
	}
	if ctx.Error != nil {
		return ctx.Error
	}
	err = ctx.PrintUsage(false)
	if err != nil {
		return err
	}
	fmt.Fprintln(realCtx.Stdout)
	return nil
}

type VersionCmd struct{}

func (v VersionCmd) Run(ctx *kong.Context) error {
	fmt.Println(Version)
	return nil
}

type VersionFlag string

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(Version)
	app.Exit(0)
	return nil
}

type CLI struct {
	Version        VersionCmd        `cmd help:"Print version information and quit"`
	VersionFlag    VersionFlag       `name:"version" help:"Print version information and quit"`
	Help           HelpCmd           `cmd help:"Show this help"`
	CheckAmi       CheckAmiCmd       `cmd name:"check-ami" help:"Check whether there's a newer AMI than the one referenced in th current LaunchTemplate"`
	ListUpgradable ListUpgradableCmd `cmd name:"list-upgradable" help:"List nodes which are running off an outdated LaunchTemplate version"`
	Upgrade        UpgradeCmd        `cmd help:"Upgrade nodes which are running off an outdated LaunchTemplate version"`
}

func main() {
	cli := CLI{}
	ctx := kong.Parse(&cli,
		kong.Name("boatswain"),
		kong.Description("Boatswain helps automate EKS node maintenance and upgrades"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}))

	err := ctx.Run()
	ctx.FatalIfErrorf(err)
	//	awsClient := aws.NewAwsClient(os.Getenv("AWS_ASSUME_ROLE_ARN"))
	//	k8sClient := k8sclient.NewK8sClient()
	//
	//	theNode := os.Getenv("NODE")
	//	if theNode != "" {
	//		fmt.Println("only considering", theNode)
	//	}
	//
	//	forceReplace, err := strconv.ParseBool(os.Getenv("FORCE_REPLACE"))
}
