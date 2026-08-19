package main

import (
	"strings"
	"testing"
)

// TestValidateAcceptsDeadSaveExitWithoutInventingRespawn preserves the distinction between rejoin and respawn evidence.
func TestValidateAcceptsDeadSaveExitWithoutInventingRespawn(t *testing.T) {
	captured := capture{
		Schema: probeSchema,
		Target: probeTarget,
		Source: "owned-runtime",
		Runtime: runtime{
			Patch:            "1.14d",
			Mode:             "expansion",
			Session:          "single-player",
			CharacterMode:    "softcore",
			CharacterOrigin:  "probe-created",
			ExecutableSHA256: strings.Repeat("a", 64),
			Observation:      "manual-frame-log",
		},
		Cases: []probeCase{{
			ID:         "dead-save-exit",
			Scenario:   "save_exit_dead",
			Difficulty: "normal",
			Class:      "amazon",
			Level:      1,
			KillerKind: "monster",
			Observations: []observation{
				newDeathObservation("before_death", 1, true, 10),
				newDeathObservation("death_started", 2, false, 0),
				newDeathObservation("death_animation_complete", 3, false, 0),
				newDeathObservation("save_exit", 4, false, 0),
				newDeathObservation("rejoined", 5, false, 0),
			},
		}},
	}

	if err := validate(captured); err != nil {
		t.Fatal(err)
	}

	got := normalize(captured.Cases[0])
	if len(got.Deaths) != 1 || got.Deaths[0].RespawnObserved || !got.SaveRejoinObserved {
		t.Fatalf("save/exit report = %#v", got)
	}
}

// newDeathObservation builds the minimal owned observation needed to keep the save/exit test focused on phase
// semantics.
func newDeathObservation(phase string, frame int, controlled bool, health int64) observation {
	return observation{
		Phase:      phase,
		DeathIndex: 1,
		Frame:      frame,
		Area:       "blood_moor",
		Controlled: controlled,
		Health:     health,
		MaxHealth:  10,
		Equipment:  []slotItem{},
		Inventory:  []slotItem{},
	}
}
