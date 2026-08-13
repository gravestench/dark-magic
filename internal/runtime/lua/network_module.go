package modruntime

import lua "github.com/yuin/gopher-lua"

// NetworkController is the presentation-safe application boundary for
// multiplayer topology. Lua requests an operation and observes copied status;
// it never receives sockets, credentials, certificates, or session handles.
type NetworkController interface {
	Host() error
	Join(address string) error
	Status() map[string]any
}

func NetworkModule(controller NetworkController) Module {
	return Module{Name: "engine.network/v1", Help: documentedModule("Request multiplayer hosting/joining and inspect safe connection progress.", map[string]CommandHelp{
		"host":   commandHelp("engine.network.host()", "Start self-hosting with the selected local-profile character."),
		"join":   commandHelp("engine.network.join(address)", "Join a self-host advertised at the supplied address."),
		"status": commandHelp("engine.network.status()", "Return copied phase, mode, address, and safe error code."),
	}), Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"host": func(state *lua.LState) int {
				accepted := false
				if controller != nil {
					accepted = controller.Host() == nil
				}
				state.Push(lua.LBool(accepted))
				return 1
			},
			"join": func(state *lua.LState) int {
				address := state.OptString(1, "")
				accepted := false
				if controller != nil {
					accepted = controller.Join(address) == nil
				}
				state.Push(lua.LBool(accepted))
				return 1
			},
			"status": func(state *lua.LState) int {
				if controller == nil {
					state.Push(state.NewTable())
					return 1
				}
				value, err := goToLua(state, controller.Status(), 0)
				if err != nil {
					state.RaiseError("copy network status: %v", err)
					return 0
				}
				state.Push(value)
				return 1
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}
