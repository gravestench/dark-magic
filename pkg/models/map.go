package models

import (
	lua "github.com/yuin/gopher-lua"
)

type Diablo2Map struct {
	ZoneGraph []Zone
}

func (m *Diablo2Map) LuaTable(state *lua.LState) *lua.LTable {
	return nil
}

type Zone struct {
}
