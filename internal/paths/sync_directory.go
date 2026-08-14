//go:build !windows

package paths

import "os"

// SyncDirectory makes a completed atomic rename durable on platforms that
// support syncing directory metadata.
func SyncDirectory(name string) error {
	directory, err := os.Open(name)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}
