// Command lua_syntax compiles Lua sources with the game's gopher-lua compiler
// (Lua 5.1 plus goto), so the check matches what the runtime can load rather
// than whichever system luac happens to be installed.
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

const (
	syntaxErrorExitCode = 1
	usageErrorExitCode  = 2
)

// main reports every failing file before exiting so one run lists all problems.
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: lua_syntax <directory or file>...")
		os.Exit(usageErrorExitCode)
	}

	failures, err := checkPaths(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(usageErrorExitCode)
	}

	for _, failure := range failures {
		fmt.Fprintln(os.Stderr, failure)
	}

	if len(failures) > 0 {
		os.Exit(syntaxErrorExitCode)
	}
}

// checkPaths walks directories for .lua files and compiles each one. Syntax
// failures are returned as values; only unreadable paths are errors.
func checkPaths(paths []string) ([]error, error) {
	var failures []error

	for _, root := range paths {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if entry.IsDir() || !strings.HasSuffix(path, ".lua") {
				return nil
			}

			source, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			if err := compile(path, string(source)); err != nil {
				failures = append(failures, err)
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("check %s: %w", root, err)
		}
	}

	return failures, nil
}

// compile parses and compiles without running, which also catches errors such
// as undefined goto labels that parsing alone accepts.
func compile(name, source string) error {
	chunk, err := parse.Parse(strings.NewReader(source), name)
	if err != nil {
		return err
	}

	if _, err := lua.Compile(chunk, name); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}

	return nil
}
