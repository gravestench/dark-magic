package main

import (
	"strings"
	"testing"
)

const ownedRuntime = `"runtime":{"patch":"1.14d","mode":"expansion","session":"single-player","executable_sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","observation":"video-frame-log"}`

func TestAnalyzeNormalizesOwned114dDefenseOutcomes(t *testing.T) {
	input := `{
  "schema":"d2legacy.defense_outcome_probe/v1","target":"diablo-ii-lod-1.14d-expansion","source":"owned-runtime",` + ownedRuntime + `,
  "cases":[
    {"id":"control","mechanism":"control","difficulty":"normal","attack_kind":"melee","attacker_kind":"monster",
     "attacker_level":10,"attack_rating":100,"defender":{"kind":"player","record":"amazon","level":10,"defense":100,"state":"standing"},
     "trials":[
       {"outcome":"damage","reaction":"gethit","health_before_raw":2560,"health_after_raw":2304},
       {"outcome":"miss","reaction":"none","health_before_raw":2304,"health_after_raw":2304}]},
    {"id":"shield","control_id":"control","mechanism":"shield_block","effect_record":"kit","displayed_chance_percent":50,
     "difficulty":"normal","attack_kind":"melee","attacker_kind":"monster","attacker_level":10,"attack_rating":100,
     "defender":{"kind":"player","record":"amazon","level":10,"defense":100,"state":"standing"},
     "trials":[
       {"outcome":"block","reaction":"block","health_before_raw":2560,"health_after_raw":2560},
       {"outcome":"damage","reaction":"gethit","health_before_raw":2560,"health_after_raw":2304}]}
  ]}`
	got, err := analyze(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	observed := got.Cases[1]
	if !observed.ContextMatch || observed.Counts["block"] != 1 || observed.Counts["damage"] != 1 ||
		observed.BlockRate.Observed != .5 || observed.TotalDamage != 256 || observed.MeanDamage != 128 {
		t.Fatalf("report = %#v", observed)
	}
	if got.ExecutableSHA256 == "" || len(got.CaptureFingerprint) != 64 {
		t.Fatalf("provenance = %#v", got)
	}
}

func TestAnalyzeRejectsClassicCommunityAndMismatchedControls(t *testing.T) {
	input := `{
  "schema":"d2legacy.defense_outcome_probe/v1","target":"classic","source":"community-tool",
  "runtime":{"patch":"1.13c","mode":"classic","session":"vanilla-server","executable_sha256":"bad","observation":"memory-tool"},
  "cases":[
    {"id":"control","mechanism":"control","difficulty":"normal","attack_kind":"melee","attacker_kind":"monster",
     "attacker_level":10,"attack_rating":100,"defender":{"kind":"player","record":"amazon","level":10,"defense":100,"state":"standing"},
     "trials":[{"outcome":"miss","reaction":"none","health_before_raw":2560,"health_after_raw":2560}]},
    {"id":"avoid","control_id":"control","mechanism":"passive_avoidance","effect_record":"Dodge","displayed_chance_percent":40,
     "difficulty":"hell","attack_kind":"missile","attacker_kind":"monster","attacker_level":10,"attack_rating":100,
     "defender":{"kind":"player","record":"amazon","level":10,"defense":100,"state":"running"},
     "trials":[{"outcome":"avoid","reaction":"none","health_before_raw":2560,"health_after_raw":2560}]}
  ]}`
	if _, err := analyze(strings.NewReader(input)); err == nil {
		t.Fatal("non-target/community/mismatched capture was accepted")
	}
}

func TestAnalyzeRejectsOutcomeThatContradictsHealthOrReaction(t *testing.T) {
	input := `{
  "schema":"d2legacy.defense_outcome_probe/v1","target":"diablo-ii-lod-1.14d-expansion","source":"owned-runtime",` + ownedRuntime + `,
  "cases":[{"id":"control","mechanism":"control","difficulty":"normal","attack_kind":"melee","attacker_kind":"monster",
  "attacker_level":10,"attack_rating":100,"defender":{"kind":"player","record":"amazon","level":10,"defense":100,"state":"standing"},
  "trials":[{"outcome":"block","reaction":"none","health_before_raw":2560,"health_after_raw":2304}]}]}`
	if _, err := analyze(strings.NewReader(input)); err == nil {
		t.Fatal("contradictory outcome was accepted")
	}
}
