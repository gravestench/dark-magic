package modruntime

import (
	"testing"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
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

func TestLuaPreloadRequestsPreservesAuthoritativeWorld(t *testing.T) {
	state := lua.NewState()
	defer state.Close()
	world := &gameworld.Map{WidthTiles: 80, HeightTiles: 80}
	userData := state.NewUserData()
	userData.Value = world
	definition := state.NewTable()
	definition.RawSetString("kind", lua.LString("world_chunks"))
	definition.RawSetString("palette", lua.LString("act1.pl2"))
	definition.RawSetString("world", userData)
	table := state.NewTable()
	table.Append(definition)
	state.Push(table)
	requests, err := luaPreloadRequests(state, state.GetTop())
	if err != nil {
		t.Fatal(err)
	}
	if len(requests) != 1 || requests[0].World != world {
		t.Fatalf("world chunk request = %#v", requests)
	}
}

func TestLuaPreloadRequestsPreservesChunkGeometry(t *testing.T) {
	state := lua.NewState()
	defer state.Close()
	if err := state.DoString(`request = {{
		kind = "ds1_chunks", path = "map.ds1", palette = "act1.pl2",
		tiles = {"floor.dt1", "walls.dt1"}, chunk_size = 256,
	}}`); err != nil {
		t.Fatal(err)
	}
	table := state.GetGlobal("request").(*lua.LTable)
	state.Push(table)
	requests, err := luaPreloadRequests(state, state.GetTop())
	if err != nil {
		t.Fatal(err)
	}
	request := requests[0]
	if request.ChunkSize != 256 || len(request.Tiles) != 2 || request.Tiles[1] != "walls.dt1" {
		t.Fatalf("chunk request = %#v", request)
	}
}
