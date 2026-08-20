package realm

import (
	"os"
	"path/filepath"
	"testing"
)

// TestOperatorTokenIsStableOwnerOnlyAndRejectsUnsafeFiles verifies operator token is stable owner only and rejects
// unsafe files. The scenario keeps the operator contract visible to maintainers.
func TestOperatorTokenIsStableOwnerOnlyAndRejectsUnsafeFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "private", "operator-token")

	first, err := LoadOrCreateOperatorToken(path)
	if err != nil || len(first) < 32 {
		t.Fatalf("first token length=%d error=%v", len(first), err)
	}

	second, err := LoadOrCreateOperatorToken(path)
	if err != nil || second != first {
		t.Fatalf("second token differs error=%v", err)
	}

	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("token mode=%v error=%v", info.Mode(), err)
	}

	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadOrCreateOperatorToken(path); err == nil {
		t.Fatal("group-readable operator token was accepted")
	}
}
