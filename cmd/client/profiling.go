package main

import (
	"log/slog"

	"github.com/gravestench/dark-magic/internal/dev/profiling"
)

// startProfiler keeps profiling opt-in and process-owned. A nil session means
// callers can run the normal client path without installing profiling hooks.
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

// stopProfiler flushes final reports after client shutdown. At that point an
// error cannot be returned through run, so logging is the remaining observable path.
func stopProfiler(profile *profiling.Session) {
	if err := profile.Stop(); err != nil {
		slog.Error("finishing profiler", "error", err)
	}
}
