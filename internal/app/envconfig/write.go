package envconfig

import (
	"os"
	"path/filepath"
)

// writePrivate atomically replaces a file with owner-only permissions.
func writePrivate(path string, data []byte) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".env-*")
	if err != nil {
		return err
	}

	temporaryPath := temporary.Name()
	defer func() {
		_ = temporary.Close()
		_ = os.Remove(temporaryPath)
	}()

	if err := temporary.Chmod(0o600); err != nil {
		return err
	}

	if _, err := temporary.Write(data); err != nil {
		return err
	}

	if err := temporary.Sync(); err != nil {
		return err
	}

	if err := temporary.Close(); err != nil {
		return err
	}

	return os.Rename(temporaryPath, path)
}
