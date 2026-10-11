package modruntime

import (
	"bytes"
	"encoding/json"

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
		Help: ModuleHelp{
			Summary: "Register and update deterministic engine-owned authoritative state.",
		},
	}
}

// registerAuthorityState registers register authority state before use so Lua cannot observe a partially configured
// capability.
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

// Cached JSON values remain private; each read converts independent Lua tables.
func readAuthorityState(stores *simulation.StateStore) lua.LGFunction {
	type cachedState struct {
		data  []byte
		value any
	}
	cache := map[string]cachedState{}
	return func(state *lua.LState) int {
		id := state.CheckString(1)

		value, found := stores.Read(id)
		if !found {
			state.Push(lua.LNil)
			return 1
		}

		cached, present := cache[id]
		changed := !present || !bytes.Equal(cached.data, value.Data)
		if changed {
			var decoded any
			if err := json.Unmarshal(value.Data, &decoded); err != nil {
				state.RaiseError("decode authoritative state %q: %v", id, err)
				return 0
			}
			cached = cachedState{data: value.Data, value: decoded}
		}

		decoded, err := goToLua(state, cached.value, 0)
		if err != nil {
			state.RaiseError("decode authoritative state %q: convert registered state: %v", id, err)
			return 0
		}
		if changed {
			cache[id] = cached
		}
		state.Push(decoded)

		return 1
	}
}

// replaceAuthorityState returns the replace authority state observation under the service lock so callers cannot
// see a partially sampled input frame.
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

// encodeLuaState returns the encode lua state observation under the service lock so callers cannot see a partially
// sampled input frame.
func encodeLuaState(value lua.LValue) ([]byte, error) {
	converted, err := luaToGo(value, 0)
	if err != nil {
		return nil, err
	}

	return json.Marshal(converted)
}
