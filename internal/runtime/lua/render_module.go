package modruntime

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/fs"
	"sync"
	"time"

	cof "github.com/gravestench/cof"
	"github.com/gravestench/dark-magic/internal/assets/decode"
	cachepkg "github.com/gravestench/dark-magic/internal/cache"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	dc6 "github.com/gravestench/dc6/pkg"
	dcc "github.com/gravestench/dcc/pkg"
	lua "github.com/yuin/gopher-lua"
)

const renderNodeType = "dm.render.node/v1"

type ownedRenderNode struct {
	composer *render.Composer
	id       render.NodeID
	resource render.ResourceID
	palette  render.ResourceID
	owned    []render.ResourceID
	assets   fs.FS
	cache    *renderAssetCache
	once     sync.Once
	err      error
}

type renderAssetCache struct {
	mu          sync.Mutex
	generation  uint64
	decoded     *cachepkg.Cache
	decodeCalls uint64
	decodeTime  time.Duration
}

// RenderDiagnostics is a stable profiling snapshot of decoded and retained
// renderer state. Cache weight estimates retained decoded bytes, while retained
// texture bytes estimates expanded RGBA residency.
type RenderDiagnostics struct {
	Decoded     cachepkg.Stats
	Retained    render.Diagnostics
	DecodeCalls uint64
	DecodeTime  time.Duration
}

// RenderCapability owns the shared asset cache behind dm.render/v1.
type RenderCapability struct {
	runtime  *Runtime
	composer *render.Composer
	assets   fs.FS
	cache    *renderAssetCache
}

func NewRenderCapability(runtime *Runtime, composer *render.Composer, assets fs.FS) *RenderCapability {
	const decodedAssetBudget = 64 * 1024 * 1024
	return &RenderCapability{runtime: runtime, composer: composer, assets: assets, cache: &renderAssetCache{decoded: cachepkg.New(decodedAssetBudget)}}
}

func (r *RenderCapability) Diagnostics() RenderDiagnostics {
	r.cache.mu.Lock()
	defer r.cache.mu.Unlock()
	return RenderDiagnostics{Decoded: r.cache.decoded.Diagnostics(), Retained: r.composer.Diagnostics(), DecodeCalls: r.cache.decodeCalls, DecodeTime: r.cache.decodeTime}
}

type generationSource interface{ Generation() uint64 }

type compositeFrame struct {
	image  image.Image
	bounds image.Rectangle
	layer  cof.CofLayer
}

type rgbaFrameDigest struct {
	width, height int
	pixels        [32]byte
}

func composeCOFFrame(asset *cof.COF, direction, frame int, components map[cof.CompositeType]compositeFrame) (image.Image, error) {
	if direction < 0 || direction >= len(asset.Priority) {
		return nil, fmt.Errorf("COF direction %d is out of range", direction)
	}
	if frame < 0 || frame >= len(asset.Priority[direction]) {
		return nil, fmt.Errorf("COF frame %d is out of range", frame)
	}
	var bounds image.Rectangle
	for _, component := range components {
		if bounds.Empty() {
			bounds = component.bounds
		} else {
			bounds = bounds.Union(component.bounds)
		}
	}
	if bounds.Empty() {
		return nil, errors.New("COF composition has no component frames")
	}
	output := image.NewRGBA(image.Rect(0, 0, bounds.Dx()+2, bounds.Dy()+2))
	for _, componentType := range asset.Priority[direction][frame] {
		component, ok := components[componentType]
		if !ok {
			continue
		}
		destination := component.bounds.Min.Sub(bounds.Min)
		if component.layer.Shadow != 0 {
			mask := image.NewUniform(color.RGBA{A: 96})
			draw.DrawMask(output, component.image.Bounds().Add(destination.Add(image.Pt(2, 2))), mask, image.Point{}, component.image, component.image.Bounds().Min, draw.Over)
		}
		alpha := uint8(255)
		switch component.layer.DrawEffect {
		case cof.DrawEffect(0):
			alpha = 191
		case cof.DrawEffect(1):
			alpha = 128
		case cof.DrawEffect(2):
			alpha = 64
		}
		if component.layer.Transparent && alpha == 255 {
			alpha = 128
		}
		if alpha == 255 {
			draw.Draw(output, component.image.Bounds().Add(destination), component.image, component.image.Bounds().Min, draw.Over)
		} else {
			mask := image.NewUniform(color.Alpha{A: alpha})
			draw.DrawMask(output, component.image.Bounds().Add(destination), component.image, component.image.Bounds().Min, mask, image.Point{}, draw.Over)
		}
	}
	return output, nil
}

func (c *renderAssetCache) refresh(assets fs.FS) {
	source, ok := assets.(generationSource)
	if !ok || source.Generation() == c.generation {
		return
	}
	c.generation = source.Generation()
	c.decoded.InvalidateNamespace("decoded", c.generation)
}

func assetWeight(assets fs.FS, names ...string) int {
	weight := 0
	for _, name := range names {
		file, err := assets.Open(name)
		if err != nil {
			continue
		}
		count, readErr := io.Copy(io.Discard, file)
		_ = file.Close()
		if readErr == nil {
			weight += int(count)
		}
	}
	if weight < 1 {
		return 1
	}
	return weight
}

