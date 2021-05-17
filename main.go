package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/mgutz/ansi"
	"github.com/slack-go/slack"
)

var (
	awsRegex     = flag.String("aws-regex", ".*", "Selects AWS instances where tag:Name contains this string. Example: (test|example)")
	awsAccessKey = MustGetenv("AWS_ACCESS_KEY_ID", "The AWS access key.")
	awsSecretKey = MustGetenv("AWS_SECRET_ACCESS_KEY", "The AWS secret key.")
	awsRegion    = MustGetenv("AWS_REGION", "The AWS region.")

	osRegex       = flag.String("os-regex", ".*", "Selects OpenStack instances where the instance name contains this string. Example: (test|example)")
	osUsername    = MustGetenv("OS_USERNAME", "")
	osPassword    = MustGetenv("OS_PASSWORD", "")
	osAuthURL     = MustGetenv("OS_AUTH_URL", "Often looks like http://host/identity/v3.")
	osProjectName = MustGetenv("OS_PROJECT_NAME", "Also called 'tenant name'.")
	osRegion      = MustGetenv("OS_REGION", "E.g., UK1 (for OVH).")
	osDomainName  = MustGetenv("OS_PROJECT_DOMAIN_NAME", `That's "Default" for most OpenStack instances.`)

	gcpRegex   = flag.String("gcp-regex", ".*", "Selects OpenStack instances where the instance name contains this string. Example: (test|example)")
	gcpJsonKey = MustGetenv("GCP_JSON_KEY", `The content of the json key in plain text, not base-64 encoded.`)

	slackChannel = flag.String("slack-channel", "", `With this argument, c4 sends a message to this channel whenever VMs are deleted (doesn't send anything when this flag isn't passed). Requires SLACK_TOKEN to be set.`)
	slackToken   = OptionalGetenv("SLACK_TOKEN", `Slack OAuth token, create one at https://api.slack.com/apps.`)

	olderThan = flag.Duration("older-than", 24*time.Hour, "Only delete resources older than this specified value. Can be any valid Go duration, such as 10m or 8h.")
	doIt      = flag.Bool("do-it", false, "By default, nothing is deleted. This flag enable deletion.")

	yel   = ansi.ColorFunc("yellow")
	green = ansi.ColorFunc("green")
	red   = ansi.ColorFunc("red")
	bold  = ansi.ColorFunc("white+b")

	showVersion = flag.Bool("version", false, "Watch out, returns 'n/a (commit none, built on unknown)' when built with 'go get'.")
	// The 'version' var is set during build, using something like:
	//  go build  -ldflags"-X main.version=$(git describe --tags)".
	// Note: "version", "commit" and "date" are set automatically by
	// goreleaser.
	version = "n/a"
	commit  = "none"
	date    = "unknown"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nEnvironment variables:\n%s\n", EnvvarUsage())
	}
	flag.Parse()
	ParseEnv()

	if *showVersion {
		fmt.Printf("%s (commit %s, built on %s)\n", version, commit, date)
		return
	}

	gcpRegex, err := regexp.Compile(*gcpRegex)
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "%s: --gcp-regex: %v\n", red("Error"), err)
		os.Exit(1)
	}
	awsRegex, err := regexp.Compile(*awsRegex)
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "%s: --aws-regex: %v\n", red("Error"), err)
		os.Exit(1)
	}
	osRegex, err := regexp.Compile(*osRegex)
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "%s: --os-regex: %v\n", red("Error"), err)
		os.Exit(1)
	}

	dryRun := !*doIt

	if dryRun {
		fmt.Printf("Showing in %s anything older than %s (to also delete them, add %s).\n", red("red"), bold(olderThan.String()), green("--do-it"))
	} else {
		fmt.Printf("Removing anything older than %s.\n", bold(olderThan.String()))
	}

	awsDeleted, err := nukeAWSInstances(*awsAccessKey, *awsSecretKey, *awsRegion, awsRegex, dryRun, *olderThan)
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "%s: while nuking AWS instances: %v\n", red("Error"), err)
		os.Exit(1)
	}

	osDeleted, err := nukeOpenStackInstances(*osRegion, *osAuthURL, *osDomainName, *osUsername, *osPassword, *osProjectName, osRegex, dryRun, *olderThan)
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "%s: while nuking OpenStack instances: %v\n", red("Error"), err)
		os.Exit(1)
	}

	gcpDeleted, err := nukeGCPInstances(*gcpJsonKey, gcpRegex, dryRun, *olderThan)
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "%s: while nuking GCP instances: %v\n", red("Error"), err)
		os.Exit(1)
	}

	if *slackToken == "" || *slackChannel == "" {
		return
	}

	if len(osDeleted) == 0 && len(awsDeleted) == 0 && len(gcpDeleted) == 0 {
		return
	}

	fmt.Printf("%s: some VMs were deleted, sending a Slack message to #%s.\n", yel("Note"), *slackChannel)

	msg := fmt.Sprintf("c4 removed instances that were older than %v:\n", *olderThan)
	for _, vm := range osDeleted {
		msg += fmt.Sprintf("- OpenStack: `%s` (%s, age: %s, `%s`)\n", vm.Name, *osRegion, osAge(vm).Truncate(time.Second), vm.Status)
	}
	for _, vm := range awsDeleted {
		msg += fmt.Sprintf("- AWS: `%s` (%s, age: %s, `%s`)\n", awsName(vm), *awsRegion, awsAge(vm).Truncate(time.Second), *vm.State.Name)
	}
	for _, vm := range gcpDeleted {
		msg += fmt.Sprintf("- GCP: `%s` (%s, age: %s, `%s`)\n", vm.Name, shorterGCPURL(vm.Zone), gcpAge(vm).Truncate(time.Second), vm.Status)
	}

	api := slack.New(*slackToken)
	_, _, _, err = api.SendMessage(*slackChannel, slack.MsgOptionText(msg, false))
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "%s: while sending a Slack message to %s: %v\n", red("Error"), *slackChannel, err)
		os.Exit(1)
	}
}
