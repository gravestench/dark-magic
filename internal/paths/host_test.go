package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandHostHomeAliases(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{"~/games/diablo", `~\games\diablo`} {
		got, err := ExpandHost(input)
		if err != nil {
			t.Fatal(err)
		}
		if want := filepath.Join(home, "games", "diablo"); got != want {
			t.Fatalf("ExpandHost(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestExpandHostEnvironmentFormsOnEveryPlatform(t *testing.T) {
	root := t.TempDir()
	t.Setenv("DARK_MAGIC_PATH_TEST", root)
	for _, input := range []string{"$DARK_MAGIC_PATH_TEST/data", "${DARK_MAGIC_PATH_TEST}/data", `%DARK_MAGIC_PATH_TEST%\data`} {
		got, err := ExpandHost(input)
		if err != nil {
			t.Fatal(err)
		}
		if want := filepath.Join(root, "data"); filepath.Clean(got) != want {
			t.Fatalf("ExpandHost(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestExpandHostRejectsMissingAliases(t *testing.T) {
	for _, input := range []string{"$DARK_MAGIC_MISSING_TEST/data", `%DARK_MAGIC_MISSING_TEST%\data`, "~someone/data"} {
		if _, err := ExpandHost(input); err == nil {
			t.Fatalf("ExpandHost(%q) succeeded", input)
		}
	}
}

func TestExpandHostPreservesOrdinaryPaths(t *testing.T) {
	for _, input := range []string{"relative/path", "/absolute/path", `C:\\Games\\Diablo II`} {
		got, err := ExpandHost(input)
		if err != nil || got != input {
			t.Fatalf("ExpandHost(%q) = %q, %v", input, got, err)
		}
	}
}
