package modruntime

import (
	"context"
	"encoding/binary"
	"image"
	"image/color"
	"testing"
	"testing/fstest"
	"time"

	cof "github.com/gravestench/cof"
	"github.com/gravestench/dark-magic/internal/host"
	"github.com/gravestench/dark-magic/internal/rendercore"
	dc6 "github.com/gravestench/dc6/pkg"
)

func TestNormalizedDC6FramesPreserveSharedAnchor(t *testing.T) {
	asset := &dc6.DC6{Directions: []*dc6.Direction{{Frames: []*dc6.Frame{
		{Width: 2, Height: 1, OffsetX: 5, OffsetY: 10, IndexData: []byte{1, 1}},
		{Width: 1, Height: 1, OffsetX: 3, OffsetY: 12, IndexData: []byte{1}},
	}}}}
	asset.SetPalette(color.Palette{color.RGBA{}, color.RGBA{R: 255, A: 255}})
	frames, bounds, err := normalizedDC6Frames(asset, 0)
	if err != nil {
		t.Fatal(err)
	}
	if bounds != image.Rect(3, -12, 7, -9) {
		t.Fatalf("normalized bounds = %v", bounds)
	}
	if got := color.RGBAModel.Convert(frames[0].At(2, 2)).(color.RGBA); got.R != 255 {
		t.Fatalf("first anchored pixel = %#v", got)
	}
	if got := color.RGBAModel.Convert(frames[1].At(0, 0)).(color.RGBA); got.R != 255 {
		t.Fatalf("second anchored pixel = %#v", got)
	}
}

func TestAssetWeightReadsAssetContents(t *testing.T) {
	t.Parallel()

	source := fstest.MapFS{
		"first.bin":  &fstest.MapFile{Data: []byte{1, 2, 3}},
		"second.bin": &fstest.MapFile{Data: []byte{4, 5}},
	}
	if got := assetWeight(source, "first.bin", "missing.bin", "second.bin"); got != 5 {
		t.Fatalf("asset weight = %d, want 5", got)
	}
}

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
			self.root:set_clip(10, 20, 300, 200)
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
	if nodes[0].Clip == nil || *nodes[0].Clip != (rendercore.Rect{X: 10, Y: 20, Width: 300, Height: 200}) {
		t.Fatalf("clip = %#v", nodes[0].Clip)
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

func TestRenderCapabilityExposesCOFCompositionMetadata(t *testing.T) {
	input := cof.New()
	input.NumberOfDirections = 1
	input.FramesPerDirection = 1
	input.NumberOfLayers = 1
	input.Speed = 128
	input.CofLayers = []cof.CofLayer{{Type: 0, Selectable: true, WeaponClass: cof.WeaponClassFromString("hth")}}
	input.AnimationFrames = []cof.FrameEvent{1}
	input.Priority = [][][]cof.CompositeType{{{0}}}
	assets := fstest.MapFS{"unit.cof": &fstest.MapFile{Data: cof.Marshal(input)}}
	var composer rendercore.Composer
	runtime := New()
	if err := runtime.RegisterModule(RenderModuleWithAssets(runtime, &composer, assets)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	script := `local r=require("dm.render/v1"); local c=assert(r.cof_info("unit.cof")); assert(c.directions==1 and c.frames==1 and c.layers[1].type=="HD" and c.priority[1][1][1]=="HD" and c.events[1]==1)`
	if err := runtime.Execute(context.Background(), fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(script)}}, "test.lua"); err != nil {
		t.Fatal(err)
	}
}

func TestCOFCompositionUsesFramePriorityAndPlacement(t *testing.T) {
	asset := cof.New()
	asset.NumberOfDirections, asset.FramesPerDirection, asset.NumberOfLayers = 1, 1, 2
	head, torso, none := cof.CompositeType(0), cof.CompositeType(1), cof.DrawEffect(8)
	asset.CofLayers = []cof.CofLayer{{Type: head, DrawEffect: none}, {Type: torso, DrawEffect: none}}
	asset.Priority = [][][]cof.CompositeType{{{torso, head}}}
	red := image.NewRGBA(image.Rect(0, 0, 2, 2))
	blue := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			red.Set(x, y, color.RGBA{R: 255, A: 255})
			blue.Set(x, y, color.RGBA{B: 255, A: 255})
		}
	}
	composed, err := composeCOFFrame(asset, 0, 0, map[cof.CompositeType]compositeFrame{
		torso: {image: red, bounds: image.Rect(-1, -1, 1, 1), layer: asset.CofLayers[1]},
		head:  {image: blue, bounds: image.Rect(0, 0, 2, 2), layer: asset.CofLayers[0]},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := color.RGBAModel.Convert(composed.At(1, 1)).(color.RGBA); got.B != 255 {
		t.Fatalf("priority pixel = %#v", got)
	}
	if composed.Bounds().Dx() != 5 || composed.Bounds().Dy() != 5 {
		t.Fatalf("bounds = %v", composed.Bounds())
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
	  self.animation = render.create("hud")
	  assert(self.animation:set_dc6_animation("one.dc6", "pal.dat", 0, 12, "loop") == 1)
	  self.animation:animation_pause()
	  self.animation:animation_seek(0.25)
	  self.animation:animation_resume()
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
	animationNode := nodes[1]
	animation, err := composer.ResourceSnapshot(animationNode.Resource)
	if err != nil || animation.Kind != rendercore.ResourceAnimation {
		t.Fatalf("animation = %#v, %v", animation, err)
	}
	if animationNode.AnimationPaused || animationNode.AnimationSeekRevision != 2 || animationNode.AnimationSeek != 250*time.Millisecond {
		t.Fatalf("animation playback state = %#v", animationNode)
	}
	if err := manager.Disable(context.Background(), definition.ID); err != nil {
		t.Fatal(err)
	}
	if nodes := composer.Snapshot(); len(nodes) != 0 {
		t.Fatalf("animation nodes leaked: %#v", nodes)
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
