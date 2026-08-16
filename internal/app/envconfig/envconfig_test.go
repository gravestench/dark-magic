package envconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBootstrapInstallsDefaultAndPreservesExportedAuthority(t *testing.T) {
	preserveEnvironment(t, "MPQ_DIRECTORY", "DARK_MAGIC_LOG_LEVEL", "DARK_MAGIC_FULLSCREEN", "DARK_MAGIC_VIEWPORT_FIT", "DARK_MAGIC_MODS", "DARK_MAGIC_OUTPUT_PALETTE", "DARK_MAGIC_PRESENTATION_PROFILE", "DARK_MAGIC_PLAYER_PROFILE", "DARK_MAGIC_PREFERENCES", "DARK_MAGIC_SHELL_CONFIG", "DARK_MAGIC_DEBUG_ADDR")
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
	info, err := os.Stat(result.DefaultPath)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("default mode = %v, error = %v", info.Mode().Perm(), err)
	}
}

func TestBootstrapExplicitFileOverridesDefaultSelection(t *testing.T) {
	preserveEnvironment(t, "DARK_MAGIC_ENVCONFIG_EXPLICIT")
	directory := t.TempDir()
	t.Setenv("DARK_MAGIC_CONFIG_DIR", directory)
	explicit := filepath.Join(t.TempDir(), "custom.env")
	if err := os.WriteFile(explicit, []byte("DARK_MAGIC_ENVCONFIG_EXPLICIT=loaded\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := Bootstrap("server", []string{"--env-file", explicit})
	if err != nil {
		t.Fatal(err)
	}
	if result.LoadedPath != explicit || os.Getenv("DARK_MAGIC_ENVCONFIG_EXPLICIT") != "loaded" {
		t.Fatalf("result = %#v, value = %q", result, os.Getenv("DARK_MAGIC_ENVCONFIG_EXPLICIT"))
	}
	if _, err := os.Stat(filepath.Join(directory, "server.env")); err != nil {
		t.Fatalf("default template was not installed: %v", err)
	}
}

func preserveEnvironment(t *testing.T, keys ...string) {
	t.Helper()
	for _, key := range keys {
		key := key
		value, found := os.LookupEnv(key)
		t.Cleanup(func() {
			if found {
				_ = os.Setenv(key, value)
			} else {
				_ = os.Unsetenv(key)
			}
		})
	}
}

func TestParseReadableDotEnvSyntax(t *testing.T) {
	values, err := Parse(strings.NewReader("# comment\nexport ONE=plain\nTWO='two words'\nTHREE=\"line\\nthree\"\nFOUR=value # comment\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"ONE": "plain", "TWO": "two words", "THREE": "line\nthree", "FOUR": "value"}
	for key, value := range want {
		if values[key] != value {
			t.Errorf("%s = %q, want %q", key, values[key], value)
		}
	}
}

func TestExplicitPathRejectsAmbiguousFlags(t *testing.T) {
	if _, err := ExplicitPath([]string{"--env-file=one", "--env-file", "two"}); err == nil {
		t.Fatal("duplicate --env-file was accepted")
	}
}

func TestUpdatePreservesTemplateAndRejectsUnknownKeys(t *testing.T) {
	directory := t.TempDir()
	t.Setenv("DARK_MAGIC_CONFIG_DIR", directory)
	path, err := Update("client", map[string]string{"MPQ_DIRECTORY": "/Games/Diablo II"})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `MPQ_DIRECTORY="/Games/Diablo II"`) || !strings.Contains(string(data), "# Dark Magic client") {
		t.Fatalf("updated file = %s", data)
	}
	if _, err := Update("client", map[string]string{"TYPO": "value"}); err == nil {
		t.Fatal("unknown template key was accepted")
	}
}

func TestDurationUsesFallbackAndRejectsInvalidValues(t *testing.T) {
	const name = "DARK_MAGIC_TEST_DURATION"
	t.Setenv(name, "")
	if value, err := Duration(name, 15*time.Second); err != nil || value != 15*time.Second {
		t.Fatalf("fallback duration=%s error=%v", value, err)
	}
	t.Setenv(name, "250ms")
	if value, err := Duration(name, 15*time.Second); err != nil || value != 250*time.Millisecond {
		t.Fatalf("configured duration=%s error=%v", value, err)
	}
	for _, invalid := range []string{"invalid", "0s", "-1s"} {
		t.Setenv(name, invalid)
		if _, err := Duration(name, 15*time.Second); err == nil {
			t.Fatalf("invalid duration %q was accepted", invalid)
		}
	}
}
