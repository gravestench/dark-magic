package modruntime

import (
	"encoding/json"
	"fmt"

	"github.com/gravestench/dark-magic/internal/game/simulation"
	lua "github.com/yuin/gopher-lua"
)

// AuthorityStateModule exposes engine-owned registered state to trusted
// authoritative scripts. It deliberately does not expose the session, ECS
// internals, files, networking, or native resources.
func AuthorityStateModule(stores *simulation.StateStore) Module {
	return Module{
		Name: "engine.authority_state/v1",
		Loader: func(state *lua.LState) int {
			module := state.NewTable()
			state.SetFuncs(module, map[string]lua.LGFunction{
				"register": registerAuthorityState(stores),
				"read":     readAuthorityState(stores),
				"replace":  replaceAuthorityState(stores),
			})
			state.Push(module)
			return 1
		},
		Help: ModuleHelp{Summary: "Register and update deterministic engine-owned authoritative state."},
	}
}

func registerAuthorityState(stores *simulation.StateStore) lua.LGFunction {
	return func(state *lua.LState) int {
		id := state.CheckString(1)
		schema := state.CheckString(2)
		data, err := encodeLuaState(state.CheckAny(3))
		if err != nil {
			state.RaiseError("encode authoritative state %q: %v", id, err)
			return 0
		}
		if err := stores.Register(id, schema, data); err != nil {
			state.RaiseError("register authoritative state %q: %v", id, err)
		}
		return 0
	}
}

func readAuthorityState(stores *simulation.StateStore) lua.LGFunction {
	return func(state *lua.LState) int {
		id := state.CheckString(1)
		value, found := stores.Read(id)
		if !found {
			state.Push(lua.LNil)
			return 1
		}
		decoded, err := decodeLuaState(state, value.Data)
		if err != nil {
			state.RaiseError("decode authoritative state %q: %v", id, err)
			return 0
		}
		state.Push(decoded)
		return 1
	}
}

func replaceAuthorityState(stores *simulation.StateStore) lua.LGFunction {
	return func(state *lua.LState) int {
		id := state.CheckString(1)
		schema := state.CheckString(2)
		data, err := encodeLuaState(state.CheckAny(3))
		if err != nil {
			state.RaiseError("encode authoritative state %q: %v", id, err)
			return 0
		}
		if err := stores.Replace(id, schema, data); err != nil {
			state.RaiseError("replace authoritative state %q: %v", id, err)
		}
		return 0
	}
}

func encodeLuaState(value lua.LValue) ([]byte, error) {
	converted, err := luaToGo(value, 0)
	if err != nil {
		return nil, err
	}
	return json.Marshal(converted)
}

func decodeLuaState(state *lua.LState, data []byte) (lua.LValue, error) {
	var decoded any
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, err
	}
	value, err := goToLua(state, decoded, 0)
	if err != nil {
		return nil, fmt.Errorf("convert registered state: %w", err)
	}
	return value, nil
}
