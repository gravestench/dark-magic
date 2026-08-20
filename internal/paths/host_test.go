package paths

import (
	"os"
	"path/filepath"
	"testing"
)

// TestExpandHostHomeAliases verifies slash styles resolve to the same current
// user directory on every supported platform.
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

// TestExpandHostEnvironmentFormsOnEveryPlatform pins shell and Windows variable
// syntax to one platform-independent expansion contract.
func TestExpandHostEnvironmentFormsOnEveryPlatform(t *testing.T) {
	root := t.TempDir()
	t.Setenv("DARK_MAGIC_PATH_TEST", root)

	inputs := []string{
		"$DARK_MAGIC_PATH_TEST/data",
		"${DARK_MAGIC_PATH_TEST}/data",
		`%DARK_MAGIC_PATH_TEST%\data`,
	}
	for _, input := range inputs {
		got, err := ExpandHost(input)
		if err != nil {
			t.Fatal(err)
		}

		if want := filepath.Join(root, "data"); filepath.Clean(got) != want {
			t.Fatalf("ExpandHost(%q) = %q, want %q", input, got, want)
		}
	}
}

// TestExpandHostRejectsMissingAliases ensures typos cannot silently redirect I/O
// to an empty environment substitution or unsupported named home.
func TestExpandHostRejectsMissingAliases(t *testing.T) {
	for _, input := range []string{"$DARK_MAGIC_MISSING_TEST/data", `%DARK_MAGIC_MISSING_TEST%\data`, "~someone/data"} {
		if _, err := ExpandHost(input); err == nil {
			t.Fatalf("ExpandHost(%q) succeeded", input)
		}
	}
}

// TestExpandHostPreservesOrdinaryPaths ensures paths without recognized aliases
// retain their original spelling, including Windows separators on other hosts.
func TestExpandHostPreservesOrdinaryPaths(t *testing.T) {
	for _, input := range []string{"relative/path", "/absolute/path", `C:\\Games\\Diablo II`} {
		got, err := ExpandHost(input)
		if err != nil || got != input {
			t.Fatalf("ExpandHost(%q) = %q, %v", input, got, err)
		}
	}
}
