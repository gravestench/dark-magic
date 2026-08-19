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

// Result describes which environment file Bootstrap installed and loaded.
type Result struct {
	Role        string
	DefaultPath string
	LoadedPath  string
	Created     bool
}

// Duration returns a positive environment duration or the supplied fallback.
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

// Bootstrap installs the role template and loads the selected environment file.
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

// selectedEnvironmentPath resolves an explicit flag or preserves the role default.
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

// environmentValue returns an environment value without surrounding whitespace.
func environmentValue(name string) string {
	return strings.TrimSpace(os.Getenv(name))
}
