package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/host"
	"github.com/gravestench/dark-magic/internal/rendercore"
)

func TestRenderNodesBelongToLuaComponentScope(t *testing.T) {
	t.Parallel()

	var composer rendercore.Composer
	runtime := New()
	if err := runtime.RegisterModule(RenderModule(runtime, &composer)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())

	source := fstest.MapFS{"screen.lua": &fstest.MapFile{Data: []byte(`
local render = require("dm.render/v1")
return {
    id = "screen.loading",
    start = function(self)
        self.root = render.create("transition")
        self.root:set_position(320, 240)
		self.root:set_z(7)
		self.root:fill_rect(16, 8, 1, 2, 3, 4)
    end,
}
`)}}
	definition, err := LoadDefinition(context.Background(), runtime, source, "screen.lua")
	if err != nil {
		t.Fatal(err)
	}
	manager := host.NewManager()
	if err := manager.Register(definition.Managed()); err != nil {
		t.Fatal(err)
	}
	if err := manager.Enable(context.Background(), definition.ID); err != nil {
		t.Fatal(err)
	}
	nodes := composer.Snapshot()
	if len(nodes) != 1 || nodes[0].X != 320 || nodes[0].Y != 240 || nodes[0].Z != 7 || nodes[0].Layer != rendercore.LayerTransition {
		t.Fatalf("nodes = %#v", nodes)
	}
	if nodes[0].Image == nil || nodes[0].Image.Bounds().Dx() != 16 || nodes[0].Image.Bounds().Dy() != 8 {
		t.Fatalf("node image = %#v", nodes[0].Image)
	}
	if err := manager.Disable(context.Background(), definition.ID); err != nil {
		t.Fatal(err)
	}
	if nodes := composer.Snapshot(); len(nodes) != 0 {
		t.Fatalf("nodes leaked after disable: %#v", nodes)
	}
}

func TestRenderCapabilityRequiresComponentScope(t *testing.T) {
	t.Parallel()

	var composer rendercore.Composer
	runtime := New()
	if err := runtime.RegisterModule(RenderModule(runtime, &composer)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	err := runtime.Execute(context.Background(), fstest.MapFS{"bad.lua": &fstest.MapFile{Data: []byte(`local render = require("dm.render/v1"); render.create("world")`)}}, "bad.lua")
	if err == nil {
		t.Fatal("expected unscoped allocation to fail")
	}
}
