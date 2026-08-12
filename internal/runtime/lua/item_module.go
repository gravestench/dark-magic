package modruntime

import (
	gameitem "github.com/gravestench/dark-magic/internal/game/item"
	lua "github.com/yuin/gopher-lua"
)

// ItemModule exposes copied item facts and queues move intents. Scripts never
// receive the mutable authority or its internal maps.
func ItemModule(controller *gameitem.Controller) Module {
	return Module{Name: "engine.items/v1", Help: documentedModule("Queue item command intents for fixed-tick admission.", map[string]CommandHelp{
		"move":              commandHelp("engine.items.move(item_id, destination[, place_held])", "Queue a move or held-item grid placement for the next simulation tick."),
		"select_weapon_set": commandHelp("engine.items.select_weapon_set(set)", "Queue selection of alternate hand-equipment set 0 or 1."),
		"sell_held":         commandHelp("engine.items.sell_held(item_id, vendor, category)", "Queue priced sale and authority-owned vendor catalog arrangement."),
		"buy_to_held":       commandHelp("engine.items.buy_to_held(item_id, vendor)", "Queue priced purchase of vendor stock into the authoritative held container."),
		"complete_service":  commandHelp("engine.items.complete_service(service)", "Queue a server-defined quest or vendor service transaction."),
	}), Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"move": func(state *lua.LState) int {
				payload := gameitem.MovePayload{ItemID: state.CheckString(1), Destination: checkPlacement(state, 2), PlaceHeld: state.OptBool(3, false)}
				if err := controller.Move(payload); err != nil {
					state.RaiseError("%v", err)
				}
				return 0
			},
			"select_weapon_set": func(state *lua.LState) int {
				if err := controller.SelectWeaponSet(state.CheckInt(1)); err != nil {
					state.RaiseError("%v", err)
				}
				return 0
			},
			"sell_held": func(state *lua.LState) int {
				if err := controller.SellHeld(state.CheckString(1), state.CheckString(2), state.CheckString(3)); err != nil {
					state.RaiseError("%v", err)
				}
				return 0
			},
			"buy_to_held": func(state *lua.LState) int {
				if err := controller.BuyToHeld(state.CheckString(1), state.CheckString(2)); err != nil {
					state.RaiseError("%v", err)
				}
				return 0
			},
			"complete_service": func(state *lua.LState) int {
				if err := controller.CompleteService(state.CheckString(1)); err != nil {
					state.RaiseError("%v", err)
				}
				return 0
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}

func checkPlacement(state *lua.LState, index int) gameitem.Placement {
	table := state.CheckTable(index)
	return gameitem.Placement{
		Container: gameitem.Container(luaString(table, "container")),
		X:         luaInt(table, "x"),
		Y:         luaInt(table, "y"),
		Slot:      luaString(table, "slot"),
		BeltSlot:  luaInt(table, "belt_slot"),
		WeaponSet: luaInt(table, "weapon_set"),
	}
}

func luaString(table *lua.LTable, name string) string {
	if value, ok := table.RawGetString(name).(lua.LString); ok {
		return string(value)
	}
	return ""
}

func luaInt(table *lua.LTable, name string) int {
	if value, ok := table.RawGetString(name).(lua.LNumber); ok {
		return int(value)
	}
	return 0
}
