package input

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	lua "github.com/yuin/gopher-lua"
)

const LuaAPIKey = "input"

// these methods are automatically invoked
// by the lua service to export stuff into the
// lua environment for use in scripts.

func (s *Service) ExportToLua(state *lua.LState, rootTable *lua.LTable) {
	table := state.NewTable()
	keys := state.NewTable()
	for name, code := range map[string]int32{
		"Grave": rl.KeyGrave, "Escape": rl.KeyEscape, "Enter": rl.KeyEnter,
		"Space": rl.KeySpace, "Up": rl.KeyUp, "Down": rl.KeyDown,
		"Left": rl.KeyLeft, "Right": rl.KeyRight,
	} {
		state.SetField(keys, name, lua.LNumber(code))
	}
	state.SetField(table, "Key", keys)
	state.SetField(table, "IsDown", state.NewFunction(s.luaIsDown))
	state.SetField(table, "Cursor", state.NewFunction(s.luaCursor))
	state.SetField(table, "OnKeyPressed", state.NewFunction(s.luaOnKeyPressed))
	state.SetField(rootTable, LuaAPIKey, table)
}

func (s *Service) UnexportFromLua(state *lua.LState, rootTable *lua.LTable) {
	state.SetField(rootTable, LuaAPIKey, lua.LNil)
	s.callbackMux.Lock()
	s.keyPressedCallbacks = make(map[int32][]*lua.LFunction)
	s.callbackMux.Unlock()
}

func (s *Service) luaIsDown(L *lua.LState) int {
	key := int32(L.CheckInt(1))
	state := s.KeyboardState()[key]
	L.Push(lua.LBool(state == StateDown || state == StatePressed))
	return 1
}

func (s *Service) luaCursor(L *lua.LState) int {
	x, y := s.MouseCursorState()
	L.Push(lua.LNumber(x))
	L.Push(lua.LNumber(y))
	return 2
}

func (s *Service) luaOnKeyPressed(L *lua.LState) int {
	key := int32(L.CheckInt(1))
	callback := L.CheckFunction(2)
	s.callbackMux.Lock()
	s.keyPressedCallbacks[key] = append(s.keyPressedCallbacks[key], callback)
	s.callbackMux.Unlock()
	return 0
}

func (s *Service) dispatchKeyPressed(key int32) {
	s.callbackMux.Lock()
	callbacks := append([]*lua.LFunction(nil), s.keyPressedCallbacks[key]...)
	s.callbackMux.Unlock()
	if len(callbacks) == 0 {
		return
	}
	if err := s.lua.WithState(func(state *lua.LState) error {
		for _, callback := range callbacks {
			if err := state.CallByParam(lua.P{Fn: callback, NRet: 0, Protect: true}, lua.LNumber(key)); err != nil {
				s.Logger().Error("running Lua input callback", "key", key, "error", err)
			}
		}
		return nil
	}); err != nil {
		s.Logger().Error("accessing Lua for input callback", "error", err)
	}
}
