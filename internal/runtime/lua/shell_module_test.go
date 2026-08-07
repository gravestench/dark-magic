package modruntime

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gravestench/dark-magic/internal/shell"
	"github.com/yuin/gopher-lua"
)

func TestShellModuleEditsResetsAndSavesSettings(t *testing.T) {
	settings, err := shell.NewSettings(filepath.Join(t.TempDir(), "shell.json"))
	if err != nil {
		t.Fatal(err)
	}
	runtime := New()
	if err := runtime.RegisterModule(ShellModule(settings)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	source := `local shell = require("dm.shell/v1")
shell.set("font_size", 24)
assert(shell.get("font_size") == 24)
assert(shell.values().font_size == 24)
assert(shell.status().dirty)
shell.save()
assert(not shell.status().dirty)
shell.reset()
assert(shell.get("font_size") == shell.defaults().font_size)`
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		return state.DoString(source)
	}); err != nil {
		t.Fatal(err)
	}
}
