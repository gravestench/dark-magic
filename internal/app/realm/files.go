package realm

import (
	"encoding/json"
	"os"
	"path/filepath"

	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

// writeRealmJSON is reserved for owner-only operational rendezvous such as a
// worker ready record. Durable Realm domain state belongs in PostgreSQL.
func writeRealmJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(directory, ".dark-magic-realm-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err == nil {
		_, err = temporary.Write(data)
	}
	if syncErr := temporary.Sync(); err == nil {
		err = syncErr
	}
	if closeErr := temporary.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return err
	}
	return darkpaths.SyncDirectory(directory)
}

// DataDirectory resolves non-database Realm service state: TLS identity,
// worker rendezvous, and generated portal caches.
func DataDirectory(configured string) (string, error) {
	if configured != "" {
		return darkpaths.ExpandHost(configured)
	}
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "dark-magic", "realm"), nil
}
