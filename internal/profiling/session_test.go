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
	for index := 0; index < 100000; index++ {
		_ = index * index
	}
	if err := session.Stop(); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"cpu.pprof", "heap.pprof"} {
		info, err := os.Stat(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() == 0 {
			t.Fatalf("%s is empty", name)
		}
	}
}
