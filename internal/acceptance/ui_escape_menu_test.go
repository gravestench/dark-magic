package acceptance

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/game/data/catalog"
	"github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
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
	} {
		if err := runtime.RegisterModule(module); err != nil {
			t.Fatal(err)
		}
	}
	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(ctx)

	scripts := fstest.MapFS{
		"assert.lua": &fstest.MapFile{Data: []byte(`
local render = require("dm.render/v1")
local escape_menu = require("darkmagic.ui.escape_menu")

local root = render.create("modal")
local closed = false
local saved = false
local changed = nil
local menu = escape_menu.new(root, {
  start_layout = "main",
  on_close = function() closed = true end,
  on_save_exit = function() saved = true end,
  on_option_change = function(layout, id, value)
    changed = layout .. ":" .. id .. "=" .. value
  end,
})

assert(menu.current_layout == "main")
assert(menu.manager.focus.id == "main:return_to_game")

-- OpenDiablo2 stops at the ends of the list rather than wrapping.
menu.manager:move_focus(1)
assert(menu.manager.focus.id == "main:return_to_game")
menu.manager:move_focus(-1)
assert(menu.manager.focus.id == "main:save_exit")
menu.manager:move_focus(-1)
assert(menu.manager.focus.id == "main:options")
menu.manager:move_focus(-1)
assert(menu.manager.focus.id == "main:options")

menu.manager:activate(menu.manager.focus)
assert(menu.current_layout == "options")
assert(menu.manager.focus.id == "options:previous_menu")

menu.manager:set_focus("options:sound_options")
menu.manager:activate(menu.manager.focus)
assert(menu.current_layout == "sound")
assert(menu.manager.focus.id == "sound:previous_menu")

menu.manager:set_focus("sound:hardware_acceleration")
local hardware = assert(menu.items_by_id["sound:hardware_acceleration"])
assert(hardware.values[hardware.value_index] == "ON")
menu.manager:activate(menu.manager.focus)
assert(hardware.values[hardware.value_index] == "OFF")
assert(changed == "sound:hardware_acceleration=OFF")

menu:set_layout("main")
menu.manager:set_focus("main:save_exit")
menu.manager:activate(menu.manager.focus)
assert(saved)

menu.manager:set_focus("main:return_to_game")
menu.manager:activate(menu.manager.focus)
assert(closed)
`)},
	}
	if err := runtime.Execute(ctx, scripts, "assert.lua"); err != nil {
		t.Fatal(err)
	}
}