func dc6DecodedWeight(asset *dc6.DC6) int {
	weight := 0
	for _, direction := range asset.Directions {
		for _, frame := range direction.Frames {
			weight += len(frame.FrameData) + len(frame.Terminator) + len(frame.IndexData)
		}
	}
	return max(weight, 1)
}

func dccDecodedWeight(asset *dcc.DCC) int {
	weight := 0
	for _, direction := range asset.Directions() {
		weight += len(direction.PixelData)
		for _, frame := range direction.Frames() {
			weight += len(frame.PixelData)
		}
	}
	return max(weight, 1)
}

func (c *renderAssetCache) loadFont(assets fs.FS, table, sheet, palette, transform string) (*assetdecode.BitmapFont, error) {
	key := table + "\x00" + sheet + "\x00" + palette + "\x00" + transform
	c.mu.Lock()
	defer c.mu.Unlock()
	c.refresh(assets)
	if cached, ok := c.decoded.RetrieveVersioned("decoded", "font\x00"+key, c.generation); ok {
		return cached.(*assetdecode.BitmapFont), nil
	}
	started := time.Now()
	font, err := assetdecode.LoadBitmapFontWithTransform(assets, table, sheet, palette, transform)
	c.decodeCalls++
	c.decodeTime += time.Since(started)
	if err != nil {
		return nil, err
	}
	if err := c.decoded.InsertVersioned("decoded", "font\x00"+key, c.generation, font, assetWeight(assets, table, sheet, palette)); err != nil {
		return nil, err
	}
	return font, nil
}

func (c *renderAssetCache) loadCOF(assets fs.FS, name string) (*cof.COF, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.refresh(assets)
	if cached, ok := c.decoded.RetrieveVersioned("decoded", "cof\x00"+name, c.generation); ok {
		return cached.(*cof.COF), nil
	}
	started := time.Now()
	asset, err := assetdecode.COF(assets, name)
	c.decodeCalls++
	c.decodeTime += time.Since(started)
	if err != nil {
		return nil, err
	}
	if err := c.decoded.InsertVersioned("decoded", "cof\x00"+name, c.generation, asset, assetWeight(assets, name)); err != nil {
		return nil, err
	}
	return asset, nil
}

func (c *renderAssetCache) loadDCC(assets fs.FS, name, palette string) (*dcc.DCC, error) {
	key := name + "\x00" + palette
	c.mu.Lock()
	defer c.mu.Unlock()
	c.refresh(assets)
	if cached, ok := c.decoded.RetrieveVersioned("decoded", "dcc\x00"+key, c.generation); ok {
		return cached.(*dcc.DCC), nil
	}
	started := time.Now()
	asset, err := assetdecode.DCC(assets, name, palette)
	c.decodeCalls++
	c.decodeTime += time.Since(started)
	if err != nil {
		return nil, err
	}
	if err := c.decoded.InsertVersioned("decoded", "dcc\x00"+key, c.generation, asset, dccDecodedWeight(asset)); err != nil {
		return nil, err
	}
	return asset, nil
}

func (c *renderAssetCache) loadDC6(assets fs.FS, name, palette string) (*dc6.DC6, error) {
	key := name + "\x00" + palette
	c.mu.Lock()
	defer c.mu.Unlock()
	c.refresh(assets)
	if cached, ok := c.decoded.RetrieveVersioned("decoded", "dc6\x00"+key, c.generation); ok {
		return cached.(*dc6.DC6), nil
	}
	started := time.Now()
	asset, err := assetdecode.DC6(assets, name, palette)
	c.decodeCalls++
	c.decodeTime += time.Since(started)
	if err != nil {
		return nil, err
	}
	if err := c.decoded.InsertVersioned("decoded", "dc6\x00"+key, c.generation, asset, dc6DecodedWeight(asset)); err != nil {
		return nil, err
	}
	return asset, nil
}

func (n *ownedRenderNode) release() error {
	n.once.Do(func() {
		n.err = n.composer.Destroy(n.id)
		if n.resource != (render.ResourceID{}) {
			n.err = errors.Join(n.err, n.composer.DestroyResource(n.resource))
			n.resource = render.ResourceID{}
		}
		if n.palette != (render.ResourceID{}) {
			n.err = errors.Join(n.err, n.composer.DestroyResource(n.palette))
			n.palette = render.ResourceID{}
		}
		for _, resource := range n.owned {
			n.err = errors.Join(n.err, n.composer.DestroyResource(resource))
		}
		n.owned = nil
	})
	return n.err
}

func (n *ownedRenderNode) setPaletteQuantization(palette color.Palette) error {
	resource, err := n.composer.CreateResource(render.ResourcePalette, palette)
	if err != nil {
		return err
	}
	previous := n.palette
	if err := n.composer.Update(n.id, func(current *render.Node) { current.Palette = resource }); err != nil {
		_ = n.composer.DestroyResource(resource)
		return err
	}
	n.palette = resource
	if previous != (render.ResourceID{}) {
		return n.composer.DestroyResource(previous)
	}
	return nil
}

func (n *ownedRenderNode) clearPaletteQuantization() error {
	previous := n.palette
	if err := n.composer.Update(n.id, func(current *render.Node) { current.Palette = render.ResourceID{} }); err != nil {
		return err
	}
	n.palette = render.ResourceID{}
	if previous != (render.ResourceID{}) {
		return n.composer.DestroyResource(previous)
	}
	return nil
}

