package fileLoader

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestAddMPQDirectorySourcesUsesPatchPrecedence(t *testing.T) {
	directory := t.TempDir()
	for _, name := range []string{"d2data.mpq", "patch_d2.mpq"} {
		file, err := os.Create(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		file.Close()
	}
	service := New()
	service.SetLogger(slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := service.addMPQDirectorySources(directory); err != nil {
		t.Fatal(err)
	}
	sources := service.fsGroups[DefaultGroup]
	if len(sources) != 2 {
		t.Fatalf("source count = %d, want 2", len(sources))
	}
	if filepath.Base(sources[0].Path) != "patch_d2.mpq" || filepath.Base(sources[1].Path) != "d2data.mpq" {
		t.Fatalf("source order = %v", sources)
	}
}
