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

func TestDC6DecodedWeightCountsRetainedFrameBuffers(t *testing.T) {
	asset := &dc6.DC6{Directions: []*dc6.Direction{{Frames: []*dc6.Frame{
		{FrameData: []byte{1, 2}, Terminator: []byte{3, 4, 5}, IndexData: []byte{6, 7, 8, 9}},
	}}}}
	if got := dc6DecodedWeight(asset); got != 9 {
		t.Fatalf("decoded weight = %d, want 9", got)
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
	if len(nodes) != 1 || nodes[0].X != 320 || nodes[0].Y != 240 || nodes[0].Z != 7 || nodes[0].Layer != render.LayerTransition {
		t.Fatalf("nodes = %#v", nodes)
	}
	if nodes[0].Clip == nil || *nodes[0].Clip != (render.Rect{X: 10, Y: 20, Width: 300, Height: 200}) {
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

func TestRenderNodePaletteQuantizationAcceptsTinyPalettes(t *testing.T) {
	t.Parallel()

	var composer render.Composer
	runtime := New()
	assets := fstest.MapFS{
		"screen.lua": {Data: []byte(`
local render = require("dm.render/v1")
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
	if got := color.RGBAModel.Convert(composed.At(1, 1)).(color.RGBA); got.R != 255 {
		t.Fatalf("priority pixel = %#v", got)
	}
	if composed.Bounds().Dx() != 5 || composed.Bounds().Dy() != 5 {
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
	if _, _, _, alpha := composed.At(2, 2).RGBA(); alpha == 0 {
		t.Fatal("shadow-enabled COF layer produced no offset shadow pixel")
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
	if got := color.RGBAModel.Convert(composed.At(2, 2)).(color.RGBA); got.R != 255 || got.A != 255 {
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
	err := runtime.Execute(context.Background(), fstest.MapFS{"bad.lua": &fstest.MapFile{Data: []byte(`local render = require("dm.render/v1"); render.create("world")`)}}, "bad.lua")
	if err == nil {
		t.Fatal("expected unscoped allocation to fail")
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
