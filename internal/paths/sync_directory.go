//go:build !windows

package paths

import "os"

// SyncDirectory flushes directory metadata on platforms that support it. Callers
// use this after separately syncing and renaming a file so the namespace update,
// not the file contents themselves, is included in the durability sequence.
func SyncDirectory(name string) error {
	directory, err := os.Open(name)
	if err != nil {
		return err
	}
	defer func() { _ = directory.Close() }()

	return directory.Sync()
}
