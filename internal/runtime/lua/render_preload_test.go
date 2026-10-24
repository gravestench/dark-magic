package modruntime

import (
	"io/fs"
	"testing"
	"time"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
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

// TestPreloaderReleasePendingJobAfterWorkerCompletes forces release to win
// the lock order before a worker can load or publish completion.
func TestPreloaderReleasePendingJobAfterWorkerCompletes(t *testing.T) {
	p := newAssetPreloader(nil, nil, nil)
	for range cap(p.work) {
		p.work <- struct{}{}
	}
	unblock := func() {
		for range cap(p.work) {
			<-p.work
		}
	}
	defer func() {
		if unblock != nil {
			unblock()
		}
	}()
	id := p.Start([]AssetPreloadRequest{{Kind: "unsupported"}})
	if got := p.Pending(); got != 1 {
		t.Fatalf("pending = %d, want 1", got)
	}
	assertPreloadRelease(t, p, id)
	if got := p.Pending(); got != 1 {
		t.Fatalf("released pending = %d, want 1", got)
	}
	assertPreloadJobCount(t, p, 1)
	unblock()
	unblock = nil
	waitPreloadPending(t, p, 0)
	assertPreloadJobCount(t, p, 0)
}

func TestPreloaderReleaseCompletedEmptyAndUnknownJobs(t *testing.T) {
	p := newAssetPreloader(nil, nil, nil)
	if p.Forget(99) {
		t.Fatal("released unknown job")
	}
	id := p.Start([]AssetPreloadRequest{{Kind: "unsupported"}})
	waitPreloadPending(t, p, 0)
	status, ok := p.Status(id)
	if !ok || !status.Done || status.Completed != 1 || status.Failed != 1 {
		t.Fatalf("completed status = %#v, present = %v", status, ok)
	}
	assertPreloadRelease(t, p, id)
	assertPreloadJobCount(t, p, 0)
	empty := p.Start(nil)
	assertPreloadRelease(t, p, empty)
	assertPreloadJobCount(t, p, 0)
}

// A single worker and gated asset reads keep the final request unfinished
// after the first completion has been published.
func TestPreloaderReleaseRetainsJobUntilFinalCompletion(t *testing.T) {
	assets := &preloadReleaseFS{opened: make(chan struct{}, 2), proceed: make(chan struct{})}
	defer close(assets.proceed)
	p := NewRenderCapability(nil, nil, assets).preloads
	p.work = make(chan struct{}, 1)
	id := p.Start([]AssetPreloadRequest{
		{Kind: "image", Path: "first.png"}, {Kind: "image", Path: "last.png"},
	})
	waitPreloadOpen(t, assets.opened)
	assertPreloadRelease(t, p, id)
	if got := p.Pending(); got != 2 {
		t.Fatalf("pending = %d, want 2", got)
	}
	assets.proceed <- struct{}{}
	waitPreloadOpen(t, assets.opened)
	if got := p.Pending(); got != 1 {
		t.Fatalf("pending after first completion = %d, want 1", got)
	}
	assertPreloadJobCount(t, p, 1)
	if _, ok := p.Status(id); ok {
		t.Fatal("released job regained status")
	}
	assets.proceed <- struct{}{}
	waitPreloadPending(t, p, 0)
	assertPreloadJobCount(t, p, 0)
}

type preloadReleaseFS struct {
	opened  chan struct{}
	proceed chan struct{}
}

func (f *preloadReleaseFS) Open(string) (fs.File, error) {
	f.opened <- struct{}{}
	<-f.proceed
	return nil, fs.ErrNotExist
}

func waitPreloadOpen(t *testing.T, opened <-chan struct{}) {
	t.Helper()
	select {
	case <-opened:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for asset read")
	}
}

func waitPreloadPending(t *testing.T, p *assetPreloader, want int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for p.Pending() != want {
		if time.Now().After(deadline) {
			t.Fatalf("pending = %d, want %d", p.Pending(), want)
		}
		time.Sleep(time.Millisecond)
	}
}

func assertPreloadJobCount(t *testing.T, p *assetPreloader, want int) {
	t.Helper()
	p.mu.Lock()
	count := len(p.jobs)
	p.mu.Unlock()
	if count != want {
		t.Fatalf("retained jobs = %d, want %d", count, want)
	}
}

func assertPreloadRelease(t *testing.T, p *assetPreloader, id uint64) {
	t.Helper()
	if !p.Forget(id) {
		t.Fatal("first release rejected")
	}
	if p.Forget(id) {
		t.Fatal("duplicate release accepted")
	}
	if _, ok := p.Status(id); ok {
		t.Fatal("released job still has status")
	}
}
