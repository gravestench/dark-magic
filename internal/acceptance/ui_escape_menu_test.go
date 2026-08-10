package acceptance

import (
	"context"
	"testing"

	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/game/data/catalog"
	"github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/preferences"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	glua "github.com/yuin/gopher-lua"
)

func TestLuaEscapeMenuRecoveredNavigation(t *testing.T) {
	ctx := context.Background()
	contentFS, err := content.New(content.Layer{Name: "darkmagic", FS: content.Shim()})
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
local render = require("dm.render/v1")
local settings = require("dm.settings/v1")
local escape_menu = require("darkmagic.ui.escape_menu")

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
