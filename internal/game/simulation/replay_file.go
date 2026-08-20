package simulation

import (
	"fmt"
	"os"
	"path/filepath"

	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

// WriteReplayContainerFile atomically replaces one replay file. A temporary file in the destination directory is
// encoded and synced before rename, then the directory is synced so readers see either complete version after a crash.
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

	// Rename removes this path on success; the deferred removal protects every
	// failure path without risking the previously complete destination file.
	defer func() { _ = os.Remove(temporaryPath) }()

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

	// Preserve the write-then-rename-then-directory-sync ordering required for
	// crash-safe replacement on filesystems that support these guarantees.
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("%w: replace file: %v", ErrReplayContainer, err)
	}

	if err := darkpaths.SyncDirectory(directory); err != nil {
		return fmt.Errorf("%w: sync destination directory: %v", ErrReplayContainer, err)
	}

	return nil
}

// ReadReplayContainerFile opens one replay for bounded decoding and leaves
// version migration policy with the caller; closing is guaranteed on every
// decode outcome.
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
	defer func() { _ = file.Close() }()

	return DecodeReplayContainerWithMigrations(file, limits, migrations)
}
