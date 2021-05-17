package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Just used for displaying help.
type envVar struct {
	key, usage string
	mandatory  bool
	value      *string
}

var envVars []envVar

func EnvvarUsage() string {
	var usage []string
	for _, elmt := range envVars {
		isMandatory := ""
		if elmt.mandatory {
			isMandatory = " (mandatory)"
		}
		usage = append(usage, fmt.Sprintf("  %s\n    \t%s%s", elmt.key, elmt.usage, isMandatory))
	}
	return strings.Join(usage, "\n")
}

func MustGetenv(key string, usage string) *string {
	value := ""
	envVars = append(envVars, envVar{key: key, usage: usage, mandatory: true, value: &value})
	return &value
}

func OptionalGetenv(key string, usage string) *string {
	value := ""
	envVars = append(envVars, envVar{key: key, usage: usage, value: &value})
	return &value
}

func ParseEnv() {
	for _, v := range envVars {
		*v.value = os.Getenv(v.key)
		if v.mandatory && *v.value == "" {
			fmt.Fprintf(flag.CommandLine.Output(), "%s: the env var %s is not set or is empty. We recommend using direnv's %s for that.\n", red("error"), yel(v.key), green(".envrc"))
			os.Exit(123)
		}
	}
}
