package movement

import (
	"math"
	"testing"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

func TestResolveUsesProductionMovementRules(t *testing.T) {
	rates := ClassRates{Walk: 6, Run: 9}
	diagonal := Resolve(gameworld.Point{}, MovePayload{X: 1, Y: 1, Running: true}, rates)
	expected := rates.Run / math.Sqrt2
	if math.Abs(diagonal.Velocity.X-expected) > 1e-12 || math.Abs(diagonal.Velocity.Y-expected) > 1e-12 || !diagonal.Moving {
		t.Fatalf("diagonal movement = %+v", diagonal)
	}
	arrived := Resolve(gameworld.Point{X: 10, Y: 20}, MovePayload{Target: &MoveTarget{X: 10.1, Y: 20}}, rates)
	if arrived.Moving || arrived.Velocity != (gameworld.Point{}) {
		t.Fatalf("arrival movement = %+v", arrived)
	}
	target := Resolve(gameworld.Point{X: 10, Y: 20}, MovePayload{Target: &MoveTarget{X: 13, Y: 24}}, rates)
	if math.Abs(target.Velocity.X-3.6) > 1e-12 || math.Abs(target.Velocity.Y-4.8) > 1e-12 {
		t.Fatalf("target velocity = %+v, want {3.6 4.8}", target.Velocity)
	}
}

func TestEffectiveRatesUseD2ItemFRWDiminishingReturnsAndSharedVelocityChannel(t *testing.T) {
	base := ClassRates{Walk: 6, Run: 9}
	effective := EffectiveRates(base, Modifiers{ItemFasterMoveVelocity: 30})
	if effective.Walk != 7.5 || effective.Run != 10.5 {
		t.Fatalf("30 item FRW rates = %+v, want 7.5/10.5", effective)
	}
	slowed := EffectiveRates(base, Modifiers{VelocityPercent: -90})
	if math.Abs(slowed.Walk-1.5) > 1e-12 || math.Abs(slowed.Run-3.6) > 1e-12 {
		t.Fatalf("minimum velocity rates = %+v, want walk floor 1.5 and run 3.6", slowed)
	}
}

func TestEffectiveVelocityPercentAlsoOwnsWalkRunAnimationRate(t *testing.T) {
	base := ClassRates{Walk: 6, Run: 9}
	if got := EffectiveVelocityPercent(base, false, Modifiers{}); got != 100 {
		t.Fatalf("walk percentage = %d", got)
	}
	if got := EffectiveVelocityPercent(base, true, Modifiers{}); got != 150 {
		t.Fatalf("run percentage = %d", got)
	}
	if got := MovementAnimationRate(base, false, Modifiers{ItemFasterMoveVelocity: 100}); got != 340 {
		t.Fatalf("100 item FRW walk animation = %d, want 340", got)
	}
	if got := MovementAnimationRate(base, true, Modifiers{}); got != 151 {
		t.Fatalf("base run animation = %d, want 151", got)
	}
	if got := MovementAnimationRate(base, false, Modifiers{VelocityPercent: -90}); got != 53 {
		t.Fatalf("floored walk animation = %d, want 53", got)
	}
}

func TestAdvanceStaminaUsesFixedPointDrainRecoveryAndTownRules(t *testing.T) {
	drained := AdvanceStamina(StaminaTick{CurrentRaw: 256, MaximumRaw: 256, RunDrain: 20, Running: true, Moving: true})
	if drained.CurrentRaw != 216 || drained.ForceWalk {
		t.Fatalf("ordinary run tick = %+v, want 216 raw", drained)
	}
	heavy := AdvanceStamina(StaminaTick{CurrentRaw: 100, MaximumRaw: 256, RunDrain: 20, ArmorRunDrain: 10, StaminaDrainPercent: 25, Running: true, Moving: true})
	if heavy.CurrentRaw != 40 {
		t.Fatalf("heavy 25%% slower-drain tick = %+v, want 40 raw", heavy)
	}
	stopped := AdvanceStamina(StaminaTick{CurrentRaw: 20, MaximumRaw: 256, RunDrain: 20, Running: true, Moving: true})
	if stopped.CurrentRaw != 0 || !stopped.ForceWalk {
		t.Fatalf("exhaustion tick = %+v", stopped)
	}
	idle := AdvanceStamina(StaminaTick{CurrentRaw: 2560, MaximumRaw: 84 * 256, CanRecover: true})
	if idle.CurrentRaw != 2644 {
		t.Fatalf("idle recovery = %+v, want 2644 raw", idle)
	}
	town := AdvanceStamina(StaminaTick{CurrentRaw: 2560, MaximumRaw: 84 * 256, Running: true, Moving: true, InTown: true, CanRecover: true})
	if town.CurrentRaw != 2602 {
		t.Fatalf("town motion recovery = %+v, want 2602 raw", town)
	}
}

func TestMaximumStaminaUsesExpansionFixedPointDependencyGraph(t *testing.T) {
	rates := ClassRates{StartingVitality: 20, StartingStamina: 84, StaminaPerLevel: 4, StaminaPerVitality: 4}
	base := MaximumStamina(rates, 10, 29, StaminaMaximumSources{})
	if base != 102*256 {
		t.Fatalf("level/vitality maximum = %d, want %d", base, 102*256)
	}
	withSources := MaximumStamina(rates, 10, 29, StaminaMaximumSources{
		BonusVitality: 5, FlatMaximum: 10, SkillStaminaPercent: 25,
		SkillPassiveStaminaPercent: 10, ItemStaminaPerLevel: 8,
	})
	if withSources != 39997 {
		t.Fatalf("source-derived maximum = %d, want 39997", withSources)
	}
}

func TestRescaleCurrentStaminaMatchesMaxStatCallback(t *testing.T) {
	if got := RescaleCurrentStamina(42*256, 84*256, 102*256); got != 51*256 {
		t.Fatalf("rescaled current = %d", got)
	}
	if got := RescaleCurrentStamina(0, 84*256, 102*256); got != 0 {
		t.Fatalf("zero current = %d", got)
	}
	if got := RescaleCurrentStamina(1, 102*256, 84*256); got != 1 {
		t.Fatalf("minimum current = %d", got)
	}
}

func TestTownLevelsAreTheFiveExpansion114dActTowns(t *testing.T) {
	for _, level := range []int64{1, 45, 80, 108, 114} {
		if !IsTownLevel(level) {
			t.Fatalf("level %d was not recognized as town", level)
		}
	}
	for _, level := range []int64{0, 2, 44, 79, 107, 115} {
		if IsTownLevel(level) {
			t.Fatalf("level %d was recognized as town", level)
		}
	}
}
