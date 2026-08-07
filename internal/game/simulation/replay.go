package simulation

import (
	"errors"
	"fmt"
	"time"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
)

const ReplayVersion = 1

var ErrReplay = errors.New("simulation: invalid replay")

type Replay struct {
	Version     uint32           `json:"version"`
	StepNanos   int64            `json:"step_nanos"`
	Initial     gameecs.Snapshot `json:"initial"`
	Commands    []Command        `json:"commands"`
	Checkpoints []Checkpoint     `json:"checkpoints"`
}

type Checkpoint struct {
	Tick     uint64            `json:"tick"`
	Checksum string            `json:"checksum"`
	Snapshot *gameecs.Snapshot `json:"snapshot,omitempty"`
}

type ReplayPrepare func(*gameecs.Engine) error
type ReplayApply func(*gameecs.Engine, Command) error

type DesyncError struct {
	Tick     uint64
	Expected string
	Actual   string
	Detail   string
}

func (err *DesyncError) Error() string {
	if err.Detail == "" {
		return fmt.Sprintf("simulation: replay desync at tick %d: expected %s, got %s", err.Tick, err.Expected, err.Actual)
	}
	return fmt.Sprintf("simulation: replay desync at tick %d: %s (expected %s, got %s)", err.Tick, err.Detail, err.Expected, err.Actual)
}

// VerifyReplay restores the initial world, applies commands before their target
// ticks, advances the fixed simulation clock, and verifies every checkpoint.
func VerifyReplay(replay Replay, prepare ReplayPrepare, apply ReplayApply) error {
	if replay.Version != ReplayVersion || replay.StepNanos <= 0 || apply == nil {
		return ErrReplay
	}
	engine, err := gameecs.RestoreSnapshot(replay.Initial)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrReplay, err)
	}
	defer engine.Close()
	if prepare != nil {
		if err := prepare(engine); err != nil {
			return fmt.Errorf("%w: prepare: %v", ErrReplay, err)
		}
	}
	commandIndex := 0
	currentTick := replay.Initial.Tick
	for _, checkpoint := range replay.Checkpoints {
		if checkpoint.Tick <= currentTick {
			return fmt.Errorf("%w: checkpoint tick %d follows %d", ErrReplay, checkpoint.Tick, currentTick)
		}
		for currentTick < checkpoint.Tick {
			nextTick := currentTick + 1
			for commandIndex < len(replay.Commands) && replay.Commands[commandIndex].Tick == nextTick {
				if err := apply(engine, replay.Commands[commandIndex]); err != nil {
					return fmt.Errorf("%w: apply command %d at tick %d: %v", ErrReplay, commandIndex, nextTick, err)
				}
				commandIndex++
			}
			if commandIndex < len(replay.Commands) && replay.Commands[commandIndex].Tick < nextTick {
				return fmt.Errorf("%w: commands are not ordered by tick", ErrReplay)
			}
			if err := engine.Update(time.Duration(replay.StepNanos)); err != nil {
				return fmt.Errorf("%w: update tick %d: %v", ErrReplay, nextTick, err)
			}
			currentTick = nextTick
		}
		snapshot, err := engine.Snapshot()
		if err != nil {
			return err
		}
		checksum, err := snapshot.Checksum()
		if err != nil {
			return err
		}
		if checksum != checkpoint.Checksum {
			detail := ""
			if checkpoint.Snapshot != nil {
				detail = gameecs.FirstDifference(*checkpoint.Snapshot, snapshot)
			}
			return &DesyncError{Tick: checkpoint.Tick, Expected: checkpoint.Checksum, Actual: checksum, Detail: detail}
		}
	}
	if commandIndex != len(replay.Commands) {
		return fmt.Errorf("%w: %d commands occur after final checkpoint", ErrReplay, len(replay.Commands)-commandIndex)
	}
	return nil
}
