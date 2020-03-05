package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/mgutz/ansi"
)

var (
	awsNameFilter = flag.String("aws-name-filter", "", "nukes AWS instances where the tag:Name matches this filter. May be a glob pattern: *test*")
	olderThan     = flag.Duration("older-than", 24*time.Hour, "Only delete resources older than this specified value. Can be any valid Go duration, such as 10m or 8h.")
	dryRun        = flag.Bool("dry-run", false, "Don't delete anything, just print what would be deleted.")
	awsAccessKey  = MustGetenv("AWS_ACCESS_KEY_ID", "The AWS access key")
	awsSecretKey  = MustGetenv("AWS_SECRET_ACCESS_KEY", "The AWS secret key")
	awsRegion     = MustGetenv("AWS_REGION", "The AWS region")

	yel   = ansi.ColorFunc("yellow")
	green = ansi.ColorFunc("green")
	red   = ansi.ColorFunc("red")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nMandatory environment variables:\n%s\n", EnvvarUsage())
	}
	flag.Parse()

	fmt.Printf("Removing anything older than %s\n", red(olderThan.String()))
	if *dryRun {
		fmt.Printf("%s\n", "(--dry-run mode)")
	}

	err := nukeAWSInstances()
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "%s: %v\n", red("error"), err)
		os.Exit(1)
	}
}
