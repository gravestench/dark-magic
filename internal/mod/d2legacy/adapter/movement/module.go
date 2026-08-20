package movement

import (
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	lua "github.com/yuin/gopher-lua"
)

const movementRulesModuleName = "d2legacy.movement_rules/v1"

// movementRulesBridge owns the catalog captured by Lua callbacks without exposing mutable package state.
type movementRulesBridge struct {
	catalog Catalog
}

// RulesModule exposes the same movement resolver to authoritative Lua that client rollback prediction uses in Go.
// Sharing the implementation prevents authority and prediction from drifting onto different movement formulas.
func RulesModule(catalog Catalog) modruntime.Module {
	bridge := movementRulesBridge{catalog: catalog}

	return modruntime.Module{
		Name:   movementRulesModuleName,
		Help:   movementRulesHelp(),
		Loader: bridge.load,
	}
}

// movementRulesHelp documents the Lua boundary independently from callback wiring so additions remain reviewable.
func movementRulesHelp() modruntime.ModuleHelp {
	return modruntime.ModuleHelp{
		Summary: "Resolve d2legacy movement intent with the production walk, run, diagonal, and arrival rules.",
		Commands: map[string]modruntime.CommandHelp{
			"animation_rate": {
				Usage:   "d2legacy.movement_rules.animation_rate(class, running, velocitypercent?, item_frw?)",
				Summary: "Return the Expansion walk/run animation rate after the shared velocity ordering.",
			},
			"by_time_adjustment": {
				Usage:   "d2legacy.movement_rules.by_time_adjustment(packed, base_time)",
				Summary: "Evaluate a Properties func 18 operand against the 360-unit environment cycle.",
			},
			"class_facts": {
				Usage: "d2legacy.movement_rules.class_facts(class)",
				Summary: "Return pinned starting stamina, RunDrain, stamina progression terms, " +
					"and starting Vitality.",
			},
			"is_town": {
				Usage:   "d2legacy.movement_rules.is_town(level_id)",
				Summary: "Report whether a level is one of the five target act towns.",
			},
			"maximum_stamina": {
				Usage: "d2legacy.movement_rules.maximum_stamina(class, level, base_vitality, bonus_vitality, " +
					"flat_maximum, skill_percent, passive_percent, item_per_level, item_by_time)",
				Summary: "Resolve Expansion 1.14d maximum stamina in 8.8 units.",
			},
			"rates": {
				Usage:   "d2legacy.movement_rules.rates(class, velocitypercent?, item_frw?)",
				Summary: "Return effective walk and run rates.",
			},
			"rescale_stamina": {
				Usage:   "d2legacy.movement_rules.rescale_stamina(current, previous_maximum, new_maximum)",
				Summary: "Apply the maxstamina source-change callback.",
			},
			"stamina": {
				Usage: "d2legacy.movement_rules.stamina(current, maximum, run_drain, armor, slower_drain, " +
					"recovery, running, moving, town, can_recover)",
				Summary: "Advance one authoritative 25 Hz 8.8 stamina tick.",
			},
			"velocity": {
				Usage:   "d2legacy.movement_rules.velocity(x, y, class, payload, velocitypercent?, item_frw?)",
				Summary: "Return authoritative x/y velocity and whether movement is active.",
			},
		},
	}
}

// load registers every callback on one table, preserving the module's established Lua names and return arities.
func (bridge movementRulesBridge) load(state *lua.LState) int {
	module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
		"animation_rate":     bridge.animationRate,
		"by_time_adjustment": bridge.byTimeAdjustment,
		"class_facts":        bridge.classFacts,
		"is_town":            bridge.isTown,
		"maximum_stamina":    bridge.maximumStamina,
		"rates":              bridge.rates,
		"rescale_stamina":    bridge.rescaleStamina,
		"stamina":            bridge.stamina,
		"velocity":           bridge.velocity,
	})
	state.Push(module)

	return 1
}

// byTimeAdjustment decodes the two required integer operands and returns exactly one Lua number.
func (movementRulesBridge) byTimeAdjustment(state *lua.LState) int {
	packed := int64(state.CheckInt(1))
	baseTime := int64(state.CheckInt(2))
	state.Push(lua.LNumber(ByTimeAdjustment(packed, baseTime)))

	return 1
}

// animationRate applies the shared velocity modifiers so Lua animation timing stays aligned with path velocity.
func (bridge movementRulesBridge) animationRate(state *lua.LState) int {
	rates := bridge.ratesOrRaise(state, 1, "d2legacy movement: class has no pinned CharStats movement facts")
	running := lua.LVAsBool(state.Get(2))
	modifiers := luaModifiers(state, 3, 4)

	state.Push(lua.LNumber(MovementAnimationRate(rates, running, modifiers)))

	return 1
}

// velocity converts the Lua payload into the same Go value consumed by authoritative and predicted movement.
func (bridge movementRulesBridge) velocity(state *lua.LState) int {
	class := state.CheckString(3)

	rates, found := bridge.catalog.Rates(class)
	if !found {
		state.RaiseError("d2legacy movement: class %q has no pinned CharStats velocity", class)
	}

	payload := movePayloadFromLua(state.CheckTable(4))
	position := gameworld.Point{
		X: float64(state.CheckNumber(1)),
		Y: float64(state.CheckNumber(2)),
	}
	resolved := Resolve(position, payload, rates, luaModifiers(state, 5, 6))

	state.Push(lua.LNumber(resolved.Velocity.X))
	state.Push(lua.LNumber(resolved.Velocity.Y))
	state.Push(lua.LBool(resolved.Moving))

	return 3
}

