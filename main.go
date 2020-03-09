package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/mgutz/ansi"
)

var (
	awsNameContains = flag.String("aws-name-contains", "", "selects AWS instances where tag:Name contains this string")
	awsAccessKey    = MustGetenv("AWS_ACCESS_KEY_ID", "The AWS access key")
	awsSecretKey    = MustGetenv("AWS_SECRET_ACCESS_KEY", "The AWS secret key")
	awsRegion       = MustGetenv("AWS_REGION", "The AWS region")

	osNameContains = flag.String("os-name-contains", "", "selects OpenStack instances where the instance name contains this string")
	osUsername     = MustGetenv("OS_USERNAME", "")
	osPassword     = MustGetenv("OS_PASSWORD", "")
	osAuthURL      = MustGetenv("OS_AUTH_URL", "looks like http://host/identity/v3")
	osProjectName  = MustGetenv("OS_PROJECT_NAME", "Also called 'tenant name'")
	osRegion       = MustGetenv("OS_REGION", "e.g., UK1 (for OVH)")
	osDomainName   = MustGetenv("OS_PROJECT_DOMAIN_NAME", `that's "Default" for most OpenStack instances`)

	olderThan = flag.Duration("older-than", 24*time.Hour, "Only delete resources older than this specified value. Can be any valid Go duration, such as 10m or 8h.")
	doIt      = flag.Bool("do-it", false, "By default, nothing is deleted. This flag enable deletion.")

	yel   = ansi.ColorFunc("yellow")
	green = ansi.ColorFunc("green")
	red   = ansi.ColorFunc("red")
	bold  = ansi.ColorFunc("white+b")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nMandatory environment variables:\n%s\n", EnvvarUsage())
	}
	flag.Parse()

	fmt.Printf("Removing anything older than %s.\n", bold(olderThan.String()))
	dryRun := !*doIt
	if dryRun {
		fmt.Printf("%s: running in dry-mode. To actually delete things, add %s.\n", yel("Note"), green("--do-it"))
	}

	err := nukeAWSInstances(awsAccessKey, awsSecretKey, awsRegion, *awsNameContains, dryRun)
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "%s: %v\n", red("Error"), err)
		os.Exit(1)
	}

	err = nukeOpenStackInstances(osRegion, osAuthURL, osDomainName, osUsername, osPassword, osProjectName, *osNameContains, dryRun)
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "%s: %v\n", red("Error"), err)
		os.Exit(1)
	}
}
