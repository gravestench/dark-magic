# Global Lua Environment Service

The purpose of this [runtime](https://github.com/gravestench/runtime) service is
to manage the state of a single lua state machine. Other services are intended to implement an exported 
integration interface in order to populate the state machine with anything
they want. Any type of lua API for a given service should be exposed through
this service. It is recommended that other services document their API in either
a README.md or in a LUA.md

## Dependencies

This service depends upon the [config file service](../config_file) and will
initialize its own default config.

## Integration with other services

This service exports an integration interface `ManagesLuaEnvironment` with an alias
`Dependencncy` which are intended to be used by other services for dependency
resolution (see runtime.HasDependencies), and expose just the methods which
other services should use.

 Another service may merely store a reference to this service and use the
lua state machine explicitly with the folloing interface:
```golang
type Dependency = ManagesLuaEnvironment

type ManagesLuaEnvironment interface {
    LuaState() *lua.LState
}
```

Other services should use the `ManagesLuaEnvironment` or `Dependency` interfaces to resolve
their dependency on this service.

__________

However, implementors of the following interface methods are invoked by the lua
service indirectly whenever the service is added or removed from the runtime.
```golang
type UsesLuaEnvironment interface {
    ExportToLua(state *lua.LState)
    UnexportFromLua(state *lua.LState)
}
```