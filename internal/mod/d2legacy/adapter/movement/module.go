package movement

import (
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	lua "github.com/yuin/gopher-lua"
)

// RulesModule exposes the same movement resolver to authoritative Lua that the
// owning client's rollback/replay predictor uses in Go.
func RulesModule(catalog Catalog) modruntime.Module {
	return modruntime.Module{
		Name: "d2legacy.movement_rules/v1",
		Help: modruntime.ModuleHelp{
			Summary: "Resolve d2legacy movement intent with the production walk, run, diagonal, and arrival rules.",
			Commands: map[string]modruntime.CommandHelp{
				"velocity":        {Usage: "d2legacy.movement_rules.velocity(x, y, class, payload, velocitypercent?, item_frw?)", Summary: "Return authoritative x/y velocity and whether movement is active."},
				"rates":           {Usage: "d2legacy.movement_rules.rates(class, velocitypercent?, item_frw?)", Summary: "Return effective walk and run rates."},
				"class_facts":     {Usage: "d2legacy.movement_rules.class_facts(class)", Summary: "Return pinned starting stamina, RunDrain, stamina progression terms, and starting Vitality."},
				"is_town":         {Usage: "d2legacy.movement_rules.is_town(level_id)", Summary: "Report whether a level is one of the five target act towns."},
				"stamina":         {Usage: "d2legacy.movement_rules.stamina(current, maximum, run_drain, armor, slower_drain, recovery, running, moving, town, can_recover)", Summary: "Advance one authoritative 25 Hz 8.8 stamina tick."},
				"maximum_stamina": {Usage: "d2legacy.movement_rules.maximum_stamina(class, level, base_vitality, bonus_vitality, flat_maximum, skill_percent, passive_percent, item_per_level)", Summary: "Resolve Expansion 1.14d maximum stamina in 8.8 units."},
				"rescale_stamina": {Usage: "d2legacy.movement_rules.rescale_stamina(current, previous_maximum, new_maximum)", Summary: "Apply the maxstamina source-change callback."},
			},
		},
		Loader: func(state *lua.LState) int {
			module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
				"velocity": func(state *lua.LState) int {
					class := state.CheckString(3)
					rates, found := catalog.Rates(class)
					if !found {
						state.RaiseError("d2legacy movement: class %q has no pinned CharStats velocity", class)
					}
					payload := state.CheckTable(4)
					move := MovePayload{
						X:       luaInteger(payload.RawGetString("x")),
						Y:       luaInteger(payload.RawGetString("y")),
						Running: lua.LVAsBool(payload.RawGetString("running")),
					}
					if target, ok := payload.RawGetString("target").(*lua.LTable); ok {
						move.Target = &MoveTarget{
							X:          luaNumber(target.RawGetString("x")),
							Y:          luaNumber(target.RawGetString("y")),
							StopRadius: luaOptionalNumber(target.RawGetString("stop_radius")),
						}
					}
					modifiers := Modifiers{
						VelocityPercent:        int64(state.OptInt(5, 0)),
						ItemFasterMoveVelocity: int64(state.OptInt(6, 0)),
					}
					resolved := Resolve(gameworld.Point{X: float64(state.CheckNumber(1)), Y: float64(state.CheckNumber(2))}, move, rates, modifiers)
					state.Push(lua.LNumber(resolved.Velocity.X))
					state.Push(lua.LNumber(resolved.Velocity.Y))
					state.Push(lua.LBool(resolved.Moving))
					return 3
				},
				"rates": func(state *lua.LState) int {
					rates, found := catalog.Rates(state.CheckString(1))
					if !found {
						state.RaiseError("d2legacy movement: class has no pinned CharStats movement facts")
					}
					effective := EffectiveRates(rates, Modifiers{
						VelocityPercent:        int64(state.OptInt(2, 0)),
						ItemFasterMoveVelocity: int64(state.OptInt(3, 0)),
					})
					state.Push(lua.LNumber(effective.Walk))
					state.Push(lua.LNumber(effective.Run))
					return 2
				},
				"class_facts": func(state *lua.LState) int {
					rates, found := catalog.Rates(state.CheckString(1))
					if !found {
						state.RaiseError("d2legacy movement: class has no pinned CharStats movement facts")
					}
					state.Push(lua.LNumber(rates.StartingStamina))
					state.Push(lua.LNumber(rates.RunDrain))
					state.Push(lua.LNumber(rates.StaminaPerLevel))
					state.Push(lua.LNumber(rates.StaminaPerVitality))
					state.Push(lua.LNumber(rates.StartingVitality))
					return 5
				},
				"maximum_stamina": func(state *lua.LState) int {
					rates, found := catalog.Rates(state.CheckString(1))
					if !found {
						state.RaiseError("d2legacy movement: class has no pinned CharStats stamina facts")
					}
					value := MaximumStamina(rates, int64(state.CheckInt(2)), int64(state.CheckInt(3)), StaminaMaximumSources{
						BonusVitality: int64(state.OptInt(4, 0)), FlatMaximum: int64(state.OptInt(5, 0)),
						SkillStaminaPercent: int64(state.OptInt(6, 0)), SkillPassiveStaminaPercent: int64(state.OptInt(7, 0)),
						ItemStaminaPerLevel: int64(state.OptInt(8, 0)),
					})
					state.Push(lua.LNumber(value))
					return 1
				},
				"rescale_stamina": func(state *lua.LState) int {
					state.Push(lua.LNumber(RescaleCurrentStamina(int64(state.CheckInt(1)), int64(state.CheckInt(2)), int64(state.CheckInt(3)))))
					return 1
				},
				"is_town": func(state *lua.LState) int {
					state.Push(lua.LBool(IsTownLevel(int64(state.CheckInt(1)))))
					return 1
				},
				"stamina": func(state *lua.LState) int {
					resolved := AdvanceStamina(StaminaTick{
						CurrentRaw: int64(state.CheckInt(1)), MaximumRaw: int64(state.CheckInt(2)),
						RunDrain: int64(state.CheckInt(3)), ArmorRunDrain: int64(state.CheckInt(4)),
						StaminaDrainPercent: int64(state.CheckInt(5)), RecoveryBonus: int64(state.CheckInt(6)),
						Running: lua.LVAsBool(state.Get(7)), Moving: lua.LVAsBool(state.Get(8)),
						InTown: lua.LVAsBool(state.Get(9)), CanRecover: lua.LVAsBool(state.Get(10)),
					})
					state.Push(lua.LNumber(resolved.CurrentRaw))
					state.Push(lua.LBool(resolved.ForceWalk))
					return 2
				},
			})
			state.Push(module)
			return 1
		},
	}
}

func luaInteger(value lua.LValue) int {
	number, _ := value.(lua.LNumber)
	return int(number)
}

func luaNumber(value lua.LValue) float64 {
	number, _ := value.(lua.LNumber)
	return float64(number)
}

func luaOptionalNumber(value lua.LValue) float64 {
	if value == lua.LNil {
		return 0
	}
	return luaNumber(value)
}
