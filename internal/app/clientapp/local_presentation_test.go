package clientapp

import (
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

func TestLocalPresentationMovesImmediatelyAndSmoothsOnlyCorrectionError(t *testing.T) {
	var state localPresentation
	if got := state.Project("player", playeradapter.HUDPosition{X: 10}, playeradapter.HUDPosition{}, false, 0); got.X != 10 {
		t.Fatalf("initial presentation = %+v", got)
	}
	if got := state.Project("player", playeradapter.HUDPosition{X: 11}, playeradapter.HUDPosition{}, false, 16*time.Millisecond); got.X != 11 {
		t.Fatalf("predicted movement was delayed: %+v", got)
	}
	if got := state.Project("player", playeradapter.HUDPosition{X: 10.5}, playeradapter.HUDPosition{X: .5}, true, 16*time.Millisecond); got.X != 11 {
		t.Fatalf("correction was visually discontinuous: %+v", got)
	}
	got := state.Project("player", playeradapter.HUDPosition{X: 10.5}, playeradapter.HUDPosition{}, false, localCorrectionHalfLife)
	if got.X != 10.75 {
		t.Fatalf("half-life presentation = %+v, want x=10.75", got)
	}
}

func TestLocalPresentationSnapsLargeCorrectionAndResetsForNewOwner(t *testing.T) {
	var state localPresentation
	state.Project("one", playeradapter.HUDPosition{X: 10}, playeradapter.HUDPosition{}, false, 0)
	if got := state.Project("one", playeradapter.HUDPosition{X: 2}, playeradapter.HUDPosition{X: 8}, true, 0); got.X != 2 {
		t.Fatalf("large correction did not snap: %+v", got)
	}
	if got := state.Project("two", playeradapter.HUDPosition{X: 40}, playeradapter.HUDPosition{X: 2}, true, 0); got.X != 40 {
		t.Fatalf("new owner retained old presentation state: %+v", got)
	}
}

func TestMergeInputHistoryRetainsAcknowledgedAndNewCommandsForCorrectionComparison(t *testing.T) {
	previous := []gameserver.CommandIntent{{Sequence: 1}, {Sequence: 2}}
	current := []gameserver.CommandIntent{{Sequence: 2}, {Sequence: 3}}
	merged := mergeInputHistory(previous, current)
	if len(merged) != 3 || merged[0].Sequence != 1 || merged[1].Sequence != 2 || merged[2].Sequence != 3 {
		t.Fatalf("merged history = %#v", merged)
	}
}
