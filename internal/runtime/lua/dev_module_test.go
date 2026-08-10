package modruntime

import (
	"context"
	"testing"
	"testing/fstest"
)

func TestDevModuleCopiesLaunchOptions(t *testing.T) {
	runtime := New()
	if err := runtime.RegisterModule(DevModule(map[string]any{"name": "AM", "direction": 3, "random": true})); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	script := `local d=require("dm.dev/v1"); assert(d.option("name")=="AM"); assert(d.option("direction")==3); assert(d.option("random")==true); assert(d.option("missing")==nil)`
	if err := runtime.Execute(context.Background(), fstest.MapFS{"test.lua": {Data: []byte(script)}}, "test.lua"); err != nil {
		t.Fatal(err)
	}
}
