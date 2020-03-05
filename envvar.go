package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Just used for displaying help.
var envvars []struct{ key, usage string }

func EnvvarUsage() string {
	var usage []string
	for _, elmt := range envvars {
		usage = append(usage, fmt.Sprintf("  %s\n    \t%s", elmt.key, elmt.usage))
	}
	return strings.Join(usage, "\n")
}

func MustGetenv(key string, usage string) string {
	envvars = append(envvars, struct{ key, usage string }{key: key, usage: usage})
	res := os.Getenv(key)
	if res == "" {
		fmt.Fprintf(flag.CommandLine.Output(), "%s: the env var %s is not set or is empty. Did you forget to '%s'?\n", red("error"), yel(key), green("source .passrc"))
		os.Exit(123)
	}
	return res
}