func (n *ownedRenderNode) setImage(decoded image.Image) error {
	resource, err := n.composer.CreateResource(render.ResourceTexture, decoded)
	if err != nil {
		return err
	}
	previous := n.resource
	previousOwned := n.owned
	if err := n.composer.Update(n.id, func(current *render.Node) { current.Resource = resource }); err != nil {
		_ = n.composer.DestroyResource(resource)
		return err
	}
	n.resource = resource
	n.owned = nil
	if previous != (render.ResourceID{}) {
		err = n.composer.DestroyResource(previous)
	}
	for _, owned := range previousOwned {
		err = errors.Join(err, n.composer.DestroyResource(owned))
	}
	return err
}

func (n *ownedRenderNode) setAnimation(frames []image.Image, duration time.Duration, loop string) error {
	textures := make([]render.ResourceID, 0, len(frames))
	owned := make([]render.ResourceID, 0, len(frames))
	duplicates := make(map[rgbaFrameDigest]render.ResourceID)
	cleanup := func() {
		for _, texture := range owned {
			_ = n.composer.DestroyResource(texture)
		}
	}
	for _, frame := range frames {
		key, shareable := rgbaFrameKey(frame)
		if shareable {
			if texture, exists := duplicates[key]; exists {
				textures = append(textures, texture)
				continue
			}
		}
		texture, err := n.composer.CreateResource(render.ResourceTexture, frame)
		if err != nil {
			cleanup()
			return err
		}
		textures = append(textures, texture)
		owned = append(owned, texture)
		if shareable {
			duplicates[key] = texture
		}
	}
	durations := make([]time.Duration, len(textures))
	for index := range durations {
		durations[index] = duration
	}
	animation, err := n.composer.CreateResource(render.ResourceAnimation, render.AnimationData{Frames: textures, Durations: durations, Loop: loop})
	if err != nil {
		cleanup()
		return err
	}
	previous, previousOwned := n.resource, n.owned
	if err := n.composer.Update(n.id, func(current *render.Node) {
		current.Resource = animation
		current.AnimationPaused = false
		current.AnimationSeek = 0
		current.AnimationSeekRevision++
	}); err != nil {
		_ = n.composer.DestroyResource(animation)
		cleanup()
		return err
	}
	n.resource, n.owned = animation, owned
	if previous != (render.ResourceID{}) {
		err = n.composer.DestroyResource(previous)
	}
	for _, owned := range previousOwned {
		err = errors.Join(err, n.composer.DestroyResource(owned))
	}
	return err
}

func rgbaFrameKey(frame image.Image) (rgbaFrameDigest, bool) {
	rgba, ok := frame.(*image.RGBA)
	if !ok || rgba.Stride != rgba.Bounds().Dx()*4 {
		return rgbaFrameDigest{}, false
	}
	bounds := rgba.Bounds()
	size := bounds.Dx() * bounds.Dy() * 4
	start := rgba.PixOffset(bounds.Min.X, bounds.Min.Y)
	if size <= 0 || start < 0 || start+size > len(rgba.Pix) {
		return rgbaFrameDigest{}, false
	}
	return rgbaFrameDigest{width: bounds.Dx(), height: bounds.Dy(), pixels: sha256.Sum256(rgba.Pix[start : start+size])}, true
}

func (n *ownedRenderNode) requireAnimation() error {
	if n.resource == (render.ResourceID{}) {
		return errors.New("render node has no animation")
	}
	resource, err := n.composer.ResourceSnapshot(n.resource)
	if err != nil {
		return err
	}
	if resource.Kind != render.ResourceAnimation {
		return errors.New("render node resource is not an animation")
	}
	return nil
}

func (n *ownedRenderNode) cofFrames(cofName, palette string, direction int, paths map[string]string) ([]image.Image, *cof.COF, error) {
	asset, err := n.cache.loadCOF(n.assets, cofName)
	if err != nil {
		return nil, nil, err
	}
	if direction < 0 || direction >= asset.NumberOfDirections {
		return nil, nil, fmt.Errorf("COF direction %d is out of range", direction)
	}
	layers := make(map[cof.CompositeType]cof.CofLayer, len(asset.CofLayers))
	decoded := make(map[cof.CompositeType]*dcc.DCC)
	for _, layer := range asset.CofLayers {
		name, ok := paths[layer.Type.String()]
		if !ok || name == "" {
			continue
		}
		component, err := n.cache.loadDCC(n.assets, name, palette)
		if err != nil {
			return nil, nil, fmt.Errorf("COF layer %s: %w", layer.Type, err)
		}
		if direction >= len(component.Directions()) {
			return nil, nil, fmt.Errorf("COF layer %s lacks direction %d", layer.Type, direction)
		}
		layers[layer.Type], decoded[layer.Type] = layer, component
	}
	frames := make([]image.Image, asset.FramesPerDirection)
	for frameIndex := range frames {
		components := make(map[cof.CompositeType]compositeFrame, len(decoded))
		for componentType, component := range decoded {
			directionFrames := component.Direction(direction).Frames()
			if frameIndex >= len(directionFrames) {
				return nil, nil, fmt.Errorf("COF layer %s lacks frame %d", componentType, frameIndex)
			}
			frame := directionFrames[frameIndex]
			normalized := image.NewRGBA(image.Rect(0, 0, frame.Bounds().Dx(), frame.Bounds().Dy()))
			draw.Draw(normalized, normalized.Bounds(), frame, frame.Bounds().Min, draw.Src)
			components[componentType] = compositeFrame{image: normalized, bounds: frame.Bounds(), layer: layers[componentType]}
		}
		frames[frameIndex], err = composeCOFFrame(asset, direction, frameIndex, components)
		if err != nil {
			return nil, nil, err
		}
	}
	return frames, asset, nil
}

