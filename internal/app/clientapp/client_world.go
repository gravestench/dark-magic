package clientapp

import (
	"context"
	"errors"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/app/gameserver"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// clientWorld owns the presentation-only state of a connected game session.
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

// presentationStateKey identifies one persistent state attached to one target.
type presentationStateKey struct {
	targetID string
	stateID  string
}

// newClientWorld creates an empty connected-world presentation cache.
func newClientWorld() *clientWorld {
	return &clientWorld{history: newPresentationBuffer()}
}

// prepareConnectedWorld creates an entity-empty replica with the offline world's schema.
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

// newConnectedReplica clones component definitions without copying authoritative entities.
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

// installConnectedReplica switches Lua and presentation adapters to a prepared replica.
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

// resetConnectedPresentation clears caches whose entities belonged to the previous replica.
func (app *application) resetConnectedPresentation() {
	app.clientWorld = newClientWorld()
	app.remoteMirrors = nil
	app.remoteMirrorKeys = nil
	app.networkRosterLogKey = ""
	app.privateProjectionKey = ""
}

// presentationSimulation returns the connected replica or the local authority when offline.
func (app *application) presentationSimulation() *gameecs.Engine {
	if app.clientSimulation != nil {
		return app.clientSimulation
	}

	return app.entitySimulation
}
