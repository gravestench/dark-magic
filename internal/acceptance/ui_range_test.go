package acceptance

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/inputstate"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

func TestLuaControlsSliderKeyboardAndDragSemantics(t *testing.T) {
	ctx := context.Background()
	var input inputstate.Store
	runtime := modruntime.New()
	if err := runtime.RegisterInstaller(modruntime.ContentRequire(content.D2Legacy(), "lua")); err != nil {
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
local ui=require("d2legacy.ui.controls")
manager=ui.new()
slider=manager:add_slider({id="volume",x=10,y=10,width=100,height=20,min=0,max=100,step=10,value=40})
`)},
		"update.lua": &fstest.MapFile{Data: []byte(`manager:update()`)},
		"assert_keyboard.lua": &fstest.MapFile{Data: []byte(`
local a=manager:accessibility()
assert(slider.value==50 and a[1].role=="slider" and a[1].focused==true)
`)},
		"assert_drag.lua": &fstest.MapFile{Data: []byte(`assert(slider.value==80)`)},
	}
	if err := runtime.Execute(ctx, scripts, "setup.lua"); err != nil {
		t.Fatal(err)
	}

	input.Publish(inputstate.Frame{Actions: map[string]inputstate.ActionState{"right": {Pressed: true}}})
	if err := runtime.Execute(ctx, scripts, "update.lua"); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Execute(ctx, scripts, "assert_keyboard.lua"); err != nil {
		t.Fatal(err)
	}

	input.Publish(inputstate.Frame{CursorX: 88, CursorY: 15, Actions: map[string]inputstate.ActionState{"pointer_primary": {Pressed: true, Down: true}}})
	if err := runtime.Execute(ctx, scripts, "update.lua"); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Execute(ctx, scripts, "assert_drag.lua"); err != nil {
		t.Fatal(err)
	}
}
