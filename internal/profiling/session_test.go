package profiling

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSessionWritesRawCPUAndHeapProfiles(t *testing.T) {
	directory := t.TempDir()
	session, err := Start(directory, false)
	if err != nil {
		t.Fatal(err)
	}
	session.ConfigureScenes("title")
	for index := 0; index < 100000; index++ {
		_ = index * index
	}
	if err := session.CaptureSceneHeap("title"); err != nil {
		t.Fatal(err)
	}
	if err := session.Stop(); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"cpu.pprof", "heap.pprof", filepath.Join("scenes", "title", "heap-001.pprof")} {
		info, err := os.Stat(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() == 0 {
			t.Fatalf("%s is empty", name)
		}
	}
}

func TestSafeName(t *testing.T) {
	if got := safeName("menu/../one"); got != "menu____one" {
		t.Fatalf("safe name = %q", got)
	}
}
