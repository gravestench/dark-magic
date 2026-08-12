package content

import (
	"archive/zip"
	"bytes"
	"io/fs"
	"testing"
)

func TestD2LegacyArchiveMatchesEmbeddedTreeAndIsDeterministic(t *testing.T) {
	var first, second bytes.Buffer
	if err := WriteD2LegacyArchive(&first); err != nil {
		t.Fatal(err)
	}
	if err := WriteD2LegacyArchive(&second); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first.Bytes(), second.Bytes()) {
		t.Fatal("d2legacy archive is not deterministic")
	}
	archive, err := zip.NewReader(bytes.NewReader(first.Bytes()), int64(first.Len()))
	if err != nil {
		t.Fatal(err)
	}
	want := D2Legacy()
	seen := 0
	if err := fs.WalkDir(want, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		expected, err := fs.ReadFile(want, name)
		if err != nil {
			return err
		}
		actual, err := fs.ReadFile(archive, name)
		if err != nil {
			return err
		}
		if !bytes.Equal(actual, expected) {
			t.Fatalf("archived %s differs", name)
		}
		seen++
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if seen == 0 {
		t.Fatal("d2legacy archive is empty")
	}
}