func luaComponentPaths(state *lua.LState, index int) map[string]string {
	result := make(map[string]string)
	table := state.CheckTable(index)
	table.ForEach(func(key, value lua.LValue) {
		if key.Type() == lua.LTString && value.Type() == lua.LTString {
			result[key.String()] = value.String()
		}
	})
	return result
}

// normalizedDC6Frames places every cropped frame on one shared canvas using
// the DC6 anchor offsets. The retained node can then animate at a fixed world
// position without jitter when individual frame bounds change.
func dc6FrameTop(frame *dc6.Frame) int {
	if frame.Flipped > 0 {
		return int(frame.OffsetY)
	}
	return int(frame.OffsetY) - int(frame.Height) + 1
}

func dc6AnimationBounds(asset *dc6.DC6, direction int) image.Rectangle {
	var bounds image.Rectangle
	for index, frame := range asset.Directions[direction].Frames {
		top := dc6FrameTop(frame)
		frameBounds := image.Rect(int(frame.OffsetX), top, int(frame.OffsetX+int32(frame.Width)), top+int(frame.Height))
		if index == 0 {
			bounds = frameBounds
		} else {
			bounds = bounds.Union(frameBounds)
		}
	}
	return bounds
}

func dc6FixedAnimationBounds(asset *dc6.DC6, direction int) image.Rectangle {
	frames := asset.Directions[direction].Frames
	if len(frames) == 0 {
		return image.Rectangle{}
	}
	var width, height int
	for _, frame := range frames {
		width = max(width, int(frame.Width))
		height = max(height, int(frame.Height))
	}
	top := int(frames[0].OffsetY) - height + 1
	if frames[0].Flipped > 0 {
		top = int(frames[0].OffsetY)
	}
	return image.Rect(int(frames[0].OffsetX), top,
		int(frames[0].OffsetX)+width, top+height)
}

func normalizedDC6Frames(asset *dc6.DC6, direction int, anchorMode string, sharedBounds ...image.Rectangle) ([]image.Image, image.Rectangle, error) {
	frames := asset.Directions[direction].Frames
	var bounds image.Rectangle
	if len(sharedBounds) > 0 {
		bounds = sharedBounds[0]
	} else if anchorMode == "first-frame" && len(frames) > 0 {
		var width, height int
		for _, frame := range frames {
			width = max(width, int(frame.Width))
			height = max(height, int(frame.Height))
		}
		top := int(frames[0].OffsetY) - height + 1
		if frames[0].Flipped > 0 {
			top = int(frames[0].OffsetY)
		}
		bounds = image.Rect(int(frames[0].OffsetX), top, int(frames[0].OffsetX)+width, top+height)
	}
	for index, frame := range frames {
		if len(sharedBounds) > 0 {
			continue
		}
		if anchorMode == "first-frame" {
			continue
		}
		top := dc6FrameTop(frame)
		frameBounds := image.Rect(int(frame.OffsetX), top, int(frame.OffsetX+int32(frame.Width)), top+int(frame.Height))
		if index == 0 {
			bounds = frameBounds
		} else {
			bounds = bounds.Union(frameBounds)
		}
	}
	if bounds.Empty() {
		return nil, image.Rectangle{}, errors.New("DC6 animation has no visible frames")
	}
	result := make([]image.Image, len(frames))
	for index, frame := range frames {
		decoded, err := assetdecode.FrameImage(asset, frame)
		if err != nil {
			return nil, image.Rectangle{}, err
		}
		canvas := image.NewRGBA(image.Rectangle{Max: bounds.Size()})
		position := image.Point{}
		if anchorMode == "first-frame" && len(sharedBounds) > 0 {
			position = image.Pt(int(frames[0].OffsetX)-bounds.Min.X, dc6FrameTop(frames[0])-bounds.Min.Y)
		} else if anchorMode != "first-frame" {
			position = image.Pt(int(frame.OffsetX)-bounds.Min.X, dc6FrameTop(frame)-bounds.Min.Y)
		}
		draw.Draw(canvas, decoded.Bounds().Add(position), decoded, decoded.Bounds().Min, draw.Src)
		result[index] = canvas
	}
	return result, bounds, nil
}

// RenderModule exposes backend-neutral retained composition to scoped Lua
// components. Nodes are automatically destroyed with their component scope.
func RenderModule(runtime *Runtime, composer *render.Composer) Module {
	return RenderModuleWithAssets(runtime, composer, nil)
}

// RenderModuleWithAssets additionally lets nodes decode standard image assets
// from the layered content filesystem. Decoding occurs on the Lua owner;
// renderer upload remains queued for the renderer thread.
func RenderModuleWithAssets(runtime *Runtime, composer *render.Composer, assets fs.FS) Module {
	return NewRenderCapability(runtime, composer, assets).Module()
}

