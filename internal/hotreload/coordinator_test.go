package hotreload

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/host"
	"github.com/gravestench/dark-magic/internal/modruntime"
	lua "github.com/yuin/gopher-lua"
)

func TestHelperChangeInvalidatesRequireAndReplacesActiveDefinition(t *testing.T) {
	root := t.TempDir()
	write := func(name, value string) {
		t.Helper()
		full := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(value), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("lua/helper.lua", `return { value = 1 }`)
	write("boot.lua", `local h=require("helper"); local observe=require("test.observe/v1"); return { id="boot", start=function() observe.set(h.value) end }`)
	contentFS, err := content.New(content.Layer{Name: "mods", FS: content.Directory(root)})
	if err != nil {
		t.Fatal(err)
	}
	runtime := modruntime.New()
	observed := 0
	if err := runtime.RegisterModule(modruntime.Module{Name: "test.observe/v1", Loader: func(state *lua.LState) int {
		state.Push(state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{"set": func(state *lua.LState) int { observed = state.CheckInt(1); return 0 }}))
		return 1
	}}); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RegisterInstaller(modruntime.ContentRequire(contentFS, "lua")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	definitions, err := modruntime.DiscoverDefinitions(context.Background(), runtime, contentFS)
	if err != nil {
		t.Fatal(err)
	}
	manager := host.NewManager()
	if err := manager.Register(definitions[0].Managed()); err != nil {
		t.Fatal(err)
	}
	if err := manager.Enable(context.Background(), "boot"); err != nil {
		t.Fatal(err)
	}
	coordinator := New(contentFS, runtime, manager, nil, definitions)
	write("lua/helper.lua", `return { value = 2 }`)
	if err := coordinator.Reload(context.Background(), "lua/helper.lua"); err != nil {
		t.Fatal(err)
	}
	if observed != 2 {
		t.Fatalf("observed = %d", observed)
	}
}
