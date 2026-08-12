package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckSnapshotRejectsExceededBudget(t *testing.T) {
	path := filepath.Join(t.TempDir(), "diagnostics.json")
	data := []byte(`{"composition":{"Decoded":{"Weight":12},"Retained":{"ActiveResources":3,"RetainedTextureBytes":40},"DecodeTime":2000000},"frame_timing":{"title":{"samples":2,"frame_p95":30000000,"update_p95":12000000}}}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	limits := budget{MaxDecodedWeight: 10, MaxActiveResources: 2, MaxRetainedTextureBytes: 30, MaxDecodeTimeMS: 1, MinFrameSamples: 3, MaxFrameP95MS: 20, MaxUpdateP95MS: 10}
	if err := checkSnapshot("title", path, limits); err == nil {
		t.Fatal("exceeded snapshot passed")
	}
}
