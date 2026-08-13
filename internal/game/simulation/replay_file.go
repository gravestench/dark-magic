package simulation

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteReplayContainerFile atomically replaces one replay file. Encoding and
// syncing complete in the destination directory before rename, so readers see
// either the previous complete replay or the new complete replay.
func WriteReplayContainerFile(path string, container ReplayContainer) error {
	if path == "" {
		return fmt.Errorf("%w: path is required", ErrReplayContainer)
	}
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".dark-magic-replay-*")
	if err != nil {
		return fmt.Errorf("%w: create temporary file: %v", ErrReplayContainer, err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("%w: protect temporary file: %v", ErrReplayContainer, err)
	}
	if err := EncodeReplayContainer(temporary, container); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("%w: sync temporary file: %v", ErrReplayContainer, err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("%w: close temporary file: %v", ErrReplayContainer, err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("%w: replace file: %v", ErrReplayContainer, err)
	}
	directoryHandle, err := os.Open(directory)
	if err != nil {
		return fmt.Errorf("%w: open destination directory: %v", ErrReplayContainer, err)
	}
	defer directoryHandle.Close()
	if err := directoryHandle.Sync(); err != nil {
		return fmt.Errorf("%w: sync destination directory: %v", ErrReplayContainer, err)
	}
	return nil
}

func ReadReplayContainerFile(path string, limits ReplayContainerLimits,
	migrations map[uint32]ReplayContainerMigration,
) (ReplayContainer, error) {
	if path == "" {
		return ReplayContainer{}, fmt.Errorf("%w: path is required", ErrReplayContainer)
	}
	file, err := os.Open(path)
	if err != nil {
		return ReplayContainer{}, fmt.Errorf("%w: open file: %v", ErrReplayContainer, err)
	}
	defer file.Close()
	return DecodeReplayContainerWithMigrations(file, limits, migrations)
}
