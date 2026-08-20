package acceptance

import (
	"context"
	"io/fs"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/runtime/lua"
)

// TestLuaControlsPointerKeyboardFocusAndAccessibility protects press capture, release, and cancellation semantics.
func TestLuaControlsPointerKeyboardFocusAndAccessibility(t *testing.T) {
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
			t.Errorf("stop pointer-control runtime: %v", err)
		}
	}()

	scripts, err := fs.Sub(content.D2Legacy(), "lua/d2legacy/tests/integration/ui_controls_pointer")
	if err != nil {
		t.Fatal(err)
	}

	if err := runtime.Execute(ctx, scripts, "setup.lua"); err != nil {
		t.Fatal(err)
	}

	// Pressing inside captures/depresses the control but does not activate it.
	input.Publish(inputstate.Frame{
		CursorX: 35, CursorY: 5,
		Actions: map[string]inputstate.ActionState{"pointer_primary": {Pressed: true}},
	})

	if err := runtime.Execute(ctx, scripts, "update.lua"); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Execute(ctx, scripts, "assert_down.lua"); err != nil {
		t.Fatal(err)
	}

	// Releasing inside the same captured control activates it.
	input.Publish(inputstate.Frame{
		CursorX: 35, CursorY: 5,
		Actions: map[string]inputstate.ActionState{"pointer_primary": {Released: true}},
	})

	if err := runtime.Execute(ctx, scripts, "update.lua"); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Execute(ctx, scripts, "assert_up.lua"); err != nil {
		t.Fatal(err)
	}

	// A press that begins inside but releases outside is cancelled.
	if err := runtime.Execute(ctx, scripts, "clear.lua"); err != nil {
		t.Fatal(err)
	}

	input.Publish(inputstate.Frame{
		CursorX: 35, CursorY: 5,
		Actions: map[string]inputstate.ActionState{"pointer_primary": {Pressed: true}},
	})

	if err := runtime.Execute(ctx, scripts, "update.lua"); err != nil {
		t.Fatal(err)
	}

	input.Publish(inputstate.Frame{
		CursorX: 100, CursorY: 100,
		Actions: map[string]inputstate.ActionState{"pointer_primary": {Released: true}},
	})

	if err := runtime.Execute(ctx, scripts, "update.lua"); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Execute(ctx, scripts, "assert_cancelled.lua"); err != nil {
		t.Fatal(err)
	}
}

// TestLuaControlsFormFieldsAndFocusScopes protects focus traversal, Unicode editing, and scoped field input.
func TestLuaControlsFormFieldsAndFocusScopes(t *testing.T) {
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
			t.Errorf("stop form-control runtime: %v", err)
		}
	}()

	scripts, err := fs.Sub(content.D2Legacy(), "lua/d2legacy/tests/integration/ui_controls_form")
	if err != nil {
		t.Fatal(err)
	}

	if err := runtime.Execute(ctx, scripts, "setup.lua"); err != nil {
		t.Fatal(err)
	}

	frames := []inputstate.Frame{
		{Actions: map[string]inputstate.ActionState{"confirm": {Pressed: true}}},
		{Actions: map[string]inputstate.ActionState{"down": {Pressed: true}}},
		{Text: "éx", Actions: map[string]inputstate.ActionState{}},
		{Actions: map[string]inputstate.ActionState{"backspace": {Pressed: true}}},
		{Actions: map[string]inputstate.ActionState{"down": {Pressed: true}}},
		{Actions: map[string]inputstate.ActionState{"right": {Pressed: true}}},
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
