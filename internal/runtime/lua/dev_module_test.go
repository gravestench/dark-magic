package modruntime

import (
	"context"
	"testing"
	"testing/fstest"
)

// TestDevModuleCopiesLaunchOptions verifies typed option reads and fresh nonzero seeds without exposing the backing
// launch-options map to Lua.
func TestDevModuleCopiesLaunchOptions(t *testing.T) {
	runtime := New()
	if err := runtime.RegisterModule(
		DevModule(map[string]any{"name": "AM", "direction": 3, "random": true, "random_seed": 7}),
	); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(context.Background()) }()

	script := `local d=require("engine.dev/v1"); assert(d.option("name")=="AM"); ` +
		`assert(d.option("direction")==3); assert(d.option("random")==true); assert(d.option("missing")==nil); ` +
		`local first=d.seed(); local second=d.seed(); assert(first>0 and second>0 and first~=second)`
	if err := runtime.Execute(
		context.Background(),
		fstest.MapFS{"test.lua": {Data: []byte(script)}},
		"test.lua",
	); err != nil {
		t.Fatal(err)
	}
}
