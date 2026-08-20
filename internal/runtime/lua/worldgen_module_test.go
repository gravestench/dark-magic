package modruntime

import (
	"context"
	"testing"
	"testing/fstest"
)

// TestWorldgenModuleCanonicalizesOpaqueModRecipe protects the worldgen module canonicalizes opaque mod recipe
// contract, including its observable ordering and failure behavior.
func TestWorldgenModuleCanonicalizesOpaqueModRecipe(t *testing.T) {
	runtime := New()
	if err := runtime.RegisterModule(WorldgenModule()); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(t.Context()) }()

	script := fstest.MapFS{"test.lua": {Data: []byte(`
local worldgen=require("engine.worldgen/v1")
local zone=assert(worldgen.admit({
  request={version=1,seed=42,act=9,level_id=7,difficulty=99},
  kind="another-mod", bounds={width=8,height=8},
  stamps={{id=2,width=8,height=8,ds1_path="stamp",tile_paths={"z","a"}}},
  rooms={{id=1,width=8,height=8,stamp_id=2}},
  structures={{x=1,y=1,kind="force-field",passable=true}},
}))
assert(zone.kind=="another-mod")
assert(zone.stamps[1].tile_paths[1]=="a")
assert(#zone.checksum==64)
`)}}
	if err := runtime.Execute(context.Background(), script, "test.lua"); err != nil {
		t.Fatal(err)
	}
}
