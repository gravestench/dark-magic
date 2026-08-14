package modcache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	cacheLockTimeout = 10 * time.Second
	cacheLockStale   = time.Minute
)

// withMutationLock serializes the short read-modify-write sections shared by
// multiple Dark Magic processes. O_EXCL is portable; stale ownership is
// recoverable after a crashed process without relying on platform-only flock.
func (store *Store) withMutationLock(operation func() error) error {
	deadline := time.Now().Add(cacheLockTimeout)
	lockPath := filepath.Join(store.root, ".mutation.lock")
	for {
		file, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err == nil {
			_, writeErr := file.WriteString(strconv.Itoa(os.Getpid()))
			closeErr := file.Close()
			if writeErr != nil || closeErr != nil {
				_ = os.Remove(lockPath)
				return errors.Join(writeErr, closeErr)
			}
			operationErr := operation()
			removeErr := os.Remove(lockPath)
			return errors.Join(operationErr, removeErr)
		}
		if !os.IsExist(err) {
			return fmt.Errorf("modcache: acquire mutation lock: %w", err)
		}
		if info, statErr := os.Stat(lockPath); statErr == nil && time.Since(info.ModTime()) > cacheLockStale {
			_ = os.Remove(lockPath)
			continue
		}
		if time.Now().After(deadline) {
			return errors.New("modcache: timed out waiting for another process to finish updating the cache")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
