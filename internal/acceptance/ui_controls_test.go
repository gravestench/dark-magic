package acceptance

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/inputcore"
	"github.com/gravestench/dark-magic/internal/modruntime"
)

func TestLuaControlsPointerKeyboardFocusAndAccessibility(t *testing.T) {
	ctx := context.Background()
	var input inputcore.Store
	runtime := modruntime.New()
	if err := runtime.RegisterInstaller(modruntime.ContentRequire(content.Shim(), "lua")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RegisterModule(modruntime.InputModule(&input)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(ctx)
	scripts := fstest.MapFS{
		"setup.lua": &fstest.MapFile{Data: []byte(`
local ui = require("darkmagic.ui.controls")
manager = ui.new()
activated = ""
manager:add({id="one", label="First", x=0, y=0, width=20, height=20, on_activate=function(c) activated=c.id end})
manager:add({id="two", label="Second", x=30, y=0, width=20, height=20, on_activate=function(c) activated=c.id end})
`)},
		"update.lua": &fstest.MapFile{Data: []byte(`manager:update()`)},
		"assert.lua": &fstest.MapFile{Data: []byte(`local a=manager:accessibility(); assert(activated=="two" and a[2].focused and a[2].role=="button")`)},
	}
	if err := runtime.Execute(ctx, scripts, "setup.lua"); err != nil {
		t.Fatal(err)
	}
	input.Publish(inputcore.Frame{CursorX: 35, CursorY: 5, Actions: map[string]inputcore.ActionState{"pointer_primary": {Pressed: true}}})
	if err := runtime.Execute(ctx, scripts, "update.lua"); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Execute(ctx, scripts, "assert.lua"); err != nil {
		t.Fatal(err)
	}
}
