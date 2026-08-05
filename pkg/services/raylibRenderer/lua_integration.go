package raylibRenderer

import (
	"image"

	rl "github.com/gen2brain/raylib-go/raylib"
	lua "github.com/yuin/gopher-lua"

	luaService "github.com/gravestench/dark-magic/pkg/services/luaManager"
)

const LuaApiKey = "renderer"

var _ luaService.LuaPlugin = &Service{}

// these methods are automatically invoked
// by the lua service to export stuff into the
// lua environment for use in scripts.

func (s *Service) ExportToLua(state *lua.LState, rootTable *lua.LTable) {
	table := state.NewTable()

	s.bindMethods(state, table)

	rootTable.RawSetString(LuaApiKey, table)
}

func (s *Service) UnexportFromLua(state *lua.LState, rootTable *lua.LTable) {
	state.SetField(rootTable, LuaApiKey, lua.LNil)
}

func (s *Service) bindMethods(state *lua.LState, table *lua.LTable) {
	fnMap := map[string]lua.LGFunction{
		"NewRenderable": func(L *lua.LState) int {
			if L.GetTop() != 0 {
				return 0
			}

			r := s.MakeLuaRenderable(s.NewRenderable(), L)

			L.Push(r)

			return 1
		},
	}

	for key, fn := range fnMap {
		state.SetField(table, key, state.NewFunction(fn))
	}

	for key, subTableFn := range map[string]func(state *lua.LState, table *lua.LTable){
		"window": s.bindWindowMethods,
	} {
		subTable := state.NewTable()
		subTableFn(state, subTable)
		state.SetField(table, key, subTable)
	}
}

func (s *Service) bindWindowMethods(state *lua.LState, table *lua.LTable) {
	fnMap := map[string]lua.LGFunction{
		"Size": s.luaWindowSizeGet,
	}

	for key, fn := range fnMap {
		table.RawSetString(key, state.NewFunction(fn))
	}
}

func (s *Service) MakeLuaRenderable(r Renderable, L *lua.LState) *lua.LTable {
	t := L.NewTable()

	ud := L.NewUserData()
	ud.Value = r

	L.SetField(t, "UserData", ud)
	L.SetField(t, "UUID", s.luaRenderableClosureUUIDGet(L, r))
	L.SetField(t, "ZIndex", s.luaRenderableClosureZIndexSetGet(L, r))
	L.SetField(t, "Position", s.luaRenderableClosurePositionSetGet(L, r))
	L.SetField(t, "Rotation", s.luaRenderableClosureRotationSetGet(L, r))
	L.SetField(t, "Scale", s.luaRenderableClosureScaleSetGet(L, r))
	L.SetField(t, "Origin", s.luaRenderableClosureOriginSetGet(L, r))
	L.SetField(t, "Opacity", s.luaRenderableClosureOpacitySetGet(L, r))
	L.SetField(t, "BlendMode", s.luaRenderableClosureBlendModeSetGet(L, r))
	L.SetField(t, "Texture", s.luaRenderableClosureTextureSetGet(L, r))
	L.SetField(t, "Image", s.luaRenderableClosureImageSetGet(L, r))
	L.SetField(t, "Enable", s.luaRenderableClosureEnableSet(L, r))
	L.SetField(t, "Parent", s.luaRenderableClosureParentSetGet(L, r))

	return t
}

func (s *Service) luaWindowSizeGet(L *lua.LState) int {
	w, h := s.WindowSize()

	L.Push(lua.LNumber(w))
	L.Push(lua.LNumber(h))

	return 2
}

func (s *Service) luaRenderableClosureUUIDGet(L *lua.LState, r Renderable) *lua.LFunction {
	fn := L.NewFunction(func(L *lua.LState) int {
		numArgs := L.GetTop()

		if numArgs < 1 {
			L.Push(lua.LString(r.UUID().String()))
			return 1
		}

		return 0
	})

	return fn
}

func (s *Service) luaRenderableClosureZIndexSetGet(L *lua.LState, r Renderable) *lua.LFunction {
	fn := L.NewFunction(func(L *lua.LState) int {
		numArgs := L.GetTop()

		if numArgs < 1 {
			L.Push(lua.LNumber(r.ZIndex()))
			return 1
		}

		val := L.CheckInt(1)
		r.SetZIndex(float32(val))

		return 0
	})

	return fn
}

func (s *Service) luaRenderableClosurePositionSetGet(L *lua.LState, r Renderable) *lua.LFunction {
	fn := L.NewFunction(func(L *lua.LState) int {
		numArgs := L.GetTop()

		if numArgs < 1 {
			x, y := r.Position()
			L.Push(lua.LNumber(x))
			L.Push(lua.LNumber(y))
			return 2
		}

		if numArgs > 1 {
			x := L.CheckNumber(1)
			y := L.CheckNumber(2)

			r.SetPosition(float32(x), float32(y))

			return 0
		}

		return 0
	})

	return fn
}

