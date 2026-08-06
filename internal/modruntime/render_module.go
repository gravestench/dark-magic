package modruntime

import (
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
	"github.com/gravestench/dark-magic/internal/rendercore"
	"github.com/gravestench/dark-magic/pkg/assetdecode"
	cachepkg "github.com/gravestench/dark-magic/pkg/cache"
	dc6 "github.com/gravestench/dc6/pkg"
	dcc "github.com/gravestench/dcc/pkg"
	lua "github.com/yuin/gopher-lua"
)

const renderNodeType = "dm.render.node/v1"

type ownedRenderNode struct {
	composer *rendercore.Composer
	id       rendercore.NodeID
	resource rendercore.ResourceID
	owned    []rendercore.ResourceID
	assets   fs.FS
	cache    *renderAssetCache
	once     sync.Once
	err      error
}

type renderAssetCache struct {
	mu         sync.Mutex
	generation uint64
	decoded    *cachepkg.Cache
}

type generationSource interface{ Generation() uint64 }

type compositeFrame struct {
	image  image.Image
	bounds image.Rectangle
	layer  cof.CofLayer
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

func (c *renderAssetCache) loadFont(assets fs.FS, table, sheet, palette string) (*assetdecode.BitmapFont, error) {
	key := table + "\x00" + sheet + "\x00" + palette
	c.mu.Lock()
	defer c.mu.Unlock()
	c.refresh(assets)
	if cached, ok := c.decoded.RetrieveVersioned("decoded", "font\x00"+key, c.generation); ok {
		return cached.(*assetdecode.BitmapFont), nil
	}
	font, err := assetdecode.LoadBitmapFont(assets, table, sheet, palette)
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
	asset, err := assetdecode.COF(assets, name)
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
	asset, err := assetdecode.DCC(assets, name, palette)
	if err != nil {
		return nil, err
	}
	if err := c.decoded.InsertVersioned("decoded", "dcc\x00"+key, c.generation, asset, assetWeight(assets, name, palette)); err != nil {
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
	asset, err := assetdecode.DC6(assets, name, palette)
	if err != nil {
		return nil, err
	}
	if err := c.decoded.InsertVersioned("decoded", "dc6\x00"+key, c.generation, asset, assetWeight(assets, name, palette)); err != nil {
		return nil, err
	}
	return asset, nil
}

func (n *ownedRenderNode) release() error {
	n.once.Do(func() {
		n.err = n.composer.Destroy(n.id)
		if n.resource != (rendercore.ResourceID{}) {
			n.err = errors.Join(n.err, n.composer.DestroyResource(n.resource))
			n.resource = rendercore.ResourceID{}
		}
		for _, resource := range n.owned {
			n.err = errors.Join(n.err, n.composer.DestroyResource(resource))
		}
		n.owned = nil
	})
	return n.err
}

func (n *ownedRenderNode) setImage(decoded image.Image) error {
	resource, err := n.composer.CreateResource(rendercore.ResourceTexture, decoded)
	if err != nil {
		return err
	}
	previous := n.resource
	previousOwned := n.owned
	if err := n.composer.Update(n.id, func(current *rendercore.Node) { current.Resource = resource }); err != nil {
		_ = n.composer.DestroyResource(resource)
		return err
	}
	n.resource = resource
	n.owned = nil
	if previous != (rendercore.ResourceID{}) {
		err = n.composer.DestroyResource(previous)
	}
	for _, owned := range previousOwned {
		err = errors.Join(err, n.composer.DestroyResource(owned))
	}
	return err
}

func (n *ownedRenderNode) setAnimation(frames []image.Image, duration time.Duration, loop string) error {
	textures := make([]rendercore.ResourceID, 0, len(frames))
	cleanup := func() {
		for _, texture := range textures {
			_ = n.composer.DestroyResource(texture)
		}
	}
	for _, frame := range frames {
		texture, err := n.composer.CreateResource(rendercore.ResourceTexture, frame)
		if err != nil {
			cleanup()
			return err
		}
		textures = append(textures, texture)
	}
	durations := make([]time.Duration, len(textures))
	for index := range durations {
		durations[index] = duration
	}
	animation, err := n.composer.CreateResource(rendercore.ResourceAnimation, rendercore.AnimationData{Frames: textures, Durations: durations, Loop: loop})
	if err != nil {
		cleanup()
		return err
	}
	previous, previousOwned := n.resource, n.owned
	if err := n.composer.Update(n.id, func(current *rendercore.Node) {
		current.Resource = animation
		current.AnimationPaused = false
		current.AnimationSeek = 0
		current.AnimationSeekRevision++
	}); err != nil {
		_ = n.composer.DestroyResource(animation)
		cleanup()
		return err
	}
	n.resource, n.owned = animation, textures
	if previous != (rendercore.ResourceID{}) {
		err = n.composer.DestroyResource(previous)
	}
	for _, owned := range previousOwned {
		err = errors.Join(err, n.composer.DestroyResource(owned))
	}
	return err
}

func (n *ownedRenderNode) requireAnimation() error {
	if n.resource == (rendercore.ResourceID{}) {
		return errors.New("render node has no animation")
	}
	resource, err := n.composer.ResourceSnapshot(n.resource)
	if err != nil {
		return err
	}
	if resource.Kind != rendercore.ResourceAnimation {
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
func normalizedDC6Frames(asset *dc6.DC6, direction int) ([]image.Image, image.Rectangle, error) {
	frames := asset.Directions[direction].Frames
	var bounds image.Rectangle
	for index, frame := range frames {
		frameBounds := image.Rect(int(frame.OffsetX), -int(frame.OffsetY), int(frame.OffsetX+int32(frame.Width)), -int(frame.OffsetY)+int(frame.Height))
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
		position := image.Pt(int(frame.OffsetX)-bounds.Min.X, -int(frame.OffsetY)-bounds.Min.Y)
		draw.Draw(canvas, decoded.Bounds().Add(position), decoded, decoded.Bounds().Min, draw.Src)
		result[index] = canvas
	}
	return result, bounds, nil
}

// RenderModule exposes backend-neutral retained composition to scoped Lua
// components. Nodes are automatically destroyed with their component scope.
func RenderModule(runtime *Runtime, composer *rendercore.Composer) Module {
	return RenderModuleWithAssets(runtime, composer, nil)
}

// RenderModuleWithAssets additionally lets nodes decode standard image assets
// from the layered content filesystem. Decoding occurs on the Lua owner;
// renderer upload remains queued for the renderer thread.
func RenderModuleWithAssets(runtime *Runtime, composer *rendercore.Composer, assets fs.FS) Module {
	const decodedAssetBudget = 256 * 1024 * 1024
	cache := &renderAssetCache{decoded: cachepkg.New(decodedAssetBudget)}
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
				var parent rendercore.NodeID
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
			if err := node.composer.Update(node.id, func(current *rendercore.Node) { current.X, current.Y = x, y }); err != nil {
				state.RaiseError("updating render node: %v", err)
			}
			return 0
		},
		"set_scale": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			x, y := float64(state.CheckNumber(2)), float64(state.CheckNumber(3))
			if err := node.composer.Update(node.id, func(current *rendercore.Node) { current.ScaleX, current.ScaleY = x, y }); err != nil {
				state.RaiseError("updating render node: %v", err)
			}
			return 0
		},
		"set_z": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			z := state.CheckInt(2)
			if err := node.composer.Update(node.id, func(current *rendercore.Node) { current.Z = z }); err != nil {
				state.RaiseError("updating render node: %v", err)
			}
			return 0
		},
		"set_rotation": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			rotation := float64(state.CheckNumber(2))
			if err := node.composer.Update(node.id, func(current *rendercore.Node) { current.Rotation = rotation }); err != nil {
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
			if err := node.composer.Update(node.id, func(current *rendercore.Node) { current.Blend = blend }); err != nil {
				state.RaiseError("updating render node: %v", err)
			}
			return 0
		},
		"set_visible": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			visible := state.CheckBool(2)
			if err := node.composer.Update(node.id, func(current *rendercore.Node) { current.Visible = visible }); err != nil {
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
			if err := node.composer.Update(node.id, func(current *rendercore.Node) {
				current.Clip = &rendercore.Rect{X: x, Y: y, Width: width, Height: height}
			}); err != nil {
				state.RaiseError("updating render clip: %v", err)
			}
			return 0
		},
		"clear_clip": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if err := node.composer.Update(node.id, func(current *rendercore.Node) { current.Clip = nil }); err != nil {
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
			state.Push(lua.LNumber(frame.OffsetY))
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
			if framesPerSecond <= 0 {
				state.ArgError(5, "frames per second must be positive")
				return 0
			}
			if loop != "loop" && loop != "once" && loop != "ping-pong" {
				state.ArgError(6, "loop mode must be loop, once, or ping-pong")
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
			frames, bounds, err := normalizedDC6Frames(asset, direction)
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
			state.Push(lua.LNumber(-bounds.Min.Y))
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
			font, err := node.cache.loadFont(node.assets, tableName, sheetName, paletteName)
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
			if err := node.composer.Update(node.id, func(current *rendercore.Node) { current.AnimationPaused = true }); err != nil {
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
			if err := node.composer.Update(node.id, func(current *rendercore.Node) { current.AnimationPaused = false }); err != nil {
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
			if err := node.composer.Update(node.id, func(current *rendercore.Node) {
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

func parseLayer(name string) (rendercore.Layer, error) {
	switch name {
	case "world":
		return rendercore.LayerWorld, nil
	case "hud":
		return rendercore.LayerHUD, nil
	case "modal":
		return rendercore.LayerModal, nil
	case "cursor":
		return rendercore.LayerCursor, nil
	case "debug":
		return rendercore.LayerDebug, nil
	case "transition":
		return rendercore.LayerTransition, nil
	default:
		return 0, fmt.Errorf("unknown render layer %q", name)
	}
}
