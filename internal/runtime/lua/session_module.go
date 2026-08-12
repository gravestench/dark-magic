package modruntime

import (
	"encoding/json"

	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	lua "github.com/yuin/gopher-lua"
)

// SessionModule exposes observational authoritative-session diagnostics. It
// does not admit commands or expose the Akara world to administration scripts.
func SessionModule(session *gamesession.Session) Module {
	return Module{Name: "engine.session/v1", Help: documentedModule("Inspect the authoritative game session and export its deterministic replay.", map[string]CommandHelp{
		"audit":  commandHelp("engine.session.audit()", "Return accepted administrator and system commands as a JSON audit record."),
		"status": commandHelp("engine.session.status()", "Return the current tick and recorded/pending command and checkpoint counts."),
		"replay": commandHelp("engine.session.replay()", "Return the current versioned replay as a JSON string."),
	}), Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"audit": func(state *lua.LState) int {
				encoded, err := json.Marshal(session.Audit())
				if err != nil {
					state.RaiseError("encode session audit: %v", err)
					return 0
				}
				state.Push(lua.LString(encoded))
				return 1
			},
			"status": func(state *lua.LState) int {
				status := session.Status()
				result := state.NewTable()
				result.RawSetString("tick", lua.LNumber(status.Tick))
				result.RawSetString("pending_commands", lua.LNumber(status.Pending))
				result.RawSetString("recorded_commands", lua.LNumber(status.Commands))
				result.RawSetString("privileged_commands", lua.LNumber(status.Privileged))
				result.RawSetString("checkpoints", lua.LNumber(status.Checkpoints))
				state.Push(result)
				return 1
			},
			"replay": func(state *lua.LState) int {
				replay, err := session.Replay()
				if err != nil {
					state.RaiseError("export session replay: %v", err)
					return 0
				}
				encoded, err := json.Marshal(replay)
				if err != nil {
					state.RaiseError("encode session replay: %v", err)
					return 0
				}
				state.Push(lua.LString(encoded))
				return 1
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}
