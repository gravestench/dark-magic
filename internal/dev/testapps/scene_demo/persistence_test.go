package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gravestench/dark-magic/internal/presentation/scene"
)

// TestSaveAndLoadSceneRoundTrip protects the demo's persisted scene format and loaded-state reconstruction together.
func TestSaveAndLoadSceneRoundTrip(t *testing.T) {
	want := scene.New(42, 800, 600)
	want.MoveHero(-125, 75)

	savePath := filepath.Join(t.TempDir(), "scene.json")

	if err := saveScene(want, savePath); err != nil {
		t.Fatalf("save scene: %v", err)
	}

	got, err := loadScene(savePath)
	if err != nil {
		t.Fatalf("load scene: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("loaded scene = %#v, want %#v", got, want)
	}
}

// TestLoadSceneReturnsOpenFailure ensures missing saves remain distinguishable from decode or validation failures.
func TestLoadSceneReturnsOpenFailure(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing.json")

	state, err := loadScene(missingPath)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("load missing scene error = %v, want os.ErrNotExist", err)
	}

	if state != nil {
		t.Fatalf("load missing scene state = %#v, want nil", state)
	}
}

// TestSaveSceneReturnsCreateFailure ensures an invalid destination fails before the codec can report a misleading
// error.
func TestSaveSceneReturnsCreateFailure(t *testing.T) {
	state := scene.New(1, 800, 600)

	if err := saveScene(state, t.TempDir()); err == nil {
		t.Fatal("save scene to directory succeeded, want create error")
	}
}
