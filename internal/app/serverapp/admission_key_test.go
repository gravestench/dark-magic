package serverapp

import (
	"os"
	"path/filepath"
	"testing"
)

// TestReadAdmissionKeyIsProtectedAndBounded covers the filesystem trust and
// size constraints before a secret can become ticket or control authority.
func TestReadAdmissionKeyIsProtectedAndBounded(t *testing.T) {
	directory := t.TempDir()
	validPath := writeAdmissionKeyFixture(
		t,
		directory,
		"valid",
		[]byte("0123456789abcdef0123456789abcdef\n"),
		0o600,
	)

	key, err := ReadAdmissionKey(validPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(key) != "0123456789abcdef0123456789abcdef" {
		t.Fatalf("key = %q", key)
	}

	invalidCases := map[string]struct {
		data []byte
		mode os.FileMode
	}{
		"short": {
			data: []byte("short"),
			mode: 0o600,
		},
		"large": {
			data: make([]byte, maximumAdmissionKeyBytes+1),
			mode: 0o600,
		},
		"exposed": {
			data: []byte("0123456789abcdef0123456789abcdef"),
			mode: 0o644,
		},
	}
	for name, testCase := range invalidCases {
		t.Run(name, func(t *testing.T) {
			path := writeAdmissionKeyFixture(t, directory, name, testCase.data, testCase.mode)
			if _, err := ReadAdmissionKey(path); err == nil {
				t.Fatalf("%s key was accepted", name)
			}
		})
	}
}

// writeAdmissionKeyFixture makes each permission mode explicit so a failing
// trust-boundary case identifies the exact on-disk input that was accepted.
func writeAdmissionKeyFixture(
	t *testing.T,
	directory string,
	name string,
	data []byte,
	mode os.FileMode,
) string {
	t.Helper()

	path := filepath.Join(directory, name+".key")
	if err := os.WriteFile(path, data, mode); err != nil {
		t.Fatal(err)
	}

	return path
}
