package envconfig

import "testing"

// TestExplicitPathRejectsAmbiguousFlags verifies duplicate selector rejection.
func TestExplicitPathRejectsAmbiguousFlags(t *testing.T) {
	if _, err := ExplicitPath([]string{"--env-file=one", "--env-file", "two"}); err == nil {
		t.Fatal("duplicate --env-file was accepted")
	}
}

// TestExplicitPathPreservesPathWhitespace locks in the selector's established
// behavior; path expansion remains responsible for interpreting the raw value.
func TestExplicitPathPreservesPathWhitespace(t *testing.T) {
	path, err := ExplicitPath([]string{"--env-file", " path with spaces "})
	if err != nil {
		t.Fatal(err)
	}

	if path != " path with spaces " {
		t.Fatalf("path = %q", path)
	}
}
