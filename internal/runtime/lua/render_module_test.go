package modruntime

import (
	"bytes"
	"context"
	"encoding/binary"
	"image"
	"image/color"
	"testing"
	"testing/fstest"
	"time"

	cof "github.com/gravestench/cof"
	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	dc6 "github.com/gravestench/dc6/pkg"
)

func TestDCCGroundOriginPreservesAuthoredCanvasAnchor(t *testing.T) {
	x, y := dccGroundOrigin(image.Rect(-12, -48, 20, 16))
	if x != 0.375 || y != 0.75 {
		t.Fatalf("ground origin = (%v, %v), want (0.375, 0.75)", x, y)
	}
}

func TestNormalizedDC6FramesPreserveSharedAnchor(t *testing.T) {
	asset := &dc6.DC6{Directions: []*dc6.Direction{{Frames: []*dc6.Frame{
		{Width: 2, Height: 1, OffsetX: 5, OffsetY: 10, IndexData: []byte{1, 1}},
		{Width: 1, Height: 1, OffsetX: 3, OffsetY: 12, IndexData: []byte{1}},
	}}}}
	asset.SetPalette(color.Palette{color.RGBA{}, color.RGBA{R: 255, A: 255}})
	frames, bounds, err := normalizedDC6Frames(asset, 0, "offsets")
	if err != nil {
		t.Fatal(err)
	}
	if bounds != image.Rect(3, 10, 7, 13) {
		t.Fatalf("normalized bounds = %v", bounds)
	}
	if got := color.RGBAModel.Convert(frames[0].At(2, 0)).(color.RGBA); got.R != 255 {
		t.Fatalf("first anchored pixel = %#v", got)
	}
	if got := color.RGBAModel.Convert(frames[1].At(0, 2)).(color.RGBA); got.R != 255 {
		t.Fatalf("second anchored pixel = %#v", got)
	}
	fixed, fixedBounds, err := normalizedDC6Frames(asset, 0, "first-frame")
	if err != nil {
		t.Fatal(err)
	}
	if fixedBounds != image.Rect(5, 10, 7, 11) {
		t.Fatalf("fixed bounds = %v", fixedBounds)
	}
	if got := color.RGBAModel.Convert(fixed[1].At(0, 0)).(color.RGBA); got.R != 255 {
		t.Fatalf("fixed second pixel = %#v", got)
	}
	shared := image.Rect(1, 8, 8, 14)
	sharedFixed, sharedFixedBounds, err := normalizedDC6Frames(asset, 0, "first-frame", shared)
	if err != nil {
		t.Fatal(err)
	}
	if sharedFixedBounds != shared {
		t.Fatalf("shared fixed bounds = %v", sharedFixedBounds)
	}
	if got := color.RGBAModel.Convert(sharedFixed[1].At(4, 2)).(color.RGBA); got.R != 255 {
		t.Fatalf("shared fixed second pixel = %#v", got)
	}
}

func TestDCCDirectionForCOFUsesIndependentLegacyOrder(t *testing.T) {
	want16 := []int{4, 8, 0, 9, 5, 10, 1, 11, 6, 12, 2, 13, 7, 14, 3, 15}
	for cofDirection, want := range want16 {
		got, err := dccDirectionForCOF(cofDirection, 16)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("16-way COF direction %d maps to DCC %d, want %d", cofDirection, got, want)
		}
	}
	want8 := []int{4, 0, 5, 1, 6, 2, 7, 3}
	for cofDirection, want := range want8 {
		got, err := dccDirectionForCOF(cofDirection, 8)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("8-way COF direction %d maps to DCC %d, want %d", cofDirection, got, want)
		}
	}
}

