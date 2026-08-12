package save

import (
	"context"
	"testing"
	"testing/fstest"

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

func execute(t *testing.T, runtime *modruntime.Runtime, script string) {
	t.Helper()
	if err := runtime.Execute(context.Background(), fstest.MapFS{"test.lua": {Data: []byte(script)}}, "test.lua"); err != nil {
		t.Fatal(err)
	}
}

func TestLuaPolicyCreatesAndSelectsCharacter(t *testing.T) {
	store := New()
	runtime := policyRuntime(t, store)
	execute(t, runtime, `local s=require("d2legacy.save/v1"); local id=assert(s.create("hero", "Hero", "Amazon")); assert(id=="hero"); assert(s.select(id)); assert(s.selected().name=="Hero")`)
	selected, ok := store.Selected()
	if !ok || selected.Name != "Hero" {
		t.Fatalf("selected = %#v", selected)
	}
}

func TestLuaPolicyOwnsNameClassAndCreationOptions(t *testing.T) {
	store := New()
	runtime := policyRuntime(t, store)
	execute(t, runtime, `
local s=require("d2legacy.save/v1")
local id=assert(s.create_named("Iron-Wolf", "paladin", false, true))
assert(id=="paladin-iron-wolf" and s.select(id))
local c=s.selected(); assert(c.class=="Paladin" and not c.expansion and c.hardcore)
assert(s.create_named("A", "amazon")==nil)
assert(s.create_named("Valid", "monk")==nil)
assert(s.create_named("Iron-Wolf", "amazon")==nil)
`)
}

func TestLuaPolicyDeletesCharacterByOpaqueID(t *testing.T) {
	store := New(Character{ID: "hero", Name: "Hero", Class: "Amazon", Level: 1})
	runtime := policyRuntime(t, store)
	execute(t, runtime, `local s=require("d2legacy.save/v1"); assert(s.select("hero")); assert(s.delete("hero")); assert(#s.characters()==0); assert(s.selected()==nil)`)
}

func TestAdapterReturnsDefensiveAppearanceAndStatsSnapshots(t *testing.T) {
	store := New(Character{
		ID: "hero", Name: "Hero", Class: "Amazon", Level: 12,
		Appearance: &Appearance{COF: "hero.cof", Palette: "units.dat", Direction: 3, Components: map[string]string{"HD": "head.dcc", "TR": "torso.dcc"}},
		Stats:      &Stats{Strength: 25, Health: 70, MaxHealth: 80, FireResistance: 15},
	})
	runtime := policyRuntime(t, store)
	execute(t, runtime, `
local c=require("d2legacy.save/v1").characters()[1]
assert(c.appearance.cof=="hero.cof" and c.appearance.palette=="units.dat")
assert(c.appearance.direction==3 and c.appearance.components.HD=="head.dcc")
assert(c.stats.strength==25 and c.stats.health==70 and c.stats.max_health==80)
c.appearance.components.HD="mutated"
assert(require("d2legacy.save/v1").characters()[1].appearance.components.HD=="head.dcc")
`)
}
