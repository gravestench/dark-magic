package modcache

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// TestMutationLockUsesDirectoryTokenAndReleasesIt verifies the cross-process
// token exists during mutation and is removed before the operation returns.
func TestMutationLockUsesDirectoryTokenAndReleasesIt(t *testing.T) {
	root := t.TempDir()

	store, err := New(root)
	if err != nil {
		t.Fatal(err)
	}

	lockPath := filepath.Join(root, ".mutation.lock")

	if err := store.withMutationLock(func() error {
		info, err := os.Stat(lockPath)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			t.Fatalf("mutation lock token is not a directory")
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Fatalf("released mutation lock remains: %v", err)
	}
}

// TestMutationLockRecognizesWindowsPermissionShapedContention protects the
// Windows case where an owned token surfaces as permission denied, not exists.
func TestMutationLockRecognizesWindowsPermissionShapedContention(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), ".mutation.lock")
	if err := os.WriteFile(lockPath, []byte("owner"), 0o600); err != nil {
		t.Fatal(err)
	}

	contended, err := mutationLockContended(lockPath, fs.ErrPermission)
	if err != nil {
		t.Fatal(err)
	}

	if !contended {
		t.Fatal("existing lock was not recognized as contention")
	}
}

// TestMutationLockDoesNotHideMissingPermissionFailure ensures a true parent
// permission problem is not retried as if another process owned the lock.
func TestMutationLockDoesNotHideMissingPermissionFailure(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), ".mutation.lock")

	contended, err := mutationLockContended(lockPath, fs.ErrPermission)
	if err != nil {
		t.Fatal(err)
	}

	if contended {
		t.Fatal("missing lock was reported as contention")
	}
}