// Module returns the versioned Lua render capability.
func (r *RenderCapability) Module() Module {
	runtime, composer, assets, cache := r.runtime, r.composer, r.assets, r.cache
	return Module{Name: "dm.render/v1", Loader: func(state *lua.LState) int {
		registerRenderNodeType(state)
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"diagnostics": func(state *lua.LState) int {
				decoded, retained := cache.decoded.Diagnostics(), composer.Diagnostics()
				result := state.NewTable()
				result.RawSetString("decoded_entries", lua.LNumber(decoded.Entries))
				result.RawSetString("decoded_weight", lua.LNumber(decoded.Weight))
				result.RawSetString("decoded_budget", lua.LNumber(decoded.Budget))
				result.RawSetString("cache_hits", lua.LNumber(decoded.Hits))
				result.RawSetString("cache_misses", lua.LNumber(decoded.Misses))
				result.RawSetString("cache_evictions", lua.LNumber(decoded.Evictions))
				result.RawSetString("active_nodes", lua.LNumber(retained.ActiveNodes))
				result.RawSetString("active_resources", lua.LNumber(retained.ActiveResources))
				result.RawSetString("pending_commands", lua.LNumber(retained.Pending))
				state.Push(result)
				return 1
			},
			"assets_available": func(state *lua.LState) int {
				state.Push(lua.LBool(assets != nil))
				return 1
			},
			"dc6_animation_bounds": func(state *lua.LState) int {
				if assets == nil {
					state.RaiseError("render asset filesystem is unavailable")
					return 0
				}
				asset, err := cache.loadDC6(assets, state.CheckString(1), state.OptString(2, ""))
				if err != nil {
					state.RaiseError("%v", err)
					return 0
				}
				direction := state.OptInt(3, 0)
				if direction < 0 || direction >= len(asset.Directions) {
					state.ArgError(3, "direction is out of range")
					return 0
				}
				var bounds image.Rectangle
				if state.OptString(4, "offsets") == "first-frame" {
					bounds = dc6FixedAnimationBounds(asset, direction)
				} else {
					bounds = dc6AnimationBounds(asset, direction)
				}
				state.Push(lua.LNumber(bounds.Min.X))
				state.Push(lua.LNumber(bounds.Min.Y))
				state.Push(lua.LNumber(bounds.Max.X))
				state.Push(lua.LNumber(bounds.Max.Y))
				return 4
			},
			"cof_info": func(state *lua.LState) int {
				if assets == nil {
					state.Push(lua.LNil)
					state.Push(lua.LString("render asset filesystem is unavailable"))
					return 2
				}
				asset, err := cache.loadCOF(assets, state.CheckString(1))
				if err != nil {
					state.Push(lua.LNil)
					state.Push(lua.LString(err.Error()))
					return 2
				}
				result := state.NewTable()
				result.RawSetString("directions", lua.LNumber(asset.NumberOfDirections))
				result.RawSetString("frames", lua.LNumber(asset.FramesPerDirection))
				result.RawSetString("speed", lua.LNumber(asset.Speed))
				layers := state.NewTable()
				for _, layer := range asset.CofLayers {
					entry := state.NewTable()
					entry.RawSetString("type", lua.LString(layer.Type.String()))
					entry.RawSetString("shadow", lua.LNumber(layer.Shadow))
					entry.RawSetString("selectable", lua.LBool(layer.Selectable))
					entry.RawSetString("transparent", lua.LBool(layer.Transparent))
					entry.RawSetString("draw_effect", lua.LNumber(layer.DrawEffect))
					entry.RawSetString("weapon_class", lua.LString(layer.WeaponClass.String()))
					layers.Append(entry)
				}
				result.RawSetString("layers", layers)
				events := state.NewTable()
				for _, event := range asset.AnimationFrames {
					events.Append(lua.LNumber(event))
				}
				result.RawSetString("events", events)
				priority := state.NewTable()
				for _, direction := range asset.Priority {
					directionTable := state.NewTable()
					for _, frame := range direction {
						frameTable := state.NewTable()
						for _, layer := range frame {
							frameTable.Append(lua.LString(layer.String()))
						}
						directionTable.Append(frameTable)
					}
					priority.Append(directionTable)
				}
				result.RawSetString("priority", priority)
				state.Push(result)
				return 1
			},
			"create": func(state *lua.LState) int {
				scope, err := runtime.requireActiveScope()
				if err != nil {
					state.RaiseError("%v", err)
					return 0
				}
				layer, err := parseLayer(state.CheckString(1))
				if err != nil {
					state.RaiseError("%v", err)
					return 0
				}
				var parent render.NodeID
				if state.GetTop() >= 2 && state.Get(2) != lua.LNil {
					parent = checkRenderNode(state, 2).id
				}
				id, err := composer.Create(parent, layer)
				if err != nil {
					state.RaiseError("creating render node: %v", err)
					return 0
				}
				node := &ownedRenderNode{composer: composer, id: id, assets: assets, cache: cache}
				if err := scope.Add(node.release); err != nil {
					_ = node.release()
					state.RaiseError("owning render node: %v", err)
					return 0
				}
				userData := state.NewUserData()
				userData.Value = node
				state.SetMetatable(userData, state.GetTypeMetatable(renderNodeType))
				state.Push(userData)
				return 1
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}

func registerRenderNodeType(state *lua.LState) {
	meta := state.NewTypeMetatable(renderNodeType)
	state.SetField(meta, "__index", state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
		"set_position": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			x, y := float64(state.CheckNumber(2)), float64(state.CheckNumber(3))
			if err := node.composer.Update(node.id, func(current *render.Node) { current.X, current.Y = x, y }); err != nil {
				state.RaiseError("updating render node: %v", err)
			}
			return 0
		},
		"set_scale": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			x, y := float64(state.CheckNumber(2)), float64(state.CheckNumber(3))
			if err := node.composer.Update(node.id, func(current *render.Node) { current.ScaleX, current.ScaleY = x, y }); err != nil {
				state.RaiseError("updating render node: %v", err)
			}
			return 0
		},
		"set_z": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			z := state.CheckInt(2)
			if err := node.composer.Update(node.id, func(current *render.Node) { current.Z = z }); err != nil {
				state.RaiseError("updating render node: %v", err)
			}
			return 0
		},
		"set_rotation": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			rotation := float64(state.CheckNumber(2))
			if err := node.composer.Update(node.id, func(current *render.Node) { current.Rotation = rotation }); err != nil {
				state.RaiseError("updating render node: %v", err)
			}
			return 0
		},
		"set_blend": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			blend := state.CheckString(2)
			switch blend {
			case "alpha", "additive", "multiply", "add-colors", "subtract-colors":
			default:
				state.ArgError(2, "blend must be alpha, additive, multiply, add-colors, or subtract-colors")
				return 0
			}
			if err := node.composer.Update(node.id, func(current *render.Node) { current.Blend = blend }); err != nil {
				state.RaiseError("updating render node: %v", err)
			}
			return 0
		},
		"set_palette_quantization": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			name := state.CheckString(2)
			palette, err := assetdecode.DisplayPalette(node.assets, name)
			if err != nil {
				state.RaiseError("loading display palette %q: %v", name, err)
				return 0
			}
			if err := node.setPaletteQuantization(palette); err != nil {
				state.RaiseError("applying display palette: %v", err)
			}
			return 0
		},
		"clear_palette_quantization": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if err := node.clearPaletteQuantization(); err != nil {
				state.RaiseError("clearing display palette: %v", err)
			}
			return 0
		},
		"set_visible": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			visible := state.CheckBool(2)
			if err := node.composer.Update(node.id, func(current *render.Node) { current.Visible = visible }); err != nil {
				state.RaiseError("updating render node: %v", err)
			}
			return 0
		},
		"set_clip": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			x, y := float64(state.CheckNumber(2)), float64(state.CheckNumber(3))
			width, height := float64(state.CheckNumber(4)), float64(state.CheckNumber(5))
			if width <= 0 || height <= 0 {
				state.ArgError(4, "clip width and height must be positive")
				return 0
			}
			if err := node.composer.Update(node.id, func(current *render.Node) {
				current.Clip = &render.Rect{X: x, Y: y, Width: width, Height: height}
			}); err != nil {
				state.RaiseError("updating render clip: %v", err)
			}
			return 0
		},
		"clear_clip": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if err := node.composer.Update(node.id, func(current *render.Node) { current.Clip = nil }); err != nil {
				state.RaiseError("clearing render clip: %v", err)
			}
			return 0
		},
		"set_image": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			fileName := state.CheckString(2)
			file, err := node.assets.Open(fileName)
			if err != nil {
				state.RaiseError("opening image %q: %v", fileName, err)
				return 0
			}
			decoded, _, err := image.Decode(file)
			_ = file.Close()
			if err != nil {
				state.RaiseError("decoding image %q: %v", fileName, err)
				return 0
			}
			if err := node.setImage(decoded); err != nil {
				state.RaiseError("updating render node: %v", err)
			}
			return 0
		},
		"set_dc6": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			fileName := state.CheckString(2)
			paletteName := state.OptString(3, "")
			direction := state.OptInt(4, 0)
			frameIndex := state.OptInt(5, 0)
			asset, err := node.cache.loadDC6(node.assets, fileName, paletteName)
			if err != nil {
				state.RaiseError("%v", err)
				return 0
			}
			frame, err := assetdecode.Frame(asset, direction, frameIndex)
			if err != nil {
				state.RaiseError("%v", err)
				return 0
			}
			decoded, err := assetdecode.FrameImage(asset, frame)
			if err != nil {
				state.RaiseError("%v", err)
				return 0
			}
			if err := node.setImage(decoded); err != nil {
				state.RaiseError("updating render node: %v", err)
				return 0
			}
			state.Push(lua.LNumber(frame.Width))
			state.Push(lua.LNumber(frame.Height))
			state.Push(lua.LNumber(frame.OffsetX))
			state.Push(lua.LNumber(dc6FrameTop(frame)))
			return 4
		},
		"set_dc6_animation": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			fileName := state.CheckString(2)
			paletteName := state.OptString(3, "")
			direction := state.OptInt(4, 0)
			framesPerSecond := float64(state.OptNumber(5, 15))
			loop := state.OptString(6, "loop")
			anchorMode := state.OptString(7, "offsets")
			if framesPerSecond <= 0 {
				state.ArgError(5, "frames per second must be positive")
				return 0
			}
			if loop != "loop" && loop != "once" && loop != "ping-pong" {
				state.ArgError(6, "loop mode must be loop, once, or ping-pong")
				return 0
			}
			if anchorMode != "offsets" && anchorMode != "first-frame" {
				state.ArgError(7, "anchor mode must be offsets or first-frame")
				return 0
			}
			asset, err := node.cache.loadDC6(node.assets, fileName, paletteName)
			if err != nil {
				state.RaiseError("%v", err)
				return 0
			}
			if direction < 0 || direction >= len(asset.Directions) {
				state.ArgError(4, "direction is out of range")
				return 0
			}
			var sharedBounds []image.Rectangle
			if state.GetTop() >= 11 {
				bounds := image.Rect(state.CheckInt(8), state.CheckInt(9), state.CheckInt(10), state.CheckInt(11))
				if bounds.Empty() {
					state.ArgError(8, "shared animation bounds must not be empty")
					return 0
				}
				sharedBounds = append(sharedBounds, bounds)
			}
			frames, bounds, err := normalizedDC6Frames(asset, direction, anchorMode, sharedBounds...)
			if err != nil {
				state.RaiseError("%v", err)
				return 0
			}
			if err := node.setAnimation(frames, time.Duration(float64(time.Second)/framesPerSecond), loop); err != nil {
				state.RaiseError("updating render animation: %v", err)
				return 0
			}
			state.Push(lua.LNumber(len(frames)))
			state.Push(lua.LNumber(bounds.Dx()))
			state.Push(lua.LNumber(bounds.Dy()))
			state.Push(lua.LNumber(bounds.Min.X))
			state.Push(lua.LNumber(bounds.Min.Y))
			return 5
		},
		"set_dcc": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			fileName, paletteName := state.CheckString(2), state.OptString(3, "")
			direction, frameIndex := state.OptInt(4, 0), state.OptInt(5, 0)
			asset, err := node.cache.loadDCC(node.assets, fileName, paletteName)
			if err != nil {
				state.RaiseError("%v", err)
				return 0
			}
			frames, err := assetdecode.DCCFrames(asset, direction)
			if err != nil {
				state.RaiseError("%v", err)
				return 0
			}
			if frameIndex < 0 || frameIndex >= len(frames) {
				state.ArgError(5, "frame is out of range")
				return 0
			}
			if err := node.setImage(frames[frameIndex]); err != nil {
				state.RaiseError("updating DCC render node: %v", err)
			}
			bounds := asset.Direction(direction).Frames()[frameIndex].Bounds()
			state.Push(lua.LNumber(bounds.Dx()))
			state.Push(lua.LNumber(bounds.Dy()))
			state.Push(lua.LNumber(bounds.Min.X))
			state.Push(lua.LNumber(bounds.Min.Y))
			return 4
		},
		"set_dcc_animation": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			fileName, paletteName := state.CheckString(2), state.OptString(3, "")
			direction := state.OptInt(4, 0)
			framesPerSecond := float64(state.OptNumber(5, 25))
			loop := state.OptString(6, "loop")
			if framesPerSecond <= 0 {
				state.ArgError(5, "frames per second must be positive")
				return 0
			}
			if loop != "loop" && loop != "once" && loop != "ping-pong" {
				state.ArgError(6, "loop mode must be loop, once, or ping-pong")
				return 0
			}
			asset, err := node.cache.loadDCC(node.assets, fileName, paletteName)
			if err != nil {
				state.RaiseError("%v", err)
				return 0
			}
			frames, err := assetdecode.DCCFrames(asset, direction)
			if err != nil {
				state.RaiseError("%v", err)
				return 0
			}
			if err := node.setAnimation(frames, time.Duration(float64(time.Second)/framesPerSecond), loop); err != nil {
				state.RaiseError("updating DCC animation: %v", err)
				return 0
			}
			state.Push(lua.LNumber(len(frames)))
			return 1
		},
		"set_cof": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			cofName, palette := state.CheckString(2), state.OptString(3, "")
			direction, frameIndex := state.OptInt(4, 0), state.OptInt(5, 0)
			frames, asset, err := node.cofFrames(cofName, palette, direction, luaComponentPaths(state, 6))
			if err != nil {
				state.RaiseError("composing COF: %v", err)
				return 0
			}
			if frameIndex < 0 || frameIndex >= len(frames) {
				state.ArgError(5, "frame is out of range")
				return 0
			}
			if err := node.setImage(frames[frameIndex]); err != nil {
				state.RaiseError("updating COF render node: %v", err)
				return 0
			}
			state.Push(lua.LNumber(frames[frameIndex].Bounds().Dx()))
			state.Push(lua.LNumber(frames[frameIndex].Bounds().Dy()))
			state.Push(lua.LNumber(asset.AnimationFrames[frameIndex]))
			return 3
		},
		"set_cof_animation": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			cofName, palette := state.CheckString(2), state.OptString(3, "")
			direction := state.OptInt(4, 0)
			paths := luaComponentPaths(state, 5)
			loop := state.OptString(6, "loop")
			frames, asset, err := node.cofFrames(cofName, palette, direction, paths)
			if err != nil {
				state.RaiseError("composing COF animation: %v", err)
				return 0
			}
			if asset.Speed <= 0 {
				state.RaiseError("COF speed must be positive")
				return 0
			}
			duration := time.Duration(float64(time.Second) * 256 / (float64(asset.Speed) * 25))
			if err := node.setAnimation(frames, duration, loop); err != nil {
				state.RaiseError("updating COF animation: %v", err)
				return 0
			}
			events := state.NewTable()
			for _, event := range asset.AnimationFrames {
				events.Append(lua.LNumber(event))
			}
			state.Push(lua.LNumber(len(frames)))
			state.Push(events)
			return 2
		},
		"set_text": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			tableName, sheetName := state.CheckString(2), state.CheckString(3)
			paletteName, text := state.OptString(4, ""), state.CheckString(5)
			red, green, blue, alpha := 255, 255, 255, 255
			transform := ""
			maxWidth, align := 0, "left"
			if state.GetTop() >= 6 && state.Get(6) != lua.LNil {
				options := state.CheckTable(6)
				integer := func(name string, fallback int) int {
					value := options.RawGetString(name)
					if value == lua.LNil {
						return fallback
					}
					return int(lua.LVAsNumber(value))
				}
				red, green, blue, alpha = integer("red", red), integer("green", green), integer("blue", blue), integer("alpha", alpha)
				if value := options.RawGetString("transform"); value != lua.LNil {
					transform = lua.LVAsString(value)
				}
				maxWidth = integer("max_width", 0)
				if value := options.RawGetString("align"); value != lua.LNil {
					align = lua.LVAsString(value)
				}
			}
			for _, channel := range []int{red, green, blue, alpha} {
				if channel < 0 || channel > 255 {
					state.ArgError(6, "text color channels must be between 0 and 255")
					return 0
				}
			}
			if maxWidth < 0 {
				state.ArgError(6, "max_width cannot be negative")
				return 0
			}
			font, err := node.cache.loadFont(node.assets, tableName, sheetName, paletteName, transform)
			if err != nil {
				state.RaiseError("loading bitmap font: %v", err)
				return 0
			}
			rendered, err := font.Render(text, color.RGBA{R: uint8(red), G: uint8(green), B: uint8(blue), A: uint8(alpha)}, maxWidth, align)
			if err != nil {
				state.RaiseError("rendering bitmap text: %v", err)
				return 0
			}
			if err := node.setImage(rendered); err != nil {
				state.RaiseError("updating text render node: %v", err)
				return 0
			}
			state.Push(lua.LNumber(rendered.Bounds().Dx()))
			state.Push(lua.LNumber(rendered.Bounds().Dy()))
			return 2
		},
		"animation_pause": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if err := node.requireAnimation(); err != nil {
				state.RaiseError("pausing animation: %v", err)
				return 0
			}
			if err := node.composer.Update(node.id, func(current *render.Node) { current.AnimationPaused = true }); err != nil {
				state.RaiseError("pausing animation: %v", err)
			}
			return 0
		},
		"animation_resume": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if err := node.requireAnimation(); err != nil {
				state.RaiseError("resuming animation: %v", err)
				return 0
			}
			if err := node.composer.Update(node.id, func(current *render.Node) { current.AnimationPaused = false }); err != nil {
				state.RaiseError("resuming animation: %v", err)
			}
			return 0
		},
		"animation_seek": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			seconds := float64(state.CheckNumber(2))
			if seconds < 0 {
				state.ArgError(2, "seek position cannot be negative")
				return 0
			}
			if err := node.requireAnimation(); err != nil {
				state.RaiseError("seeking animation: %v", err)
				return 0
			}
			position := time.Duration(seconds * float64(time.Second))
			if err := node.composer.Update(node.id, func(current *render.Node) {
				current.AnimationSeek = position
				current.AnimationSeekRevision++
			}); err != nil {
				state.RaiseError("seeking animation: %v", err)
			}
			return 0
		},
		"fill_rect": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			width, height := state.CheckInt(2), state.CheckInt(3)
			if width <= 0 || height <= 0 {
				state.ArgError(2, "positive width and height required")
				return 0
			}
			channel := func(index int, fallback int) uint8 {
				value := fallback
				if state.GetTop() >= index {
					value = state.CheckInt(index)
				}
				if value < 0 || value > 255 {
					state.ArgError(index, "color channel must be between 0 and 255")
				}
				return uint8(value)
			}
			fill := image.NewRGBA(image.Rect(0, 0, width, height))
			draw.Draw(fill, fill.Bounds(), &image.Uniform{C: color.RGBA{R: channel(4, 0), G: channel(5, 0), B: channel(6, 0), A: channel(7, 255)}}, image.Point{}, draw.Src)
			if err := node.setImage(fill); err != nil {
				state.RaiseError("updating render node: %v", err)
			}
			return 0
		},
		"destroy": func(state *lua.LState) int {
			if err := checkRenderNode(state, 1).release(); err != nil {
				state.RaiseError("destroying render node: %v", err)
			}
			return 0
		},
	}))
}

func checkRenderNode(state *lua.LState, index int) *ownedRenderNode {
	userData := state.CheckUserData(index)
	node, ok := userData.Value.(*ownedRenderNode)
	if !ok {
		state.ArgError(index, "dm.render/v1 node expected")
		return nil
	}
	return node
}

func parseLayer(name string) (render.Layer, error) {
	switch name {
	case "world":
		return render.LayerWorld, nil
	case "hud":
		return render.LayerHUD, nil
	case "modal":
		return render.LayerModal, nil
	case "cursor":
		return render.LayerCursor, nil
	case "debug":
		return render.LayerDebug, nil
	case "transition":
		return render.LayerTransition, nil
	default:
		return 0, fmt.Errorf("unknown render layer %q", name)
	}
}
