package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/savecore"
)

func TestSaveModuleSelectsCharacter(t *testing.T) {
	runtime := New()
	store := savecore.New()
	if err := runtime.RegisterModule(SaveModule(store)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	if err := runtime.Execute(context.Background(), fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(`local s=require("dm.save/v1"); assert(s.create("hero", "Hero", "Amazon")); assert(s.select(s.characters()[1].id)); name=s.selected().name`)}}, "test.lua"); err != nil {
		t.Fatal(err)
	}
	selected, ok := store.Selected()
	if !ok || selected.Name != "Hero" {
		t.Fatalf("selected = %#v", selected)
	}
}

func TestSaveModuleCreatesNamedCharacter(t *testing.T) {
	runtime := New()
	store := savecore.New()
	if err := runtime.RegisterModule(SaveModule(store)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	if err := runtime.Execute(context.Background(), fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(`local s=require("dm.save/v1"); id=assert(s.create_named("Iron-Wolf", "paladin", false, true)); assert(s.select(id)); c=s.selected(); assert(not c.expansion and c.hardcore)`)}}, "test.lua"); err != nil {
		t.Fatal(err)
	}
	selected, ok := store.Selected()
	if !ok || selected.ID != "paladin-iron-wolf" || selected.Class != "Paladin" || selected.Expansion || !selected.Hardcore {
		t.Fatalf("selected = %#v", selected)
	}
}

func TestSaveModuleDeletesCharacterByOpaqueID(t *testing.T) {
	runtime := New()
	store := savecore.New(savecore.Character{ID: "hero", Name: "Hero", Class: "Amazon", Level: 1})
	if err := runtime.RegisterModule(SaveModule(store)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	script := `local s=require("dm.save/v1"); assert(s.select("hero")); assert(s.delete("hero")); assert(#s.characters()==0); assert(s.selected()==nil)`
	if err := runtime.Execute(context.Background(), fstest.MapFS{"test.lua": {Data: []byte(script)}}, "test.lua"); err != nil {
		t.Fatal(err)
	}
}

func TestSaveModuleExposesImmutableAppearanceSnapshot(t *testing.T) {
	t.Parallel()

	store := savecore.New(savecore.Character{
		ID: "hero", Name: "Hero", Class: "Amazon", Level: 12,
		Appearance: &savecore.Appearance{
			COF: "hero.cof", Palette: "units.dat", Direction: 3,
			Components: map[string]string{"HD": "head.dcc", "TR": "torso.dcc"},
		},
	})
	runtime := New()
	if err := runtime.RegisterModule(SaveModule(store)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	script := `local c=require("dm.save/v1").characters()[1]
assert(c.appearance.cof=="hero.cof")
assert(c.appearance.palette=="units.dat")
assert(c.appearance.direction==3)
assert(c.appearance.components.HD=="head.dcc")
assert(c.appearance.components.TR=="torso.dcc")`
	if err := runtime.Execute(context.Background(), fstest.MapFS{"test.lua": {Data: []byte(script)}}, "test.lua"); err != nil {
		t.Fatal(err)
	}
}

func TestSaveModuleExposesCharacterStats(t *testing.T) {
	t.Parallel()

	store := savecore.New(savecore.Character{
		ID: "hero", Name: "Hero", Class: "Amazon", Level: 12,
		Stats: &savecore.Stats{Strength: 25, Health: 70, MaxHealth: 80, FireResistance: 15},
	})
	runtime := New()
	if err := runtime.RegisterModule(SaveModule(store)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	script := `local s=require("dm.save/v1").characters()[1].stats
assert(s.strength==25 and s.health==70 and s.max_health==80 and s.fire_resistance==15)`
	if err := runtime.Execute(context.Background(), fstest.MapFS{"test.lua": {Data: []byte(script)}}, "test.lua"); err != nil {
		t.Fatal(err)
	}
}
