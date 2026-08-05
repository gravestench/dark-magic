package fileLoader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSourceType(t *testing.T) {
	directory := t.TempDir()
	archive := filepath.Join(directory, "data.MPQ")
	if err := os.WriteFile(archive, []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}

	if got := NewSource(directory).Type(); got != SourceDirectory {
		t.Fatalf("directory type = %v", got)
	}
	if got := NewSource(archive).Type(); got != SourceArchive {
		t.Fatalf("archive type = %v", got)
	}
	if got := NewSource(filepath.Join(directory, "missing.mpq")).Type(); got != SourceUnknown {
		t.Fatalf("missing path type = %v", got)
	}
}

func TestDirectoryFilesystemNormalizesGamePaths(t *testing.T) {
	directory := t.TempDir()
	if err := os.MkdirAll(filepath.Join(directory, "data", "global"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "data", "global", "test.txt"), []byte("ok"), 0o600); err != nil {
		t.Fatal(err)
	}

	filesystem, err := NewSource(directory).Filesystem()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/data/global/test.txt", `data\global\test.txt`} {
		file, err := filesystem.Open(path)
		if err != nil {
			t.Fatalf("opening normalized path %q: %v", path, err)
		}
		file.Close()
	}
}
