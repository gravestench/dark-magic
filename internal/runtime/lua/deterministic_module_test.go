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
local digest=require("engine.deterministic/v1").digest
assert(digest("")=="e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
assert(digest("abc")=="ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad")
assert(#digest("a\0b")==64 and digest("a\0b")~=digest("ab"))
assert(digest(" abc ")~=digest("abc"))
local first=draw(42,"room",17,3)
assert(first==draw(42,"room",17,3))
assert(first>=0 and first<17)
assert(draw(42,"room",17,3)~=draw(42,"other",17,3) or draw(42,"room",997,3)~=draw(42,"other",997,3))
`)}}
	if err := runtime.Execute(context.Background(), script, "test.lua"); err != nil {
		t.Fatal(err)
	}
}
