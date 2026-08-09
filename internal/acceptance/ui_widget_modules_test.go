package acceptance

import (
	"context"
	"testing"
	"testing/fstest"

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
	shim := content.Shim()

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

	scripts := fstest.MapFS{
		"load.lua": &fstest.MapFile{Data: []byte(`
local modules = {
  "darkmagic.ui.slider",
  "darkmagic.ui.scrollbar",
  "darkmagic.ui.list",
  "darkmagic.ui.tabs",
  "darkmagic.ui.panel",
  "darkmagic.ui.progress_bar",
}
for _, name in ipairs(modules) do
  local loaded = require(name)
  assert(type(loaded) == "table", name .. " did not return a module")
end
`)},
	}
	if err := runtime.Execute(ctx, scripts, "load.lua"); err != nil {
		t.Fatal(err)
	}
}
