// Package envconfig installs and loads the per-process environment files used
// by Dark Magic composition roots. Exported process variables remain
// authoritative; files provide local defaults without replacing CLI flags.
package envconfig

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

// Result records both the role default and the file actually loaded. Commands
// need the default for flag help even when an explicit file won selection.
type Result struct {
	Role        string
	DefaultPath string
	LoadedPath  string
	Created     bool
}

// Duration centralizes duration validation for composition roots. Rejecting zero
// and negative values prevents maintenance or timeout loops from becoming busy loops.
func Duration(name string, fallback time.Duration) (time.Duration, error) {
	if strings.TrimSpace(name) == "" || fallback <= 0 {
		return 0, errors.New("environment duration requires a name and positive fallback")
	}

	value := environmentValue(name)
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("invalid %s duration %q", name, value)
	}

	return parsed, nil
}

// Bootstrap guarantees a private role template exists before selecting a file,
// then loads file values without overriding already-exported process authority.
func Bootstrap(role string, arguments []string) (Result, error) {
	defaultPath, created, err := Install(role)
	if err != nil {
		return Result{}, err
	}

	loadedPath, err := selectedEnvironmentPath(defaultPath, arguments)
	if err != nil {
		return Result{}, err
	}

	if err := Load(loadedPath); err != nil {
		return Result{}, err
	}

	return Result{
		Role:        role,
		DefaultPath: defaultPath,
		LoadedPath:  loadedPath,
		Created:     created,
	}, nil
}

// selectedEnvironmentPath gives an explicit --env-file selection precedence while
// retaining the installed role default when the command line makes no choice.
func selectedEnvironmentPath(defaultPath string, arguments []string) (string, error) {
	explicitPath, err := ExplicitPath(arguments)
	if err != nil {
		return "", err
	}

	if explicitPath == "" {
		return defaultPath, nil
	}

	expandedPath, err := darkpaths.ExpandHost(explicitPath)
	if err != nil {
		return "", fmt.Errorf("environment file: %w", err)
	}

	return expandedPath, nil
}

// environmentValue treats surrounding whitespace as configuration formatting,
// not part of role/path policy interpreted by the composition root.
func environmentValue(name string) string {
	return strings.TrimSpace(os.Getenv(name))
}
