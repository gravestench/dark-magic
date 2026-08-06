package acceptance

import (
	"context"
	"reflect"
	"testing"
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
	"github.com/gravestench/dark-magic/pkg/scene"
)

func TestEmbeddedShimNavigationAndResourceLifetime(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	contentFS, err := content.New(content.Layer{Name: "darkmagic", FS: content.Shim()})
	if err != nil {
		t.Fatal(err)
	}
	runtime := modruntime.New()
	navigator := navigation.New()
	scenes := modruntime.NewScenes(runtime, navigator)
	var composer rendercore.Composer
	var input inputcore.Store
	var mixer audiocore.Mixer
	saves := savecore.New(savecore.Character{ID: "hero", Name: "Hero", Class: "Amazon", Level: 1})
	simulation := modruntime.NewSimulation(scene.New(11, 1000, 1000))
	if err := runtime.RegisterInstaller(modruntime.ContentRequire(contentFS, "lua")); err != nil {
		t.Fatal(err)
	}
	for _, module := range []modruntime.Module{
		modruntime.VFSModule(contentFS),
		modruntime.DataModule(contentFS),
		modruntime.InputModule(&input),
		modruntime.AudioModule(runtime, &mixer, contentFS),
		modruntime.LocaleModule(localecore.New(contentFS, "English")),
		modruntime.RenderModule(runtime, &composer),
		modruntime.SaveModule(saves),
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
	defer runtime.Stop(ctx)
	components := host.NewManager()
	boot, err := modruntime.LoadDefinition(ctx, runtime, contentFS, "boot.lua")
	if err != nil {
		t.Fatal(err)
	}
	if err := components.Register(boot.Managed()); err != nil {
		t.Fatal(err)
	}
	if err := components.Enable(ctx, boot.ID); err != nil {
		t.Fatal(err)
	}
	if err := scenes.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	assertStack(t, navigator, "loading")
	assertNodes(t, &composer, 2)

	publishAction(&input, "confirm")
	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}
	assertStack(t, navigator, "title")
	assertNodes(t, &composer, 2)

	input.Publish(inputcore.Frame{})
	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}
	publishAction(&input, "confirm")
	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}
	assertStack(t, navigator, "main_menu")
	assertNodes(t, &composer, 2)
	input.Publish(inputcore.Frame{})
	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}
	publishAction(&input, "down")
	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}
	input.Publish(inputcore.Frame{})
	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}
	publishAction(&input, "confirm")
	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}
	assertStack(t, navigator, "tcpip")
	publishAction(&input, "cancel")
	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}
	assertStack(t, navigator, "main_menu")

	input.Publish(inputcore.Frame{})
	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}
	publishAction(&input, "confirm")
	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}
	assertStack(t, navigator, "character_select")
	assertNodes(t, &composer, 2)

	input.Publish(inputcore.Frame{})
	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}
	publishAction(&input, "confirm")
	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}
	assertStack(t, navigator, "game_world")
	assertNodes(t, &composer, 3)
	if selected, ok := saves.Selected(); !ok || selected.ID != "hero" {
		t.Fatalf("selected character = %#v, %v", selected, ok)
	}

	for _, overlay := range []string{"inventory", "character", "skills", "automap", "options", "pause"} {
		input.Publish(inputcore.Frame{})
		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}
		publishAction(&input, overlay)
		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}
		assertStack(t, navigator, "game_world", overlay)
		assertNodes(t, &composer, 4)
		input.Publish(inputcore.Frame{})
		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}
		publishAction(&input, "cancel")
		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}
		assertStack(t, navigator, "game_world")
		assertNodes(t, &composer, 3)
	}
	for cycle := 0; cycle < 50; cycle++ {
		input.Publish(inputcore.Frame{})
		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}
		publishAction(&input, "inventory")
		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}
		input.Publish(inputcore.Frame{})
		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}
		publishAction(&input, "cancel")
		if err := scenes.Update(ctx, time.Second/60); err != nil {
			t.Fatal(err)
		}
	}
	if diagnostics := composer.Diagnostics(); diagnostics.ActiveNodes != 3 || diagnostics.NodeSlots > 4 {
		t.Fatalf("composer diagnostics after rapid transitions = %#v", diagnostics)
	}

	before := simulation.Snapshot().Hero.X
	input.Publish(inputcore.Frame{Actions: map[string]inputcore.ActionState{"right": {Down: true}}})
	if err := scenes.Update(ctx, time.Second); err != nil {
		t.Fatal(err)
	}
	if after := simulation.Snapshot().Hero.X; after <= before {
		t.Fatalf("hero did not move: %v -> %v", before, after)
	}

	if err := scenes.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if err := components.Disable(ctx, boot.ID); err != nil {
		t.Fatal(err)
	}
	assertNodes(t, &composer, 0)
	if diagnostics := composer.Diagnostics(); diagnostics.ActiveNodes != 0 || diagnostics.ActiveResources != 0 {
		t.Fatalf("composer leaked resources: %#v", diagnostics)
	}
	if diagnostics := mixer.Diagnostics(); diagnostics.Active != 0 {
		t.Fatalf("mixer leaked sounds: %#v", diagnostics)
	}
}

func publishAction(input *inputcore.Store, name string) {
	input.Publish(inputcore.Frame{Actions: map[string]inputcore.ActionState{name: {Pressed: true}}})
}

func assertStack(t *testing.T, navigator *navigation.Manager, want ...string) {
	t.Helper()
	if got := navigator.Stack(); !reflect.DeepEqual(got, want) {
		t.Fatalf("stack = %v, want %v", got, want)
	}
}

func assertNodes(t *testing.T, composer *rendercore.Composer, want int) {
	t.Helper()
	if got := len(composer.Snapshot()); got != want {
		t.Fatalf("render node count = %d, want %d; nodes=%#v", got, want, composer.Snapshot())
	}
}
