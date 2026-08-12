package acceptance

import (
	"context"
	"testing"

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
	if err := runtime.RegisterInstaller(modruntime.ContentRequire(content.D2Legacy(), "lua")); err != nil {
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

	if err := runtime.Execute(ctx, content.D2Legacy(), "lua/d2legacy/tests/integration/cursor_visibility.lua"); err != nil {
		t.Fatal(err)
	}
	input.Publish(inputstate.Frame{
		Actions: map[string]inputstate.ActionState{"pointer_primary": {Down: true}},
		Owner:   inputstate.FocusOwner{Domain: inputstate.FocusScene, ID: "cursor-test"},
	})
	if err := runtime.Execute(ctx, content.D2Legacy(), "lua/d2legacy/tests/integration/cursor_pressed.lua"); err != nil {
		t.Fatal(err)
	}
}
