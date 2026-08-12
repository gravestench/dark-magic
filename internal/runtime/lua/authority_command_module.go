package modruntime

import (
	"context"
	"encoding/json"
	"fmt"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	lua "github.com/yuin/gopher-lua"
)

// AuthorityCommandModule lets a trusted mod own command policy without giving
// it ownership of admission. Go still checks player identity, tick windows,
// monotonic sequence numbers, and allowed authority classes before Lua runs.
func AuthorityCommandModule(runtime *Runtime, session *gamesession.Session) Module {
	return Module{
		Name: "dm.authority_command/v1",
		Loader: func(state *lua.LState) int {
			module := state.NewTable()
			state.SetFuncs(module, map[string]lua.LGFunction{
				"register": registerAuthorityCommand(runtime, session),
			})
			state.Push(module)
			return 1
		},
		Help: ModuleHelp{Summary: "Register deterministic authoritative command validation and application."},
	}
}

func registerAuthorityCommand(runtime *Runtime, session *gamesession.Session) lua.LGFunction {
	return func(state *lua.LState) int {
		definition := state.CheckTable(1)
		kind := requiredTableString(state, definition, "kind")
		validate, ok := definition.RawGetString("validate").(*lua.LFunction)
		if !ok {
			state.ArgError(1, "validate must be a function")
		}
		apply, ok := definition.RawGetString("apply").(*lua.LFunction)
		if !ok {
			state.ArgError(1, "apply must be a function")
		}
		authorities, err := luaAuthorities(state, definition.RawGetString("authorities"))
		if err != nil {
			state.ArgError(1, err.Error())
		}

		handler := gamesession.CommandHandler{
			Allowed: authorities,
			Validate: func(command simulation.Command) error {
				return callLuaCommand(runtime, validate, command)
			},
			Apply: func(_ *gameecs.Engine, command simulation.Command) error {
				return callLuaCommand(runtime, apply, command)
			},
		}
		if err := session.Register(kind, handler); err != nil {
			state.RaiseError("register authoritative command %q: %v", kind, err)
		}
		state.Push(lua.LString(kind))
		return 1
	}
}

func callLuaCommand(runtime *Runtime, function *lua.LFunction, command simulation.Command) error {
	return runtime.Run(context.Background(), func(state *lua.LState) error {
		value, err := luaCommandValue(state, command)
		if err != nil {
			return err
		}
		if err := state.CallByParam(lua.P{Fn: function, NRet: 0, Protect: true}, value); err != nil {
			return err
		}
		return nil
	})
}

func luaCommandValue(state *lua.LState, command simulation.Command) (*lua.LTable, error) {
	var payload any
	if err := json.Unmarshal(command.Payload, &payload); err != nil {
		return nil, fmt.Errorf("decode command payload: %w", err)
	}
	converted, err := goToLua(state, payload, 0)
	if err != nil {
		return nil, fmt.Errorf("convert command payload: %w", err)
	}
	value := state.NewTable()
	value.RawSetString("tick", lua.LNumber(command.Tick))
	value.RawSetString("player", lua.LString(command.Player))
	value.RawSetString("authority", lua.LString(command.Authority))
	value.RawSetString("sequence", lua.LNumber(command.Sequence))
	value.RawSetString("kind", lua.LString(command.Kind))
	value.RawSetString("payload", converted)
	return value, nil
}

func luaAuthorities(state *lua.LState, value lua.LValue) ([]simulation.Authority, error) {
	if value == lua.LNil {
		return []simulation.Authority{simulation.AuthorityPlayer}, nil
	}
	table, ok := value.(*lua.LTable)
	if !ok || table.Len() == 0 {
		return nil, fmt.Errorf("authorities must be a non-empty array")
	}
	result := make([]simulation.Authority, 0, table.Len())
	for index := 1; index <= table.Len(); index++ {
		name, ok := table.RawGetInt(index).(lua.LString)
		if !ok {
			return nil, fmt.Errorf("authority %d must be a string", index)
		}
		authority := simulation.Authority(name)
		if authority != simulation.AuthorityPlayer && authority != simulation.AuthorityAdmin && authority != simulation.AuthoritySystem {
			return nil, fmt.Errorf("unknown authority %q", name)
		}
		result = append(result, authority)
	}
	return result, nil
}
