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
	write("boot.lua", `local h=require("helper"); return { id="boot", start=function() observed=h.value end }`)
	contentFS, err := content.New(content.Layer{Name: "mods", FS: content.Directory(root)})
	if err != nil {
		t.Fatal(err)
	}
	runtime := modruntime.New()
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
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		if state.GetGlobal("observed") != lua.LNumber(2) {
			t.Fatalf("observed = %s", state.GetGlobal("observed"))
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
