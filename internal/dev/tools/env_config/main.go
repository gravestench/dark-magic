// Command env_config installs or updates per-process Dark Magic environment files.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gravestench/dark-magic/internal/app/envconfig"
)

type assignments map[string]string

func (values assignments) String() string { return "NAME=VALUE" }

func (values assignments) Set(value string) error {
	separator := strings.IndexByte(value, '=')
	if separator <= 0 {
		return fmt.Errorf("set value must be NAME=VALUE")
	}
	values[value[:separator]] = value[separator+1:]
	return nil
}

func main() {
	role := flag.String("role", "", "environment role: client, server, realm, or all")
	updates := assignments{}
	flag.Var(updates, "set", "set a template variable (repeatable NAME=VALUE)")
	flag.Parse()
	roles := []string{*role}
	if *role == "all" {
		if len(updates) != 0 {
			fmt.Fprintln(os.Stderr, "--set requires one explicit role")
			os.Exit(2)
		}
		roles = []string{"client", "server", "realm"}
	}
	if *role == "" {
		fmt.Fprintln(os.Stderr, "--role is required")
		os.Exit(2)
	}
	for _, current := range roles {
		path, err := envconfig.Update(current, updates)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(path)
	}
}
