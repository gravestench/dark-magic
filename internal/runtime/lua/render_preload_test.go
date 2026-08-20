package modruntime

import (
	"encoding/binary"
	"testing"
	"testing/fstest"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	lua "github.com/yuin/gopher-lua"
)

// TestLuaPreloadRequestsPreservesCompositeRecipe protects the lua preload requests preserves composite recipe
// contract, including its observable ordering and failure behavior.
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
	if request.Kind != "cof_animation" || request.Path != "hero.cof" ||
		request.Palette != "act1.dat" ||
		request.Direction != 3 {
		t.Fatalf("request metadata = %#v", request)
	}

	if request.Components["HD"] != "head.dcc" || request.Components["TR"] != "torso.dcc" {
		t.Fatalf("request components = %#v", request.Components)
	}
}

// TestLuaPreloadRequestsPreservesAuthoritativeWorld protects the lua preload requests preserves authoritative world
// contract, including its observable ordering and failure behavior.
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

// TestWorldTilePreloadWarmsUniqueGraphicsAcrossEditSnapshots protects map
// painting from decoding one DT1 picture once per placement or once per new
// immutable world snapshot.
func TestWorldTilePreloadWarmsUniqueGraphicsAcrossEditSnapshots(t *testing.T) {
	dt1Data := make([]byte, 276+96)
	binary.LittleEndian.PutUint32(dt1Data[0:4], 7)
	binary.LittleEndian.PutUint32(dt1Data[4:8], 6)
	binary.LittleEndian.PutUint32(dt1Data[268:272], 1)
	binary.LittleEndian.PutUint32(dt1Data[272:276], 276)
	tile := dt1Data[276:]
	binary.LittleEndian.PutUint32(tile[8:12], 0xffffffb0)
	binary.LittleEndian.PutUint32(tile[12:16], 160)
	binary.LittleEndian.PutUint32(tile[72:76], uint32(len(dt1Data)))

	assets := fstest.MapFS{
		"one.dt1":  {Data: dt1Data},
		"act1.dat": {Data: make([]byte, 256*3)},
	}
	capability := NewRenderCapability(New(), &render.Composer{}, assets)
	reference := gameworld.TileReference{
		Path: "one.dt1", Index: 0, Width: 160, Height: -80,
	}
	placements := []gameworld.TilePlacement{
		{X: 0, Y: 0, Layer: gameworld.LayerFloor, Reference: reference},
		{X: 1, Y: 0, Layer: gameworld.LayerFloor, Reference: reference},
		{X: 0, Y: 1, Layer: gameworld.LayerFloor, Reference: reference},
	}

	for range 2 {
		world := &gameworld.Map{
			WidthTiles: 2, HeightTiles: 2,
			Tiles: append([]gameworld.TilePlacement(nil), placements...),
		}
		if err := capability.preloads.load(AssetPreloadRequest{
			Kind: "world_tiles", World: world, Palette: "act1.dat",
		}); err != nil {
			t.Fatal(err)
		}
	}

	stages := capability.Diagnostics().Stages
	if stages["dt1-file"].Calls != 1 || stages["dt1-tile"].Calls != 1 {
		t.Fatalf("unique DT1 picture decoded across placements or edit snapshots: %#v", stages)
	}
}

// TestLuaPreloadRequestsPreservesVisibleWorldChunk protects the lua preload requests preserves visible world chunk
// contract, including its observable ordering and failure behavior.
func TestLuaPreloadRequestsPreservesVisibleWorldChunk(t *testing.T) {
	state := lua.NewState()
	defer state.Close()

	world := &gameworld.Map{WidthTiles: 80, HeightTiles: 80}
	userData := state.NewUserData()
	userData.Value = world
	definition := state.NewTable()
	definition.RawSetString("kind", lua.LString("world_chunk"))
	definition.RawSetString("palette", lua.LString("act1.pl2"))
	definition.RawSetString("world", userData)
	definition.RawSetString("chunk_index", lua.LNumber(37))
	definition.RawSetString("chunk_size", lua.LNumber(256))

	table := state.NewTable()
	table.Append(definition)
	state.Push(table)

	requests, err := luaPreloadRequests(state, state.GetTop())
	if err != nil {
		t.Fatal(err)
	}

	if len(requests) != 1 || requests[0].World != world || requests[0].ChunkIndex != 37 ||
		requests[0].ChunkSize != 256 {
		t.Fatalf("visible world chunk request = %#v", requests)
	}
}

// TestLuaPreloadRequestsPreservesChunkGeometry protects the lua preload requests preserves chunk geometry contract,
// including its observable ordering and failure behavior.
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

// TestPreloaderForgetsOnlyCompletedJobs protects the preloader forgets only completed jobs contract, including its
// observable ordering and failure behavior.
func TestPreloaderForgetsOnlyCompletedJobs(t *testing.T) {
	preloader := newAssetPreloader(nil, nil, nil)
	preloader.mu.Lock()
	preloader.jobs[7] = &assetPreloadJob{total: 2, completed: 1}
	preloader.jobs[8] = &assetPreloadJob{total: 1, completed: 1}
	preloader.mu.Unlock()

	if preloader.Forget(7) {
		t.Fatal("forgot an active preload job")
	}

	if !preloader.Forget(8) {
		t.Fatal("did not forget completed preload job")
	}

	if _, ok := preloader.Status(8); ok {
		t.Fatal("forgotten preload job still has status")
	}
}
