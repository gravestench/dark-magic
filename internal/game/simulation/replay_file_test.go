package simulation

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestReplayContainerFileRoundTripsWithPrivatePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.dmr")
	if err := WriteReplayContainerFile(path, replayContainerFixture()); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Fatalf("replay permissions = %o, want 600", info.Mode().Perm())
	}
	container, err := ReadReplayContainerFile(path, ReplayContainerLimits{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if container.Format != ReplayContainerFormat || len(container.Events) != 1 {
		t.Fatalf("replay container = %#v", container)
	}
}

func TestReplayContainerFileDoesNotReplaceExistingFileWhenEncodingFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.dmr")
	if err := os.WriteFile(path, []byte("previous"), 0o600); err != nil {
		t.Fatal(err)
	}
	invalid := replayContainerFixture()
	invalid.Version++
	if err := WriteReplayContainerFile(path, invalid); err == nil {
		t.Fatal("invalid replay write succeeded")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "previous" {
		t.Fatalf("existing replay changed to %q", data)
	}
}
