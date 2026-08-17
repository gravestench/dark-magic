package main

import (
	"strings"
	"testing"
)

func TestAnalyzeNormalizesOwned114dPartyPoolObservation(t *testing.T) {
	input := `{
  "schema":"d2legacy.party_xp_probe/v1","target":"lod-1.14d","source":"owned-runtime",
  "cases":[
    {"id":"neutral","difficulty":"normal","area":"Blood Moor","monster":"fallen","monster_level":5,
     "game_players":2,"party":false,"members":[
       {"id":"alice","level":5,"killer":true,"same_area":true,"distance_subtiles":3,"experience_before":100,"experience_after":200},
       {"id":"bob","level":5,"killer":false,"same_area":true,"distance_subtiles":4,"experience_before":100,"experience_after":100}]},
    {"id":"party","baseline_id":"neutral","difficulty":"normal","area":"Blood Moor","monster":"fallen","monster_level":5,
     "game_players":2,"party":true,"members":[
       {"id":"alice","level":5,"killer":true,"same_area":true,"distance_subtiles":3,"experience_before":200,"experience_after":268},
       {"id":"bob","level":5,"killer":false,"same_area":true,"distance_subtiles":4,"experience_before":100,"experience_after":167}]}
  ]}`
	got, err := analyze(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Cases) != 2 || got.Cases[1].TotalDelta != 135 || got.Cases[1].BaselineKillerDelta != 100 {
		t.Fatalf("report = %#v", got)
	}
	if candidates := got.Cases[1].PoolCandidates; len(candidates) != 2 || candidates[0].ObservedFit != "floor" {
		t.Fatalf("pool candidates = %#v", candidates)
	}
}

func TestAnalyzeRejectsNonTargetAndMismatchedBaseline(t *testing.T) {
	input := `{
  "schema":"d2legacy.party_xp_probe/v1","target":"classic","source":"community-tool",
  "cases":[{"id":"party","baseline_id":"missing","difficulty":"normal","area":"Blood Moor",
  "monster":"fallen","monster_level":5,"game_players":1,"party":true,"members":[
  {"id":"alice","level":5,"killer":true,"same_area":true,"distance_subtiles":0,
  "experience_before":0,"experience_after":1}]}]}`
	if _, err := analyze(strings.NewReader(input)); err == nil {
		t.Fatal("invalid target/source/baseline was accepted")
	}
}
