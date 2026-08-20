package modruntime

import (
	"context"
	"strings"
	"testing"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	lua "github.com/yuin/gopher-lua"
)

// TestSessionModuleReportsStatusAndReplay verifies that Lua observes the stepped tick and receives serialized replay
// and audit snapshots from the session.
func TestSessionModuleReportsStatusAndReplay(t *testing.T) {
	engine := gameecs.New()

	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = session.Close() })

	runtime := New()
	if err := runtime.RegisterModule(SessionModule(session)); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = runtime.Stop(context.Background()) })

	if err := session.Step(); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		return state.DoString(
			`local s=require("engine.session/v1"); local v=s.status(); tick=v.tick; ` +
				`privileged=v.privileged_commands; replay=s.replay(); audit=s.audit()`,
		)
	}); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		if state.GetGlobal("tick") != lua.LNumber(1) {
			t.Fatalf("tick = %s", state.GetGlobal("tick"))
		}

		if state.GetGlobal("privileged") != lua.LNumber(0) ||
			state.GetGlobal("audit") != lua.LString("[]") {
			t.Fatalf("audit = %s / %s", state.GetGlobal("privileged"), state.GetGlobal("audit"))
		}

		if replay := state.GetGlobal("replay").
			String(); !strings.Contains(replay, `"version":1`) ||
			!strings.Contains(replay, `"tick":1`) {
			t.Fatalf("replay = %s", replay)
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
