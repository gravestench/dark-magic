package main

import (
	"fmt"
	"testing"
)

// TestOpenMPQStackRejectsDirectoryWithoutSupportedArchives preserves the distinction between an empty installation
// and a usable partial installation, including the user-facing directory diagnostic.
func TestOpenMPQStackRejectsDirectoryWithoutSupportedArchives(t *testing.T) {
	directory := t.TempDir()

	_, err := openMPQStack(directory)

	want := fmt.Sprintf("no supported MPQs found in %q", directory)
	if err == nil || err.Error() != want {
		t.Fatalf("openMPQStack() error = %v, want %q", err, want)
	}
}
