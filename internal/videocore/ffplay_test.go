package videocore

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"testing/fstest"
	"time"
)

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

func TestFFplayUnavailable(t *testing.T) {
	if (FFplay{Executable: filepath.Join(t.TempDir(), "missing")}).Available() {
		t.Fatal("missing player reported available")
	}
}

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
	if err := playback.Stop(); err != nil {
		t.Fatal(err)
	}
	if got := playback.Snapshot(); got.State != Stopped {
		t.Fatalf("snapshot = %#v", got)
	}
}
