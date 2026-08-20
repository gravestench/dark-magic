// Command client is the native client composition root. It parses process
// policy, constructs owned capabilities, and translates final errors to exit
// status; game rules and reusable behavior belong under internal packages.
package main

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"

	"github.com/gravestench/dark-magic/internal/app/envconfig"
	"github.com/gravestench/dark-magic/internal/dev/capture"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

// main owns the two policies that only the executable can apply: pinning the
// renderer to the initial OS thread and converting runMain's result into exit status.
func main() {
	// macOS requires the window to live on the first operating-system thread.
	// Locking here keeps that rule out of the rest of the application.
	runtime.LockOSThread()
	os.Exit(runMain())
}

// runMain acquires process resources in dependency order and defers their cleanup
// immediately. Returning an integer keeps os.Exit out of deferred-cleanup code.
func runMain() int {
	environment, err := envconfig.Bootstrap("client", os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	config, err := parseClientConfig(environment.DefaultPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	logs := configureLogging(config.logLevel)

	profile, err := startProfiler(config.profileDirectory, config.profileScenes)
	if err != nil {
		slog.Error("starting profiler", "error", err)
		return 1
	}

	if profile != nil {
		defer stopProfiler(profile)
	}

	mods, contentFS, err := prepareContent(config.mods)
	if err != nil {
		slog.Error("preparing client content", "error", err)
		return 1
	}
	defer closeContent(mods)

	config.captureDirectory, config.captureScenes = capture.Defaults(
		config.captureDirectory,
		config.captureScenes,
	)

	config.captureDirectory, err = darkpaths.ExpandHost(config.captureDirectory)
	if err != nil {
		slog.Error("expanding capture directory", "error", err)
		return 1
	}

	if err := run(contentFS, mods, profile, config, logs); err != nil {
		slog.Error("running Dark Magic", "error", err)
		return 1
	}

	return 0
}
