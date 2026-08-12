package modruntime

import (
	"context"
	"testing"
	"testing/fstest"
)

func TestWorldgenModuleCanonicalizesOpaqueModRecipe(t *testing.T) {
	runtime := New()
	if err := runtime.RegisterModule(WorldgenModule()); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(t.Context())
	script := fstest.MapFS{"test.lua": {Data: []byte(`
local worldgen=require("engine.worldgen/v1")
local zone=assert(worldgen.admit({
  Request={version=1,seed=42,act=9,level_id=7,difficulty=99},
  Kind="another-mod", Bounds={Width=8,Height=8},
  Stamps={{ID=2,Width=8,Height=8,DS1Path="stamp",TilePaths={"z","a"}}},
  Rooms={{ID=1,Width=8,Height=8,StampID=2}},
  Structures={{X=1,Y=1,Kind="force-field",Passable=true}},
}))
assert(zone.Kind=="another-mod")
assert(zone.Stamps[1].TilePaths[1]=="a")
assert(#zone.checksum==64)
`)}}
	if err := runtime.Execute(context.Background(), script, "test.lua"); err != nil {
		t.Fatal(err)
	}
}
