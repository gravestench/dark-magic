package main

import (
	"strings"
	"testing"
)

const ownedRuntimeJSON = `"runtime":{
  "patch":"1.14d",
  "mode":"expansion",
  "session":"single-player",
  "executable_sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "observation":"video-frame-log"
}`

// TestAnalyzeNormalizesOwned114dDefenseOutcomes verifies ordered control comparison, rates, damage, and provenance.
func TestAnalyzeNormalizesOwned114dDefenseOutcomes(t *testing.T) {
	input := `{
  "schema":"d2legacy.defense_outcome_probe/v1",
  "target":"diablo-ii-lod-1.14d-expansion",
  "source":"owned-runtime",
  ` + ownedRuntimeJSON + `,
  "cases":[
    {
      "id":"control",
      "mechanism":"control",
      "difficulty":"normal",
      "attack_kind":"melee",
      "attacker_kind":"monster",
      "attacker_level":10,
      "attack_rating":100,
      "defender":{"kind":"player","record":"amazon","level":10,"defense":100,"state":"standing"},
      "trials":[
        {"outcome":"damage","reaction":"gethit","health_before_raw":2560,"health_after_raw":2304},
        {"outcome":"miss","reaction":"none","health_before_raw":2304,"health_after_raw":2304}
      ]
    },
    {
      "id":"shield",
      "control_id":"control",
      "mechanism":"shield_block",
      "effect_record":"kit",
      "displayed_chance_percent":50,
      "difficulty":"normal",
      "attack_kind":"melee",
      "attacker_kind":"monster",
      "attacker_level":10,
      "attack_rating":100,
      "defender":{"kind":"player","record":"amazon","level":10,"defense":100,"state":"standing"},
      "trials":[
        {"outcome":"block","reaction":"block","health_before_raw":2560,"health_after_raw":2560},
        {"outcome":"damage","reaction":"gethit","health_before_raw":2560,"health_after_raw":2304}
      ]
    }
  ]
}`

	got := requireAnalyzedCapture(t, input)
	if len(got.Cases) != 2 {
		t.Fatalf("cases = %#v", got.Cases)
	}

	// Selecting by position deliberately verifies that analysis preserves capture ordering.
	observed := got.Cases[1]
	if !observed.ContextMatch || observed.Counts["block"] != 1 || observed.Counts["damage"] != 1 ||
		observed.BlockRate.Observed != .5 || observed.TotalDamage != 256 || observed.MeanDamage != 128 {
		t.Fatalf("report = %#v", observed)
	}

	if got.ExecutableSHA256 == "" || len(got.CaptureFingerprint) != 64 {
		t.Fatalf("provenance = %#v", got)
	}
}

// TestAnalyzeRejectsClassicCommunityAndMismatchedControls proves all provenance and comparison boundaries are enforced.
func TestAnalyzeRejectsClassicCommunityAndMismatchedControls(t *testing.T) {
	input := `{
  "schema":"d2legacy.defense_outcome_probe/v1",
  "target":"classic",
  "source":"community-tool",
  "runtime":{
    "patch":"1.13c",
    "mode":"classic",
    "session":"vanilla-server",
    "executable_sha256":"bad",
    "observation":"memory-tool"
  },
  "cases":[
    {
      "id":"control",
      "mechanism":"control",
      "difficulty":"normal",
      "attack_kind":"melee",
      "attacker_kind":"monster",
      "attacker_level":10,
      "attack_rating":100,
      "defender":{"kind":"player","record":"amazon","level":10,"defense":100,"state":"standing"},
      "trials":[
        {"outcome":"miss","reaction":"none","health_before_raw":2560,"health_after_raw":2560}
      ]
    },
    {
      "id":"avoid",
      "control_id":"control",
      "mechanism":"passive_avoidance",
      "effect_record":"Dodge",
      "displayed_chance_percent":40,
      "difficulty":"hell",
      "attack_kind":"missile",
      "attacker_kind":"monster",
      "attacker_level":10,
      "attack_rating":100,
      "defender":{"kind":"player","record":"amazon","level":10,"defense":100,"state":"running"},
      "trials":[
        {"outcome":"avoid","reaction":"none","health_before_raw":2560,"health_after_raw":2560}
      ]
    }
  ]
}`

	requireRejectedCapture(t, input, "non-target/community/mismatched capture was accepted")
}

// TestAnalyzeRejectsOutcomeThatContradictsHealthOrReaction protects the visual-evidence consistency invariant.
func TestAnalyzeRejectsOutcomeThatContradictsHealthOrReaction(t *testing.T) {
	input := `{
  "schema":"d2legacy.defense_outcome_probe/v1",
  "target":"diablo-ii-lod-1.14d-expansion",
  "source":"owned-runtime",
  ` + ownedRuntimeJSON + `,
  "cases":[
    {
      "id":"control",
      "mechanism":"control",
      "difficulty":"normal",
      "attack_kind":"melee",
      "attacker_kind":"monster",
      "attacker_level":10,
      "attack_rating":100,
      "defender":{"kind":"player","record":"amazon","level":10,"defense":100,"state":"standing"},
      "trials":[
        {"outcome":"block","reaction":"none","health_before_raw":2560,"health_after_raw":2304}
      ]
    }
  ]
}`

	requireRejectedCapture(t, input, "contradictory outcome was accepted")
}

// requireAnalyzedCapture keeps successful scenarios focused on assertions while preserving the complete input bytes.
func requireAnalyzedCapture(t *testing.T, input string) report {
	t.Helper()

	got, err := analyze(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	return got
}

// requireRejectedCapture makes rejection scenarios fail with their original contract-specific diagnostic.
func requireRejectedCapture(t *testing.T, input, failureMessage string) {
	t.Helper()

	if _, err := analyze(strings.NewReader(input)); err == nil {
		t.Fatal(failureMessage)
	}
}
