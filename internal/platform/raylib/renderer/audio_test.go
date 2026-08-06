package raylibRenderer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStageMusicDataPreservesExtensionAndPayload(t *testing.T) {
	path, err := stageMusicData(".wav", []byte("music"))
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
	if filepath.Ext(path) != ".wav" {
		t.Fatalf("staged extension = %q", filepath.Ext(path))
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "music" {
		t.Fatalf("staged data = %q, error = %v", data, err)
	}
}
