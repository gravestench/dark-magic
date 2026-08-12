package save

import (
	"context"
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

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
	t.Cleanup(func() { _ = runtime.Stop(context.Background()) })
	return runtime
}

func execute(t *testing.T, runtime *modruntime.Runtime, name string) {
	t.Helper()
	if err := runtime.Execute(context.Background(), os.DirFS("."), "testdata/"+name); err != nil {
		t.Fatal(err)
	}
}

func TestLuaPolicyCreatesAndSelectsCharacter(t *testing.T) {
	store := New()
	runtime := policyRuntime(t, store)
	execute(t, runtime, "create_select_test.lua")
	selected, ok := store.Selected()
	if !ok || selected.Name != "Hero" {
		t.Fatalf("selected = %#v", selected)
	}
}

func TestLuaPolicyOwnsNameClassAndCreationOptions(t *testing.T) {
	store := New()
	runtime := policyRuntime(t, store)
	execute(t, runtime, "name_class_test.lua")
}

func TestLuaPolicyDeletesCharacterByOpaqueID(t *testing.T) {
	store := New(Character{ID: "hero", Name: "Hero", Class: "Amazon", Level: 1})
	runtime := policyRuntime(t, store)
	execute(t, runtime, "delete_test.lua")
}

func TestAdapterReturnsDefensiveAppearanceAndStatsSnapshots(t *testing.T) {
	store := New(Character{
		ID: "hero", Name: "Hero", Class: "Amazon", Level: 12,
		Appearance: &Appearance{COF: "hero.cof", Palette: "units.dat", Direction: 3, Components: map[string]string{"HD": "head.dcc", "TR": "torso.dcc"}},
		Stats:      &Stats{Strength: 25, Health: 70, MaxHealth: 80, FireResistance: 15},
	})
	runtime := policyRuntime(t, store)
	execute(t, runtime, "snapshots_test.lua")
}
