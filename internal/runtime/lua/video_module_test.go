package modruntime

import (
	"context"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/video"
	lua "github.com/yuin/gopher-lua"
)

type testVideoBackend struct{ playback *testPlayback }

// Available keeps this test backend selectable so playback exercises the scoped-handle path.
func (b *testVideoBackend) Available() bool { return true }

// Play records the created playback so the test can observe scope-driven cleanup.
func (b *testVideoBackend) Play(_ fs.FS, path string) (video.Playback, error) {
	b.playback = &testPlayback{snapshot: video.Snapshot{State: video.Playing}}
	return b.playback, nil
}

type testPlayback struct {
	snapshot video.Snapshot
	stops    int
}

// Snapshot reports the test playback's current state through the production backend contract.
func (p *testPlayback) Snapshot() video.Snapshot { return p.snapshot }

// Stop records cleanup and transitions the playback to the stopped state.
func (p *testPlayback) Stop() error {
	p.stops++
	p.snapshot.State = video.Stopped

	return nil
}

// TestVideoModulePlaybackIsScoped verifies that closing the active Lua scope stops its backend playback exactly
// once.
func TestVideoModulePlaybackIsScoped(t *testing.T) {
	runtime := New()
	backend := &testVideoBackend{}

	source := fstest.MapFS{"intro.bik": {Data: []byte("BIK")}}
	if err := runtime.RegisterModule(VideoModule(runtime, backend, source)); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(context.Background()) }()

	scope := &Scope{}
	if err := runtime.runScoped(context.Background(), scope, func(state *lua.LState) error {
		return state.DoString(
			`local video=require("engine.video/v1"); assert(video.available()); ` +
				`local p=video.play("intro.bik"); assert(p:status().state=="playing")`,
		)
	}); err != nil {
		t.Fatal(err)
	}

	if backend.playback == nil || backend.playback.stops != 0 {
		t.Fatalf("playback before close = %#v", backend.playback)
	}

	if err := scope.Close(); err != nil {
		t.Fatal(err)
	}

	if backend.playback.stops != 1 {
		t.Fatalf("stop calls = %d", backend.playback.stops)
	}
}
