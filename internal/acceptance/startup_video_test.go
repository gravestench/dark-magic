package acceptance

import (
	"context"
	"io/fs"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gravestench/dark-magic/internal/audiocore"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/host"
	"github.com/gravestench/dark-magic/internal/inputcore"
	"github.com/gravestench/dark-magic/internal/localecore"
	"github.com/gravestench/dark-magic/internal/modruntime"
	"github.com/gravestench/dark-magic/internal/navigation"
	"github.com/gravestench/dark-magic/internal/rendercore"
	"github.com/gravestench/dark-magic/internal/savecore"
	"github.com/gravestench/dark-magic/internal/videocore"
	"github.com/gravestench/dark-magic/pkg/scene"
)

const (
	blizzardMovie      = "data/local/video/New_Bliz640x480.bik"
	blizzardNorthMovie = "data/local/video/BlizNorth640x480.bik"
)

type startupVideoBackend struct {
	paths     []string
	playbacks []*startupPlayback
}

func (*startupVideoBackend) Available() bool { return true }

func (b *startupVideoBackend) Play(_ fs.FS, path string) (videocore.Playback, error) {
	playback := &startupPlayback{snapshot: videocore.Snapshot{State: videocore.Playing}}
	b.paths = append(b.paths, path)
	b.playbacks = append(b.playbacks, playback)
	return playback, nil
}

type startupPlayback struct {
	snapshot videocore.Snapshot
	stops    int
}

func (p *startupPlayback) Snapshot() videocore.Snapshot { return p.snapshot }

func (p *startupPlayback) Stop() error {
	p.stops++
	p.snapshot.State = videocore.Stopped
	return nil
}

func TestStartupVideoSequenceCompletionFailureAndSkip(t *testing.T) {
	t.Run("failed movie follows skip policy and sequence continues", func(t *testing.T) {
		harness := newStartupHarness(t)
		harness.backend.playbacks[0].snapshot = videocore.Snapshot{State: videocore.Failed, Error: "decode failed"}
		harness.update(t)
		harness.assertPaths(t, blizzardMovie, blizzardNorthMovie)
		if harness.backend.playbacks[0].stops != 1 {
			t.Fatalf("failed playback stop calls = %d", harness.backend.playbacks[0].stops)
		}

		harness.backend.playbacks[1].snapshot = videocore.Snapshot{State: videocore.Complete}
		harness.update(t)
		assertStack(t, harness.navigator, "title")
		if harness.backend.playbacks[1].stops != 1 {
			t.Fatalf("completed playback stop calls = %d", harness.backend.playbacks[1].stops)
		}
	})

	t.Run("skip advances each movie without bypassing ordering", func(t *testing.T) {
		harness := newStartupHarness(t)
		harness.skip(t)
		harness.assertPaths(t, blizzardMovie, blizzardNorthMovie)
		assertStack(t, harness.navigator, "loading")

		harness.skip(t)
		assertStack(t, harness.navigator, "title")
		for index, playback := range harness.backend.playbacks {
			if playback.stops != 1 {
				t.Fatalf("playback %d stop calls = %d", index, playback.stops)
			}
		}
	})
}

type startupHarness struct {
	runtime   *modruntime.Runtime
	scenes    *modruntime.Scenes
	navigator *navigation.Manager
	input     *inputcore.Store
	backend   *startupVideoBackend
}

func newStartupHarness(t *testing.T) *startupHarness {
	t.Helper()
	ctx := context.Background()
	videos := fstest.MapFS{
		blizzardMovie:      {Data: []byte("BIK")},
		blizzardNorthMovie: {Data: []byte("BIK")},
	}
	contentFS, err := content.New(
		content.Layer{Name: "videos", FS: videos},
		content.Layer{Name: "darkmagic", FS: content.Shim()},
	)
	if err != nil {
		t.Fatal(err)
	}
	runtime := modruntime.New()
	navigator := navigation.New()
	scenes := modruntime.NewScenes(runtime, navigator)
	backend := &startupVideoBackend{}
	var input inputcore.Store
	var composer rendercore.Composer
	var mixer audiocore.Mixer
	simulation := modruntime.NewSimulation(scene.New(1, 100, 100))
	if err := runtime.RegisterInstaller(modruntime.ContentRequire(contentFS, "lua")); err != nil {
		t.Fatal(err)
	}
	for _, module := range []modruntime.Module{
		modruntime.AppModule("test", func() {}),
		modruntime.VFSModule(contentFS),
		modruntime.DataModule(contentFS),
		modruntime.InputModule(&input),
		modruntime.AudioModule(runtime, &mixer, contentFS),
		modruntime.VideoModule(runtime, backend, contentFS),
		modruntime.LocaleModule(localecore.New(contentFS, "English")),
		modruntime.RenderModule(runtime, &composer),
		modruntime.SaveModule(savecore.New()),
		modruntime.SimulationModule(simulation),
		scenes.Module(),
	} {
		if err := runtime.RegisterModule(module); err != nil {
			t.Fatal(err)
		}
	}
	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := runtime.Stop(ctx); err != nil {
			t.Errorf("stop runtime: %v", err)
		}
	})
	manager := host.NewManager()
	boot, err := modruntime.LoadDefinition(ctx, runtime, contentFS, "boot.lua")
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.Register(boot.Managed()); err != nil {
		t.Fatal(err)
	}
	if err := manager.Enable(ctx, boot.ID); err != nil {
		t.Fatal(err)
	}
	if err := scenes.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	harness := &startupHarness{runtime: runtime, scenes: scenes, navigator: navigator, input: &input, backend: backend}
	assertStack(t, navigator, "loading")
	harness.assertPaths(t, blizzardMovie)
	return harness
}

func (h *startupHarness) update(t *testing.T) {
	t.Helper()
	if err := h.scenes.Update(context.Background(), time.Second/60); err != nil {
		t.Fatal(err)
	}
}

func (h *startupHarness) skip(t *testing.T) {
	t.Helper()
	publishAction(h.input, "skip")
	h.update(t)
	h.input.Publish(inputcore.Frame{})
}

func (h *startupHarness) assertPaths(t *testing.T, paths ...string) {
	t.Helper()
	if len(h.backend.paths) != len(paths) {
		t.Fatalf("played paths = %v, want %v", h.backend.paths, paths)
	}
	for index := range paths {
		if h.backend.paths[index] != paths[index] {
			t.Fatalf("played paths = %v, want %v", h.backend.paths, paths)
		}
	}
}