func TestCombinedDC6PagesJoinsFrameTiles(t *testing.T) {
	asset := &dc6.DC6{Directions: []*dc6.Direction{{Frames: []*dc6.Frame{
		{Width: 256, Height: 1, OffsetX: 0, OffsetY: 0, IndexData: bytes.Repeat([]byte{1}, 256)},
		{Width: 1, Height: 1, OffsetX: 0, OffsetY: 0, IndexData: []byte{2}},
	}}}}
	asset.SetPalette(color.Palette{
		color.RGBA{},
		color.RGBA{R: 255, A: 255},
		color.RGBA{G: 255, A: 255},
	})

	pages, err := combinedDC6Pages(asset, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 1 {
		t.Fatalf("combined pages = %d, want 1", len(pages))
	}
	composite := pages[0]
	if composite.Bounds() != image.Rect(0, 0, 257, 1) {
		t.Fatalf("composite bounds = %v", composite.Bounds())
	}
	if got := color.RGBAModel.Convert(composite.At(255, 0)).(color.RGBA); got.R != 255 {
		t.Fatalf("first tile pixel = %#v", got)
	}
	if got := color.RGBAModel.Convert(composite.At(256, 0)).(color.RGBA); got.G != 255 {
		t.Fatalf("second tile pixel = %#v", got)
	}
}

func TestHorizontalDC6StripPreservesExplicitRightCap(t *testing.T) {
	asset := &dc6.DC6{Directions: []*dc6.Direction{{Frames: []*dc6.Frame{
		{Width: 2, Height: 1, IndexData: []byte{1, 1}},
		{Width: 1, Height: 1, IndexData: []byte{2}},
	}}}}
	asset.SetPalette(color.Palette{
		color.RGBA{},
		color.RGBA{R: 255, A: 255},
		color.RGBA{G: 255, A: 255},
	})
	strip, err := horizontalDC6Strip(asset, 0)
	if err != nil {
		t.Fatal(err)
	}
	if strip.Bounds() != image.Rect(0, 0, 3, 1) {
		t.Fatalf("strip bounds = %v", strip.Bounds())
	}
	if got := color.RGBAModel.Convert(strip.At(2, 0)).(color.RGBA); got.G != 255 {
		t.Fatalf("right cap pixel = %#v", got)
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

	var composer render.Composer
	runtime := New()
	if err := runtime.RegisterModule(RenderModule(runtime, &composer)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())

	source := fstest.MapFS{"screen.lua": &fstest.MapFile{Data: []byte(`
local render = require("engine.render/v1")
return {
    id = "screen.loading",
    start = function(self)
        self.root = render.create("transition")
		self.root:set_position(320, 240)
			self.root:set_z(7)
			self.root:set_tint(96, 112, 128)
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
	if len(nodes) != 1 || nodes[0].X != 320 || nodes[0].Y != 240 || nodes[0].Z != 7 || nodes[0].Layer != render.LayerTransition {
		t.Fatalf("nodes = %#v", nodes)
	}
	if nodes[0].Clip == nil || *nodes[0].Clip != (render.Rect{X: 10, Y: 20, Width: 300, Height: 200}) {
		t.Fatalf("clip = %#v", nodes[0].Clip)
	}
	if nodes[0].Tint != (color.RGBA{R: 96, G: 112, B: 128, A: 255}) {
		t.Fatalf("tint = %#v", nodes[0].Tint)
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

func TestRenderNodePaletteQuantizationAcceptsTinyPalettes(t *testing.T) {
	t.Parallel()

	var composer render.Composer
	runtime := New()
	assets := fstest.MapFS{
		"screen.lua": {Data: []byte(`
local render = require("engine.render/v1")
return {
    id = "screen.palette",
    start = function(self)
        self.root = render.create("hud")
        self.root:fill_rect(4, 4, 255, 0, 0, 255)
        self.root:set_palette_quantization("black-white.pal")
    end,
}
`)},
		"black-white.pal": {Data: []byte{0, 0, 0, 255, 255, 255}},
	}
	if err := runtime.RegisterModule(RenderModuleWithAssets(runtime, &composer, assets)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())

	definition, err := LoadDefinition(context.Background(), runtime, assets, "screen.lua")
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
	if len(nodes) != 1 || nodes[0].Palette == (render.ResourceID{}) {
		t.Fatalf("palette node = %#v", nodes)
	}
	resource, err := composer.ResourceSnapshot(nodes[0].Palette)
	if err != nil {
		t.Fatal(err)
	}
	palette, ok := resource.Payload.(color.Palette)
	if !ok || len(palette) != 2 {
		t.Fatalf("palette resource = %#v", resource)
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
	var composer render.Composer
	runtime := New()
	if err := runtime.RegisterModule(RenderModuleWithAssets(runtime, &composer, assets)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	script := `local r=require("engine.render/v1"); local c=assert(r.cof_info("unit.cof")); assert(c.directions==1 and c.frames==1 and c.layers[1].type=="HD" and c.priority[1][1][1]=="HD" and c.events[1]==1)`
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
	if composed.Bounds().Dx() != 3 || composed.Bounds().Dy() != 3 {
		t.Fatalf("bounds = %v", composed.Bounds())
	}
}

func TestCOFCompositionDrawsOnlyShadowEnabledLayers(t *testing.T) {
	asset := cof.New()
	asset.NumberOfDirections, asset.FramesPerDirection, asset.NumberOfLayers = 1, 1, 1
	head := cof.CompositeType(0)
	asset.CofLayers = []cof.CofLayer{{Type: head, Shadow: 1, DrawEffect: cof.DrawEffect(8)}}
	asset.Priority = [][][]cof.CompositeType{{{head}}}
	source := image.NewRGBA(image.Rect(0, 0, 1, 1))
	source.Set(0, 0, color.RGBA{R: 255, A: 255})
	composed, err := composeCOFFrame(asset, 0, 0, map[cof.CompositeType]compositeFrame{
		head: {image: source, bounds: source.Bounds(), layer: asset.CofLayers[0]},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, alpha := composed.At(0, 0).RGBA(); alpha == 0 {
		t.Fatal("shadow-enabled COF layer produced no projected shadow pixel")
	}
}

func TestCompositeShadowUsesLegacyHalfHeightShear(t *testing.T) {
	componentType := cof.CompositeType(0)
	asset := cof.New()
	asset.NumberOfDirections, asset.FramesPerDirection, asset.NumberOfLayers = 1, 1, 1
	asset.CofLayers = []cof.CofLayer{{Type: componentType, Shadow: 1}}
	asset.Priority = [][][]cof.CompositeType{{{componentType}}}
	source := image.NewRGBA(image.Rect(10, 20, 12, 24))
	source.Set(10, 20, color.RGBA{R: 255, A: 255}) // top-left
	source.Set(10, 23, color.RGBA{R: 255, A: 255}) // baseline-left
	component := compositeFrame{image: source, bounds: source.Bounds(), layer: asset.CofLayers[0]}
	canvas := shadowCanvasBounds(source.Bounds(), map[cof.CompositeType]compositeFrame{componentType: component})
	if canvas.Min.X != 8 || canvas.Max.X != 14 || canvas.Min.Y != 20 || canvas.Max.Y != 24 {
		t.Fatalf("center-preserving shadow canvas = %v", canvas)
	}
	composed, err := composeCOFFrame(asset, 0, 0, map[cof.CompositeType]compositeFrame{componentType: component})
	if err != nil {
		t.Fatal(err)
	}
	// Top-left shears two pixels left and projects two pixels above baseline;
	// baseline-left remains anchored. Coordinates are translated by canvas.Min.
	for _, point := range []image.Point{{0, 1}, {2, 3}} {
		if _, _, _, alpha := composed.At(point.X, point.Y).RGBA(); alpha == 0 {
			t.Fatalf("projected shadow pixel %v is transparent", point)
		}
	}
}

func TestCompositeShadowProjectsPartsFromOneSharedBaseline(t *testing.T) {
	headType, torsoType := cof.CompositeType(0), cof.CompositeType(1)
	layers := []cof.CofLayer{{Type: headType, Shadow: 1}, {Type: torsoType, Shadow: 1}}
	head := image.NewRGBA(image.Rect(0, 0, 1, 2))
	torso := image.NewRGBA(image.Rect(0, 0, 1, 2))
	head.SetRGBA(0, 0, color.RGBA{A: 255})
	torso.SetRGBA(0, 1, color.RGBA{A: 255})
	bounds := image.Rect(0, 0, 1, 4)
	components := map[cof.CompositeType]compositeFrame{
		headType:  {image: head, bounds: head.Bounds(), layer: layers[0]},
		torsoType: {image: torso, bounds: image.Rect(0, 2, 1, 4), layer: layers[1]},
	}
	mask := compositeShadowMask(bounds, []cof.CompositeType{headType, torsoType}, components)
	canvas := shadowCanvasBounds(bounds, components)
	output := image.NewRGBA(image.Rect(0, 0, canvas.Dx(), canvas.Dy()))
	drawCompositeShadow(output, mask, bounds, canvas, 255)
	// The head's top and torso's foot use the same y=3 ground baseline. If each
	// part projected from its own bottom, the head pixel would land a row higher.
	for _, absolute := range []image.Point{{-2, 1}, {0, 3}} {
		point := absolute.Sub(canvas.Min)
		if output.RGBAAt(point.X, point.Y).A == 0 {
			t.Fatalf("shared-baseline shadow pixel %v is transparent", absolute)
		}
	}
}

func TestCOFCompositionDrawsAllShadowsBehindAllVisibleLayers(t *testing.T) {
	asset := cof.New()
	asset.NumberOfDirections, asset.FramesPerDirection, asset.NumberOfLayers = 1, 1, 2
	back, front := cof.CompositeType(3), cof.CompositeType(4)
	asset.CofLayers = []cof.CofLayer{
		{Type: back, DrawEffect: cof.DrawEffect(8)},
		{Type: front, Shadow: 1, DrawEffect: cof.DrawEffect(8)},
	}
	asset.Priority = [][][]cof.CompositeType{{{back, front}}}
	red := image.NewRGBA(image.Rect(0, 0, 3, 3))
	red.Set(2, 2, color.RGBA{R: 255, A: 255})
	shadowSource := image.NewRGBA(image.Rect(0, 0, 1, 1))
	shadowSource.Set(0, 0, color.RGBA{B: 255, A: 255})
	composed, err := composeCOFFrame(asset, 0, 0, map[cof.CompositeType]compositeFrame{
		back:  {image: red, bounds: red.Bounds(), layer: asset.CofLayers[0]},
		front: {image: shadowSource, bounds: shadowSource.Bounds(), layer: asset.CofLayers[1]},
	})
	if err != nil {
		t.Fatal(err)
	}
	canvas := shadowCanvasBounds(red.Bounds(), map[cof.CompositeType]compositeFrame{
		back:  {image: red, bounds: red.Bounds(), layer: asset.CofLayers[0]},
		front: {image: shadowSource, bounds: shadowSource.Bounds(), layer: asset.CofLayers[1]},
	})
	point := image.Pt(2, 2).Sub(canvas.Min)
	if got := color.RGBAModel.Convert(composed.At(point.X, point.Y)).(color.RGBA); got.R != 255 || got.A != 255 {
		t.Fatalf("visible back layer was covered by a later component shadow: %#v", got)
	}
}

func TestCOFCompositionIgnoresDrawEffectOnOpaqueLayer(t *testing.T) {
	asset := cof.New()
	asset.NumberOfDirections, asset.FramesPerDirection, asset.NumberOfLayers = 1, 1, 1
	head := cof.CompositeType(0)
	asset.CofLayers = []cof.CofLayer{{Type: head, Transparent: false, DrawEffect: cof.DrawEffect(0)}}
	asset.Priority = [][][]cof.CompositeType{{{head}}}
	source := image.NewRGBA(image.Rect(0, 0, 1, 1))
	source.Set(0, 0, color.RGBA{R: 255, A: 255})
	composed, err := composeCOFFrame(asset, 0, 0, map[cof.CompositeType]compositeFrame{
		head: {image: source, bounds: source.Bounds(), layer: asset.CofLayers[0]},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := color.RGBAModel.Convert(composed.At(0, 0)).(color.RGBA); got.A != 255 {
		t.Fatalf("opaque layer alpha = %d, want 255", got.A)
	}
}

func TestCOFAnimationFramesShareCanvasButKeepDistinctPixels(t *testing.T) {
	asset := cof.New()
	asset.NumberOfDirections, asset.FramesPerDirection, asset.NumberOfLayers = 1, 2, 1
	head := cof.CompositeType(0)
	asset.CofLayers = []cof.CofLayer{{Type: head, DrawEffect: cof.DrawEffect(8)}}
	asset.Priority = [][][]cof.CompositeType{{{head}, {head}}}
	first := image.NewRGBA(image.Rect(0, 0, 2, 2))
	first.Set(0, 0, color.RGBA{R: 255, A: 255})
	second := image.NewRGBA(image.Rect(0, 0, 3, 1))
	second.Set(2, 0, color.RGBA{G: 255, A: 255})
	shared := image.Rect(-2, -1, 4, 3)
	frame0, err := composeCOFFrame(asset, 0, 0, map[cof.CompositeType]compositeFrame{
		head: {image: first, bounds: image.Rect(-2, -1, 0, 1), layer: asset.CofLayers[0]},
	}, shared)
	if err != nil {
		t.Fatal(err)
	}
	frame1, err := composeCOFFrame(asset, 0, 1, map[cof.CompositeType]compositeFrame{
		head: {image: second, bounds: image.Rect(1, 2, 4, 3), layer: asset.CofLayers[0]},
	}, shared)
	if err != nil {
		t.Fatal(err)
	}
	if frame0.Bounds() != frame1.Bounds() {
		t.Fatalf("animation canvases differ: %v and %v", frame0.Bounds(), frame1.Bounds())
	}
	firstKey, secondKey := render.TextureKey(frame0), render.TextureKey(frame1)
	if firstKey == secondKey {
		t.Fatal("visually distinct composite frames share a texture identity")
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

	var composer render.Composer
	runtime := New()
	if err := runtime.RegisterModule(RenderModuleWithAssets(runtime, &composer, assets)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	source := fstest.MapFS{"screen.lua": &fstest.MapFile{Data: []byte(`
local render = require("engine.render/v1")
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
	if got != (color.RGBA{R: 30, G: 20, B: 10, A: 0xff}) {
		t.Fatalf("pixel = %#v", got)
	}
	animationNode := nodes[1]
	animation, err := composer.ResourceSnapshot(animationNode.Resource)
	if err != nil || animation.Kind != render.ResourceAnimation {
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

func TestRenderCapabilityLazyDecodesDT1LabTile(t *testing.T) {
	data := make([]byte, 276+96)
	binary.LittleEndian.PutUint32(data[0:4], 7)
	binary.LittleEndian.PutUint32(data[4:8], 6)
	binary.LittleEndian.PutUint32(data[268:272], 1)
	binary.LittleEndian.PutUint32(data[272:276], 276)
	tile := data[276:]
	binary.LittleEndian.PutUint32(tile[8:12], 0xffffffb0)
	binary.LittleEndian.PutUint32(tile[12:16], 160)
	binary.LittleEndian.PutUint32(tile[20:24], 2)
	binary.LittleEndian.PutUint32(tile[24:28], 3)
	binary.LittleEndian.PutUint32(tile[28:32], 4)
	binary.LittleEndian.PutUint32(tile[72:76], uint32(len(data)))
	assets := fstest.MapFS{"one.dt1": {Data: data}}
	capability := NewRenderCapability(New(), &render.Composer{}, assets)

	prepared, err := capability.cache.loadDT1Tile(assets, "one.dt1", "", 0, "composite")
	if err != nil {
		t.Fatal(err)
	}
	if prepared.total != 1 || prepared.tile.Type != 2 || prepared.tile.Style != 3 || prepared.tile.Sequence != 4 {
		t.Fatalf("DT1 metadata = %#v, total %d", prepared.tile, prepared.total)
	}
	if prepared.image.Bounds().Dx() != 160 || prepared.image.Bounds().Dy() != 80 {
		t.Fatalf("DT1 fallback bounds = %v", prepared.image.Bounds())
	}
	diagnostics := capability.Diagnostics()
	if diagnostics.Stages["dt1-file"].Calls != 1 || diagnostics.Stages["dt1-tile"].Calls != 1 {
		t.Fatalf("DT1 stage diagnostics = %#v", diagnostics.Stages)
	}
}

func TestRenderCapabilityPreloadsAssetsAndReportsProgress(t *testing.T) {
	palette := make([]byte, 256*3)
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
		"one.dc6": {Data: dc6Data},
		"pal.dat": {Data: palette},
	}

	capability := NewRenderCapability(New(), &render.Composer{}, assets)
	job := capability.preloads.Start([]AssetPreloadRequest{{Kind: "dc6_animation", Path: "one.dc6", Palette: "pal.dat", Anchor: "offsets"}})
	deadline := time.Now().Add(time.Second)
	for {
		status, ok := capability.preloads.Status(job)
		if !ok {
			t.Fatal("preload job disappeared")
		}
		if status.Done {
			if status.Completed != 1 || status.Failed != 0 {
				t.Fatalf("preload status = %#v", status)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("preload did not finish")
		}
		time.Sleep(time.Millisecond)
	}

	before := capability.Diagnostics().DecodeCalls
	if _, err := capability.cache.loadDC6Animation(assets, "one.dc6", "pal.dat", 0, "offsets"); err != nil {
		t.Fatal(err)
	}
	if after := capability.Diagnostics().DecodeCalls; after != before {
		t.Fatalf("warm asset decoded again: before=%d after=%d", before, after)
	}
	stage := capability.Diagnostics().Stages["dc6-animation"]
	if stage.Calls != 1 || stage.Time <= 0 {
		t.Fatalf("dc6 animation stage diagnostics = %#v", stage)
	}
	diagnostics := capability.Diagnostics()
	if got := diagnostics.Stages["dc6-file"].Calls; got != 1 {
		t.Fatalf("lazy DC6 file opens = %d, want 1", got)
	}
	if got := diagnostics.Stages["dc6-direction"].Calls; got != 1 {
		t.Fatalf("decoded DC6 directions = %d, want 1", got)
	}
	if got := diagnostics.Stages["dc6"].Calls; got != 0 {
		t.Fatalf("full DC6 decodes = %d, want 0", got)
	}
	if diagnostics.Encoded.Weight == 0 {
		t.Fatal("lazy DC6 file was not retained in the encoded cache tier")
	}
}

func TestDC6DirectionLoadingDoesNotDecodeOtherDirections(t *testing.T) {
	palette := make([]byte, 256*3)
	data := make([]byte, 70+32)
	put := func(offset int, value uint32) { binary.LittleEndian.PutUint32(data[offset:offset+4], value) }
	put(0, 6)
	put(16, 2)
	put(20, 1)
	put(24, 32)
	put(28, 70)
	put(36, 1)
	put(40, 1)
	put(60, 3)
	data[64], data[65], data[66] = 1, 1, 0x80
	assets := fstest.MapFS{"two.dc6": {Data: data}, "pal.dat": {Data: palette}}
	capability := NewRenderCapability(New(), &render.Composer{}, assets)

	if _, err := capability.cache.loadDC6Direction(assets, "two.dc6", "pal.dat", 0); err != nil {
		t.Fatalf("valid requested direction failed: %v", err)
	}
	if _, err := capability.cache.loadDC6Direction(assets, "two.dc6", "pal.dat", 1); err == nil {
		t.Fatal("malformed second direction unexpectedly decoded")
	}
}

func TestRenderCapabilityRequiresComponentScope(t *testing.T) {
	t.Parallel()

	var composer render.Composer
	runtime := New()
	if err := runtime.RegisterModule(RenderModule(runtime, &composer)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	err := runtime.Execute(context.Background(), fstest.MapFS{"bad.lua": &fstest.MapFile{Data: []byte(`local render = require("engine.render/v1"); render.create("world")`)}}, "bad.lua")
	if err == nil {
		t.Fatal("expected unscoped allocation to fail")
	}
}

func TestChunkResourcesUseTheirResidencyCacheBudgets(t *testing.T) {
	capability := NewRenderCapability(New(), &render.Composer{}, fstest.MapFS{})
	tier, namespace := capability.cache.tier("ds1-chunks\x00map")
	if tier != capability.cache.composed || namespace != "composed" {
		t.Fatalf("DS1 chunk set uses %q cache tier", namespace)
	}
	tier, namespace = capability.cache.tier("world-chunk\x00world\x000")
	if tier != capability.cache.world || namespace != "world" {
		t.Fatalf("visible world chunk uses %q cache tier", namespace)
	}
	tier, namespace = capability.cache.tier("world-chunks\x00world")
	if tier != capability.cache.decoded || namespace != "decoded" {
		t.Fatalf("world chunk index uses %q cache tier", namespace)
	}
}

func TestOwnedRenderNodeReleaseToleratesRecursiveParentDestruction(t *testing.T) {
	var composer render.Composer
	parentID, err := composer.Create(render.NodeID{}, render.LayerWorld)
	if err != nil {
		t.Fatal(err)
	}
	childID, err := composer.Create(parentID, render.LayerWorld)
	if err != nil {
		t.Fatal(err)
	}
	parent := &ownedRenderNode{composer: &composer, id: parentID}
	child := &ownedRenderNode{composer: &composer, id: childID}
	if err := parent.release(); err != nil {
		t.Fatal(err)
	}
	if err := child.release(); err != nil {
		t.Fatalf("releasing recursively destroyed child: %v", err)
	}
}

func TestAnimationSharesIdenticalRGBAFrames(t *testing.T) {
	composer := &render.Composer{}
	nodeID, err := composer.Create(render.NodeID{}, render.LayerHUD)
	if err != nil {
		t.Fatal(err)
	}
	node := &ownedRenderNode{composer: composer, id: nodeID}
	first := image.NewRGBA(image.Rect(0, 0, 2, 2))
	first.Pix[0] = 1
	duplicate := image.NewRGBA(image.Rect(0, 0, 2, 2))
	copy(duplicate.Pix, first.Pix)
	different := image.NewRGBA(image.Rect(0, 0, 2, 2))
	different.Pix[0] = 2
	if err := node.setAnimation([]image.Image{first, duplicate, different}, time.Millisecond, "loop"); err != nil {
		t.Fatal(err)
	}
	if len(node.owned) != 2 {
		t.Fatalf("owned textures = %d, want 2", len(node.owned))
	}
	resource, err := composer.ResourceSnapshot(node.resource)
	if err != nil {
		t.Fatal(err)
	}
	animation := resource.Payload.(render.AnimationData)
	if animation.Frames[0] != animation.Frames[1] || animation.Frames[1] == animation.Frames[2] {
		t.Fatalf("animation frames were not deduplicated: %v", animation.Frames)
	}
}
