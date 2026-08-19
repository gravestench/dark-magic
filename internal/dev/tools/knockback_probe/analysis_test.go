package main

import (
	"strings"
	"testing"
)

const ownedRuntimeCapture = `{
  "schema": "d2legacy.knockback_probe/v1",
  "target": "diablo-ii-lod-1.14d-expansion",
  "source": "owned-runtime",
  "cases": [
    {
      "id": "control",
      "mechanism": "control",
      "difficulty": "normal",
      "attacker_kind": "player",
      "target": {
        "kind": "monster",
        "record": "fallen1",
        "size_class": "small",
        "mode_supported": true
      },
      "open_distance_subtiles": 10,
      "trials": [
        {
          "hit": true,
          "combat_blocked": false,
          "lethal": false,
          "uninterruptible": false,
          "displacement_subtiles": 0,
          "reaction": "gethit"
        },
        {
          "hit": true,
          "combat_blocked": false,
          "lethal": false,
          "uninterruptible": false,
          "displacement_subtiles": 0,
          "reaction": "gethit"
        }
      ]
    },
    {
      "id": "item-small",
      "control_id": "control",
      "mechanism": "item_knockback",
      "difficulty": "normal",
      "attacker_kind": "player",
      "target": {
        "kind": "monster",
        "record": "fallen1",
        "size_class": "small",
        "mode_supported": true
      },
      "open_distance_subtiles": 10,
      "trials": [
        {
          "hit": true,
          "combat_blocked": false,
          "lethal": false,
          "uninterruptible": false,
          "displacement_subtiles": 5,
          "reaction": "knockback"
        },
        {
          "hit": true,
          "combat_blocked": false,
          "lethal": false,
          "uninterruptible": false,
          "displacement_subtiles": 0,
          "reaction": "knockback"
        }
      ]
    }
  ]
}`

const invalidCapture = `{
  "schema": "d2legacy.knockback_probe/v1",
  "target": "classic",
  "source": "community-tool",
  "cases": [
    {
      "id": "missile",
      "control_id": "missing",
      "mechanism": "missile_knockback",
      "difficulty": "normal",
      "attacker_kind": "player",
      "target": {
        "kind": "monster",
        "record": "fallen1",
        "size_class": "small",
        "mode_supported": true
      },
      "missile_knockback_value": 33,
      "open_distance_subtiles": 10,
      "trials": [
        {
          "hit": true,
          "combat_blocked": false,
          "lethal": false,
          "uninterruptible": false,
          "displacement_subtiles": 0,
          "reaction": "none"
        }
      ]
    }
  ]
}`

// TestAnalyzeNormalizesOwned114dKnockbackTrials verifies that eligible reactions
// distinguish successful movement from a knockback animation blocked by geometry.
func TestAnalyzeNormalizesOwned114dKnockbackTrials(t *testing.T) {
	got, err := analyze(strings.NewReader(ownedRuntimeCapture))
	if err != nil {
		t.Fatal(err)
	}

	if len(got.Cases) != 2 {
		t.Fatalf("case count = %d, want 2", len(got.Cases))
	}

	observed := got.Cases[1]
	if observed.EligibleTrials != 2 ||
		observed.KnockbackReactions != 2 ||
		observed.MovedTrials != 1 ||
		observed.BlockedMotionTrials != 1 {
		t.Fatalf("report = %#v", observed)
	}

	if len(observed.Candidates) != 2 ||
		observed.Candidates[0].Name != "size_weighted_128_roll" ||
		!observed.Candidates[0].InsideObserved95Band {
		t.Fatalf("candidates = %#v", observed.Candidates)
	}
}

// TestAnalyzeRejectsNonExpansionCommunityAndMismatchedControl ensures evidence
// from the wrong runtime, source, and experimental context cannot yield a report.
func TestAnalyzeRejectsNonExpansionCommunityAndMismatchedControl(t *testing.T) {
	if _, err := analyze(strings.NewReader(invalidCapture)); err == nil {
		t.Fatal("invalid target/source/control was accepted")
	}
}

// TestNormalizeKeepsModeIncapableTargetsOutOfChanceInference ensures an eligible
// hit is still counted when the target cannot visually reveal knockback chance.
func TestNormalizeKeepsModeIncapableTargetsOutOfChanceInference(t *testing.T) {
	observed := normalize(probeCase{
		ID:                    "no-kb-mode",
		Mechanism:             "missile_knockback",
		MissileKnockbackValue: 75,
		Target: target{
			Kind:          "monster",
			Record:        "gorgon1",
			SizeClass:     "normal",
			ModeSupported: false,
		},
		Trials: []trial{{Hit: true, Reaction: "gethit"}},
	})
	if observed.ChanceObservable || len(observed.Candidates) != 0 || observed.EligibleTrials != 1 {
		t.Fatalf("report = %#v", observed)
	}
}
