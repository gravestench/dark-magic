package acceptance

import (
	"context"
	"testing"

	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/game/data/catalog"
	"github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/inputstate"
	d2presentation "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/presentation"
	"github.com/gravestench/dark-magic/internal/preferences"
	"github.com/gravestench/dark-magic/internal/presentation/navigation"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	glua "github.com/yuin/gopher-lua"
)

func TestLuaEscapeMenuRecoveredNavigation(t *testing.T) {
	ctx := context.Background()
	contentFS, err := content.New(content.Layer{Name: "darkmagic", FS: content.D2Legacy()})
	if err != nil {
		t.Fatal(err)
	}

	runtime := modruntime.New()
	var input inputstate.Store
	var mixer audio.Mixer
	var composer render.Composer
	if err := runtime.RegisterInstaller(modruntime.ContentRequire(contentFS, "lua")); err != nil {
		t.Fatal(err)
	}
	for _, module := range []modruntime.Module{
		modruntime.InputModule(&input),
		modruntime.DataModule(contentFS),
		modruntime.RenderModule(runtime, &composer),
		modruntime.AudioModule(runtime, &mixer, contentFS, gamedata.New(recordstore.New(contentFS))),
		modruntime.SettingsModule(preferences.NewTransient(), &mixer),
	} {
		if err := runtime.RegisterModule(module); err != nil {
			t.Fatal(err)
		}
	}
	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(ctx)

	script := `
local render = require("engine.render/v1")
local settings = require("engine.settings/v1")
local escape_menu = require("d2legacy.ui.escape_menu")

local function value(value)
  if value == nil then return "<nil>" end
  return tostring(value)
end

local function expect(actual, expected, label)
  assert(actual == expected, label .. ": got " .. value(actual) .. ", want " .. value(expected))
end

local function focus_id(menu)
  return menu.manager.focus and menu.manager.focus.id or "<nil>"
end

local root = render.create("modal")
local closed = false
local saved = false
local changed = nil
local menu = escape_menu.new(root, {
  start_layout = "main",
  on_close = function() closed = true end,
  on_save_exit = function() saved = true end,
  on_option_change = function(layout, id, selected)
    changed = layout .. ":" .. id .. "=" .. selected
  end,
})

expect(menu.current_layout, "main", "initial layout")
expect(focus_id(menu), "main:return_to_game", "initial focus")

-- OpenDiablo2 stops at the ends of the list rather than wrapping.
menu.manager:move_focus(1)
expect(focus_id(menu), "main:return_to_game", "focus stays at bottom")
menu.manager:move_focus(-1)
expect(focus_id(menu), "main:save_exit", "focus moves to save/exit")
menu.manager:move_focus(-1)
expect(focus_id(menu), "main:options", "focus moves to options")
menu.manager:move_focus(-1)
expect(focus_id(menu), "main:options", "focus stays at top")

menu.manager:activate(menu.manager.focus)
expect(menu.current_layout, "options", "options activation layout")
expect(focus_id(menu), "options:previous_menu", "options default focus")

menu.manager:set_focus("options:sound_options")
menu.manager:activate(menu.manager.focus)
expect(menu.current_layout, "sound", "sound activation layout")
expect(focus_id(menu), "sound:previous_menu", "sound default focus")

local sound = assert(menu.items_by_id["sound:sound_volume"])
sound.control:set_value(0.25)
expect(settings.get("sound_volume"), 0.25, "sound slider preference")

local hardware = assert(menu.items_by_id["sound:hardware_acceleration"])
expect(hardware.control.enabled, false, "unsupported hardware acceleration is disabled")

menu:set_layout("main")
menu.manager:set_focus("main:save_exit")
menu.manager:activate(menu.manager.focus)
expect(saved, true, "save/exit callback")

menu.manager:set_focus("main:return_to_game")
menu.manager:activate(menu.manager.focus)
expect(closed, true, "return callback")
`
	scope := &modruntime.Scope{}
	defer scope.Close()
	if err := runtime.RunScoped(ctx, scope, func(state *glua.LState) error {
		return state.DoString(script)
	}); err != nil {
		t.Fatal(err)
	}
}

func TestLuaOptionsOverlayKeepsAuthoredCoordinatesViewportRelative(t *testing.T) {
	assertLuaOptionsBackdropCenter(t, "", 400, 300)
}

func TestLuaOptionsOverlayCentersInClassicViewport(t *testing.T) {
	assertLuaOptionsBackdropCenter(t, "lod-english-640x480-gameplay", 320, 240)
}

func assertLuaOptionsBackdropCenter(t *testing.T, profile string, wantX, wantY float64) {
	t.Helper()
	ctx := context.Background()
	contentFS, err := content.New(content.Layer{Name: "darkmagic", FS: content.D2Legacy()})
	if err != nil {
		t.Fatal(err)
	}

	runtime := modruntime.New()
	var input inputstate.Store
	var mixer audio.Mixer
	var composer render.Composer
	if err := runtime.RegisterInstaller(modruntime.ContentRequire(contentFS, "lua")); err != nil {
		t.Fatal(err)
	}
	scenes := modruntime.NewScenes(runtime, navigation.New())
	for _, module := range []modruntime.Module{
		modruntime.InputModule(&input),
		modruntime.DataModule(contentFS, d2presentation.ManifestTransforms(profile)),
		modruntime.RenderModule(runtime, &composer),
		modruntime.AudioModule(runtime, &mixer, contentFS, gamedata.New(recordstore.New(contentFS))),
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
	defer runtime.Stop(ctx)

	scope := &modruntime.Scope{}
	defer scope.Close()
	if err := runtime.RunScoped(ctx, scope, func(state *glua.LState) error {
		return state.DoString(`local overlay = require("d2legacy.overlays.options"); overlay:create()`)
	}); err != nil {
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
