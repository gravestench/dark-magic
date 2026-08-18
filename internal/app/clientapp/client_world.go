package clientapp

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/app/clientsession"
	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/networkclock"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

type clientWorld struct {
	lastCorrection        uint64
	lastViewRevision      uint64
	lastPresentedRevision uint64
	history               *presentationBuffer
	local                 localPresentation
	lastHUD               playeradapter.HUD
	lastPending           []gameserver.CommandIntent
	hasHUD                bool
	lastEventEpoch        uint64
	lastEventViewTick     uint64
	eventCursorTick       uint64
	eventCursorID         uint64
	semanticEventEntities []akara.Entity
	missileEntities       map[string]akara.Entity
}

func newClientWorld() *clientWorld {
	return &clientWorld{history: newPresentationBuffer()}
}

// prepareConnectedWorld creates a schema-compatible, entity-empty client ECS
// and rebinds Lua presentation to it before the game-world scene opens. The
// offline authority remains owned by offlineSession and is never used as a
// network replica or correction cache.
func (app *application) prepareConnectedWorld(ctx context.Context) error {
	if app.entitySimulation == nil || app.ecsCapability == nil {
		return errors.New("connected client world: offline schema source is unavailable")
	}
	snapshot, err := app.entitySimulation.Snapshot()
	if err != nil {
		return err
	}
	snapshot.Tick = 0
	snapshot.Entities = nil
	for index := range snapshot.Components {
		snapshot.Components[index].Instances = nil
	}
	replica, err := gameecs.RestoreSnapshot(snapshot)
	if err != nil {
		return err
	}
	previous := app.clientSimulation
	if err := app.ecsCapability.SetEngine(ctx, replica); err != nil {
		_ = replica.Close()
		return err
	}
	app.clientSimulation = replica
	app.clientWorld = newClientWorld()
	app.remoteMirrors = nil
	app.remoteMirrorKeys = nil
	app.networkRosterLogKey = ""
	app.privateProjectionKey = ""
	if app.movementSource != nil {
		app.movementSource.SetEngine(replica)
	}
	if previous != nil {
		_ = previous.Close()
	}
	return nil
}

// reconcile installs each canonical correction once, replays unacknowledged
// local movement through production rules, and samples peers from delayed
// snapshot history.
func (world *clientWorld) reconcile(app *application, session *clientsession.Session, elapsed time.Duration) error {
	return world.reconcileAt(app, session, elapsed, time.Now())
}

func (world *clientWorld) reconcileAt(app *application, session *clientsession.Session, elapsed time.Duration, now time.Time) error {
	if world == nil {
		return errors.New("connected client world is unavailable")
	}
	presentation := session.PresentationSnapshot()
	if presentation == nil {
		return nil
	}
	hud, projection, revision := presentation.HUD, presentation.World, presentation.Revision
	corrected := projection.Tick > world.lastCorrection
	if revision > world.lastViewRevision {
		// Corrections update presentation state and interpolation targets. They
		// must not overwrite an already-rendered position: doing so turns the
		// 10 Hz authoritative stream into visible 100 ms teleports.
		if err := app.installRemoteProjection(hud, presentation.Private, presentation.Party, world.lastCorrection == 0); err != nil {
			return err
		}
		if err := world.reconcileMissiles(app, projection.Missiles); err != nil {
			return err
		}
		world.lastCorrection = max(world.lastCorrection, projection.Tick)
		world.lastViewRevision = revision
		world.history.Upsert(projection)
	}
	if err := world.reconcileSemanticEvents(app, presentation.Events, presentation.EventEpoch); err != nil {
		return err
	}
	timeline := session.NetworkTimeline(now)
	if sample, found := world.history.Sample(timeline.Interpolation); found {
		if sample.discreteRevision != world.lastPresentedRevision {
			if err := app.syncRemoteMirrors(sample.entities, hud.Location); err != nil {
				return err
			}
			world.lastPresentedRevision = sample.discreteRevision
			app.logNetworkRoster(hud)
		}
		if err := app.applySampledWorldPositions(sample.entities); err != nil {
			return err
		}
	}
	// Input is sampled at 25 Hz, independently of the 10 Hz correction stream.
	// Start from the newest canonical owner state and replay saved inputs to the
	// prediction clock every frame.
	pending := session.PendingInputs()
	collision := app.gameWorlds[int(hud.Location.LevelID)]
	predicted := predictPosition(hud, pending, timeline.Prediction, collision, networkInputStep, app.movementCatalog)
	correction := playeradapter.HUDPosition{}
	sameOwnerWorld := world.hasHUD && world.lastHUD.Player.PlayerID == hud.Player.PlayerID && world.lastHUD.Location == hud.Location
	if corrected && sameOwnerWorld {
		previous := predictPosition(world.lastHUD, mergeInputHistory(world.lastPending, pending), timeline.Prediction, collision, networkInputStep, app.movementCatalog)
		correction = playeradapter.HUDPosition{X: previous.X - predicted.X, Y: previous.Y - predicted.Y}
	} else if !sameOwnerWorld {
		world.local = localPresentation{}
	}
	presented := world.local.Project(hud.Player.PlayerID, predicted, correction, corrected, elapsed)
	if err := app.applyLocalPredictedPosition(hud.Player.PlayerID, presented); err != nil {
		return err
	}
	if err := app.applyAnimationTimeline(hud.Player.PlayerID, timeline, session.StepDuration()); err != nil {
		return err
	}
	world.lastHUD = hud
	world.lastPending = pending
	world.hasHUD = true
	return nil
}

