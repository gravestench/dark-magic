package main

import (
	"log/slog"

	"github.com/gravestench/dark-magic/internal/dev/profiling"
)

// startProfiler creates an optional profiling session from the command configuration.
func startProfiler(directory, scenes string) (*profiling.Session, error) {
	if directory == "" {
		return nil, nil
	}

	profile, err := profiling.Start(directory, true)
	if err != nil {
		return nil, err
	}

	profile.ConfigureScenes(scenes)

	return profile, nil
}

// stopProfiler reports cleanup failures that occur after the client has stopped.
func stopProfiler(profile *profiling.Session) {
	if err := profile.Stop(); err != nil {
		slog.Error("finishing profiler", "error", err)
	}
}
