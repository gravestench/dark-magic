package lua

import (
	"github.com/gravestench/servicemesh"
	"github.com/yuin/gopher-lua"

	"github.com/gravestench/dark-magic/pkg/services/fileWatcher"
)

var (
	_ servicemesh.Service          = &Service{}
	_ servicemesh.HasLogger        = &Service{}
	_ fileWatcher.NeedsFileWatcher = &Service{}
	_ ManagesLuaEnvironment        = &Service{}
)

type Dependency = ManagesLuaEnvironment

type ManagesLuaEnvironment interface {
	LuaState() *lua.LState
}

type UsesLuaEnvironment interface {
	ExportToLua(state *lua.LState)
	UnexportFromLua(state *lua.LState)
}
