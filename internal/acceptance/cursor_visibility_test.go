package acceptance

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	"github.com/gravestench/dark-magic/internal/runtime/lua"
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
local cursor = require("darkmagic.ui.cursor")
local definition = {
    sheet = "unused.dc6",
    palette = "units",
    direction = 0,
    frame = 0,
    hotspot = { x = 0, y = 0 },
}
local palettes = { units = "unused.dat" }

-- Only the focused cursor remains visible.
local first = cursor.new(nil, definition, palettes)
local second = cursor.new(nil, definition, palettes)
cursor.focus(first, true)
assert(first.visible == true)
assert(second.visible == false)

-- A scene wrapper supplies a cursor even when the screen itself never creates
-- one, and dynamic visibility can hide it during cinematic playback.
local scene = cursor.wrap({ playing = false }, definition, palettes, {
    visible_when = function(current) return not current.playing end,
})
scene:create()
scene:enter()
assert(scene.cursor ~= nil and scene.cursor.visible == true)
assert(first.visible == false and second.visible == false)

scene.playing = true
scene:update(0, true)
assert(scene.cursor.visible == false)

scene.playing = false
scene:update(0, true)
assert(scene.cursor.visible == true)

-- Losing focus hides this cursor even if the scene below a blocking overlay
-- would otherwise stop receiving updates.
scene:update(0, false)
assert(scene.cursor.visible == false)

-- Hidden scenes (startup cinematics/loading) create no software pointer and
-- synchronously suppress every cursor already registered.
local hidden = cursor.wrap({}, definition, palettes, { hidden = true })
hidden:create()
hidden:enter()
assert(hidden.cursor == nil)
assert(first.visible == false and second.visible == false)
assert(scene.cursor.visible == false)

scene:destroy()
hidden:destroy()
`)},
	}
	if err := runtime.Execute(ctx, scripts, "cursor.lua"); err != nil {
		t.Fatal(err)
	}
}
