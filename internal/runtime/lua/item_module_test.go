package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	gameitem "github.com/gravestench/dark-magic/internal/game/item"
)

func TestItemModuleQueuesIntent(t *testing.T) {
	controller := &gameitem.Controller{}
	runtime := New()
	if err := runtime.RegisterModule(ItemModule(controller)); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(ctx)
	script := fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(`
local items=require("engine.items/v1")
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
