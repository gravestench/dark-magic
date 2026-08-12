package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	lua "github.com/yuin/gopher-lua"
)

func TestContentRequireLoadsModuleFromVFS(t *testing.T) {
	t.Parallel()

	source := fstest.MapFS{
		"lua/d2/screens/loading.lua": &fstest.MapFile{Data: []byte(`return { id = "loading" }`)},
		"boot.lua":                   &fstest.MapFile{Data: []byte(`screen_id = require("d2.screens.loading").id`)},
	}
	runtime := New()
	if err := runtime.RegisterInstaller(ContentRequire(source, "lua")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	if err := runtime.Execute(context.Background(), source, "boot.lua"); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		if got := state.GetGlobal("screen_id").String(); got != "loading" {
			t.Fatalf("screen_id = %q", got)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestInvalidateContentModuleReloadsNextRequire(t *testing.T) {
	files := fstest.MapFS{"lua/example.lua": &fstest.MapFile{Data: []byte(`return { value = 1 }`)}}
	runtime := New()
	if err := runtime.RegisterInstaller(ContentRequire(files, "lua")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	if err := runtime.Execute(context.Background(), fstest.MapFS{"first.lua": &fstest.MapFile{Data: []byte(`first = require("example").value`)}}, "first.lua"); err != nil {
		t.Fatal(err)
	}
	files["lua/example.lua"] = &fstest.MapFile{Data: []byte(`return { value = 2 }`)}
	if err := runtime.InvalidateModule(context.Background(), "example"); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Execute(context.Background(), fstest.MapFS{"second.lua": &fstest.MapFile{Data: []byte(`second = require("example").value`)}}, "second.lua"); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		if state.GetGlobal("first") != lua.LNumber(1) || state.GetGlobal("second") != lua.LNumber(2) {
			t.Fatalf("values = %s/%s", state.GetGlobal("first"), state.GetGlobal("second"))
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestContentModulesHaveIsolatedEnvironments(t *testing.T) {
	source := fstest.MapFS{
		"lua/first.lua":  &fstest.MapFile{Data: []byte(`private_value = 42; return { value = private_value }`)},
		"lua/second.lua": &fstest.MapFile{Data: []byte(`return { leaked = private_value ~= nil }`)},
		"boot.lua":       &fstest.MapFile{Data: []byte(`first_value = require("first").value; leaked = require("second").leaked`)},
	}
	runtime := New()
	if err := runtime.RegisterInstaller(ContentRequire(source, "lua")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	if err := runtime.Execute(context.Background(), source, "boot.lua"); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		if state.GetGlobal("first_value") != lua.LNumber(42) || state.GetGlobal("leaked") != lua.LFalse || state.GetGlobal("private_value") != lua.LNil {
			t.Fatalf("module environment leaked: first=%s leaked=%s global=%s", state.GetGlobal("first_value"), state.GetGlobal("leaked"), state.GetGlobal("private_value"))
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
