package screenManager

import (
	lua "github.com/yuin/gopher-lua"
)

// these methods are automatically invoked
// by the lua service to export stuff into the
// lua environment for use in scripts.

func (s *Service) ExportToLua(state *lua.LState) {
	// add stuff here to the global lua state machine
}

func (s *Service) UnexportFromLua(state *lua.LState) {
	// remove stuff you added in your export method above
}
