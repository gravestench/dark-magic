package acceptance

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/host"
	"github.com/gravestench/dark-magic/internal/inputcore"
	"github.com/gravestench/dark-magic/internal/modruntime"
	"github.com/gravestench/dark-magic/internal/navigation"
	"github.com/gravestench/dark-magic/internal/rendercore"
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
	if err := runtime.RegisterInstaller(modruntime.ContentRequire(contentFS, "lua")); err != nil {
		t.Fatal(err)
	}
	for _, module := range []modruntime.Module{
		modruntime.VFSModule(contentFS),
		modruntime.InputModule(&input),
		modruntime.RenderModule(runtime, &composer),
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
	assertStack(t, navigator, "main_menu")
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
	assertNodes(t, &composer, 2)

	input.Publish(inputcore.Frame{})
	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}
	publishAction(&input, "inventory")
	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}
	assertStack(t, navigator, "game_world", "inventory")
	assertNodes(t, &composer, 3)

	input.Publish(inputcore.Frame{})
	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}
	publishAction(&input, "cancel")
	if err := scenes.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}
	assertStack(t, navigator, "game_world")
	assertNodes(t, &composer, 2)

	if err := scenes.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if err := components.Disable(ctx, boot.ID); err != nil {
		t.Fatal(err)
	}
	assertNodes(t, &composer, 0)
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
