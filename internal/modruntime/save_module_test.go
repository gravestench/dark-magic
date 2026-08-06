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
