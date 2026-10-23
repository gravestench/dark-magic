package movement

import (
	"math"
	"testing"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

func TestResolveMatchesOwned114dStraightAxisSteps(t *testing.T) {
	base := ClassRates{Walk: 6, Run: 9}
	for _, tc := range []struct {
		name     string
		running  bool
		modifier int64
		stepRaw  int64
	}{
		{"walk", false, 0, 24576},
		{"run", true, 0, 36864},
		{"burst_rank_one_walk", false, 23, 30224},
		{"burst_rank_one_run", true, 23, 42512},
		{"burst_rank_twenty_walk", false, 61, 39552},
		{"burst_rank_twenty_run", true, 61, 51840},
	} {
		t.Run(tc.name, func(t *testing.T) {
			payload := MovePayload{X: 1, Running: tc.running}
			got := Resolve(gameworld.Point{}, payload, base, 1.0/25, Modifiers{VelocityPercent: tc.modifier})
			if step := got.Velocity.X / 25 * 65536; step != float64(tc.stepRaw) {
				t.Fatalf("straight-axis step = %v raw, want %d", step, tc.stepRaw)
			}
			if tc.modifier == 0 && Resolve(gameworld.Point{}, payload, base, 1.0/25) != got {
				t.Fatal("omitted modifiers changed the movement rate")
			}
		})
	}
}

// TestResolveUsesProductionMovementRules protects diagonal normalization, arrival, and target-vector resolution.
func TestResolveUsesProductionMovementRules(t *testing.T) {
	rates := ClassRates{Walk: 6, Run: 9}
	diagonal := Resolve(gameworld.Point{}, MovePayload{X: 1, Y: 1, Running: true}, rates, 1.0/25)

	expected := 14.0625 / math.Sqrt2
	if math.Abs(diagonal.Velocity.X-expected) > 1e-12 ||
		math.Abs(diagonal.Velocity.Y-expected) > 1e-12 ||
		!diagonal.Moving {
		t.Fatalf("diagonal movement = %+v", diagonal)
	}

	arrived := Resolve(gameworld.Point{X: 10, Y: 20}, MovePayload{Target: &MoveTarget{X: 10.1, Y: 20}}, rates, 1.0/25)
	if arrived.Moving || arrived.Velocity != (gameworld.Point{}) {
		t.Fatalf("arrival movement = %+v", arrived)
	}

	target := Resolve(gameworld.Point{X: 10, Y: 20}, MovePayload{Target: &MoveTarget{X: 13, Y: 24}}, rates, 1.0/25)
	if math.Abs(target.Velocity.X-5.625) > 1e-12 || math.Abs(target.Velocity.Y-7.5) > 1e-12 {
		t.Fatalf("target velocity = %+v, want {5.625 7.5}", target.Velocity)
	}
}

func TestResolveSettlesCloseTargetsAtTheActualStep(t *testing.T) {
	for _, elapsed := range []float64{.04, .08} {
		payload := MovePayload{Running: true, Target: &MoveTarget{X: 10.3, Y: 20}}
		position := gameworld.Point{X: 10, Y: 20}
		for range 3 {
			resolved := Resolve(position, payload, ClassRates{Walk: 6, Run: 9}, elapsed)
			position.X += resolved.Velocity.X * elapsed
			position.Y += resolved.Velocity.Y * elapsed
		}
		if position.X != 10.3 || position.Y != 20 {
			t.Fatalf("target at step %v settled at %+v", elapsed, position)
		}
	}
}

// TestEffectiveRatesUseD2ItemFRWDiminishingReturnsAndSharedVelocityChannel pins modifier ordering and floors.
func TestEffectiveRatesUseD2ItemFRWDiminishingReturnsAndSharedVelocityChannel(t *testing.T) {
	base := ClassRates{Walk: 6, Run: 9}

	effective := EffectiveRates(base, Modifiers{ItemFasterMoveVelocity: 30})
	if effective.Walk != 11.71875 || effective.Run != 16.40625 {
		t.Fatalf("30 item FRW rates = %+v, want 11.71875/16.40625", effective)
	}

	slowed := EffectiveRates(base, Modifiers{VelocityPercent: -90})
	if slowed.Walk != 2.34375 || slowed.Run != 5.621337890625 {
		t.Fatalf("minimum velocity rates = %+v, want walk floor 2.34375 and run 5.621337890625", slowed)
	}
}

// TestEffectiveVelocityPercentAlsoOwnsWalkRunAnimationRate keeps animation and path velocity on one percentage model.
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

// TestAdvanceStaminaUsesFixedPointDrainRecoveryAndTownRules protects the mutually exclusive tick outcomes.
func TestAdvanceStaminaUsesFixedPointDrainRecoveryAndTownRules(t *testing.T) {
	drained := AdvanceStamina(StaminaTick{
		CurrentRaw: 256,
		MaximumRaw: 256,
		RunDrain:   20,
		Running:    true,
		Moving:     true,
	})
	if drained.CurrentRaw != 216 || drained.ForceWalk {
		t.Fatalf("ordinary run tick = %+v, want 216 raw", drained)
	}

	heavy := AdvanceStamina(StaminaTick{
		CurrentRaw:          100,
		MaximumRaw:          256,
		RunDrain:            20,
		ArmorRunDrain:       10,
		StaminaDrainPercent: 25,
		Running:             true,
		Moving:              true,
	})
	if heavy.CurrentRaw != 40 {
		t.Fatalf("heavy 25%% slower-drain tick = %+v, want 40 raw", heavy)
	}

	stopped := AdvanceStamina(StaminaTick{
		CurrentRaw: 20,
		MaximumRaw: 256,
		RunDrain:   20,
		Running:    true,
		Moving:     true,
	})
	if stopped.CurrentRaw != 0 || !stopped.ForceWalk {
		t.Fatalf("exhaustion tick = %+v", stopped)
	}

	idle := AdvanceStamina(StaminaTick{CurrentRaw: 2560, MaximumRaw: 84 * 256, CanRecover: true})
	if idle.CurrentRaw != 2644 {
		t.Fatalf("idle recovery = %+v, want 2644 raw", idle)
	}

	town := AdvanceStamina(StaminaTick{
		CurrentRaw: 2560,
		MaximumRaw: 84 * 256,
		Running:    true,
		Moving:     true,
		InTown:     true,
		CanRecover: true,
	})
	if town.CurrentRaw != 2602 {
		t.Fatalf("town motion recovery = %+v, want 2602 raw", town)
	}
}

func TestAdvanceStaminaMatchesSigned32DrainArithmetic(t *testing.T) {
	for _, tc := range []struct {
		name        string
		runDrain    int64
		armorSpeed  int64
		percentage  int64
		wantStamina int64
	}{
		{name: "percentage product wraps", runDrain: 20, percentage: 107374183, wantStamina: 23512},
		{name: "torso weight wraps", runDrain: 20, armorSpeed: 1000000000, wantStamina: 23551},
		{name: "positive percentage truncates", runDrain: 20, percentage: 33, wantStamina: 23525},
		{name: "negative percentage truncates", runDrain: 20, percentage: -33, wantStamina: 23499},
		{name: "positive torso multiplier", runDrain: 20, armorSpeed: 10, wantStamina: 23472},
		{name: "negative torso division", runDrain: 20, armorSpeed: -9, wantStamina: 23512},
		{name: "full slower drain has minimum", runDrain: 20, percentage: 100, wantStamina: 23551},
		{name: "zero base has minimum", runDrain: 0, wantStamina: 23551},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := AdvanceStamina(StaminaTick{
				CurrentRaw:          23552,
				MaximumRaw:          23552,
				RunDrain:            tc.runDrain,
				ArmorRunDrain:       tc.armorSpeed,
				StaminaDrainPercent: tc.percentage,
				Running:             true,
				Moving:              true,
			})
			if got.CurrentRaw != tc.wantStamina || got.ForceWalk {
				t.Fatalf("running stamina = %+v, want raw=%d and no force-walk", got, tc.wantStamina)
			}
		})
	}
}

