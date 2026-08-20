package modruntime

import (
	"context"
	"testing"
	"testing/fstest"
)

// TestDiscoverDefinitionsRegistersWithoutStarting protects the discover definitions registers without starting
// contract, including its observable ordering and failure behavior.
func TestDiscoverDefinitionsRegistersWithoutStarting(t *testing.T) {
	runtime := New()
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(context.Background()) }()

	source := fstest.MapFS{
		"boot.lua":             &fstest.MapFile{Data: []byte(`return { id = "boot" }`)},
		"components/audio.lua": &fstest.MapFile{Data: []byte(`return { id = "audio" }`)},
		"lua/helper.lua":       &fstest.MapFile{Data: []byte(`error("not a component")`)},
	}

	definitions, err := DiscoverDefinitions(context.Background(), runtime, source)
	if err != nil {
		t.Fatal(err)
	}

	if len(definitions) != 2 || definitions[0].ID != "boot" || definitions[1].ID != "audio" {
		t.Fatalf("definitions = %#v", definitions)
	}
}

// TestDiscoverDefinitionsRejectsDuplicateIDs protects the discover definitions rejects duplicate ids contract,
// including its observable ordering and failure behavior.
func TestDiscoverDefinitionsRejectsDuplicateIDs(t *testing.T) {
	runtime := New()
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(context.Background()) }()

	_, err := DiscoverDefinitions(context.Background(), runtime, fstest.MapFS{
		"boot.lua":         &fstest.MapFile{Data: []byte(`return { id = "same" }`)},
		"components/x.lua": &fstest.MapFile{Data: []byte(`return { id = "same" }`)},
	})
	if err == nil {
		t.Fatal("expected duplicate ID error")
	}
}
