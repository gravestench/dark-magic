package acceptance

import (
	"context"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

func TestLuaReusableWidgetModulesLoad(t *testing.T) {
	ctx := context.Background()
	var input inputstate.Store
	runtime := modruntime.New()
	composer := &render.Composer{}
	shim := content.D2Legacy()

	if err := runtime.RegisterInstaller(modruntime.ContentRequire(shim, "lua")); err != nil {
		t.Fatal(err)
	}
	for _, module := range []modruntime.Module{
		modruntime.InputModule(&input),
		modruntime.DataModule(shim),
		modruntime.RenderModule(runtime, composer),
	} {
		if err := runtime.RegisterModule(module); err != nil {
			t.Fatal(err)
		}
	}
	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(ctx)

	if err := runtime.Execute(ctx, content.D2Legacy(), "lua/d2legacy/tests/integration/ui_widgets.lua"); err != nil {
		t.Fatal(err)
	}
}