// rates returns effective movement rates after applying Lua's optional stat channels in their established order.
func (bridge movementRulesBridge) rates(state *lua.LState) int {
	rates := bridge.ratesOrRaise(state, 1, "d2legacy movement: class has no pinned CharStats movement facts")
	effective := EffectiveRates(rates, luaModifiers(state, 2, 3))

	state.Push(lua.LNumber(effective.Walk))
	state.Push(lua.LNumber(effective.Run))

	return 2
}

// classFacts returns the five pinned record fields in the legacy positional order expected by Lua callers.
func (bridge movementRulesBridge) classFacts(state *lua.LState) int {
	rates := bridge.ratesOrRaise(state, 1, "d2legacy movement: class has no pinned CharStats movement facts")

	state.Push(lua.LNumber(rates.StartingStamina))
	state.Push(lua.LNumber(rates.RunDrain))
	state.Push(lua.LNumber(rates.StaminaPerLevel))
	state.Push(lua.LNumber(rates.StaminaPerVitality))
	state.Push(lua.LNumber(rates.StartingVitality))

	return 5
}

// maximumStamina translates positional Lua operands into named sources before applying fixed-point rules.
func (bridge movementRulesBridge) maximumStamina(state *lua.LState) int {
	rates := bridge.ratesOrRaise(state, 1, "d2legacy movement: class has no pinned CharStats stamina facts")
	sources := StaminaMaximumSources{
		BonusVitality:              int64(state.OptInt(4, 0)),
		FlatMaximum:                int64(state.OptInt(5, 0)),
		SkillStaminaPercent:        int64(state.OptInt(6, 0)),
		SkillPassiveStaminaPercent: int64(state.OptInt(7, 0)),
		ItemStaminaPerLevel:        int64(state.OptInt(8, 0)),
		ItemStaminaByTime:          int64(state.OptInt(9, 0)),
	}
	value := MaximumStamina(rates, int64(state.CheckInt(2)), int64(state.CheckInt(3)), sources)

	state.Push(lua.LNumber(value))

	return 1
}

// rescaleStamina preserves the legacy callback's three integer inputs and single integer result.
func (movementRulesBridge) rescaleStamina(state *lua.LState) int {
	current := int64(state.CheckInt(1))
	previousMaximum := int64(state.CheckInt(2))
	newMaximum := int64(state.CheckInt(3))
	state.Push(lua.LNumber(RescaleCurrentStamina(current, previousMaximum, newMaximum)))

	return 1
}

// isTown exposes the pinned act-town classification without duplicating level IDs in Lua.
func (movementRulesBridge) isTown(state *lua.LState) int {
	state.Push(lua.LBool(IsTownLevel(int64(state.CheckInt(1)))))
	return 1
}

// stamina maps positional Lua tick state into named fields so drain and recovery inputs remain distinguishable.
func (movementRulesBridge) stamina(state *lua.LState) int {
	resolved := AdvanceStamina(StaminaTick{
		CurrentRaw:          int64(state.CheckInt(1)),
		MaximumRaw:          int64(state.CheckInt(2)),
		RunDrain:            int64(state.CheckInt(3)),
		ArmorRunDrain:       int64(state.CheckInt(4)),
		StaminaDrainPercent: int64(state.CheckInt(5)),
		RecoveryBonus:       int64(state.CheckInt(6)),
		Running:             lua.LVAsBool(state.Get(7)),
		Moving:              lua.LVAsBool(state.Get(8)),
		InTown:              lua.LVAsBool(state.Get(9)),
		CanRecover:          lua.LVAsBool(state.Get(10)),
	})

	state.Push(lua.LNumber(resolved.CurrentRaw))
	state.Push(lua.LBool(resolved.ForceWalk))

	return 2
}

// ratesOrRaise centralizes catalog lookup while preserving each callback's established Lua error message.
func (bridge movementRulesBridge) ratesOrRaise(state *lua.LState, index int, message string) ClassRates {
	rates, found := bridge.catalog.Rates(state.CheckString(index))
	if !found {
		state.RaiseError("%s", message)
	}

	return rates
}

// luaModifiers reads the adjacent optional velocity channels used by all movement-rate callbacks.
func luaModifiers(state *lua.LState, velocityIndex, itemIndex int) Modifiers {
	return Modifiers{
		VelocityPercent:        int64(state.OptInt(velocityIndex, 0)),
		ItemFasterMoveVelocity: int64(state.OptInt(itemIndex, 0)),
	}
}

// movePayloadFromLua preserves missing numeric fields as zero, matching the previous permissive table conversion.
func movePayloadFromLua(table *lua.LTable) MovePayload {
	payload := MovePayload{
		X:       luaInteger(table.RawGetString("x")),
		Y:       luaInteger(table.RawGetString("y")),
		Running: lua.LVAsBool(table.RawGetString("running")),
	}

	if target, ok := table.RawGetString("target").(*lua.LTable); ok {
		payload.Target = &MoveTarget{
			X:          luaNumber(target.RawGetString("x")),
			Y:          luaNumber(target.RawGetString("y")),
			StopRadius: luaOptionalNumber(target.RawGetString("stop_radius")),
		}
	}

	return payload
}

// luaInteger returns zero for absent or non-number table fields, retaining Lua table compatibility.
func luaInteger(value lua.LValue) int {
	number, _ := value.(lua.LNumber)
	return int(number)
}

// luaNumber returns zero for absent or non-number table fields, retaining Lua table compatibility.
func luaNumber(value lua.LValue) float64 {
	number, _ := value.(lua.LNumber)
	return float64(number)
}

// luaOptionalNumber distinguishes an omitted optional value while preserving zero as its legacy default.
func luaOptionalNumber(value lua.LValue) float64 {
	if value == lua.LNil {
		return 0
	}

	return luaNumber(value)
}
