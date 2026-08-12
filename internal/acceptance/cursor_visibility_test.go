package acceptance

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	"github.com/gravestench/dark-magic/internal/runtime/lua"
	glua "github.com/yuin/gopher-lua"
)

func TestLuaSoftwareCursorFocusAndSuppressionPolicy(t *testing.T) {
	ctx := context.Background()
	var input inputstate.Store
	var composer render.Composer
	runtime := modruntime.New()
	if err := runtime.RegisterInstaller(modruntime.ContentRequire(content.Shim(), "lua")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RegisterModule(modruntime.InputModule(&input)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RegisterModule(modruntime.RenderModule(runtime, &composer)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(ctx)

	scripts := fstest.MapFS{
		"cursor.lua": &fstest.MapFile{Data: []byte(`
local cursor = require("d2.ui.cursor")
local definition = {
    palette = "units",
    direction = 0,
    default_mode = "default",
    modes = {
        default = { sheet = "unused.dc6", frame = 0, hotspot = { x = 0, y = 0 } },
        pressed = { sheet = "unused.dc6", frame = 5, hotspot = { x = 0, y = -2 } },
        hand = { sheet = "unused.dc6", frame = 6, hotspot = { x = 0, y = -2 } },
    },
}
local palettes = { units = "unused.dat" }

-- Only the focused cursor remains visible.
local first = cursor.new(nil, definition, palettes)
local second = cursor.new(nil, definition, palettes)
assert(first.mode == "default")
first:set_mode("hand")
assert(first.mode == "hand" and first.requested_mode == "hand")
cursor.focus(first, true)
assert(first.visible == true)
assert(second.visible == false)

-- A scene wrapper supplies a shell cursor even when the screen itself never
-- creates an authored self.cursor, and dynamic visibility can hide it during
-- cinematic playback. The shell cursor must not mutate scene-owned cursor state.
local scene = cursor.wrap({ playing = false }, definition, palettes, {
    visible_when = function(current) return not current.playing end,
})
scene:create()
scene:enter()
assert(scene.cursor == nil)
assert(scene.__darkmagic_shell_cursor ~= nil and scene.__darkmagic_shell_cursor.visible == true)
assert(first.visible == false and second.visible == false)

scene.playing = true
scene:update(0, true)
assert(scene.__darkmagic_shell_cursor.visible == false)

scene.playing = false
scene:update(0, true)
assert(scene.__darkmagic_shell_cursor.visible == true)

-- Losing focus hides this cursor even if the scene below a blocking overlay
-- would otherwise stop receiving updates.
scene:update(0, false)
assert(scene.__darkmagic_shell_cursor.visible == false)

-- Regression: character_select and similar scenes may use self.cursor==nil as
-- their own incomplete/redirecting lifecycle guard. Automatic shell cursor
-- ownership must never defeat that guard.
local guarded_updates = 0
local guarded = cursor.wrap({
    update = function(self)
        if not self.cursor then return end
        guarded_updates = guarded_updates + 1
    end,
}, definition, palettes)
guarded:create()
guarded:enter()
guarded:update(0.016, true)
assert(guarded.cursor == nil)
assert(guarded.__darkmagic_shell_cursor ~= nil)
assert(guarded_updates == 0)

-- Hidden scenes (startup cinematics/loading) create no software pointer and
-- synchronously suppress every cursor already registered.
local hidden = cursor.wrap({}, definition, palettes, { hidden = true })
hidden:create()
hidden:enter()
assert(hidden.cursor == nil and hidden.__darkmagic_shell_cursor == nil)
assert(first.visible == false and second.visible == false)
assert(scene.__darkmagic_shell_cursor.visible == false)
assert(guarded.__darkmagic_shell_cursor.visible == false)

scene:destroy()
guarded:destroy()
hidden:destroy()
pressed_cursor = cursor.new(nil, definition, palettes)
`)},
	}
	if err := runtime.Execute(ctx, scripts, "cursor.lua"); err != nil {
		t.Fatal(err)
	}
	input.Publish(inputstate.Frame{
		Actions: map[string]inputstate.ActionState{"pointer_primary": {Down: true}},
		Owner:   inputstate.FocusOwner{Domain: inputstate.FocusScene, ID: "cursor-test"},
	})
	if err := runtime.Run(ctx, func(state *glua.LState) error {
		return state.DoString(`pressed_cursor:update(); assert(pressed_cursor.mode == "pressed")`)
	}); err != nil {
		t.Fatal(err)
	}
}
