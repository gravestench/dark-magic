package acceptance

import (
	"context"
	"testing"

	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/inputstate"
	d2presentation "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/presentation"
	"github.com/gravestench/dark-magic/internal/preferences"
	"github.com/gravestench/dark-magic/internal/presentation/navigation"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// TestLuaEscapeMenuRecoveredNavigation protects the recovered menu's real capability composition.
func TestLuaEscapeMenuRecoveredNavigation(t *testing.T) {
	ctx := context.Background()

	contentFS, err := content.New(content.Layer{Name: "darkmagic", FS: content.D2Legacy()})
	if err != nil {
		t.Fatal(err)
	}

	runtime := modruntime.New()

	var (
		input    inputstate.Store
		mixer    audio.Mixer
		composer render.Composer
	)

	if err := runtime.RegisterInstaller(modruntime.ContentRequire(contentFS, "lua")); err != nil {
		t.Fatal(err)
	}

	for _, module := range []modruntime.Module{
		modruntime.InputModule(&input),
		modruntime.DataModule(contentFS),
		modruntime.RenderModule(runtime, &composer),
		modruntime.AudioModule(runtime, &mixer, contentFS),
		modruntime.SettingsModule(preferences.NewTransient(), &mixer),
	} {
		if err := runtime.RegisterModule(module); err != nil {
			t.Fatal(err)
		}
	}

	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := runtime.Stop(ctx); err != nil {
			t.Errorf("stop escape-menu runtime: %v", err)
		}
	}()

	scope := &modruntime.Scope{}

	defer func() {
		if err := scope.Close(); err != nil {
			t.Errorf("close escape-menu scope: %v", err)
		}
	}()

	const escapeMenuScript = "lua/d2legacy/tests/integration/ui_escape_menu.lua"
	if err := runtime.ExecuteScoped(ctx, scope, content.D2Legacy(), escapeMenuScript); err != nil {
		t.Fatal(err)
	}
}

// TestLuaOptionsOverlayKeepsAuthoredCoordinatesViewportRelative protects the modern viewport default.
func TestLuaOptionsOverlayKeepsAuthoredCoordinatesViewportRelative(t *testing.T) {
	assertLuaOptionsBackdropCenter(t, "", 400, 300)
}

// TestLuaOptionsOverlayCentersInClassicViewport protects manifest-transformed classic coordinates.
func TestLuaOptionsOverlayCentersInClassicViewport(t *testing.T) {
	assertLuaOptionsBackdropCenter(t, "lod-english-640x480-gameplay", 320, 240)
}

// assertLuaOptionsBackdropCenter assembles the real overlay and checks the transformed backdrop position.
func assertLuaOptionsBackdropCenter(t *testing.T, profile string, wantX, wantY float64) {
	t.Helper()

	ctx := context.Background()

	contentFS, err := content.New(content.Layer{Name: "darkmagic", FS: content.D2Legacy()})
	if err != nil {
		t.Fatal(err)
	}

	runtime := modruntime.New()

	var (
		input    inputstate.Store
		mixer    audio.Mixer
		composer render.Composer
	)

	if err := runtime.RegisterInstaller(modruntime.ContentRequire(contentFS, "lua")); err != nil {
		t.Fatal(err)
	}

	scenes := modruntime.NewScenes(runtime, navigation.New())
	for _, module := range []modruntime.Module{
		modruntime.InputModule(&input),
		modruntime.DataModule(contentFS, d2presentation.ManifestTransforms(profile)),
		modruntime.RenderModule(runtime, &composer),
		modruntime.AudioModule(runtime, &mixer, contentFS),
		modruntime.SettingsModule(preferences.NewTransient(), &mixer),
		scenes.Module(),
	} {
		if err := runtime.RegisterModule(module); err != nil {
			t.Fatal(err)
		}
	}

	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := runtime.Stop(ctx); err != nil {
			t.Errorf("stop options-overlay runtime: %v", err)
		}
	}()

	scope := &modruntime.Scope{}

	defer func() {
		if err := scope.Close(); err != nil {
			t.Errorf("close options-overlay scope: %v", err)
		}
	}()

	const overlayScript = "lua/d2legacy/tests/integration/options_overlay_create.lua"
	if err := runtime.ExecuteScoped(ctx, scope, content.D2Legacy(), overlayScript); err != nil {
		t.Fatal(err)
	}

	nodes := composer.Snapshot()
	if len(nodes) < 2 {
		t.Fatalf("render nodes = %#v", nodes)
	}

	root := nodes[0]
	if root.X != 0 || root.Y != 0 {
		t.Fatalf("overlay root position = (%v, %v), want viewport origin", root.X, root.Y)
	}

	foundBackdrop := false

	for _, node := range nodes[1:] {
		if node.Parent == root.ID && node.X == wantX && node.Y == wantY {
			foundBackdrop = true
			break
		}
	}

	if !foundBackdrop {
		t.Fatalf("centered backdrop child not found below root %#v in nodes %#v", root.ID, nodes)
	}
}
