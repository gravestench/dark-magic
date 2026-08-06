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
visual_state = ""
manager:add({id="one", label="First", x=0, y=0, width=20, height=20, on_activate=function(c) activated=c.id end})
manager:add({id="two", label="Second", x=30, y=0, width=20, height=20,
    on_activate=function(c) activated=c.id end,
    on_state=function(_, state) visual_state=state end})
`)},
		"update.lua": &fstest.MapFile{Data: []byte(`manager:update()`)},
		"assert.lua": &fstest.MapFile{Data: []byte(`local a=manager:accessibility(); assert(activated=="two" and visual_state=="pressed" and a[2].focused and a[2].role=="button")`)},
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

func TestLuaControlsFormFieldsAndFocusScopes(t *testing.T) {
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
local ui=require("darkmagic.ui.controls")
manager=ui.new()
manager:add({id="outside",x=0,y=0,width=10,height=10})
manager:add_checkbox({id="check",scope="form",x=20,y=0,width=10,height=10})
manager:add_text_field({id="name",scope="form",x=40,y=0,width=10,height=10,max_length=3})
manager:add_scrollbar({id="volume",scope="form",x=60,y=0,width=20,height=10,min=0,max=10,value=5,step=2})
manager:set_scope("form")
`)},
		"update.lua": &fstest.MapFile{Data: []byte(`manager:update()`)},
		"assert.lua": &fstest.MapFile{Data: []byte(`
local a=manager:accessibility()
assert(a[1].focused==false and a[2].checked==true and a[3].value=="é" and a[4].value==7 and a[4].focused==true)
`)},
	}
	if err := runtime.Execute(ctx, scripts, "setup.lua"); err != nil {
		t.Fatal(err)
	}
	frames := []inputcore.Frame{
		{Actions: map[string]inputcore.ActionState{"confirm": {Pressed: true}}},
		{Actions: map[string]inputcore.ActionState{"down": {Pressed: true}}},
		{Text: "éx", Actions: map[string]inputcore.ActionState{}},
		{Actions: map[string]inputcore.ActionState{"backspace": {Pressed: true}}},
		{Actions: map[string]inputcore.ActionState{"down": {Pressed: true}}},
		{Actions: map[string]inputcore.ActionState{"right": {Pressed: true}}},
	}
	for _, frame := range frames {
		input.Publish(frame)
		if err := runtime.Execute(ctx, scripts, "update.lua"); err != nil {
			t.Fatal(err)
		}
	}
	if err := runtime.Execute(ctx, scripts, "assert.lua"); err != nil {
		t.Fatal(err)
	}
}
