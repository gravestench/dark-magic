package assetLoader

import (
	"io"

	lua "github.com/yuin/gopher-lua"
)

func (s *Service) ExportToLua(state *lua.LState) {
	fn := state.NewFunction(func(L *lua.LState) int {
		// check argument count
		if L.GetTop() != 1 {
			return 0
		}

		mpqInternalPath := L.CheckString(1)

		reader, err := s.Load(mpqInternalPath)
		if err != nil {
			return 0
		}

		data, err := io.ReadAll(reader)
		if err != nil {
			s.logger.Error().Msgf("LUA: loading asset: %v", err)
			return 0
		}

		ud := state.NewUserData()
		ud.Value = data
		L.Push(ud)

		return 1
	})

	table := state.NewTable()
	state.SetField(table, "load", fn)
	state.SetGlobal("assets", table)
}

func (s *Service) UnexportFromLua(state *lua.LState) {
	state.SetGlobal("assets", lua.LNil)
}
