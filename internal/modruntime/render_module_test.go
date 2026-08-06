package modruntime

import (
	"context"
	"encoding/binary"
	"image"
	"image/color"
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
	resource, err := composer.ResourceSnapshot(nodes[0].Resource)
	if err != nil {
		t.Fatal(err)
	}
	decoded, ok := resource.Payload.(image.Image)
	if !ok || decoded.Bounds().Dx() != 16 || decoded.Bounds().Dy() != 8 {
		t.Fatalf("node resource = %#v", resource)
	}
	if err := manager.Disable(context.Background(), definition.ID); err != nil {
		t.Fatal(err)
	}
	if nodes := composer.Snapshot(); len(nodes) != 0 {
		t.Fatalf("nodes leaked after disable: %#v", nodes)
	}
}

func TestRenderNodeDecodesPaletteAwareDC6(t *testing.T) {
	t.Parallel()

	palette := make([]byte, 256*3)
	palette[3], palette[4], palette[5] = 10, 20, 30
	dc6Data := make([]byte, 16+8+4+32+3+3)
	put := func(offset int, value uint32) { binary.LittleEndian.PutUint32(dc6Data[offset:offset+4], value) }
	put(0, 6)
	put(16, 1)
	put(20, 1)
	put(24, 28)
	put(32, 1)
	put(36, 1)
	put(56, 3)
	dc6Data[60], dc6Data[61], dc6Data[62] = 1, 1, 0x80
	assets := fstest.MapFS{
		"one.dc6": &fstest.MapFile{Data: dc6Data},
		"pal.dat": &fstest.MapFile{Data: palette},
	}

	var composer rendercore.Composer
	runtime := New()
	if err := runtime.RegisterModule(RenderModuleWithAssets(runtime, &composer, assets)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	source := fstest.MapFS{"screen.lua": &fstest.MapFile{Data: []byte(`
local render = require("dm.render/v1")
return { id = "screen.dc6", start = function(self)
  self.root = render.create("hud")
  local w, h, ox, oy = self.root:set_dc6("one.dc6", "pal.dat", 0, 0)
  assert(w == 1 and h == 1 and ox == 0 and oy == 0)
end }
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
	resource, err := composer.ResourceSnapshot(nodes[0].Resource)
	if err != nil {
		t.Fatal(err)
	}
	decoded := resource.Payload.(image.Image)
	got := color.RGBAModel.Convert(decoded.At(0, 0)).(color.RGBA)
	if got != (color.RGBA{R: 10, G: 20, B: 30, A: 0xff}) {
		t.Fatalf("pixel = %#v", got)
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
