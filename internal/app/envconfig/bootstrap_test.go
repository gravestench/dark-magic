package envconfig

import (
	"os"
	"path/filepath"
	"testing"
)

var clientEnvironmentKeys = []string{
	"MPQ_DIRECTORY",
	"DARK_MAGIC_LOG_LEVEL",
	"DARK_MAGIC_FULLSCREEN",
	"DARK_MAGIC_VIEWPORT_FIT",
	"DARK_MAGIC_MODS",
	"DARK_MAGIC_OUTPUT_PALETTE",
	"DARK_MAGIC_PRESENTATION_PROFILE",
	"DARK_MAGIC_PLAYER_PROFILE",
	"DARK_MAGIC_PREFERENCES",
	"DARK_MAGIC_SHELL_CONFIG",
	"DARK_MAGIC_DEBUG_ADDR",
}

// TestBootstrapInstallsDefaultAndPreservesExportedAuthority verifies that
// process variables remain authoritative over values in an installed file.
func TestBootstrapInstallsDefaultAndPreservesExportedAuthority(t *testing.T) {
	preserveEnvironment(t, clientEnvironmentKeys...)
	directory := t.TempDir()
	t.Setenv("DARK_MAGIC_CONFIG_DIR", directory)
	t.Setenv("DARK_MAGIC_LOG_LEVEL", "trace")

	result, err := Bootstrap("client", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Created || result.LoadedPath != filepath.Join(directory, "client.env") {
		t.Fatalf("result = %#v", result)
	}
	if got := os.Getenv("DARK_MAGIC_LOG_LEVEL"); got != "trace" {
		t.Fatalf("exported authority = %q", got)
	}
	assertPrivateFile(t, result.DefaultPath)
}

// TestBootstrapExplicitFileOverridesDefaultSelection verifies that Bootstrap
// still installs the role default before loading an explicit file.
func TestBootstrapExplicitFileOverridesDefaultSelection(t *testing.T) {
	const key = "DARK_MAGIC_ENVCONFIG_EXPLICIT"
	preserveEnvironment(t, key)
	directory := t.TempDir()
	t.Setenv("DARK_MAGIC_CONFIG_DIR", directory)
	explicitPath := filepath.Join(t.TempDir(), "custom.env")
	if err := os.WriteFile(explicitPath, []byte(key+"=loaded\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := Bootstrap("server", []string{"--env-file", explicitPath})
	if err != nil {
		t.Fatal(err)
	}
	if result.LoadedPath != explicitPath || os.Getenv(key) != "loaded" {
		t.Fatalf("result = %#v, value = %q", result, os.Getenv(key))
	}
	if _, err := os.Stat(filepath.Join(directory, "server.env")); err != nil {
		t.Fatalf("default template was not installed: %v", err)
	}
}

// assertPrivateFile verifies that an installed environment file is owner-only.
func assertPrivateFile(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if permission := info.Mode().Perm(); permission != 0o600 {
		t.Fatalf("default mode = %v", permission)
	}
}
