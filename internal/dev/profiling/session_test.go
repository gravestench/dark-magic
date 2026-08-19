package profiling

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestSessionWritesRawCPUAndHeapProfiles verifies that a non-rendering session still emits every raw artifact and the
// diagnostics snapshots associated with both the scene capture and final shutdown.
func TestSessionWritesRawCPUAndHeapProfiles(t *testing.T) {
	directory := t.TempDir()
	session := startRawTestSession(t, directory)

	session.ConfigureScenes("title")
	session.SetDiagnostics(testResourceDiagnostics)

	for index := 0; index < 100000; index++ {
		_ = index * index
	}

	if err := session.CaptureSceneHeap("title"); err != nil {
		t.Fatal(err)
	}

	if err := session.Stop(); err != nil {
		t.Fatal(err)
	}

	assertNonemptyProfileArtifacts(t, directory)
	assertResourceDiagnostics(t, filepath.Join(directory, "diagnostics.json"))
	assertResourceDiagnostics(t, filepath.Join(directory, "scenes", "title", "diagnostics-001.json"))
}

// startRawTestSession starts the process-global profiler and registers cleanup so an earlier test failure cannot leave
// profiling active for later tests in the same process.
func startRawTestSession(t *testing.T, directory string) *Session {
	t.Helper()

	session, err := Start(directory, false)
	if err != nil {
		t.Fatal(err)
	}

	// Explicit Stop in the test remains the asserted path; cleanup owns only early-exit and failure cases.
	t.Cleanup(func() {
		if err := session.Stop(); err != nil {
			t.Errorf("clean up profiling session: %v", err)
		}
	})

	return session
}

// testResourceDiagnostics returns a stable fixture whose JSON value can be checked after both snapshot phases.
func testResourceDiagnostics() any {
	return map[string]int{"resources": 3}
}

// assertNonemptyProfileArtifacts checks the original artifact set and retains size assertions so zero-byte profiles
// cannot pass merely because their paths exist.
func assertNonemptyProfileArtifacts(t *testing.T, directory string) {
	t.Helper()

	names := []string{
		"cpu.pprof",
		"heap.pprof",
		"diagnostics.json",
		filepath.Join("scenes", "title", "heap-001.pprof"),
		filepath.Join("scenes", "title", "diagnostics-001.json"),
	}
	for _, name := range names {
		info, err := os.Stat(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}

		if info.Size() == 0 {
			t.Fatalf("%s is empty", name)
		}
	}
}

// assertResourceDiagnostics decodes the artifact instead of matching its formatting, preserving the behavioral check
// that the configured provider's value survives serialization.
func assertResourceDiagnostics(t *testing.T, path string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var diagnostics map[string]int
	if err := json.Unmarshal(data, &diagnostics); err != nil || diagnostics["resources"] != 3 {
		t.Fatalf("diagnostics = %v, error = %v", diagnostics, err)
	}
}

// TestSafeName verifies both traversal sanitization and the exact portable alphabet retained in artifact paths.
func TestSafeName(t *testing.T) {
	testCases := []struct {
		name string
		want string
	}{
		{name: "menu/../one", want: "menu____one"},
		{name: "Title-screen_2", want: "Title-screen_2"},
		{name: "café", want: "caf_"},
	}

	for _, testCase := range testCases {
		if got := safeName(testCase.name); got != testCase.want {
			t.Fatalf("safeName(%q) = %q, want %q", testCase.name, got, testCase.want)
		}
	}
}
