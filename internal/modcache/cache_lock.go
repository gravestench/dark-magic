package modcache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	cacheLockTimeout = 10 * time.Second
	cacheLockStale   = time.Minute
)

// withMutationLock serializes the short read-modify-write sections shared by
// multiple Dark Magic processes. Atomically creating a directory is portable,
// avoids Windows lock-file open/delete races, and leaves stale ownership
// recoverable after a crashed process without relying on platform-only flock.
func (store *Store) withMutationLock(operation func() error) error {
	// Avoid making the filesystem arbitrate goroutines sharing one Store. The
	// directory token remains necessary for independent Store instances and
	// processes, but Windows has a brief remove/create interval in which it may
	// report ACCESS_DENIED instead of either EXISTS or success.
	store.mutation.Lock()
	defer store.mutation.Unlock()

	deadline := time.Now().Add(cacheLockTimeout)
	lockPath := filepath.Join(store.root, ".mutation.lock")
	permissionRaces := 0
	for {
		err := os.Mkdir(lockPath, 0o700)
		if err == nil {
			operationErr := operation()
			removeErr := os.Remove(lockPath)
			return errors.Join(operationErr, removeErr)
		}
		contended, inspectErr := mutationLockContended(lockPath, err)
		if inspectErr != nil {
			return fmt.Errorf("modcache: acquire mutation lock: %w", inspectErr)
		}
		if !contended {
			// If another Windows owner removed the directory between our failed
			// Mkdir and Stat, the path is already gone but Mkdir's error remains
			// ACCESS_DENIED. Retry that narrow race briefly. A persistent parent
			// permission failure still returns its original error promptly.
			if os.IsPermission(err) && permissionRaces < 10 {
				permissionRaces++
				time.Sleep(time.Millisecond)
				continue
			}
			return fmt.Errorf("modcache: acquire mutation lock: %w", err)
		}
		permissionRaces = 0
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
