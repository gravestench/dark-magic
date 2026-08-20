package save

import (
	"fmt"
	"os"
	"path/filepath"

	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

// WriteProfileFile durably replaces one private profile through a same-directory temporary file.
// Same-directory rename preserves atomicity, and directory sync makes the replacement survive a host crash.
func WriteProfileFile(path string, profile Profile) error {
	if path == "" {
		return fmt.Errorf("%w: path is required", ErrProfile)
	}

	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("%w: create profile directory: %v", ErrProfile, err)
	}

	temporary, err := os.CreateTemp(directory, ".dark-magic-profile-*")
	if err != nil {
		return fmt.Errorf("%w: create temporary file: %v", ErrProfile, err)
	}

	temporaryPath := temporary.Name()
	defer removeTemporaryProfile(temporaryPath)

	// Restrict permissions before writing any player data so a partial write is never broadly readable.
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("%w: protect temporary file: %v", ErrProfile, err)
	}

	if err := EncodeProfile(temporary, profile); err != nil {
		_ = temporary.Close()
		return err
	}

	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("%w: sync temporary file: %v", ErrProfile, err)
	}

	if err := temporary.Close(); err != nil {
		return fmt.Errorf("%w: close temporary file: %v", ErrProfile, err)
	}

	// Rename publishes the complete file atomically; syncing the directory persists that namespace update.
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("%w: replace profile: %v", ErrProfile, err)
	}

	if err := darkpaths.SyncDirectory(directory); err != nil {
		return fmt.Errorf("%w: sync profile directory: %v", ErrProfile, err)
	}

	return nil
}

// removeTemporaryProfile performs best-effort cleanup after either failure or successful rename.
func removeTemporaryProfile(path string) {
	_ = os.Remove(path)
}

// ReadProfileFile opens one profile and delegates bounded decoding without transferring file ownership to callers.
func ReadProfileFile(path string) (Profile, error) {
	if path == "" {
		return Profile{}, fmt.Errorf("%w: path is required", ErrProfile)
	}

	file, err := os.Open(path)
	if err != nil {
		return Profile{}, fmt.Errorf("%w: open profile: %w", ErrProfile, err)
	}
	defer closeProfileFile(file)

	return DecodeProfile(file)
}

// closeProfileFile ignores a read-only close error because decoding has already determined the operation result.
func closeProfileFile(file *os.File) {
	_ = file.Close()
}
