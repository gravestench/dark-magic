package modruntime

import (
	"context"
	"strings"
	"testing"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	lua "github.com/yuin/gopher-lua"
)

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
		return state.DoString(`local s=require("dm.session/v1"); local v=s.status(); tick=v.tick; replay=s.replay()`)
	}); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		if state.GetGlobal("tick") != lua.LNumber(1) {
			t.Fatalf("tick = %s", state.GetGlobal("tick"))
		}
		if replay := state.GetGlobal("replay").String(); !strings.Contains(replay, `"version":1`) || !strings.Contains(replay, `"tick":1`) {
			t.Fatalf("replay = %s", replay)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
