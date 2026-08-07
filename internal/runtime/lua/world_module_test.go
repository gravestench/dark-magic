package modruntime

import (
	"context"
	"testing"
	"testing/fstest"
)

func TestWorldModuleReportsAssetErrorsToLua(t *testing.T) {
	runtime := New()
	if err := runtime.RegisterModule(WorldModule(fstest.MapFS{})); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())

	script := fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(`
local world = require("dm.world/v1")
local decoded, err = world.load("missing.ds1", {"missing.dt1"})
assert(decoded == nil)
assert(string.find(err, "missing.ds1"))
`)}}
	if err := runtime.Execute(context.Background(), script, "test.lua"); err != nil {
		t.Fatal(err)
	}
}

func TestWorldModuleRejectsNonStringTilePaths(t *testing.T) {
	runtime := New()
	if err := runtime.RegisterModule(WorldModule(fstest.MapFS{})); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())

	script := fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(`
local world = require("dm.world/v1")
local ok, err = pcall(world.load, "missing.ds1", {42})
assert(not ok and string.find(err, "sequence of strings"))
`)}}
	if err := runtime.Execute(context.Background(), script, "test.lua"); err != nil {
		t.Fatal(err)
	}
}
