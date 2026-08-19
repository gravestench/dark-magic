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

// Advance reconciles corrections, samples movement, and submits queued intents.
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

// submitMovementSamples emits fixed-step movement commands for elapsed time.
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

// submitMovementCommands filters redundant stops while preserving other commands.
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

// movementCommandActivity decodes whether one valid move command is active.
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

// submitPendingIntents drains non-movement commands onto the next input tick.
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

// inputTicks converts renderer time into bounded fixed-step authority samples.
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

// sampleMovement claims one authority tick for movement sampling.
func (controller *networkController) sampleMovement(targetTick uint64) bool {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	if targetTick <= controller.lastMovementTick {
		return false
	}

	controller.lastMovementTick = targetTick

	return true
}

// movementRequired keeps one stop command after active movement ends.
func (controller *networkController) movementRequired(active bool) bool {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	return active || controller.lastMovementActive
}

// markMovement records whether the last submitted move remained active.
func (controller *networkController) markMovement(active bool) {
	controller.mu.Lock()
	controller.lastMovementActive = active
	controller.mu.Unlock()
}

// submit stages an intent and transfers it to the bounded send queue.
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
