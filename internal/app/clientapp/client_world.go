package clientapp

import (
	"context"
	"errors"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/app/gameserver"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// clientWorld owns caches used to present a connected game, never gameplay authority. Its entity
// IDs and histories are valid only for the currently installed presentation replica.
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
	projectedStates       []playeradapter.WorldState
	stateEntities         map[presentationStateKey]akara.Entity
}

// presentationStateKey includes both target and state identity because different actors may carry
// the same state and one actor may carry several states simultaneously.
type presentationStateKey struct {
	targetID string
	stateID  string
}

// newClientWorld creates fresh interpolation history for one presentation replica. Reusing history
// across replicas would apply corrections to unrelated entity IDs.
func newClientWorld() *clientWorld {
	return &clientWorld{history: newPresentationBuffer()}
}

// prepareConnectedWorld requires the offline ECS only as a schema source, then installs an
// entity-empty replica. Copying offline entities would leak local authority into connected play.
func (app *application) prepareConnectedWorld(ctx context.Context) error {
	if app.entitySimulation == nil || app.ecsCapability == nil {
		return errors.New("connected client world: offline schema source is unavailable")
	}

	replica, err := app.newConnectedReplica()
	if err != nil {
		return err
	}

	return app.installConnectedReplica(ctx, replica)
}

// newConnectedReplica preserves registered component schemas while removing ticks, entities, and
// instances. Lua can therefore address the same component types without inheriting offline state.
func (app *application) newConnectedReplica() (*gameecs.Engine, error) {
	snapshot, err := app.entitySimulation.Snapshot()
	if err != nil {
		return nil, err
	}

	snapshot.Tick = 0
	snapshot.Entities = nil
	for index := range snapshot.Components {
		snapshot.Components[index].Instances = nil
	}

	return gameecs.RestoreSnapshot(snapshot)
}

// installConnectedReplica moves the shared ECS capability before publishing the application field,
// making Lua and Go observe one engine boundary. The previous replica closes only after all
// adapters point at the replacement.
func (app *application) installConnectedReplica(ctx context.Context, replica *gameecs.Engine) error {
	previous := app.clientSimulation
	if err := app.ecsCapability.SetEngine(ctx, replica); err != nil {
		_ = replica.Close()

		return err
	}

	app.clientSimulation = replica
	app.resetConnectedPresentation()

	if app.movementSource != nil {
		app.movementSource.SetEngine(replica)
	}
	if previous != nil {
		_ = previous.Close()
	}

	return nil
}

// resetConnectedPresentation invalidates every cache containing entity IDs or revision cursors from
// the previous replica; retaining even one would cross-wire corrections into the new world.
func (app *application) resetConnectedPresentation() {
	app.clientWorld = newClientWorld()
	app.remoteMirrors = nil
	app.remoteMirrorKeys = nil
	app.networkRosterLogKey = ""
	app.privateProjectionKey = ""
}

// presentationSimulation selects the replica during connected play and the authoritative local ECS
// offline. Callers can render either mode without gaining authority over a remote simulation.
func (app *application) presentationSimulation() *gameecs.Engine {
	if app.clientSimulation != nil {
		return app.clientSimulation
	}

	return app.entitySimulation
}
