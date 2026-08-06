package content

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenSourceExpandsEnvironmentAlias(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "asset.txt"), []byte("ok"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DARK_MAGIC_SOURCE_TEST", directory)
	source, err := OpenSource("$DARK_MAGIC_SOURCE_TEST")
	if err != nil {
		t.Fatal(err)
	}
	data, err := fs.ReadFile(source, "asset.txt")
	if err != nil || string(data) != "ok" {
		t.Fatalf("data = %q, err = %v", data, err)
	}
}

type missingMPQFS struct{ wrapped bool }

func (m missingMPQFS) Open(string) (fs.File, error) {
	if m.wrapped {
		return nil, errors.New("getting file stream: file not found")
	}
	return nil, errors.New("file not found")
}

func TestNormalizedMPQTranslatesMissingFile(t *testing.T) {
	source := normalizedFS{FS: missingMPQFS{}, backslash: true}
	_, err := source.Open("data/global/example.dc6")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected fs.ErrNotExist, got %v", err)
	}
}

func TestNormalizedMPQTranslatesWrappedMissingFile(t *testing.T) {
	source := normalizedFS{FS: missingMPQFS{wrapped: true}, backslash: true}
	_, err := source.Open("data/global/example.dc6")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected fs.ErrNotExist, got %v", err)
	}
}
