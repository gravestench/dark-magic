package save

import (
	"context"
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// policyRuntime starts the real Lua capability with test content and ensures every test releases runtime resources.
func policyRuntime(t *testing.T, store *Store) *modruntime.Runtime {
	t.Helper()

	runtime := modruntime.New()
	if err := runtime.RegisterInstaller(modruntime.ContentRequire(content.D2Legacy(), "lua")); err != nil {
		t.Fatal(err)
	}

	if err := runtime.RegisterModule(Module(store)); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	// Cleanup must stop the runtime even when script assertions fail, preventing goroutine leakage between tests.
	t.Cleanup(func() { _ = runtime.Stop(context.Background()) })

	return runtime
}

// execute runs one authored Lua contract fixture through the same filesystem boundary used by runtime scripts.
func execute(t *testing.T, runtime *modruntime.Runtime, name string) {
	t.Helper()

	if err := runtime.Execute(context.Background(), os.DirFS("."), "testdata/"+name); err != nil {
		t.Fatal(err)
	}
}

// TestLuaPolicyCreatesAndSelectsCharacter verifies the atomic create-selected callback from Lua through Store.
func TestLuaPolicyCreatesAndSelectsCharacter(t *testing.T) {
	store := New()
	runtime := policyRuntime(t, store)
	execute(t, runtime, "create_select_test.lua")

	selected, ok := store.Selected()
	if !ok || selected.Name != "Hero" {
		t.Fatalf("selected = %#v", selected)
	}
}

// TestLuaPolicyOwnsNameClassAndCreationOptions keeps untrusted Lua input behind canonical creation policy.
func TestLuaPolicyOwnsNameClassAndCreationOptions(t *testing.T) {
	store := New()
	runtime := policyRuntime(t, store)
	execute(t, runtime, "name_class_test.lua")
}

// TestLuaPolicyDeletesCharacterByOpaqueID verifies that Lua deletes by identity without receiving mutable records.
func TestLuaPolicyDeletesCharacterByOpaqueID(t *testing.T) {
	store := New(Character{ID: "hero", Name: "Hero", Class: "Amazon", Level: 1})
	runtime := policyRuntime(t, store)
	execute(t, runtime, "delete_test.lua")
}

// TestAdapterReturnsDefensiveAppearanceAndStatsSnapshots protects both nested table schemas from aliasing Store data.
func TestAdapterReturnsDefensiveAppearanceAndStatsSnapshots(t *testing.T) {
	store := New(Character{
		ID:    "hero",
		Name:  "Hero",
		Class: "Amazon",
		Level: 12,
		Appearance: &Appearance{
			COF:        "hero.cof",
			Palette:    "units.dat",
			Direction:  3,
			Components: map[string]string{"HD": "head.dcc", "TR": "torso.dcc"},
		},
		Stats: &Stats{Strength: 25, Health: 70, MaxHealth: 80, FireResistance: 15},
	})
	runtime := policyRuntime(t, store)
	execute(t, runtime, "snapshots_test.lua")
}
