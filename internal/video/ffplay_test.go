package video

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"testing/fstest"
	"time"
)

// TestFFplayCompletesAndCleansTemporaryFile exercises natural child exit and the
// idempotent Stop wait used by callers during deferred cleanup.
func TestFFplayCompletesAndCleansTemporaryFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is POSIX-only")
	}

	directory := t.TempDir()

	player := filepath.Join(directory, "ffplay-test")
	if err := os.WriteFile(player, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	backend := FFplay{Executable: player}

	playback, err := backend.Play(fstest.MapFS{"movie.bik": {Data: []byte("payload")}}, "movie.bik")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		// Stop also waits for temporary-file removal, keeping failed assertions
		// from leaking either the child process or its extracted input.
		if err := playback.Stop(); err != nil {
			t.Errorf("stop playback: %v", err)
		}
	})

	// Polling observes the public contract while the deadline prevents a broken
	// fixture process from hanging the test suite.
	deadline := time.Now().Add(2 * time.Second)
	for playback.Snapshot().State == Playing && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}

	if got := playback.Snapshot(); got.State != Complete {
		t.Fatalf("snapshot = %#v", got)
	}

	if err := playback.Stop(); err != nil {
		t.Fatal(err)
	}
}

// TestFFplayUnavailable verifies executable lookup fails closed without loading media.
func TestFFplayUnavailable(t *testing.T) {
	if (FFplay{Executable: filepath.Join(t.TempDir(), "missing")}).Available() {
		t.Fatal("missing player reported available")
	}
}

// TestFFplayStopTerminatesPlayer verifies explicit cancellation wins the final
// state transition even though process reaping completes asynchronously.
func TestFFplayStopTerminatesPlayer(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is POSIX-only")
	}

	directory := t.TempDir()

	player := filepath.Join(directory, "ffplay-test")
	if err := os.WriteFile(player, []byte("#!/bin/sh\nsleep 30\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	backend := FFplay{Executable: player}

	playback, err := backend.Play(fstest.MapFS{"movie.bik": {Data: []byte("payload")}}, "movie.bik")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		// The sleeping fixture otherwise outlives a test that fails before its
		// explicit Stop assertion.
		if err := playback.Stop(); err != nil {
			t.Errorf("stop playback: %v", err)
		}
	})

	if err := playback.Stop(); err != nil {
		t.Fatal(err)
	}

	if got := playback.Snapshot(); got.State != Stopped {
		t.Fatalf("snapshot = %#v", got)
	}
}
