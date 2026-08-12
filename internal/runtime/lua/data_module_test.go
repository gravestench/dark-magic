package modruntime

import (
	"context"
	"testing"
	"testing/fstest"
)

func TestDataModuleLoadsJSONAsLuaValues(t *testing.T) {
	runtime := New()
	source := fstest.MapFS{
		"data.json":     &fstest.MapFile{Data: []byte(`{"version":1,"enabled":true,"names":["one","two"],"nothing":null}`)},
		"manifest.json": &fstest.MapFile{Data: []byte(`{"schema":"test/v1","version":1,"game_version":"test","language":"neutral","confidence":"verified","resolution":{"width":800,"height":600}}`)},
	}
	if err := runtime.RegisterModule(DataModule(source)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	if err := runtime.Execute(context.Background(), fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(`
local data = require("engine.data/v1")
local document = assert(data.load("data.json"))
assert(document.version == 1 and document.enabled and document.names[2] == "two" and document.nothing == nil)
local manifest = assert(data.load_manifest("manifest.json", "test/v1"))
assert(manifest.resolution.width == 800)
local incompatible, schema_err = data.load_manifest("manifest.json", "test/v2")
assert(incompatible == nil and string.find(schema_err, "want"))
local value, err = data.load("manifest.lua")
assert(value == nil and string.find(err, "only JSON"))
`)}}, "test.lua"); err != nil {
		t.Fatal(err)
	}
}

func TestManifestRejectsInvalidSupportedProfile(t *testing.T) {
	decoded, err := readDataJSON(fstest.MapFS{
		"manifest.json": &fstest.MapFile{Data: []byte(`{
            "schema":"test/v1", "version":1, "game_version":"test",
            "language":"neutral", "confidence":"verified",
            "resolution":{"width":800,"height":600},
            "supported_profiles":[{"id":"broken","game_version":"test","language":"English","resolution":{"width":0,"height":600}}]
        }`)},
	}, "manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := validateManifest(decoded, "test/v1"); err == nil {
		t.Fatal("invalid supported profile was accepted")
	}
}

func TestDataModuleAppliesSelectedPresentationProfile(t *testing.T) {
	runtime := New()
	source := fstest.MapFS{"manifest.json": &fstest.MapFile{Data: []byte(`{
        "schema":"d2legacy.presentation/v1","version":1,"game_version":"test","language":"neutral","confidence":"verified",
        "resolution":{"width":800,"height":600},"screens":{"world":{"hud":{"sheet":"800.dc6","x":400}}},
        "supported_profiles":[
          {"id":"wide","game_version":"test","language":"English","resolution":{"width":800,"height":600}},
          {"id":"classic","game_version":"test","language":"English","resolution":{"width":640,"height":480},
           "overrides":{"screens":{"world":{"hud":{"sheet":"640.dc6","x":320}}}}}
        ]}`)}}
	if err := runtime.RegisterModule(DataModule(source, "classic")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	if err := runtime.Execute(context.Background(), fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(`
local manifest=assert(require("engine.data/v1").load_manifest("manifest.json","d2legacy.presentation/v1"))
assert(manifest.active_profile=="classic")
assert(manifest.resolution.width==640 and manifest.resolution.height==480)
assert(manifest.screens.world.hud.sheet=="640.dc6" and manifest.screens.world.hud.x==320)
`)}}, "test.lua"); err != nil {
		t.Fatal(err)
	}
}
