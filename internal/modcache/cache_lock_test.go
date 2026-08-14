package modcache

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

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
