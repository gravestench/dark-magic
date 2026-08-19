package envconfig

import "testing"

// TestExplicitPathRejectsAmbiguousFlags prevents two file selectors from making
// bootstrap precedence depend on argument order.
func TestExplicitPathRejectsAmbiguousFlags(t *testing.T) {
	if _, err := ExplicitPath([]string{"--env-file=one", "--env-file", "two"}); err == nil {
		t.Fatal("duplicate --env-file was accepted")
	}
}

// TestExplicitPathPreservesPathWhitespace keeps flag extraction lexical; host
// path expansion remains the single layer responsible for interpreting path text.
func TestExplicitPathPreservesPathWhitespace(t *testing.T) {
	path, err := ExplicitPath([]string{"--env-file", " path with spaces "})
	if err != nil {
		t.Fatal(err)
	}

	if path != " path with spaces " {
		t.Fatalf("path = %q", path)
	}
}
