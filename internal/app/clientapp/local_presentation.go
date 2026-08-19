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

// Project returns immediate predicted motion plus a decaying authority correction for the same admitted owner.
func (state *localPresentation) Project(
	playerID string,
	predicted playeradapter.HUDPosition,
	correction playeradapter.HUDPosition,
	corrected bool,
	elapsed time.Duration,
) playeradapter.HUDPosition {
	if !state.initialized || state.playerID != playerID {
		return state.reset(playerID, predicted)
	}

	if corrected {
		state.accumulateCorrection(correction)
	} else if elapsed > 0 {
		state.decayCorrection(elapsed)
	}

	state.predicted = predicted
	state.presented = playeradapter.HUDPosition{X: predicted.X + state.error.X, Y: predicted.Y + state.error.Y}

	return state.presented
}

// reset prevents smoothing state from one admitted player leaking into another player's presentation.
func (state *localPresentation) reset(
	playerID string,
	predicted playeradapter.HUDPosition,
) playeradapter.HUDPosition {
	state.initialized = true
	state.playerID = playerID
	state.predicted = predicted
	state.presented = predicted
	state.error = playeradapter.HUDPosition{}

	return predicted
}

// accumulateCorrection smooths small reconciliation errors but snaps large errors to authoritative prediction.
func (state *localPresentation) accumulateCorrection(correction playeradapter.HUDPosition) {
	state.error = playeradapter.HUDPosition{
		X: state.error.X + correction.X,
		Y: state.error.Y + correction.Y,
	}

	if math.Hypot(state.error.X, state.error.Y) > localCorrectionSnap {
		state.error = playeradapter.HUDPosition{}
	}
}

// decayCorrection uses a frame-rate-independent half-life and clears negligible residual drift.
func (state *localPresentation) decayCorrection(elapsed time.Duration) {
	decay := math.Exp(-math.Ln2 * float64(elapsed) / float64(localCorrectionHalfLife))
	state.error = playeradapter.HUDPosition{X: state.error.X * decay, Y: state.error.Y * decay}

	if math.Hypot(state.error.X, state.error.Y) < 1e-5 {
		state.error = playeradapter.HUDPosition{}
	}
}

// mergeInputHistory retains the newest intent for each sequence and restores deterministic replay order.
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
