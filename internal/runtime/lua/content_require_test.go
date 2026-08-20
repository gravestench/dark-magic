package modruntime

import (
	"context"
	"strings"
	"testing"
	"testing/fstest"

	lua "github.com/yuin/gopher-lua"
)

// TestContentRequireLoadsModuleFromVFS protects the content require loads module from vfs contract, including its
// observable ordering and failure behavior.
func TestContentRequireLoadsModuleFromVFS(t *testing.T) {
	t.Parallel()

	source := fstest.MapFS{
		"lua/example/screens/loading.lua": &fstest.MapFile{
			Data: []byte(`return { id = "loading" }`),
		},
		"boot.lua": &fstest.MapFile{
			Data: []byte(`screen_id = require("example.screens.loading").id`),
		},
	}

	runtime := New()
	if err := runtime.RegisterInstaller(ContentRequire(source, "lua")); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(context.Background()) }()

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

// TestInvalidateContentModuleReloadsNextRequire protects the invalidate content module reloads next require
// contract, including its observable ordering and failure behavior.
func TestInvalidateContentModuleReloadsNextRequire(t *testing.T) {
	files := fstest.MapFS{"lua/example.lua": &fstest.MapFile{Data: []byte(`return { value = 1 }`)}}

	runtime := New()
	if err := runtime.RegisterInstaller(ContentRequire(files, "lua")); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(context.Background()) }()

	if err := runtime.Execute(
		context.Background(),
		fstest.MapFS{
			"first.lua": &fstest.MapFile{Data: []byte(`first = require("example").value`)},
		},
		"first.lua",
	); err != nil {
		t.Fatal(err)
	}

	files["lua/example.lua"] = &fstest.MapFile{Data: []byte(`return { value = 2 }`)}

	if err := runtime.InvalidateModule(context.Background(), "example"); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Execute(
		context.Background(),
		fstest.MapFS{
			"second.lua": &fstest.MapFile{Data: []byte(`second = require("example").value`)},
		},
		"second.lua",
	); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		if state.GetGlobal("first") != lua.LNumber(1) ||
			state.GetGlobal("second") != lua.LNumber(2) {
			t.Fatalf("values = %s/%s", state.GetGlobal("first"), state.GetGlobal("second"))
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

// TestContentModulesHaveIsolatedEnvironments protects the content modules have isolated environments contract,
// including its observable ordering and failure behavior.
func TestContentModulesHaveIsolatedEnvironments(t *testing.T) {
	source := fstest.MapFS{
		"lua/first.lua": &fstest.MapFile{
			Data: []byte(`private_value = 42; return { value = private_value }`),
		},
		"lua/second.lua": &fstest.MapFile{Data: []byte(`return { leaked = private_value ~= nil }`)},
		"boot.lua": &fstest.MapFile{
			Data: []byte(`first_value = require("first").value; leaked = require("second").leaked`),
		},
	}

	runtime := New()
	if err := runtime.RegisterInstaller(ContentRequire(source, "lua")); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(context.Background()) }()

	if err := runtime.Execute(context.Background(), source, "boot.lua"); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		if state.GetGlobal("first_value") != lua.LNumber(42) ||
			state.GetGlobal("leaked") != lua.LFalse ||
			state.GetGlobal("private_value") != lua.LNil {
			t.Fatalf(
				"module environment leaked: first=%s leaked=%s global=%s",
				state.GetGlobal("first_value"),
				state.GetGlobal("leaked"),
				state.GetGlobal("private_value"),
			)
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

// TestPackageRequireBindsModulesToOwningNamespace protects the package require binds modules to owning namespace
// contract, including its observable ordering and failure behavior.
func TestPackageRequireBindsModulesToOwningNamespace(t *testing.T) {
	source := fstest.MapFS{
		"mods/base/lua/base/value.lua": &fstest.MapFile{Data: []byte(`return {value="base"}`)},
		"mods/extension/lua/base/value.lua": &fstest.MapFile{
			Data: []byte(`return {value="spoofed"}`),
		},
		"mods/extension/lua/extension.lua": &fstest.MapFile{
			Data: []byte(`return {value="extension"}`),
		},
		"test.lua": &fstest.MapFile{Data: []byte(`
base_value = require("base.value").value
extension_value = require("extension").value
`)},
	}

	runtime := New()
	if err := runtime.RegisterInstaller(
		PackageRequire(source, []string{"base", "extension"}),
	); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(t.Context()) }()

	if err := runtime.Execute(t.Context(), source, "test.lua"); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Run(t.Context(), func(state *lua.LState) error {
		if state.GetGlobal("base_value").String() != "base" ||
			state.GetGlobal("extension_value").String() != "extension" {
			t.Fatalf("package modules resolved outside their owners")
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

// TestPackageRegistryChangeCannotReuseRemovedCachedModule protects the package registry change cannot reuse removed
// cached module contract, including its observable ordering and failure behavior.
func TestPackageRegistryChangeCannotReuseRemovedCachedModule(t *testing.T) {
	source := fstest.MapFS{
		"mods/extension/lua/extension/value.lua": &fstest.MapFile{
			Data: []byte(`return {value="old"}`),
		},
		"test.lua": &fstest.MapFile{
			Data: []byte(`value = require("extension.value").value`),
		},
	}
	registry := NewPackageRegistry([]string{"extension"})

	runtime := New()
	if err := runtime.RegisterInstaller(PackageRequireRegistry(source, registry)); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(t.Context()) }()

	if err := runtime.Execute(t.Context(), source, "test.lua"); err != nil {
		t.Fatal(err)
	}

	registry.Replace(nil)

	if err := InvalidatePackageModules(t.Context(), runtime, "extension"); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Execute(t.Context(), source, "test.lua"); err == nil {
		t.Fatal("removed package module remained available through package.loaded")
	}
}

// TestDefinitionDependenciesRequireDeclaredPackageOwnership protects the definition dependencies require declared
// package ownership contract, including its observable ordering and failure behavior.
func TestDefinitionDependenciesRequireDeclaredPackageOwnership(t *testing.T) {
	definitions := []Definition{
		{ID: "extension.boot", DependsOn: []string{"dependency.service", "extension.internal"}},
	}
	if err := ValidateDefinitionDependencies(
		definitions,
		"extension",
		[]string{"dependency"},
	); err != nil {
		t.Fatal(err)
	}

	if err := ValidateDefinitionDependencies(definitions, "extension", nil); err == nil {
		t.Fatal("undeclared component dependency was accepted")
	}
}

// TestDefinitionEntrypointsMustExist protects the definition entrypoints must exist contract, including its
// observable ordering and failure behavior.
func TestDefinitionEntrypointsMustExist(t *testing.T) {
	definitions := []Definition{{ID: "extension.client"}}
	if err := ValidateDefinitionEntrypoints(definitions, []string{"extension.client"}); err != nil {
		t.Fatal(err)
	}

	if err := ValidateDefinitionEntrypoints(
		definitions,
		[]string{"extension.missing"},
	); err == nil {
		t.Fatal("missing manifest entrypoint was accepted")
	}
}

// TestDefinitionDomainsRejectOppositeEntrypointDependency protects the definition domains reject opposite
// entrypoint dependency contract, including its observable ordering and failure behavior.
func TestDefinitionDomainsRejectOppositeEntrypointDependency(t *testing.T) {
	definitions := []Definition{
		{ID: "example.client", DependsOn: []string{"example.shared"}},
		{ID: "example.shared", DependsOn: []string{"example.authority"}},
		{ID: "example.authority"},
	}
	if err := ValidateDefinitionDomains(
		definitions,
		[]string{"example.client"},
		[]string{"example.authority"},
	); err == nil {
		t.Fatal("client dependency closure enabled an authority entrypoint")
	}

	definitions[1].DependsOn = nil
	if err := ValidateDefinitionDomains(
		definitions,
		[]string{"example.client"},
		[]string{"example.authority"},
	); err != nil {
		t.Fatal(err)
	}
}

// TestValidatePackageSyntaxCompilesAllLuaWithoutExecutingIt protects the validate package syntax compiles all lua
// without executing it contract, including its observable ordering and failure behavior.
func TestValidatePackageSyntaxCompilesAllLuaWithoutExecutingIt(t *testing.T) {
	source := fstest.MapFS{
		"boot.lua": &fstest.MapFile{Data: []byte(`error("must not execute")`)},
		"lua/example/good.lua": &fstest.MapFile{
			Data: []byte(`return function(value) return value + 1 end`),
		},
	}
	if err := ValidatePackageSyntax(source); err != nil {
		t.Fatal(err)
	}

	source["lua/example/bad.lua"] = &fstest.MapFile{Data: []byte(`return function(`)}
	if err := ValidatePackageSyntax(
		source,
	); err == nil ||
		!strings.Contains(err.Error(), "lua/example/bad.lua") {
		t.Fatalf("invalid package syntax error = %v", err)
	}
}