// TestMaximumStaminaUsesExpansionFixedPointDependencyGraph pins source order and integer scaling.
func TestMaximumStaminaUsesExpansionFixedPointDependencyGraph(t *testing.T) {
	rates := ClassRates{StartingVitality: 20, StartingStamina: 84, StaminaPerLevel: 4, StaminaPerVitality: 4}

	base := MaximumStamina(rates, 10, 29, StaminaMaximumSources{})
	if base != 102*256 {
		t.Fatalf("level/vitality maximum = %d, want %d", base, 102*256)
	}

	withSources := MaximumStamina(rates, 10, 29, StaminaMaximumSources{
		BonusVitality: 5, FlatMaximum: 10, SkillStaminaPercent: 25,
		SkillPassiveStaminaPercent: 10, ItemStaminaPerLevel: 8,
		ItemStaminaByTime: 20,
	})
	if withSources != 45117 {
		t.Fatalf("source-derived maximum = %d, want 45117", withSources)
	}
}

// TestByTimeAdjustmentDecodesRoundsWrapsAndInterpolates covers both halves of the recovered environment cycle.
func TestByTimeAdjustmentDecodesRoundsWrapsAndInterpolates(t *testing.T) {
	packed := int64(0 + 4*((-10+256)+(20+256)*1024))
	for baseTime, want := range map[int64]int64{0: 20, 7: 20, 8: 18, 90: 5, 180: -10, 270: 5, 359: 20} {
		if got := ByTimeAdjustment(packed, baseTime); got != want {
			t.Fatalf("base time %d adjustment = %d, want %d", baseTime, got, want)
		}
	}
}

// TestRescaleCurrentStaminaMatchesMaxStatCallback protects ratio truncation, exhaustion, and the one-unit floor.
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

// TestTownLevelsMatchRecovered113cPredicate protects the original five-town lookup.
func TestTownLevelsMatchRecovered113cPredicate(t *testing.T) {
	towns := map[int64]bool{1: true, 40: true, 75: true, 103: true, 109: true}
	for level := int64(0); level <= 120; level++ {
		if got := IsTownLevel(level); got != towns[level] {
			t.Errorf("level %d town = %t, want %t", level, got, towns[level])
		}
	}
	for _, level := range []int64{-1, -100, 121, 1024, 1 << 31} {
		if IsTownLevel(level) {
			t.Errorf("out-of-range level %d was recognized as town", level)
		}
	}
}

func TestResolveRejectsInvalidElapsedTime(t *testing.T) {
	for _, elapsed := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		if got := Resolve(gameworld.Point{}, MovePayload{X: 1}, ClassRates{Walk: 6, Run: 9}, elapsed); got != (ResolvedMovement{}) {
			t.Fatalf("elapsed %v resolved movement %+v", elapsed, got)
		}
	}
}
