package main

import (
	"flag"
	"fmt"
	"os"

	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

// main reports startup failures at the process boundary so the interactive runtime can keep explicit resource
// ownership.
func main() {
	if err := runSceneDemo(flag.CommandLine, os.Args[1:]); err != nil {
		fatal(err.Error())
	}
}

// runSceneDemo prepares host and content inputs before opening a window, preventing startup errors from leaking
// graphics resources.
func runSceneDemo(flagSet *flag.FlagSet, arguments []string) error {
	config, err := parseDemoConfig(flagSet, arguments)
	if err != nil {
		return err
	}

	savePath, err := darkpaths.ExpandHost(config.savePath)
	if err != nil {
		return err
	}

	mapPNG, err := loadMapPreview(config)
	if err != nil {
		return err
	}

	startInteractiveScene(config, savePath, mapPNG)

	return nil
}

// fatal preserves the command's historical stderr text and exit status for startup failures.
func fatal(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
