package modruntime

import (
	"sort"

	gameitem "github.com/gravestench/dark-magic/internal/game/item"
	lua "github.com/yuin/gopher-lua"
)

// ItemModule exposes copied item facts and queues move intents. Scripts never
// receive the mutable authority or its internal maps.
func ItemModule(authority *gameitem.Authority, controller *gameitem.Controller, owner string) Module {
	return Module{Name: "dm.items/v1", Help: documentedModule("Inspect authoritative item containers and request fixed-tick moves.", map[string]CommandHelp{
		"snapshot":          commandHelp("dm.items.snapshot()", "Return copied item identities, placements, container layout, and active weapon set."),
		"move":              commandHelp("dm.items.move(item_id, destination[, place_held])", "Queue a move or held-item grid placement for the next simulation tick."),
		"select_weapon_set": commandHelp("dm.items.select_weapon_set(set)", "Queue selection of alternate hand-equipment set 0 or 1."),
		"sell_held":         commandHelp("dm.items.sell_held(item_id, category)", "Queue sale and authority-owned vendor catalog arrangement."),
		"buy_to_held":       commandHelp("dm.items.buy_to_held(item_id)", "Queue purchase of vendor stock into the authoritative held container."),
	}), Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"snapshot": func(state *lua.LState) int {
				layout, items, placements, err := authority.Snapshot(owner)
				if err != nil {
					return pushLuaError(state, err)
				}
				state.Push(itemSnapshotTable(state, layout, items, placements))
				return 1
			},
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
				if err := controller.SellHeld(state.CheckString(1), state.CheckString(2)); err != nil {
					state.RaiseError("%v", err)
				}
				return 0
			},
			"buy_to_held": func(state *lua.LState) int {
				if err := controller.BuyToHeld(state.CheckString(1)); err != nil {
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

func itemSnapshotTable(state *lua.LState, layout gameitem.Layout, items map[string]gameitem.Item, placements map[string]gameitem.Placement) *lua.LTable {
	result := state.NewTable()
	result.RawSetString("belt_capacity", lua.LNumber(layout.BeltCapacity))
	result.RawSetString("active_weapon_set", lua.LNumber(layout.ActiveWeaponSet))
	vendorGrid := state.NewTable()
	vendorGrid.RawSetString("width", lua.LNumber(layout.VendorGrid.Width))
	vendorGrid.RawSetString("height", lua.LNumber(layout.VendorGrid.Height))
	result.RawSetString("vendor_grid", vendorGrid)
	grids := state.NewTable()
	for container, grid := range layout.Grids {
		entry := state.NewTable()
		entry.RawSetString("width", lua.LNumber(grid.Width))
		entry.RawSetString("height", lua.LNumber(grid.Height))
		grids.RawSetString(string(container), entry)
	}
	result.RawSetString("grids", grids)
	entries := state.NewTable()
	ids := make([]string, 0, len(items))
	for id := range items {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		candidate, placement := items[id], placements[id]
		entry := state.NewTable()
		entry.RawSetString("id", lua.LString(id))
		entry.RawSetString("code", lua.LString(candidate.Code))
		entry.RawSetString("width", lua.LNumber(candidate.Width))
		entry.RawSetString("height", lua.LNumber(candidate.Height))
		entry.RawSetString("inventory_dc6", lua.LString(candidate.Presentation.InventoryDC6))
		entry.RawSetString("world_dc6", lua.LString(candidate.Presentation.WorldDC6))
		entry.RawSetString("container", lua.LString(placement.Container))
		entry.RawSetString("x", lua.LNumber(placement.X))
		entry.RawSetString("y", lua.LNumber(placement.Y))
		entry.RawSetString("slot", lua.LString(placement.Slot))
		entry.RawSetString("belt_slot", lua.LNumber(placement.BeltSlot))
		entry.RawSetString("weapon_set", lua.LNumber(placement.WeaponSet))
		entry.RawSetString("page", lua.LNumber(placement.Page))
		entries.Append(entry)
	}
	result.RawSetString("items", entries)
	return result
}
