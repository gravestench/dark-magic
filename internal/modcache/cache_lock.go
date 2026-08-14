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
		contended, inspectErr := mutationLockContended(lockPath, err)
		if inspectErr != nil {
			return fmt.Errorf("modcache: acquire mutation lock: %w", inspectErr)
		}
		if !contended {
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

// Windows may report ERROR_ACCESS_DENIED, rather than ERROR_FILE_EXISTS, when
// another goroutine or process owns an O_EXCL lock file. Stat distinguishes
// that contention from an actual permission failure without weakening the
// cross-process lock contract on other platforms.
func mutationLockContended(lockPath string, createErr error) (bool, error) {
	if os.IsExist(createErr) {
		return true, nil
	}
	_, statErr := os.Stat(lockPath)
	if statErr == nil {
		return true, nil
	}
	if os.IsNotExist(statErr) {
		return false, nil
	}
	return false, errors.Join(createErr, statErr)
}
