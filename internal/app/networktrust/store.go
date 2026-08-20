// Package networktrust persists direct-game host identity and client TOFU pins.
package networktrust

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Store serializes access to host identities and client trust pins in one directory.
type Store struct {
	mu  sync.Mutex
	dir string
}

// New validates a trust directory and returns a store without touching the filesystem.
func New(dir string) (*Store, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, errors.New("network trust: directory is required")
	}

	return &Store{dir: dir}, nil
}

// Directory derives the trust directory beside explicit preferences, or under the user's default configuration root.
func Directory(preferencesPath string) (string, error) {
	if preferencesPath != "" {
		return filepath.Join(filepath.Dir(preferencesPath), "network"), nil
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("network trust: user config directory: %w", err)
	}

	return filepath.Join(dir, "dark-magic", "network"), nil
}
