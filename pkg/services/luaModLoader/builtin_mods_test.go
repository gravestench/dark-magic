package luaModLoader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallBuiltinModsDoesNotOverwriteUserFiles(t *testing.T) {
	modDirectory := t.TempDir()
	service := &Service{Config: &Config{ModDirectory: modDirectory}}
	target := filepath.Join(modDirectory, "terminal", "init.lua")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("-- user version"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := service.installBuiltinMods(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "-- user version" {
		t.Fatal("built-in installer overwrote a user file")
	}
	if _, err := os.Stat(filepath.Join(modDirectory, "terminal", "manifest.json")); err != nil {
		t.Fatalf("expected missing built-in file to be installed: %v", err)
	}
}

func TestExpandSourcePath(t *testing.T) {
	t.Setenv("MPQ_DIRECTORY", "/game/mpq")
	if got, want := expandSourcePath("{{MPQ_DIRECTORY}}/d2data.mpq", "/mod"), "/game/mpq/d2data.mpq"; got != want {
		t.Fatalf("expanded path = %q, want %q", got, want)
	}
	if got, want := expandSourcePath("data/mod.mpq", "/mod"), "/mod/data/mod.mpq"; got != want {
		t.Fatalf("relative path = %q, want %q", got, want)
	}
}
