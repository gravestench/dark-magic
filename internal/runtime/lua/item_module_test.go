package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	gameitem "github.com/gravestench/dark-magic/internal/game/item"
)

func TestItemModuleReturnsCopiesAndQueuesIntent(t *testing.T) {
	state, err := gameitem.NewState(gameitem.Layout{Grids: map[gameitem.Container]gameitem.Grid{gameitem.ContainerInventory: {Width: 10, Height: 4}}, BeltCapacity: 4}, []gameitem.Item{{ID: "potion", Code: "hp1", Width: 1, Height: 1, Presentation: gameitem.Presentation{Composite: map[string]string{"RH": "ssd"}, WeaponClass: "1HS"}}}, map[string]gameitem.Placement{"potion": {Container: gameitem.ContainerInventory}})
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
assert(snapshot.belt_capacity==4 and snapshot.active_weapon_set==0)
assert(#snapshot.items==1 and snapshot.items[1].container=="inventory" and snapshot.items[1].weapon_set==0)
assert(snapshot.items[1].weapon_class=="1HS" and snapshot.items[1].composite.RH=="ssd")
items.move("potion", {container="held"})
items.select_weapon_set(1)
items.sell_held("potion", "Akara", "misc")
items.buy_to_held("stock", "Akara")
items.complete_service("socket")
`)}}
	if err := runtime.Execute(ctx, script, "test.lua"); err != nil {
		t.Fatal(err)
	}
	source, _ := gameitem.NewSource(controller, "alice")
	if commands := source.Commands(1); len(commands) != 5 || commands[0].Kind != gameitem.MoveCommand || commands[1].Kind != gameitem.WeaponSetCommand || commands[2].Kind != gameitem.VendorSellCommand || commands[3].Kind != gameitem.VendorBuyCommand || commands[4].Kind != gameitem.ServiceCommand {
		t.Fatalf("queued commands = %#v", commands)
	}
}
