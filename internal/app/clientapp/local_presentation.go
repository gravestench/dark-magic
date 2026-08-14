package clientapp

import (
	"math"
	"sort"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

const (
	localCorrectionHalfLife = 100 * time.Millisecond
	localCorrectionSnap     = 4.0
)

// localPresentation keeps the predicted simulation transform distinct from
// the transform written into the render ECS. Input motion is immediate; only
// authoritative error is accumulated and decayed visually.
type localPresentation struct {
	initialized bool
	playerID    string
	predicted   playeradapter.HUDPosition
	presented   playeradapter.HUDPosition
	error       playeradapter.HUDPosition
}

func (state *localPresentation) Project(playerID string, predicted, correction playeradapter.HUDPosition, corrected bool, elapsed time.Duration) playeradapter.HUDPosition {
	if !state.initialized || state.playerID != playerID {
		state.initialized = true
		state.playerID = playerID
		state.predicted = predicted
		state.presented = predicted
		state.error = playeradapter.HUDPosition{}
		return predicted
	}
	if corrected {
		state.error.X += correction.X
		state.error.Y += correction.Y
		if math.Hypot(state.error.X, state.error.Y) > localCorrectionSnap {
			state.error = playeradapter.HUDPosition{}
		}
	} else if elapsed > 0 {
		decay := math.Exp(-math.Ln2 * float64(elapsed) / float64(localCorrectionHalfLife))
		state.error.X *= decay
		state.error.Y *= decay
		if math.Hypot(state.error.X, state.error.Y) < 1e-5 {
			state.error = playeradapter.HUDPosition{}
		}
	}
	state.predicted = predicted
	state.presented = playeradapter.HUDPosition{X: predicted.X + state.error.X, Y: predicted.Y + state.error.Y}
	return state.presented
}

func mergeInputHistory(previous, current []gameserver.CommandIntent) []gameserver.CommandIntent {
	bySequence := make(map[uint64]gameserver.CommandIntent, len(previous)+len(current))
	for _, intent := range previous {
		bySequence[intent.Sequence] = intent
	}
	for _, intent := range current {
		bySequence[intent.Sequence] = intent
	}
	result := make([]gameserver.CommandIntent, 0, len(bySequence))
	for _, intent := range bySequence {
		result = append(result, intent)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Sequence < result[j].Sequence })
	return result
}
