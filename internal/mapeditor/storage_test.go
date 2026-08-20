package mapeditor

import (
	"os"
	"path/filepath"
	"testing"
)

// TestStorageSavesOnlyInsideConfiguredRoot covers atomic in-project writes and traversal rejection.
func TestStorageSavesOnlyInsideConfiguredRoot(t *testing.T) {
	storage, err := NewStorage(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	path, err := storage.Save("data/global/tiles/test.ds1", []byte("map"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "map" {
		t.Fatalf("saved = %q, %v", data, err)
	}
	if _, err := storage.Save(filepath.Join("..", "outside.ds1"), []byte("nope")); err == nil {
		t.Fatal("path escape succeeded")
	}
	if _, err := storage.Save("wrong.txt", []byte("nope")); err == nil {
		t.Fatal("non-DS1 save succeeded")
	}
}

// TestStorageRejectsMountedGameDataDestination prevents an editor save from replacing original assets.
func TestStorageRejectsMountedGameDataDestination(t *testing.T) {
	gameRoot := t.TempDir()
	storage, err := NewStorage(t.TempDir(), gameRoot)
	if err != nil {
		t.Fatal(err)
	}
	storage.root = filepath.Dir(gameRoot)
	if _, err := storage.Save(filepath.Base(gameRoot)+"/data/global/tiles/map.ds1", []byte("ds1")); err == nil {
		t.Fatal("Save allowed a destination below mounted source data")
	}
}

// TestStorageRejectsMountedGameDataAsOutputRoot catches unsafe configuration before any edit begins.
func TestStorageRejectsMountedGameDataAsOutputRoot(t *testing.T) {
	gameRoot := t.TempDir()
	if _, err := NewStorage(gameRoot, gameRoot); err == nil {
		t.Fatal("NewStorage accepted a mounted source as output")
	}
}
