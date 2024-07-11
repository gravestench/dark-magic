package modalGameUI

import (
	lua "github.com/yuin/gopher-lua"
)

// these methods are automatically invoked
// by the lua service to export stuff into the
// lua environment for use in scripts.

func (s *Service) LuaPluginLoadIntoTable(state *lua.LState, rootTable *lua.LTable) {
	// add stuff here to the global lua state machine
}

func (s *Service) LuaPluginUnloadFromTable(state *lua.LState, rootTable *lua.LTable) {
	// remove stuff you added in your export method above
}
