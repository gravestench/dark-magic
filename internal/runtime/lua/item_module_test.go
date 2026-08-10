package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	gameitem "github.com/gravestench/dark-magic/internal/game/item"
)

func TestItemModuleReturnsCopiesAndQueuesIntent(t *testing.T) {
	state, err := gameitem.NewState(gameitem.Layout{Grids: map[gameitem.Container]gameitem.Grid{gameitem.ContainerInventory: {Width: 10, Height: 4}}, BeltCapacity: 4}, []gameitem.Item{{ID: "potion", Code: "hp1", Width: 1, Height: 1}}, map[string]gameitem.Placement{"potion": {Container: gameitem.ContainerInventory}})
	if err != nil {
		t.Fatal(err)
	}
	authority := gameitem.NewAuthority()
	if err := authority.Register("alice", state); err != nil {
		t.Fatal(err)
	}
	controller := &gameitem.Controller{}
	runtime := New()
	if err := runtime.RegisterModule(ItemModule(authority, controller, "alice")); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(ctx)
	script := fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(`
local items=require("dm.items/v1")
local snapshot=assert(items.snapshot())
assert(snapshot.belt_capacity==4 and #snapshot.items==1 and snapshot.items[1].container=="inventory")
items.move("potion", {container="held"})
`)}}
	if err := runtime.Execute(ctx, script, "test.lua"); err != nil {
		t.Fatal(err)
	}
	source, _ := gameitem.NewSource(controller, "alice")
	if commands := source.Commands(1); len(commands) != 1 {
		t.Fatalf("queued commands = %#v", commands)
	}
}
