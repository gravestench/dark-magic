package envconfig

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

//go:embed templates/*.env
var templates embed.FS

var roles = map[string]struct{}{"client": {}, "server": {}, "realm": {}}

// Install materializes a documented role template with owner-only permissions.
// Existing files are never overwritten because they may contain operator secrets.
func Install(role string) (string, bool, error) {
	if _, found := roles[role]; !found {
		return "", false, fmt.Errorf("unknown environment role %q", role)
	}

	directory, err := ensureConfigDirectory()
	if err != nil {
		return "", false, err
	}

	path := filepath.Join(directory, role+".env")

	exists, err := secureExistingFile(path)
	if err != nil {
		return "", false, err
	}

	if exists {
		return path, false, nil
	}

	data, err := templates.ReadFile("templates/" + role + ".env")
	if err != nil {
		return "", false, err
	}

	if err := writePrivate(path, data); err != nil {
		return "", false, fmt.Errorf("install environment file %q: %w", path, err)
	}

	return path, true, nil
}

// ensureConfigDirectory enforces owner-only access before any template or secret
// is written, including when the directory predates this process.
func ensureConfigDirectory() (string, error) {
	directory, err := configDirectory()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(directory, 0o700); err != nil {
		return "", fmt.Errorf("create environment directory: %w", err)
	}

	if err := os.Chmod(directory, 0o700); err != nil {
		return "", fmt.Errorf("secure environment directory: %w", err)
	}

	return directory, nil
}

// secureExistingFile preserves existing contents while repairing permissions.
// The Boolean lets Install distinguish an operator file from a missing default.
func secureExistingFile(path string) (bool, error) {
	_, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("inspect environment file %q: %w", path, err)
	}

	if err := os.Chmod(path, 0o600); err != nil {
		return false, fmt.Errorf("secure environment file %q: %w", path, err)
	}

	return true, nil
}

// configDirectory supports portable deployments through an explicit host path
// and otherwise keeps per-user process policy in the platform configuration area.
func configDirectory() (string, error) {
	if configured := environmentValue("DARK_MAGIC_CONFIG_DIR"); configured != "" {
		expanded, err := darkpaths.ExpandHost(configured)
		if err != nil {
			return "", fmt.Errorf("environment directory: %w", err)
		}

		return expanded, nil
	}

	root, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve environment directory: %w", err)
	}

	return filepath.Join(root, "dark-magic"), nil
}
