package acceptance

import (
	"context"
	"io/fs"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/game/data/catalog"
	"github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/localization"
	"github.com/gravestench/dark-magic/internal/persistence"
	"github.com/gravestench/dark-magic/internal/preferences"
	"github.com/gravestench/dark-magic/internal/presentation/navigation"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	"github.com/gravestench/dark-magic/internal/presentation/scene"
	"github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/video"
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

func (b *startupVideoBackend) Play(_ fs.FS, path string) (video.Playback, error) {
	playback := &startupPlayback{snapshot: video.Snapshot{State: video.Playing}}
	b.paths = append(b.paths, path)
	b.playbacks = append(b.playbacks, playback)
	return playback, nil
}

type startupPlayback struct {
	snapshot video.Snapshot
	stops    int
}

func (p *startupPlayback) Snapshot() video.Snapshot { return p.snapshot }

func (p *startupPlayback) Stop() error {
	p.stops++
	p.snapshot.State = video.Stopped
	return nil
}

func TestStartupVideoSequenceCompletionFailureAndSkip(t *testing.T) {
	t.Run("failed movie follows skip policy and sequence continues", func(t *testing.T) {
		harness := newStartupHarness(t)
		harness.backend.playbacks[0].snapshot = video.Snapshot{State: video.Failed, Error: "decode failed"}
		harness.update(t)
		harness.assertPaths(t, blizzardMovie, blizzardNorthMovie)
		if harness.backend.playbacks[0].stops != 1 {
			t.Fatalf("failed playback stop calls = %d", harness.backend.playbacks[0].stops)
		}

		harness.backend.playbacks[1].snapshot = video.Snapshot{State: video.Complete}
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

	t.Run("class selection walks forward before naming and switches cleanly", func(t *testing.T) {
		harness := newStartupHarness(t)
		harness.skip(t)
		harness.skip(t)
		harness.skip(t) // Leave the trademark scene for the main menu.
		assertStack(t, harness.navigator, "main_menu")

		harness.action(t, "confirm")
		assertStack(t, harness.navigator, "character_select")
		harness.update(t) // Empty save list redirects into character creation.
		assertStack(t, harness.navigator, "character_create")

		harness.action(t, "confirm") // Start the initially focused Amazon walking forward.
		assertStack(t, harness.navigator, "character_create")
		harness.updateFor(t, 3*time.Second)
		assertStack(t, harness.navigator, "character_create")
		harness.updateFor(t, time.Second) // Forward walk completes and opens the dialog.

		harness.action(t, "cancel") // Keep the Amazon, then choose another class.
		harness.action(t, "right")
		harness.action(t, "confirm") // Amazon walks back while Sorceress walks forward.
		harness.updateFor(t, 3*time.Second)
		assertStack(t, harness.navigator, "character_create")
		harness.updateFor(t, time.Second) // Sorceress reaches her selected pose.

		harness.input.Publish(inputstate.Frame{Text: "Hero"})
		harness.update(t)
		harness.input.Publish(inputstate.Frame{})
		harness.action(t, "down")    // Move dialog focus from the field to OK.
		harness.action(t, "confirm") // Accept after the authored walk is complete.
		assertStack(t, harness.navigator, "game_loading")
		selected, ok := harness.saves.Selected()
		if !ok || selected.Name != "Hero" || selected.Class != "Sorceress" {
			t.Fatalf("created selection = %#v, selected=%v", selected, ok)
		}
	})

	t.Run("character deletion is confirmed before leaving the list", func(t *testing.T) {
		harness := newStartupHarnessWithSaves(t, persistence.Character{
			ID: "hero", Name: "Hero", Class: "Amazon", Level: 1,
		})
		harness.skip(t)
		harness.skip(t)
		harness.skip(t)
		harness.action(t, "confirm")
		assertStack(t, harness.navigator, "character_select")

		// Slot -> scrollbar -> New -> Delete.
		for range 3 {
			harness.action(t, "down")
		}
		harness.action(t, "confirm")
		assertStack(t, harness.navigator, "character_select")
		harness.action(t, "confirm") // Confirm the focused Yes action.
		assertStack(t, harness.navigator, "character_create")
	})
}

type startupHarness struct {
	runtime   *modruntime.Runtime
	scenes    *modruntime.Scenes
	navigator *navigation.Manager
	input     *inputstate.Store
	backend   *startupVideoBackend
	saves     *persistence.Store
}

func newStartupHarness(t *testing.T) *startupHarness {
	return newStartupHarnessWithSaves(t)
}

func newStartupHarnessWithSaves(t *testing.T, entries ...persistence.Character) *startupHarness {
	t.Helper()
	ctx := context.Background()
	videos := fstest.MapFS{
		blizzardMovie:      {Data: []byte("BIK")},
		blizzardNorthMovie: {Data: []byte("BIK")},
	}
	contentFS, err := content.New(
		content.Layer{Name: "videos", FS: videos},
		content.Layer{Name: "darkmagic", FS: content.D2Legacy()},
	)
	if err != nil {
		t.Fatal(err)
	}
	runtime := modruntime.New()
	navigator := navigation.New()
	scenes := modruntime.NewScenes(runtime, navigator)
	backend := &startupVideoBackend{}
	var input inputstate.Store
	var composer render.Composer
	var mixer audio.Mixer
	simulation := modruntime.NewSimulation(scene.New(1, 100, 100))
	loading := acceptanceLoadingCoordinator()
	saves := persistence.New(entries...)
	t.Cleanup(loading.Close)
	if err := runtime.RegisterInstaller(modruntime.ContentRequire(contentFS, "lua")); err != nil {
		t.Fatal(err)
	}
	for _, module := range []modruntime.Module{
		modruntime.AppModule("test", func() {}),
		modruntime.VFSModule(contentFS),
		modruntime.DataModule(contentFS),
		modruntime.InputModule(&input),
		modruntime.AudioModule(runtime, &mixer, contentFS, gamedata.New(recordstore.New(contentFS))),
		modruntime.SettingsModule(preferences.NewTransient(), &mixer),
		modruntime.VideoModule(runtime, backend, contentFS),
		modruntime.LocaleModule(localization.New(contentFS, "English")),
		modruntime.RenderModule(runtime, &composer),
		modruntime.SaveModule(saves),
		modruntime.SimulationModule(simulation),
		modruntime.LoadingModule(loading),
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
	harness := &startupHarness{runtime: runtime, scenes: scenes, navigator: navigator, input: &input, backend: backend, saves: saves}
	assertStack(t, navigator, "loading")
	harness.assertPaths(t, blizzardMovie)
	return harness
}

func (h *startupHarness) update(t *testing.T) {
	t.Helper()
	h.updateFor(t, time.Second/60)
}

func (h *startupHarness) updateFor(t *testing.T, elapsed time.Duration) {
	t.Helper()
	if err := h.scenes.Update(context.Background(), elapsed); err != nil {
		t.Fatal(err)
	}
}

func (h *startupHarness) action(t *testing.T, name string) {
	t.Helper()
	publishAction(h.input, name)
	h.update(t)
	h.input.Publish(inputstate.Frame{})
}

func (h *startupHarness) skip(t *testing.T) {
	t.Helper()
	publishAction(h.input, "skip")
	h.update(t)
	h.input.Publish(inputstate.Frame{})
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
