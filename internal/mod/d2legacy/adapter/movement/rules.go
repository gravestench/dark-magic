package movement

import (
	"math"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

const (
	ArrivalDistance = 0.2
	diagonalScale   = 0.7071067811865476
)

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

type ResolvedStamina struct {
	CurrentRaw int64
	ForceWalk  bool
}

// AdvanceStamina applies one 25 Hz Diablo II stamina event in 8.8 units.
func AdvanceStamina(tick StaminaTick) ResolvedStamina {
	current := max(int64(0), min(tick.CurrentRaw, tick.MaximumRaw))
	if tick.Running && tick.Moving && !tick.InTown {
		drain := 2 * tick.RunDrain * (tick.ArmorRunDrain/10 + 1)
		if tick.StaminaDrainPercent != 0 {
			drain += drain * tick.StaminaDrainPercent / -100
		}
		drain = max(1, drain)
		current = max(0, current-drain)
		return ResolvedStamina{CurrentRaw: current, ForceWalk: current == 0}
	}
	if tick.CanRecover && (tick.InTown || !tick.Moving || current >= 256) && current < tick.MaximumRaw {
		divisor := int64(256)
		if tick.Moving {
			divisor = 512
		}
		recovery := tick.MaximumRaw / divisor
		if tick.RecoveryBonus != 0 {
			recovery += tick.RecoveryBonus * recovery / 100
		}
		current = min(tick.MaximumRaw, current+recovery)
	}
	return ResolvedStamina{CurrentRaw: current}
}

func IsTownLevel(level int64) bool {
	return level == 1 || level == 45 || level == 80 || level == 108 || level == 114
}

func EffectiveRates(base ClassRates, modifiers Modifiers) ClassRates {
	item := modifiers.ItemFasterMoveVelocity
	effectiveItem := int64(0)
	if item != 0 && item+150 != 0 {
		effectiveItem = item * 150 / (item + 150)
	}
	bonus := float64(effectiveItem+modifiers.VelocityPercent) / 100
	walk := math.Max(base.Walk*.25, base.Walk*(1+bonus))
	run := math.Max(base.Walk*.25, base.Run+base.Walk*bonus)
	base.Walk, base.Run = walk, run
	return base
}

// Resolve applies the production d2legacy input policy without touching ECS.
// Both authoritative Lua and client prediction call this implementation.
func Resolve(position gameworld.Point, payload MovePayload, rates ClassRates, modifiers ...Modifiers) ResolvedMovement {
	if len(modifiers) > 0 {
		rates = EffectiveRates(rates, modifiers[0])
	}
	x, y := float64(payload.X), float64(payload.Y)
	if payload.Target != nil {
		x, y = payload.Target.X-position.X, payload.Target.Y-position.Y
		distance := math.Hypot(x, y)
		if distance <= ArrivalDistance {
			x, y = 0, 0
		} else {
			x, y = x/distance, y/distance
		}
	} else if x != 0 && y != 0 {
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
