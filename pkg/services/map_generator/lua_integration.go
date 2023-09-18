package map_generator

import (
	"strconv"

	lua "github.com/yuin/gopher-lua"
)

func (s *Service) ExportToLua(state *lua.LState) {
	table := state.NewTable()

	state.SetField(table, "seed", state.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(s.Seed()))
		return 1
	}))

	state.SetField(table, "setSeed", state.NewFunction(func(L *lua.LState) int {
		if L.GetTop() != 1 {
			return 0
		}

		luaNumber := L.CheckNumber(1)

		seed, err := strconv.ParseInt(luaNumber.String(), 10, 64)
		if err != nil {
			return 0
		}

		s.SetSeed(uint64(seed))

		return 0
	}))

	state.SetField(table, "generateMap", state.NewFunction(func(L *lua.LState) int {
		// check argument count
		if L.GetTop() != 2 {
			return 0
		}

		act := L.CheckInt(1)
		difficulty := L.CheckInt(2)

		mapObject, err := s.GenerateMap(uint(act), uint(difficulty))
		if err != nil {
			L.Push(nil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(mapObject.LuaTable(state))
		L.Push(nil)
		return 2
	}))

	state.SetGlobal("map_generator", table)
}

func (s *Service) UnexportFromLua(state *lua.LState) {
	state.SetGlobal("map_generator", lua.LNil)
}
