package clientapp

import (
	"encoding/json"
	"math"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/networkclock"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// positionPrediction is a disposable replay of unacknowledged movement from one canonical HUD.
// Mutating it affects presentation only and cannot advance the authoritative simulation.
type positionPrediction struct {
	hud        playeradapter.HUD
	pending    []gameserver.CommandIntent
	collision  *gameworld.Map
	step       time.Duration
	position   gameworld.Point
	velocity   gameworld.Point
	bounds     gameworld.Point
	radius     float64
	staminaRaw int64
	running    bool
	applied    map[uint64]bool
	rates      movement.ClassRates
	classKnown bool
}

// predictPosition advances full authority ticks and then a presentation-only fractional remainder.
// It returns canonical position unchanged when the timeline cannot safely move forward.
func predictPosition(
	hud playeradapter.HUD,
	pending []gameserver.CommandIntent,
	moment networkclock.Moment,
	collision *gameworld.Map,
	step time.Duration,
	catalog movement.Catalog,
) playeradapter.HUDPosition {
	if step <= 0 || moment.Tick < hud.Tick {
		return hud.Position
	}

	prediction := newPositionPrediction(hud, pending, collision, step, catalog)
	for tick := hud.Tick + 1; tick <= moment.Tick; tick++ {
		prediction.advanceTick(tick)
	}

	if moment.Fraction > 0 {
		prediction.advanceFraction(moment.Tick+1, moment.Fraction)
	}

	return playeradapter.HUDPosition{
		X: prediction.position.X,
		Y: prediction.position.Y,
	}
}

// newPositionPrediction seeds every movement rule from the authoritative HUD and trusted collision
// map. Collision dimensions replace HUD bounds when available because loaded world data is the more
// complete constraint.
func newPositionPrediction(
	hud playeradapter.HUD,
	pending []gameserver.CommandIntent,
	collision *gameworld.Map,
	step time.Duration,
	catalog movement.Catalog,
) *positionPrediction {
	bounds := gameworld.Point{
		X: movementBound(hud.Movement.Bounds.X),
		Y: movementBound(hud.Movement.Bounds.Y),
	}
	if collision != nil {
		bounds = gameworld.Point{
			X: float64(collision.WidthSubtiles),
			Y: float64(collision.HeightSubtiles),
		}
	}

	rates, classKnown := catalog.Rates(hud.Player.Class)

	return &positionPrediction{
		hud:        hud,
		pending:    pending,
		collision:  collision,
		step:       step,
		position:   gameworld.Point{X: hud.Position.X, Y: hud.Position.Y},
		velocity:   gameworld.Point{X: hud.Movement.Velocity.X, Y: hud.Movement.Velocity.Y},
		bounds:     bounds,
		radius:     movementRadius(hud.Movement.Radius),
		staminaRaw: hud.Vitals.StaminaRaw,
		running:    hud.Movement.Running,
		applied:    make(map[uint64]bool, len(pending)),
		rates:      rates,
		classKnown: classKnown,
	}
}

// advanceTick mirrors authority ordering: apply eligible input, advance stamina once, then integrate
// one fixed step. Changing that order would create recurring prediction corrections.
func (prediction *positionPrediction) advanceTick(tick uint64) {
	prediction.applyInputs(tick)
	prediction.advanceStamina()
	prediction.integrate(1)
}

// advanceFraction applies next-tick direction for responsive rendering but deliberately skips
// stamina advancement; fractional presentation time must not consume authoritative resources.
func (prediction *positionPrediction) advanceFraction(tick uint64, fraction float64) {
	prediction.applyInputs(tick)
	prediction.integrate(fraction)
}

// applyInputs processes pending movement by target tick and sequence. Invalid payloads are marked
// consumed for this replay so one malformed command cannot be reconsidered on every later tick.
func (prediction *positionPrediction) applyInputs(tick uint64) {
	for _, intent := range prediction.pending {
		if prediction.skipIntent(intent, tick) {
			continue
		}

		var payload movement.MovePayload
		if json.Unmarshal(intent.Payload, &payload) == nil && prediction.classKnown {
			payload.Running = payload.Running && prediction.staminaRaw > 0
			resolved := movement.Resolve(
				prediction.position,
				payload,
				prediction.rates,
				prediction.movementModifiers(),
			)
			prediction.velocity = resolved.Velocity
			prediction.running = payload.Running
		}

		prediction.applied[intent.Sequence] = true
	}
}

// skipIntent excludes already applied, non-movement, and future input while preserving original
// pending order for commands eligible on the same tick.
func (prediction *positionPrediction) skipIntent(intent gameserver.CommandIntent, tick uint64) bool {
	return prediction.applied[intent.Sequence] ||
		intent.Kind != movement.MoveCommand ||
		intent.TargetTick > tick
}

// advanceStamina invokes the production stamina rules with HUD-projected modifiers. ForceWalk also
// adjusts velocity immediately so prediction does not continue running after local stamina reaches
// zero.
func (prediction *positionPrediction) advanceStamina() {
	moving := prediction.velocity.X != 0 || prediction.velocity.Y != 0
	inTown := movement.IsTownLevel(prediction.hud.Location.LevelID)
	canRecover := prediction.canRecoverStamina(moving, inTown)

	resolved := movement.AdvanceStamina(movement.StaminaTick{
		CurrentRaw:          prediction.staminaRaw,
		MaximumRaw:          prediction.hud.Vitals.MaxStaminaRaw,
		RunDrain:            prediction.hud.Movement.RunDrain,
		ArmorRunDrain:       prediction.hud.Movement.ArmorRunDrain,
		StaminaDrainPercent: prediction.hud.Movement.StaminaDrainPercent,
		RecoveryBonus:       prediction.hud.Movement.StaminaRecoveryBonus,
		Running:             prediction.running,
		Moving:              moving,
		InTown:              inTown,
		CanRecover:          canRecover,
	})
	prediction.staminaRaw = resolved.CurrentRaw

	if resolved.ForceWalk {
		prediction.forceWalk()
	}
}

// canRecoverStamina duplicates authority recovery exceptions explicitly because omitting any one
// creates deterministic drift during long movement sequences.
func (prediction *positionPrediction) canRecoverStamina(moving, inTown bool) bool {
	return prediction.hud.Animation.Mode == "NU" ||
		(moving && !prediction.running) ||
		(inTown && moving) ||
		prediction.hud.Movement.StaminaRecoveryBonus >= 1000
}

// forceWalk preserves direction while reducing magnitude to effective walking speed. This avoids a
// visible heading change at the exact tick stamina is exhausted.
func (prediction *positionPrediction) forceWalk() {
	prediction.running = false
	effective := movement.EffectiveRates(prediction.rates, prediction.movementModifiers())

	magnitude := math.Hypot(prediction.velocity.X, prediction.velocity.Y)
	if magnitude == 0 {
		return
	}

	prediction.velocity.X = prediction.velocity.X / magnitude * effective.Walk
	prediction.velocity.Y = prediction.velocity.Y / magnitude * effective.Walk
}

// movementModifiers restricts prediction to modifiers projected by authority; client-only stats
// must not influence the movement result.
func (prediction *positionPrediction) movementModifiers() movement.Modifiers {
	return movement.Modifiers{
		VelocityPercent:        prediction.hud.Movement.VelocityPercent,
		ItemFasterMoveVelocity: prediction.hud.Movement.ItemFasterMoveVelocity,
	}
}

// integrate delegates bounds and collision resolution to the same world primitive used by
// authority, which keeps prediction differences limited to network timing rather than geometry.
func (prediction *positionPrediction) integrate(fraction float64) {
	prediction.position = gameworld.IntegrateVelocity(
		prediction.collision,
		prediction.position,
		prediction.velocity,
		prediction.bounds,
		prediction.radius,
		prediction.step.Seconds()*fraction,
	)
}
