package clientapp

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gravestench/dark-magic/internal/app/clientsession"
	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
)

const networkInputStep = 40 * time.Millisecond

// Advance first applies authenticated corrections, then samples the resulting presentation input.
// This ordering prevents pointer projection from starting at a position the authority has already
// corrected.
func (controller *networkController) Advance(
	ctx context.Context,
	elapsed time.Duration,
) error {
	controller.mu.Lock()
	client := controller.client
	controller.mu.Unlock()

	if client == nil {
		return nil
	}

	// Pointer projection must begin from the newest authenticated position.
	if err := controller.app.clientWorld.reconcile(controller.app, client, elapsed); err != nil {
		return err
	}

	now := time.Now()
	if err := controller.submitMovementSamples(client, elapsed, now); err != nil {
		return err
	}

	return controller.submitPendingIntents(client, now)
}

// submitMovementSamples converts renderer time into the authority's 25 Hz input cadence. Sampling
// independently of frame rate keeps command density stable across fast and slow renderers.
func (controller *networkController) submitMovementSamples(
	client *clientsession.Session,
	elapsed time.Duration,
	now time.Time,
) error {
	if controller.app.movementSource == nil {
		return nil
	}

	for _, targetTick := range controller.inputTicks(client, elapsed, now) {
		if err := controller.submitMovementCommands(targetTick); err != nil {
			return err
		}
	}

	return nil
}

// submitMovementCommands suppresses repeated idle movement while preserving one active-to-idle
// transition. The authority needs that final stop, but sending one every tick wastes queue space.
func (controller *networkController) submitMovementCommands(targetTick uint64) error {
	commands := controller.app.movementSource.Commands(targetTick)

	for _, command := range commands {
		active, movementCommand := movementCommandActivity(command)
		if !movementCommand {
			if err := controller.submit(targetTick, command.Kind, command.Payload); err != nil {
				return err
			}

			continue
		}

		if !controller.movementRequired(active) {
			continue
		}

		if err := controller.submit(targetTick, command.Kind, command.Payload); err != nil {
			return err
		}

		controller.markMovement(active)
	}

	return nil
}

// movementCommandActivity recognizes only well-formed movement payloads. Other commands remain
// opaque and malformed movement cannot accidentally change stop-suppression state.
func movementCommandActivity(command simulation.Command) (bool, bool) {
	if command.Kind != movement.MoveCommand {
		return false, false
	}

	var payload movement.MovePayload
	if err := json.Unmarshal(command.Payload, &payload); err != nil {
		return false, false
	}

	active := payload.X != 0 || payload.Y != 0 || payload.Target != nil

	return active, true
}

// submitPendingIntents assigns all commands drained in one frame to the next admissible authority
// tick. An empty drain does not advance sequencing or consume a tick.
func (controller *networkController) submitPendingIntents(
	client *clientsession.Session,
	now time.Time,
) error {
	intents := controller.app.commandIntents.Drain()
	if len(intents) == 0 {
		return nil
	}

	targetTick := client.NextInputTick(now)

	for _, intent := range intents {
		payload, err := json.Marshal(intent.Payload)
		if err != nil {
			return err
		}

		if err := controller.submit(targetTick, intent.Kind, payload); err != nil {
			return err
		}
	}

	return nil
}

// inputTicks converts renderer time into fixed-step samples while capping catch-up at five ticks.
// Dropping excess lag after a hitch is preferable to flooding the bounded transport queue with
// stale movement.
func (controller *networkController) inputTicks(
	client *clientsession.Session,
	elapsed time.Duration,
	now time.Time,
) []uint64 {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	if elapsed < 0 {
		elapsed = 0
	}

	// Bound catch-up so a renderer hitch cannot overflow transport queues.
	controller.inputLag = min(controller.inputLag+elapsed, 5*networkInputStep)
	result := make([]uint64, 0, 5)

	for controller.inputLag >= networkInputStep {
		target := client.NextInputTick(now)
		if target > controller.lastMovementTick {
			controller.lastMovementTick = target
			result = append(result, target)
		}

		controller.inputLag -= networkInputStep
	}

	return result
}

// sampleMovement atomically claims an authority tick so concurrent or repeated callers cannot
// submit duplicate movement samples for the same point on the simulation timeline.
func (controller *networkController) sampleMovement(targetTick uint64) bool {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	if targetTick <= controller.lastMovementTick {
		return false
	}

	controller.lastMovementTick = targetTick

	return true
}

// movementRequired keeps active input and exactly one stop after movement ends; subsequent idle
// samples can be omitted without leaving the authority moving indefinitely.
func (controller *networkController) movementRequired(active bool) bool {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	return active || controller.lastMovementActive
}

// markMovement updates stop-suppression state only after submission succeeds. A full queue must not
// make the controller believe an unsent stop reached authority.
func (controller *networkController) markMovement(active bool) {
	controller.mu.Lock()
	controller.lastMovementActive = active
	controller.mu.Unlock()
}

// submit stages input in the session before publishing it to the bounded send queue. If publication
// fails, it discards the staged sequence so prediction never replays a command that transport will
// not send.
func (controller *networkController) submit(
	targetTick uint64,
	kind string,
	payload json.RawMessage,
) error {
	controller.mu.Lock()

	intent := gameserver.CommandIntent{
		TargetTick: targetTick,
		Sequence:   controller.sequence + 1,
		Kind:       kind,
		Payload:    append(json.RawMessage(nil), payload...),
	}

	client := controller.client
	if client == nil {
		controller.mu.Unlock()

		return errors.New("network client is unavailable")
	}

	if err := client.StageInput(intent); err != nil {
		controller.mu.Unlock()

		return err
	}

	select {
	case controller.submissions <- intent:
		controller.sequence++
		controller.mu.Unlock()

		return nil
	default:
		client.DiscardInput(intent.Sequence)
		controller.mu.Unlock()

		return errors.New("network input queue is full")
	}
}
