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

// To keep the CLI consistent, we keep a slice of the env vars. If we were
// to just rely on the map, we would get different ordering (e.g., in
// --help) every time.
var envVarsKeys []string
var envVarsMap map[string]envVar = make(map[string]envVar)

func EnvvarUsage() string {
	var usage []string
	for _, key := range envVarsKeys {
		elmt := envVarsMap[key]
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
	envVarsKeys = append(envVarsKeys, key)
	envVarsMap[key] = envVar{key: key, usage: usage, mandatory: true, value: &value}
	return &value
}

func OptionalGetenv(key string, usage string) *string {
	value := ""
	envVarsKeys = append(envVarsKeys, key)
	envVarsMap[key] = envVar{key: key, usage: usage, mandatory: false, value: &value}
	return &value
}

func OptionalFlagOrEnv(flagName, envVarKey, usage string) *string {
	value := flag.String(flagName, "", usage+". Alternatively, you can use "+envVarKey+".")

	envVarsKeys = append(envVarsKeys, envVarKey)
	envVarsMap[envVarKey] = envVar{key: envVarKey, usage: usage, mandatory: false, value: value}

	return value
}

func ParseEnv() {
	for _, key := range envVarsKeys {
		v := envVarsMap[key]
		*v.value = os.Getenv(v.key)
		if v.mandatory && *v.value == "" {
			fmt.Fprintf(flag.CommandLine.Output(), "%s: the env var %s is not set or is empty. We recommend using direnv's %s for that.\n", red("error"), yel(v.key), green(".envrc"))
			os.Exit(123)
		}
	}
}

func AreSet(keys ...string) (ok bool, missingKeys []string) {
	var missing []string
	for _, key := range keys {
		v, ok := envVarsMap[key]
		if !ok {
			fmt.Fprintf(flag.CommandLine.Output(), "programmer mistake: IsSet is called with the var '%s' that has not been registered with MustGetenv or OptionalGetenv", key)
		}
		if *v.value == "" {
			missing = append(missing, key)
		}
	}

	return len(missing) == 0, missing
}