func (s *Service) luaRenderableClosureRotationSetGet(L *lua.LState, r Renderable) *lua.LFunction {
	fn := L.NewFunction(func(L *lua.LState) int {
		numArgs := L.GetTop()

		if numArgs < 1 {
			L.Push(lua.LNumber(r.Rotation()))
			return 1
		}

		rotation := L.CheckNumber(1)
		r.SetRotation(float32(rotation))

		return 0
	})

	return fn
}

func (s *Service) luaRenderableClosureScaleSetGet(L *lua.LState, r Renderable) *lua.LFunction {
	fn := L.NewFunction(func(L *lua.LState) int {
		numArgs := L.GetTop()

		if numArgs < 1 {
			L.Push(lua.LNumber(r.Scale()))
			return 1
		}

		scale := L.CheckNumber(1)
		r.SetScale(float32(scale))

		return 0
	})

	return fn
}

func (s *Service) luaRenderableClosureOriginSetGet(L *lua.LState, r Renderable) *lua.LFunction {
	fn := L.NewFunction(func(L *lua.LState) int {
		numArgs := L.GetTop()

		if numArgs < 1 {
			v := r.Origin()
			L.Push(lua.LNumber(v.X))
			L.Push(lua.LNumber(v.Y))
			return 2
		}

		if numArgs < 2 {
			x := L.CheckNumber(1)
			y := x
			r.SetOrigin(float64(x), float64(y))
			return 0
		}

		x := L.CheckNumber(1)
		y := L.CheckNumber(2)
		r.SetOrigin(float64(x), float64(y))

		return 0
	})

	return fn
}

func (s *Service) luaRenderableClosureOpacitySetGet(L *lua.LState, r Renderable) *lua.LFunction {
	fn := L.NewFunction(func(L *lua.LState) int {
		numArgs := L.GetTop()

		if numArgs < 1 {
			L.Push(lua.LNumber(r.Opacity()))
			return 1
		}

		rotation := L.CheckNumber(1)
		r.SetOpacity(float32(rotation))

		return 0
	})

	return fn
}

func (s *Service) luaRenderableClosureBlendModeSetGet(L *lua.LState, r Renderable) *lua.LFunction {
	fn := L.NewFunction(func(L *lua.LState) int {
		numArgs := L.GetTop()

		if numArgs < 1 {
			blendeMode := r.BlendMode()

			L.Push(lua.LNumber(blendeMode))

			return 1
		}

		blendMode := L.CheckNumber(1)
		r.SetBlendMode(rl.BlendMode(blendMode))

		return 0
	})

	return fn
}

func (s *Service) luaRenderableClosureTextureSetGet(L *lua.LState, r Renderable) *lua.LFunction {
	fn := L.NewFunction(func(L *lua.LState) int {
		numArgs := L.GetTop()

		if numArgs < 1 {
			t := L.NewUserData()
			t.Value = r.Texture()

			L.Push(t)

			return 1
		}

		d := L.CheckUserData(1)

		switch v := d.Value.(type) {
		case rl.Texture2D:
			r.SetTexture(v)
		default:
			s.Logger().Error("setting texture: supplied value is not a rl.Texture2D")
		}

		return 0
	})

	return fn
}

func (s *Service) luaRenderableClosureImageSetGet(L *lua.LState, r Renderable) *lua.LFunction {
	fn := L.NewFunction(func(L *lua.LState) int {
		numArgs := L.GetTop()

		if numArgs < 1 {
			data := L.NewUserData()
			data.Value = r.Image()

			L.Push(data)

			return 1
		}

		d := L.CheckUserData(1)

		switch v := d.Value.(type) {
		case image.Image:
			r.SetImage(v)
		default:
			s.Logger().Error("setting image: supplied value is not an image.Image")
		}

		return 0
	})

	return fn
}

func (s *Service) luaRenderableClosureEnableSet(L *lua.LState, r Renderable) *lua.LFunction {
	fn := L.NewFunction(func(L *lua.LState) int {
		numArgs := L.GetTop()

		if numArgs < 1 {
			L.Push(lua.LBool(r.IsEnabled()))
			return 1
		}

		set := L.CheckBool(1)

		if set {
			r.Enable()
		} else {
			r.Disable()
		}

		return 0
	})

	return fn
}

func (s *Service) luaRenderableClosureParentSetGet(L *lua.LState, r Renderable) *lua.LFunction {
	fn := L.NewFunction(func(L *lua.LState) int {
		numArgs := L.GetTop()

		if numArgs < 1 {
			parent := r.Parent()
			parentTable := s.MakeLuaRenderable(parent, L)

			L.Push(parentTable)

			return 1
		}

		renderableTable := L.CheckAny(1)
		userdata := L.GetField(renderableTable, "UserData")

		switch v := userdata.(type) {
		case *lua.LNilType:
			return 0
		case *lua.LUserData:
			parentAny := v.Value

			if parent, ok := parentAny.(Renderable); ok {
				r.SetParent(parent)
			}
		}

		return 0
	})

	return fn
}
