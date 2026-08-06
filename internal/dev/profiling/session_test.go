package profiling

import (
	"encoding/json"
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
	session.SetDiagnostics(func() any { return map[string]int{"resources": 3} })
	for index := 0; index < 100000; index++ {
		_ = index * index
	}
	if err := session.CaptureSceneHeap("title"); err != nil {
		t.Fatal(err)
	}
	if err := session.Stop(); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"cpu.pprof", "heap.pprof", "diagnostics.json", filepath.Join("scenes", "title", "heap-001.pprof"), filepath.Join("scenes", "title", "diagnostics-001.json")} {
		info, err := os.Stat(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() == 0 {
			t.Fatalf("%s is empty", name)
		}
	}
	data, err := os.ReadFile(filepath.Join(directory, "diagnostics.json"))
	if err != nil {
		t.Fatal(err)
	}
	var diagnostics map[string]int
	if err := json.Unmarshal(data, &diagnostics); err != nil || diagnostics["resources"] != 3 {
		t.Fatalf("diagnostics = %v, error = %v", diagnostics, err)
	}
}

func TestSafeName(t *testing.T) {
	if got := safeName("menu/../one"); got != "menu____one" {
		t.Fatalf("safe name = %q", got)
	}
}
