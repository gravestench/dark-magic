package hotreload

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/runtime/lua"
	lua "github.com/yuin/gopher-lua"
)

type helperReloadFixture struct {
	root        string
	coordinator *Coordinator
	observed    int
}

// TestHelperChangeInvalidatesRequireAndReplacesActiveDefinition verifies that dependents observe a fresh module value.
func TestHelperChangeInvalidatesRequireAndReplacesActiveDefinition(t *testing.T) {
	fixture := newHelperReloadFixture(t)
	fixture.writeSource(t, "lua/helper.lua", `return { value = 2 }`)

	if err := fixture.coordinator.Reload(context.Background(), "lua/helper.lua"); err != nil {
		t.Fatal(err)
	}

	if fixture.observed != 2 {
		t.Fatalf("observed = %d", fixture.observed)
	}
}

// newHelperReloadFixture starts an active boot definition and gives the test ownership of runtime cleanup.
func newHelperReloadFixture(t *testing.T) *helperReloadFixture {
	t.Helper()

	fixture := &helperReloadFixture{root: t.TempDir()}
	fixture.writeSource(t, "lua/helper.lua", `return { value = 1 }`)
	fixture.writeSource(t, "boot.lua", `
local helper = require("helper")
local observe = require("test.observe/v1")

return {
    id = "boot",
    start = function()
        observe.set(helper.value)
    end,
}
`)

	contentFS := newContentFS(t, fixture.root)
	runtime := newObservedRuntime(t, contentFS, &fixture.observed)
	definitions := discoverDefinitions(t, runtime, contentFS)
	manager := enableBootDefinition(t, definitions)

	fixture.coordinator = New(contentFS, runtime, manager, nil, definitions)

	return fixture
}

// writeSource creates a fixture source with private permissions so each scenario remains isolated in its temp tree.
func (f *helperReloadFixture) writeSource(t *testing.T, name, value string) {
	t.Helper()

	full := filepath.Join(f.root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(full, []byte(value), 0o600); err != nil {
		t.Fatal(err)
	}
}

// newContentFS mounts the scenario directory as the sole content layer, making file invalidation observable.
func newContentFS(t *testing.T, root string) *content.FS {
	t.Helper()

	contentFS, err := content.New(content.Layer{Name: "mods", FS: content.Directory(root)})
	if err != nil {
		t.Fatal(err)
	}

	return contentFS
}

// newObservedRuntime starts a Lua runtime whose test module records values emitted by component startup.
func newObservedRuntime(t *testing.T, contentFS *content.FS, observed *int) *modruntime.Runtime {
	t.Helper()

	runtime := modruntime.New()
	// The module bridges Lua startup into a Go assertion without exposing fixture internals to the source under test.
	if err := runtime.RegisterModule(modruntime.Module{Name: "test.observe/v1", Loader: func(state *lua.LState) int {
		// setObserved captures the value synchronously because component startup completes before Reload returns.
		setObserved := func(state *lua.LState) int {
			*observed = state.CheckInt(1)
			return 0
		}
		state.Push(state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{"set": setObserved}))

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
	// Runtime shutdown belongs to the fixture even if a later setup assertion aborts the test.
	t.Cleanup(func() {
		_ = runtime.Stop(context.Background())
	})

	return runtime
}

// discoverDefinitions loads fixture sources through the same runtime and VFS path used by production discovery.
func discoverDefinitions(
	t *testing.T,
	runtime *modruntime.Runtime,
	contentFS *content.FS,
) []modruntime.Definition {
	t.Helper()

	definitions, err := modruntime.DiscoverDefinitions(context.Background(), runtime, contentFS)
	if err != nil {
		t.Fatal(err)
	}

	return definitions
}

// enableBootDefinition registers and activates the fixture definition so replacement exercises manager lifecycle logic.
func enableBootDefinition(t *testing.T, definitions []modruntime.Definition) *host.Manager {
	t.Helper()

	manager := host.NewManager()
	if err := manager.Register(definitions[0].Managed()); err != nil {
		t.Fatal(err)
	}

	if err := manager.Enable(context.Background(), "boot"); err != nil {
		t.Fatal(err)
	}

	return manager
}

// TestAuthoritativeD2LegacyChangeRequiresNewSession checks both guarded source forms before any cache is invalidated.
func TestAuthoritativeD2LegacyChangeRequiresNewSession(t *testing.T) {
	root := t.TempDir()
	contentFS := newContentFS(t, root)
	coordinator := New(contentFS, modruntime.New(), host.NewManager(), nil, nil)

	for _, changed := range []string{
		"lua/d2legacy/policy/damage.lua",
		"components/d2legacy.lua",
	} {
		if err := coordinator.Reload(context.Background(), changed); !errors.Is(err, ErrAuthoritativeReload) {
			t.Fatalf("Reload(%q) error = %v", changed, err)
		}
	}
}
