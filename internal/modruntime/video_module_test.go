package modruntime

import (
	"context"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/videocore"
	lua "github.com/yuin/gopher-lua"
)

type testVideoBackend struct{ playback *testPlayback }

func (b *testVideoBackend) Available() bool { return true }
func (b *testVideoBackend) Play(_ fs.FS, path string) (videocore.Playback, error) {
	b.playback = &testPlayback{snapshot: videocore.Snapshot{State: videocore.Playing}}
	return b.playback, nil
}

type testPlayback struct {
	snapshot videocore.Snapshot
	stops    int
}

func (p *testPlayback) Snapshot() videocore.Snapshot { return p.snapshot }
func (p *testPlayback) Stop() error {
	p.stops++
	p.snapshot.State = videocore.Stopped
	return nil
}

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
	defer runtime.Stop(context.Background())
	scope := &Scope{}
	if err := runtime.runScoped(context.Background(), scope, func(state *lua.LState) error {
		return state.DoString(`local video=require("dm.video/v1"); assert(video.available()); local p=video.play("intro.bik"); assert(p:status().state=="playing")`)
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
