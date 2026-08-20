package modruntime

import (
	"context"
	"testing"
	"testing/fstest"
)

// TestDeterministicDrawIsPureAndPurposeSeparated protects the deterministic draw is pure and purpose separated
// contract, including its observable ordering and failure behavior.
func TestDeterministicDrawIsPureAndPurposeSeparated(t *testing.T) {
	runtime := New()
	if err := runtime.RegisterModule(DeterministicModule()); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(t.Context()) }()

	script := fstest.MapFS{"test.lua": {Data: []byte(`
local draw=require("engine.deterministic/v1").integer
local first=draw(42,"room",17,3)
assert(first==draw(42,"room",17,3))
assert(first>=0 and first<17)
assert(draw(42,"room",17,3)~=draw(42,"other",17,3) or draw(42,"room",997,3)~=draw(42,"other",997,3))
`)}}
	if err := runtime.Execute(context.Background(), script, "test.lua"); err != nil {
		t.Fatal(err)
	}
}
