package luaManager

import (
	"github.com/gravestench/servicemesh"
	"github.com/yuin/gopher-lua"
)

var (
	_ servicemesh.Service             = &Service{}
	_ servicemesh.HasLogger           = &Service{}
	_ servicemesh.HasGracefulShutdown = &Service{}
	_ ManagesLuaEnvironment           = &Service{}
)

type Dependency = ManagesLuaEnvironment

type ManagesLuaEnvironment interface {
	Ready() bool
	WithState(fn func(state *lua.LState) error) error
	GlobalsExist(globals ...string) bool
	RebuildState()
}

type LuaPlugin interface {
	ExportToLua(state *lua.LState, apiRootTable *lua.LTable)
	UnexportFromLua(state *lua.LState, apiRootTable *lua.LTable)
}
