package acceptance

import (
	"context"
	"io/fs"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/inputstate"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// TestLuaControlsSliderKeyboardAndDragSemantics keeps keyboard steps and pointer dragging behavior aligned.
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

	defer func() {
		if err := runtime.Stop(ctx); err != nil {
			t.Errorf("stop range-control runtime: %v", err)
		}
	}()

	scripts, err := fs.Sub(content.D2Legacy(), "lua/d2legacy/tests/integration/ui_range")
	if err != nil {
		t.Fatal(err)
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

	input.Publish(inputstate.Frame{
		CursorX: 88, CursorY: 15,
		Actions: map[string]inputstate.ActionState{"pointer_primary": {Pressed: true, Down: true}},
	})

	if err := runtime.Execute(ctx, scripts, "update.lua"); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Execute(ctx, scripts, "assert_drag.lua"); err != nil {
		t.Fatal(err)
	}
}
