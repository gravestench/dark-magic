package clientapp

import (
	"errors"
	"time"

	"github.com/gravestench/dark-magic/internal/app/clientsession"
	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/networkclock"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// reconcile samples the network timeline against wall-clock time. Keeping clock acquisition at this
// edge makes reconcileAt deterministic for tests and replayable scenarios.
func (world *clientWorld) reconcile(
	app *application,
	session *clientsession.Session,
	elapsed time.Duration,
) error {
	return world.reconcileAt(app, session, elapsed, time.Now())
}

// reconcileAt installs each authenticated revision, presents peers on the delayed interpolation
// timeline, and predicts only the local player. Those timelines intentionally differ to trade peer
// latency for smoothness without adding local input latency.
func (world *clientWorld) reconcileAt(
	app *application,
	session *clientsession.Session,
	elapsed time.Duration,
	now time.Time,
) error {
	if world == nil {
		return errors.New("connected client world is unavailable")
	}

	presentation := session.PresentationSnapshot()
	if presentation == nil {
		return nil
	}

	corrected, err := world.installCanonicalPresentation(app, presentation)
	if err != nil {
		return err
	}

	if err := world.reconcileSemanticEvents(app, presentation.Events, presentation.EventEpoch); err != nil {
		return err
	}

	timeline := session.NetworkTimeline(now)
	if err := world.presentRemoteHistory(app, presentation.HUD, timeline.Interpolation); err != nil {
		return err
	}

	return world.presentLocalPrediction(app, session, presentation.HUD, timeline, elapsed, corrected)
}

// installCanonicalPresentation applies discrete projection changes once per view revision while
// retaining every newer world tick in interpolation history. Separating those revisions prevents
// repeated entity churn when only transform sampling advances.
func (world *clientWorld) installCanonicalPresentation(
	app *application,
	presentation *clientsession.PresentationSnapshot,
) (bool, error) {
	projection := presentation.World

	corrected := projection.Tick > world.lastCorrection
	if presentation.Revision <= world.lastViewRevision {
		return corrected, nil
	}

	// A correction updates interpolation targets but must not overwrite the
	// position already rendered this frame. Doing so exposes the 10 Hz stream as
	// visible 100 ms teleports.
	firstCorrection := world.lastCorrection == 0
	if err := app.installRemoteProjection(
		presentation.HUD,
		presentation.Private,
		presentation.Party,
		firstCorrection,
	); err != nil {
		return false, err
	}

	if err := world.reconcileMissiles(app, projection.Missiles); err != nil {
		return false, err
	}

	world.projectedStates = append(world.projectedStates[:0], projection.States...)
	world.lastCorrection = max(world.lastCorrection, projection.Tick)
	world.lastViewRevision = presentation.Revision
	world.history.Upsert(projection)

	return corrected, nil
}

// presentRemoteHistory updates the remote roster only when discrete membership changes, but applies
// sampled transforms every frame. This avoids rebuilding mirrors merely to interpolate movement.
func (world *clientWorld) presentRemoteHistory(
	app *application,
	hud playeradapter.HUD,
	moment networkclock.Moment,
) error {
	sample, found := world.history.Sample(moment)
	if !found {
		return nil
	}

	if sample.discreteRevision != world.lastPresentedRevision {
		if err := app.syncRemoteMirrors(sample.entities, hud.Location); err != nil {
			return err
		}

		world.lastPresentedRevision = sample.discreteRevision

		app.logNetworkRoster(hud)
	}

	return app.applySampledWorldPositions(sample.entities)
}

// presentLocalPrediction replays unacknowledged input from the canonical HUD, then decays any newly
// observed correction error through presentation. Persistent states and animation still follow the
// authenticated timeline rather than speculative input.
func (world *clientWorld) presentLocalPrediction(
	app *application,
	session *clientsession.Session,
	hud playeradapter.HUD,
	timeline networkclock.Timeline,
	elapsed time.Duration,
	corrected bool,
) error {
	pending := session.PendingInputs()
	collision := app.gameWorlds[int(hud.Location.LevelID)]
	predicted := predictPosition(
		hud,
		pending,
		timeline.Prediction,
		collision,
		networkInputStep,
		app.movementCatalog,
	)
	correction := world.predictionCorrection(
		app,
		hud,
		pending,
		predicted,
		timeline.Prediction,
		collision,
		corrected,
	)
	presented := world.local.Project(hud.Player.PlayerID, predicted, correction, corrected, elapsed)

	if err := app.applyLocalPredictedPosition(hud.Player.PlayerID, presented); err != nil {
		return err
	}

	if err := app.applyAnimationTimeline(hud.Player.PlayerID, timeline, session.StepDuration()); err != nil {
		return err
	}

	if err := world.reconcilePersistentStates(app, world.projectedStates); err != nil {
		return err
	}

	world.lastHUD = hud
	world.lastPending = pending
	world.hasHUD = true

	return nil
}

// predictionCorrection compares the old and new replay results only for the same player in the same
// location. Crossing owner or world boundaries resets smoothing so stale error cannot drag a newly
// loaded character or level.
func (world *clientWorld) predictionCorrection(
	app *application,
	hud playeradapter.HUD,
	pending []gameserver.CommandIntent,
	predicted playeradapter.HUDPosition,
	moment networkclock.Moment,
	collision *gameworld.Map,
	corrected bool,
) playeradapter.HUDPosition {
	sameOwnerWorld := world.hasHUD &&
		world.lastHUD.Player.PlayerID == hud.Player.PlayerID &&
		world.lastHUD.Location == hud.Location
	if !sameOwnerWorld {
		world.local = localPresentation{}

		return playeradapter.HUDPosition{}
	}

	if !corrected {
		return playeradapter.HUDPosition{}
	}

	history := mergeInputHistory(world.lastPending, pending)
	previous := predictPosition(
		world.lastHUD,
		history,
		moment,
		collision,
		networkInputStep,
		app.movementCatalog,
	)

	return playeradapter.HUDPosition{
		X: previous.X - predicted.X,
		Y: previous.Y - predicted.Y,
	}
}
