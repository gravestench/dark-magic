package luaModLoader

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/yuin/gopher-lua"

	"github.com/gravestench/dark-magic/pkg/services/fileLoader"
	"github.com/gravestench/dark-magic/pkg/services/luaManager"
)

func TestExternalModLoadsEndToEnd(t *testing.T) {
	root := filepath.Join(t.TempDir(), "example")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	manifestJSON := `{"Name":"Example Mod","Version":"1.0","Enabled":true}`
	initLua := `local Mod = {}; function Mod:Init() self.initialized = true end; return Mod`
	if err := os.WriteFile(filepath.Join(root, "manifest.json"), []byte(manifestJSON), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "init.lua"), []byte(initLua), 0o600); err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	luaService := &luaManager.Service{}
	luaService.SetLogger(logger)
	luaService.RebuildState()
	defer luaService.OnShutdown()
	if err := luaService.WithState(func(state *lua.LState) error {
		api := state.GetGlobal("api").(*lua.LTable)
		state.SetField(api, "mods", state.NewTable())
		setupPackagePath(state, filepath.Dir(root))
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	service := &Service{
		Config: &Config{EnabledMods: map[string]bool{}},
		lua:    luaService,
		loader: fileLoader.New(),
	}
	service.SetLogger(logger)
	if err := service.loadMod(root, os.DirFS(root)); err != nil {
		t.Fatal(err)
	}

	if err := luaService.WithState(func(state *lua.LState) error {
		api := state.GetGlobal("api")
		mods := state.GetField(api, "mods")
		mod := state.GetField(mods, "examplemod10")
		if got := state.GetField(mod, "initialized"); got != lua.LTrue {
			t.Fatalf("initialized = %v, want true", got)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestEmbeddedModLoadsEndToEnd(t *testing.T) {
	modDirectory := t.TempDir()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	luaService := &luaManager.Service{}
	luaService.SetLogger(logger)
	luaService.RebuildState()
	defer luaService.OnShutdown()

	service := &Service{
		Config: &Config{ModDirectory: modDirectory, EnabledMods: map[string]bool{}},
		lua:    luaService,
		loader: fileLoader.New(),
	}
	service.SetLogger(logger)
	if err := service.installBuiltinMods(); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(modDirectory, "smoke")
	if err := luaService.WithState(func(state *lua.LState) error {
		api := state.GetGlobal("api").(*lua.LTable)
		state.SetField(api, "mods", state.NewTable())
		setupPackagePath(state, modDirectory)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := service.loadMod(root, os.DirFS(root)); err != nil {
		t.Fatal(err)
	}
	if err := luaService.WithState(func(state *lua.LState) error {
		mod := state.GetField(state.GetField(state.GetGlobal("api"), "mods"), "darkmagicsmoketest10")
		if got := state.GetField(mod, "initialized"); got != lua.LTrue {
			t.Fatalf("initialized = %v, want true", got)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
