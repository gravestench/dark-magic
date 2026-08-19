// Command env_config installs or updates per-process Dark Magic environment files.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gravestench/dark-magic/internal/app/envconfig"
)

const (
	updateFailureExitCode = 1
	usageErrorExitCode    = 2
)

// assignments collects repeated --set values by name so later values retain the flag package's last-write behavior.
type assignments map[string]string

// String describes the accepted flag value without exposing configured values in usage output.
func (values assignments) String() string { return "NAME=VALUE" }

// Set splits only the first separator so environment values may contain equals signs without additional escaping.
func (values assignments) Set(value string) error {
	separator := strings.IndexByte(value, '=')
	if separator <= 0 {
		return fmt.Errorf("set value must be NAME=VALUE")
	}

	values[value[:separator]] = value[separator+1:]

	return nil
}

// main keeps flag parsing at the process boundary so the standard flag package continues to own usage and help output.
func main() {
	role := flag.String("role", "", "environment role: client, server, realm, or all")
	updates := assignments{}
	flag.Var(updates, "set", "set a template variable (repeatable NAME=VALUE)")
	flag.Parse()

	exitCode := updateEnvironmentFiles(*role, updates, os.Stdout, os.Stderr)
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

// updateEnvironmentFiles validates the cross-role request and applies updates in stable, fail-fast command order.
func updateEnvironmentFiles(role string, updates assignments, stdout, stderr io.Writer) int {
	roles, err := selectUpdateRoles(role, updates)
	if err != nil {
		writeCommandLine(stderr, err)

		return usageErrorExitCode
	}

	// Roles remain sequential so printed paths reflect completed writes and the first failure stops later mutations.
	for _, current := range roles {
		path, err := envconfig.Update(current, updates)
		if err != nil {
			writeCommandLine(stderr, err)

			return updateFailureExitCode
		}

		writeCommandLine(stdout, path)
	}

	return 0
}

// writeCommandLine preserves command outcome semantics when a terminal or pipe cannot accept informational output.
func writeCommandLine(destination io.Writer, value any) {
	_, _ = fmt.Fprintln(destination, value)
}

// selectUpdateRoles expands "all" deterministically while preventing one update map from crossing role schemas.
func selectUpdateRoles(role string, updates assignments) ([]string, error) {
	if role == "all" {
		if len(updates) != 0 {
			return nil, errors.New("--set requires one explicit role")
		}

		return []string{"client", "server", "realm"}, nil
	}

	if role == "" {
		return nil, errors.New("--role is required")
	}

	// Concrete role validation belongs to envconfig.Update so its existing error remains the command contract.
	return []string{role}, nil
}
