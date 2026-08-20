package main

import (
	"strings"
	"testing"
)

const validPartyCapture = `{
  "schema": "d2legacy.party_xp_probe/v1",
  "target": "lod-1.14d",
  "source": "owned-runtime",
  "cases": [
    {
      "id": "neutral",
      "difficulty": "normal",
      "area": "Blood Moor",
      "monster": "fallen",
      "monster_level": 5,
      "game_players": 2,
      "party": false,
      "members": [
        {
          "id": "alice",
          "level": 5,
          "killer": true,
          "same_area": true,
          "distance_subtiles": 3,
          "experience_before": 100,
          "experience_after": 200
        },
        {
          "id": "bob",
          "level": 10,
          "killer": false,
          "same_area": true,
          "distance_subtiles": 4,
          "experience_before": 100,
          "experience_after": 100
        }
      ]
    },
    {
      "id": "party",
      "baseline_id": "neutral",
      "difficulty": "normal",
      "area": "Blood Moor",
      "monster": "fallen",
      "monster_level": 5,
      "game_players": 2,
      "party": true,
      "members": [
        {
          "id": "alice",
          "level": 5,
          "killer": true,
          "same_area": true,
          "distance_subtiles": 3,
          "experience_before": 200,
          "experience_after": 245
        },
        {
          "id": "bob",
          "level": 10,
          "killer": false,
          "same_area": true,
          "distance_subtiles": 4,
          "experience_before": 100,
          "experience_after": 190
        }
      ]
    }
  ]
}`

const invalidTargetAndBaselineCapture = `{
  "schema": "d2legacy.party_xp_probe/v1",
  "target": "classic",
  "source": "community-tool",
  "cases": [
    {
      "id": "party",
      "baseline_id": "missing",
      "difficulty": "normal",
      "area": "Blood Moor",
      "monster": "fallen",
      "monster_level": 5,
      "game_players": 1,
      "party": true,
      "members": [
        {
          "id": "alice",
          "level": 5,
          "killer": true,
          "same_area": true,
          "distance_subtiles": 0,
          "experience_before": 0,
          "experience_after": 1
        }
      ]
    }
  ]
}`

// TestAnalyzeNormalizesOwned114dPartyPoolObservation locks candidate order and rounding labels used in evidence review.
func TestAnalyzeNormalizesOwned114dPartyPoolObservation(t *testing.T) {
	got := mustAnalyze(t, validPartyCapture)

	if len(got.Cases) != 2 {
		t.Fatalf("case count = %d, want 2; report = %#v", len(got.Cases), got)
	}

	party := got.Cases[1]
	if party.TotalDelta != 135 {
		t.Fatalf("party total delta = %d, want 135; report = %#v", party.TotalDelta, party)
	}

	if party.BaselineKillerDelta != 100 {
		t.Fatalf("baseline killer delta = %d, want 100; report = %#v", party.BaselineKillerDelta, party)
	}

	if len(party.PoolCandidates) != 2 {
		t.Fatalf("pool candidate count = %d, want 2; candidates = %#v", len(party.PoolCandidates), party.PoolCandidates)
	}

	if party.PoolCandidates[0].ObservedFit != "floor" {
		t.Fatalf(
			"first pool fit = %q, want floor; candidates = %#v",
			party.PoolCandidates[0].ObservedFit,
			party.PoolCandidates,
		)
	}

	shares := party.ShareCandidates
	if len(shares) != 3 {
		t.Fatalf("share candidate count = %d, want 3; candidates = %#v", len(shares), shares)
	}

	if shares[0].Name != "direct_character_level" || shares[0].ObservedFit != "floor" {
		t.Fatalf("direct-level candidate = %#v, want floor fit", shares[0])
	}

	if shares[1].Name != "inverse_character_level" || shares[1].ObservedFit != "none" {
		t.Fatalf("inverse-level candidate = %#v, want no fit", shares[1])
	}
}

// TestAnalyzeRejectsNonTargetAndMismatchedBaseline proves provenance and pairing failures are reported together.
func TestAnalyzeRejectsNonTargetAndMismatchedBaseline(t *testing.T) {
	_, err := analyze(strings.NewReader(invalidTargetAndBaselineCapture))
	if err == nil {
		t.Fatal("invalid target/source/baseline was accepted")
	}

	requireErrorContains(t, err, `target "classic", want "lod-1.14d"`)
	requireErrorContains(t, err, "source must be owned-runtime")
	requireErrorContains(t, err, `case "party" has invalid neutral baseline`)
}

// mustAnalyze keeps success-path tests focused on report behavior while preserving the complete failure as diagnostics.
func mustAnalyze(t *testing.T, input string) report {
	t.Helper()

	result, err := analyze(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	return result
}

// requireErrorContains verifies each independent validation consequence without coupling the test to join formatting.
func requireErrorContains(t *testing.T, err error, expected string) {
	t.Helper()

	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("error %q does not contain %q", err, expected)
	}
}
