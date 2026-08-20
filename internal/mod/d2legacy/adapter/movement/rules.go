package movement

import (
	"math"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

const (
	// ArrivalDistance is the world-space threshold where exact-target movement resolves to a stop.
	ArrivalDistance         = 0.2
	diagonalScale           = 0.7071067811865476
	walkAnimationRate int64 = 213
	runAnimationRate  int64 = 101
)

// ResolvedMovement carries velocity and mode flags derived without mutating ECS state.
type ResolvedMovement struct {
	Velocity gameworld.Point
	Running  bool
	Moving   bool
}

// Modifiers are resolved authoritative stats. Item Faster Run/Walk receives
// Diablo II's 150-point diminishing-return conversion; velocitypercent is the
// additive skill, state, and equipped-armor channel after its 100 baseline.
type Modifiers struct {
	VelocityPercent        int64
	ItemFasterMoveVelocity int64
}

// StaminaTick gathers the authoritative fixed-point state needed for one ordered drain-or-recovery event.
type StaminaTick struct {
	CurrentRaw          int64
	MaximumRaw          int64
	RunDrain            int64
	ArmorRunDrain       int64
	StaminaDrainPercent int64
	RecoveryBonus       int64
	Running             bool
	Moving              bool
	InTown              bool
	CanRecover          bool
}

// ResolvedStamina carries the updated fixed-point value and whether exhaustion must override running.
type ResolvedStamina struct {
	CurrentRaw int64
	ForceWalk  bool
}

// StaminaMaximumSources are normalized ItemStatCost operands. Whole-point
// maxstamina and the evaluated by-time operand are shifted to 8.8 here;
// per-level remains in the record's eighths encoding and the two skill
// percentages remain separate op-1 families.
type StaminaMaximumSources struct {
	BonusVitality              int64
	FlatMaximum                int64
	SkillStaminaPercent        int64
	SkillPassiveStaminaPercent int64
	ItemStaminaPerLevel        int64
	ItemStaminaByTime          int64
}

// MaximumStamina reproduces the Expansion 1.14d maxstamina dependency graph.
// CharStats level/vitality terms are quarters (hence << 6), while maxstamina
// itself is 8.8. Op-1 percentages use the direct value before op-derived
// vitality/per-level/by-time contributions, matching ItemStatCost evaluation.
func MaximumStamina(rates ClassRates, level, baseVitality int64, sources StaminaMaximumSources) int64 {
	level = max(int64(1), level)
	baseVitality = max(rates.StartingVitality, baseVitality)

	// Direct maximum-stamina sources establish the percentage base before derived operands are applied.
	direct := rates.StartingStamina*256 +
		(level-1)*rates.StaminaPerLevel*64 +
		(baseVitality-rates.StartingVitality)*rates.StaminaPerVitality*64 +
		sources.FlatMaximum*256

	// Derived operands remain separate because combining them would change integer truncation behavior.
	result := direct + sources.BonusVitality*rates.StaminaPerVitality*64
	result += direct * sources.SkillStaminaPercent / 100
	result += direct * sources.SkillPassiveStaminaPercent / 100
	result += level * sources.ItemStaminaPerLevel >> 3
	result += sources.ItemStaminaByTime * 256

	return max(int64(256), result)
}

// ByTimeAdjustment decodes Properties func 18's signed 10-bit min/max values
// and linearly interpolates them around a 360-unit environment cycle. The
// period argument accepted by the recovered runtime is informational; the
// packed center period and base time determine the result.
func ByTimeAdjustment(packed, baseTime int64) int64 {
	center := packed & 3
	minimum := signedTimeOperand(packed >> 2)
	maximum := signedTimeOperand(packed >> 12)

	difference := baseTime - center*90
	if difference < 0 {
		difference = -difference
	}

	// The recovered runtime rounds to 15-unit boundaries before folding the second half of the day.
	rounded := ((difference + 7) / 15) * 15

	rounded = max(int64(0), min(int64(359), rounded))
	if rounded > 180 {
		rounded = 360 - rounded
	}

	return maximum - (maximum-minimum)*rounded/180
}

// signedTimeOperand extracts the recovered signed ten-bit value used by Properties function 18.
func signedTimeOperand(packed int64) int64 {
	return (packed & 0x3ff) - 0x100
}

// RescaleCurrentStamina is the maxstamina value-change callback: positive
// current stamina preserves its ratio using a double calculation, truncates,
// and clamps to [1,new]. Zero remains exhausted.
func RescaleCurrentStamina(currentRaw, previousMaximumRaw, newMaximumRaw int64) int64 {
	newMaximumRaw = max(int64(0), newMaximumRaw)

	if currentRaw <= 0 {
		return 0
	}

	if previousMaximumRaw <= 0 {
		return min(currentRaw, newMaximumRaw)
	}

	// Floating-point ratio calculation is intentional compatibility with the recovered max-stat callback.
	rescaled := int64(float64(currentRaw) / float64(max(previousMaximumRaw, int64(256))) * float64(newMaximumRaw))

	return max(int64(1), min(rescaled, newMaximumRaw))
}

// AdvanceStamina applies one 25 Hz Diablo II stamina event in 8.8 units.
// Drain takes precedence over recovery, so an eligible running tick cannot also regenerate stamina.
func AdvanceStamina(tick StaminaTick) ResolvedStamina {
	current := max(int64(0), min(tick.CurrentRaw, tick.MaximumRaw))
	if tick.Running && tick.Moving && !tick.InTown {
		current = max(0, current-staminaDrain(tick))
		return ResolvedStamina{CurrentRaw: current, ForceWalk: current == 0}
	}

	if tick.CanRecover && (tick.InTown || !tick.Moving || current >= 256) && current < tick.MaximumRaw {
		current = min(tick.MaximumRaw, current+staminaRecovery(tick))
	}

	return ResolvedStamina{CurrentRaw: current}
}

// staminaDrain preserves the runtime's armor multiplier, signed slower-drain channel, and one-unit minimum.
func staminaDrain(tick StaminaTick) int64 {
	drain := 2 * tick.RunDrain * (tick.ArmorRunDrain/10 + 1)
	if tick.StaminaDrainPercent != 0 {
		drain += drain * tick.StaminaDrainPercent / -100
	}

	return max(1, drain)
}

// staminaRecovery halves regeneration while moving and applies the bonus after fixed-point base truncation.
func staminaRecovery(tick StaminaTick) int64 {
	divisor := int64(256)
	if tick.Moving {
		divisor = 512
	}

	recovery := tick.MaximumRaw / divisor
	if tick.RecoveryBonus != 0 {
		recovery += tick.RecoveryBonus * recovery / 100
	}

	return recovery
}

// IsTownLevel recognizes only the five Expansion 1.14d act towns used by stamina recovery policy.
func IsTownLevel(level int64) bool {
	return level == 1 || level == 45 || level == 80 || level == 108 || level == 114
}

// EffectiveRates converts shared percentage channels into world-subtile walk and run rates.
// Run is derived from the original walk rate because the percentage already embeds the class run/walk ratio.
func EffectiveRates(base ClassRates, modifiers Modifiers) ClassRates {
	walkPercentage := EffectiveVelocityPercent(base, false, modifiers)
	runPercentage := EffectiveVelocityPercent(base, true, modifiers)
	walk := base.Walk
	base.Walk = walk * float64(walkPercentage) / 100
	base.Run = walk * float64(runPercentage) / 100

	return base
}

// EffectiveVelocityPercent reproduces the Expansion movement ordering. Walk
// starts at 100; run starts at the class RunSpeed/WalkSpeed ratio. Item FRW is
// diminished before joining the additive velocitypercent stat channel, then
// the final percentage is floored at 25.
func EffectiveVelocityPercent(base ClassRates, running bool, modifiers Modifiers) int64 {
	percentage := int64(100)
	if running && base.Walk > 0 {
		percentage = int64(100 * base.Run / base.Walk)
	}

	// Item FRW has its own diminishing-return transform before joining the additive stat channel.
	item := modifiers.ItemFasterMoveVelocity
	if item != 0 && item+150 != 0 {
		percentage += item * 150 / (item + 150)
	}

	percentage += modifiers.VelocityPercent

	return max(int64(25), percentage)
}

// MovementAnimationRate is the runtime WL/RN override applied after the same
// effective velocity percentage that drives path velocity. AnimData still owns
// the frame count and events for these modes, but not their playback rate.
func MovementAnimationRate(base ClassRates, running bool, modifiers Modifiers) int64 {
	rate := walkAnimationRate
	if running {
		rate = runAnimationRate
	}

	return rate * EffectiveVelocityPercent(base, running, modifiers) / 100
}

// Resolve applies the production d2legacy input policy without touching ECS.
// Both authoritative Lua and client prediction call this implementation.
func Resolve(position gameworld.Point, payload MovePayload, rates ClassRates, modifiers ...Modifiers) ResolvedMovement {
	if len(modifiers) > 0 {
		rates = EffectiveRates(rates, modifiers[0])
	}

	x, y := float64(payload.X), float64(payload.Y)
	if payload.Target != nil {
		x, y = targetDirection(position, *payload.Target)
	} else if x != 0 && y != 0 {
		// Normalize diagonal keyboard input so two active axes do not increase movement speed.
		x, y = x*diagonalScale, y*diagonalScale
	}

	speed := rates.Walk
	if payload.Running {
		speed = rates.Run
	}

	return ResolvedMovement{
		Velocity: gameworld.Point{X: x * speed, Y: y * speed},
		Running:  payload.Running,
		Moving:   x != 0 || y != 0,
	}
}

// targetDirection normalizes world-space target movement and stops inside the simulation arrival threshold.
func targetDirection(position gameworld.Point, target MoveTarget) (float64, float64) {
	x := target.X - position.X
	y := target.Y - position.Y

	distance := math.Hypot(x, y)
	if distance <= ArrivalDistance {
		return 0, 0
	}

	return x / distance, y / distance
}
