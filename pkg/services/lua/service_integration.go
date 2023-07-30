package lua

import (
	"github.com/gravestench/runtime"
	"github.com/yuin/gopher-lua"
)

var (
	_ runtime.Service       = &Service{}
	_ runtime.HasLogger     = &Service{}
	_ ManagesLuaEnvironment = &Service{}
)

type Dependency = ManagesLuaEnvironment

type ManagesLuaEnvironment interface {
	LuaState() *lua.LState
}

type UsesLuaEnvironment interface {
	ExportToLua(state *lua.LState)
}
