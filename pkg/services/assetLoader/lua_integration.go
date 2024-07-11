package assetLoader

import (
	"io"

	lua "github.com/yuin/gopher-lua"

	"github.com/gravestench/dark-magic/pkg/services/luaManager"
)

var _ luaManager.LuaPlugin = &Service{}

func (s *Service) LuaPluginPreload(state *lua.LState) {

}

func (s *Service) LuaPluginLoadIntoTable(state *lua.LState, rootTable *lua.LTable) {
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
			s.logger.Error("LUA: loading asset", "error", err)
			return 0
		}

		ud := state.NewUserData()
		ud.Value = data
		L.Push(ud)

		return 1
	})

	table := state.NewTable()
	state.SetField(table, "Open", fn)

	rootTable.RawSetString("assets", table)
}

func (s *Service) LuaPluginUnloadFromTable(state *lua.LState, rootTable *lua.LTable) {
	rootTable.RawSetString("assets", lua.LNil)
}
