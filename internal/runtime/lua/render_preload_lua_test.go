package modruntime

import (
	"context"
	"io/fs"
	"sync"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	lua "github.com/yuin/gopher-lua"
)

// Exercise the first-party adapter through the real capability while native
// work cannot finish, then prove released records disappear on completion.
func TestChunkedMapDestroyReleasesBlockedNativePreload(t *testing.T) {
	assets := &chunkedMapPreloadFS{}
	runtime := New()
	capability := NewRenderCapability(runtime, &render.Composer{}, assets)
	if err := runtime.RegisterInstaller(ContentRequire(content.D2Legacy(), "lua")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RegisterModule(capability.Module()); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = runtime.Stop(context.Background()) })
	scope := &Scope{}
	t.Cleanup(func() { _ = scope.Close() })
	run := func(code string) {
		t.Helper()
		if err := runtime.RunScoped(context.Background(), scope, func(state *lua.LState) error { return state.DoString(code) }); err != nil {
			t.Fatal(err)
		}
	}
	p := capability.preloads
	for cycle := 0; cycle < 2; cycle++ {
		gate := &chunkedMapPreloadGate{opened: make(chan struct{}), proceed: make(chan struct{})}
		t.Cleanup(gate.unblock)
		assets.mu.Lock()
		assets.gate = gate
		assets.mu.Unlock()
		run(`chunks = require("d2legacy.presentation.chunked_map")
adapter = chunks.create(nil, {ds1="map.ds1", dt1={}, palette=""})
job_id = adapter.job`)
		waitPreloadOpen(t, gate.opened)
		var id uint64
		if err := runtime.RunScoped(context.Background(), scope, func(state *lua.LState) error {
			id = uint64(state.GetGlobal("job_id").(lua.LNumber))
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		assertPreloadJobCount(t, p, 1)
		run(`chunks.destroy(adapter)
chunks.destroy(adapter)
assert(adapter.job == nil and next(adapter.pending) == nil)`)
		if _, ok := p.Status(id); ok {
			t.Fatal("destroyed adapter job still has native status")
		}
		if got := p.Pending(); got != 1 {
			t.Fatalf("blocked pending = %d, want 1", got)
		}
		assertPreloadJobCount(t, p, 1)
		gate.unblock()
		deadline := time.Now().Add(5 * time.Second)
		for {
			p.mu.Lock()
			count := len(p.jobs)
			p.mu.Unlock()
			if count == 0 && p.Pending() == 0 {
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("cycle %d: retained jobs = %d, pending = %d", cycle, count, p.Pending())
			}
			time.Sleep(time.Millisecond)
		}
		assertPreloadJobCount(t, p, 0)
	}
}

type chunkedMapPreloadGate struct {
	opened  chan struct{}
	proceed chan struct{}
	once    sync.Once
}

func (g *chunkedMapPreloadGate) unblock() { g.once.Do(func() { close(g.proceed) }) }

type chunkedMapPreloadFS struct {
	mu   sync.Mutex
	gate *chunkedMapPreloadGate
}

func (f *chunkedMapPreloadFS) Open(name string) (fs.File, error) {
	if name != "map.ds1" {
		return nil, fs.ErrNotExist
	}
	f.mu.Lock()
	gate := f.gate
	f.mu.Unlock()
	close(gate.opened)
	<-gate.proceed
	return nil, fs.ErrNotExist
}
