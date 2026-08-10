package modruntime

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestLuaPreloadRequestsPreservesCompositeRecipe(t *testing.T) {
	state := lua.NewState()
	defer state.Close()
	if err := state.DoString(`request = {{
		kind = "cof_animation",
		path = "hero.cof",
		palette = "act1.dat",
		direction = 3,
		components = {HD = "head.dcc", TR = "torso.dcc"},
	}}`); err != nil {
		t.Fatal(err)
	}
	table, ok := state.GetGlobal("request").(*lua.LTable)
	if !ok {
		t.Fatal("request is not a table")
	}
	state.Push(table)
	requests, err := luaPreloadRequests(state, state.GetTop())
	if err != nil {
		t.Fatal(err)
	}
	if len(requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(requests))
	}
	request := requests[0]
	if request.Kind != "cof_animation" || request.Path != "hero.cof" || request.Palette != "act1.dat" || request.Direction != 3 {
		t.Fatalf("request metadata = %#v", request)
	}
	if request.Components["HD"] != "head.dcc" || request.Components["TR"] != "torso.dcc" {
		t.Fatalf("request components = %#v", request.Components)
	}
}