func predictPosition(hud playeradapter.HUD, pending []gameserver.CommandIntent, moment networkclock.Moment, collision *gameworld.Map, step time.Duration, catalog movement.Catalog) playeradapter.HUDPosition {
	position := gameworld.Point{X: hud.Position.X, Y: hud.Position.Y}
	velocity := gameworld.Point{X: hud.Movement.Velocity.X, Y: hud.Movement.Velocity.Y}
	running := hud.Movement.Running
	staminaRaw := hud.Vitals.StaminaRaw
	bounds := gameworld.Point{X: movementBound(hud.Movement.Bounds.X), Y: movementBound(hud.Movement.Bounds.Y)}
	radius := movementRadius(hud.Movement.Radius)
	if collision != nil {
		bounds = gameworld.Point{X: float64(collision.WidthSubtiles), Y: float64(collision.HeightSubtiles)}
	}
	if step <= 0 || moment.Tick < hud.Tick {
		return hud.Position
	}
	applied := make(map[uint64]bool, len(pending))
	rates, classKnown := catalog.Rates(hud.Player.Class)
	applyInputs := func(tick uint64) {
		for _, intent := range pending {
			if applied[intent.Sequence] || intent.Kind != movement.MoveCommand || intent.TargetTick > tick {
				continue
			}
			var payload movement.MovePayload
			if json.Unmarshal(intent.Payload, &payload) == nil && classKnown {
				payload.Running = payload.Running && staminaRaw > 0
				velocity = movement.Resolve(position, payload, rates, movement.Modifiers{
					VelocityPercent:        hud.Movement.VelocityPercent,
					ItemFasterMoveVelocity: hud.Movement.ItemFasterMoveVelocity,
				}).Velocity
				running = payload.Running
			}
			applied[intent.Sequence] = true
		}
	}
	advanceStamina := func() {
		moving := velocity.X != 0 || velocity.Y != 0
		inTown := movement.IsTownLevel(hud.Location.LevelID)
		canRecover := hud.Animation.Mode == "NU" || (moving && !running) || (inTown && moving) || hud.Movement.StaminaRecoveryBonus >= 1000
		resolved := movement.AdvanceStamina(movement.StaminaTick{
			CurrentRaw: staminaRaw, MaximumRaw: hud.Vitals.MaxStaminaRaw,
			RunDrain: hud.Movement.RunDrain, ArmorRunDrain: hud.Movement.ArmorRunDrain,
			StaminaDrainPercent: hud.Movement.StaminaDrainPercent,
			RecoveryBonus:       hud.Movement.StaminaRecoveryBonus,
			Running:             running, Moving: moving, InTown: inTown, CanRecover: canRecover,
		})
		staminaRaw = resolved.CurrentRaw
		if resolved.ForceWalk {
			running = false
			effective := movement.EffectiveRates(rates, movement.Modifiers{
				VelocityPercent:        hud.Movement.VelocityPercent,
				ItemFasterMoveVelocity: hud.Movement.ItemFasterMoveVelocity,
			})
			magnitude := math.Hypot(velocity.X, velocity.Y)
			if magnitude > 0 {
				velocity.X, velocity.Y = velocity.X/magnitude*effective.Walk, velocity.Y/magnitude*effective.Walk
			}
		}
	}
	for tick := hud.Tick + 1; tick <= moment.Tick; tick++ {
		applyInputs(tick)
		advanceStamina()
		position = gameworld.IntegrateVelocity(collision, position, velocity, bounds, radius, step.Seconds())
	}
	if moment.Fraction > 0 {
		applyInputs(moment.Tick + 1)
		position = gameworld.IntegrateVelocity(collision, position, velocity, bounds, radius, step.Seconds()*moment.Fraction)
	}
	return playeradapter.HUDPosition{X: position.X, Y: position.Y}
}

func (app *application) presentationSimulation() *gameecs.Engine {
	if app.clientSimulation != nil {
		return app.clientSimulation
	}
	return app.entitySimulation
}
