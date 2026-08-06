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
	"io/fs"
	"sync"

	"github.com/gravestench/dark-magic/internal/rendercore"
	"github.com/gravestench/dark-magic/pkg/assetdecode"
	dc6 "github.com/gravestench/dc6/pkg"
	lua "github.com/yuin/gopher-lua"
)

const renderNodeType = "dm.render.node/v1"

type ownedRenderNode struct {
	composer *rendercore.Composer
	id       rendercore.NodeID
	resource rendercore.ResourceID
	assets   fs.FS
	cache    *renderAssetCache
	once     sync.Once
	err      error
}

type renderAssetCache struct {
	mu  sync.Mutex
	dc6 map[string]*dc6.DC6
}

func (c *renderAssetCache) loadDC6(assets fs.FS, name, palette string) (*dc6.DC6, error) {
	key := name + "\x00" + palette
	c.mu.Lock()
	defer c.mu.Unlock()
	if asset := c.dc6[key]; asset != nil {
		return asset, nil
	}
	asset, err := assetdecode.DC6(assets, name, palette)
	if err != nil {
		return nil, err
	}
	c.dc6[key] = asset
	return asset, nil
}

func (n *ownedRenderNode) release() error {
	n.once.Do(func() {
		n.err = n.composer.Destroy(n.id)
		if n.resource != (rendercore.ResourceID{}) {
			n.err = errors.Join(n.err, n.composer.DestroyResource(n.resource))
			n.resource = rendercore.ResourceID{}
		}
	})
	return n.err
}

func (n *ownedRenderNode) setImage(decoded image.Image) error {
	resource, err := n.composer.CreateResource(rendercore.ResourceTexture, decoded)
	if err != nil {
		return err
	}
	previous := n.resource
	if err := n.composer.Update(n.id, func(current *rendercore.Node) { current.Resource = resource }); err != nil {
		_ = n.composer.DestroyResource(resource)
		return err
	}
	n.resource = resource
	if previous != (rendercore.ResourceID{}) {
		return n.composer.DestroyResource(previous)
	}
	return nil
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
	cache := &renderAssetCache{dc6: make(map[string]*dc6.DC6)}
	return Module{Name: "dm.render/v1", Loader: func(state *lua.LState) int {
		registerRenderNodeType(state)
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"assets_available": func(state *lua.LState) int {
				state.Push(lua.LBool(assets != nil))
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
			if err := node.setImage(frame.ToImageRGBA()); err != nil {
				state.RaiseError("updating render node: %v", err)
				return 0
			}
			state.Push(lua.LNumber(frame.Width))
			state.Push(lua.LNumber(frame.Height))
			state.Push(lua.LNumber(frame.OffsetX))
			state.Push(lua.LNumber(frame.OffsetY))
			return 4
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
