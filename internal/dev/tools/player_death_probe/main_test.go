package main

import (
	"strings"
	"testing"
)

const ownedRuntime = `"runtime":{"patch":"1.14d","mode":"expansion","session":"single-player","character_mode":"softcore","character_origin":"probe-created","executable_sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","observation":"video-frame-log"}`

func TestAnalyzeNormalizesOwned114dPlayerDeath(t *testing.T) {
	input := `{
  "schema":"d2legacy.player_death_probe/v1","target":"diablo-ii-lod-1.14d-expansion","source":"owned-runtime",` + ownedRuntime + `,
  "cases":[{"id":"hell-recovery","scenario":"single_recovery","difficulty":"hell","class":"amazon","level":80,"killer_kind":"monster",
    "observations":[
      {"phase":"before_death","death_index":1,"frame":10,"area":"blood_moor","controlled":true,"health":100,"max_health":100,"experience":10000,"carried_gold":1000,"stashed_gold":5000,"ground_gold":0,"corpse_count":0,"equipment":[{"slot":"head","id":"cap-1"}],"inventory":[{"slot":"0,0","id":"key-1"}]},
      {"phase":"death_started","death_index":1,"frame":20,"area":"blood_moor","controlled":false,"health":0,"max_health":100,"experience":10000,"carried_gold":1000,"stashed_gold":5000,"ground_gold":0,"corpse_count":1,"equipment":[{"slot":"head","id":"cap-1"}],"inventory":[{"slot":"0,0","id":"key-1"}]},
      {"phase":"death_animation_complete","death_index":1,"frame":40,"area":"blood_moor","controlled":false,"health":0,"max_health":100,"experience":10000,"carried_gold":1000,"stashed_gold":5000,"ground_gold":0,"corpse_count":1,"equipment":[{"slot":"head","id":"cap-1"}],"inventory":[{"slot":"0,0","id":"key-1"}]},
      {"phase":"respawn_input","death_index":1,"frame":50,"area":"blood_moor","controlled":false,"health":0,"max_health":100,"experience":10000,"carried_gold":1000,"stashed_gold":5000,"ground_gold":0,"corpse_count":1,"equipment":[{"slot":"head","id":"cap-1"}],"inventory":[{"slot":"0,0","id":"key-1"}]},
      {"phase":"town_control","death_index":1,"frame":70,"area":"rogue_encampment","controlled":true,"health":100,"max_health":100,"experience":9000,"carried_gold":500,"stashed_gold":5000,"ground_gold":500,"corpse_count":1,"equipment":[],"inventory":[{"slot":"0,0","id":"key-1"}]},
      {"phase":"corpse_recovered","death_index":1,"frame":100,"area":"blood_moor","controlled":true,"health":100,"max_health":100,"experience":9750,"carried_gold":500,"stashed_gold":5000,"ground_gold":500,"corpse_count":0,"equipment":[{"slot":"head","id":"cap-1"}],"inventory":[{"slot":"0,0","id":"key-1"}]}
    ]}]}`
	got, err := analyze(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	death := got.Cases[0].Deaths[0]
	if death.DeathAnimationFrames != 20 || death.RespawnInputToControlFrames != 20 ||
		death.ExperienceLoss != 1000 || death.ExperienceRecovered != 750 ||
		death.CarriedGoldLoss != 500 || death.GroundGoldAfterRespawn != 500 ||
		death.EquipmentRemovedAtRespawn != 1 || death.EquipmentRestored != 1 ||
		death.InventoryRemovedAtRespawn != 0 || !death.RecoveryObserved {
		t.Fatalf("death report = %#v", death)
	}
	if got.ExecutableSHA256 == "" || len(got.CaptureFingerprint) != 64 || !got.Cases[0].StashedGoldUnchanged {
		t.Fatalf("provenance/case = %#v", got)
	}
}

func TestAnalyzeRejectsClassicCommunityServerAndImportedSave(t *testing.T) {
	input := `{"schema":"d2legacy.player_death_probe/v1","target":"classic","source":"community-tool",
  "runtime":{"patch":"1.13c","mode":"classic","session":"vanilla-server","character_mode":"hardcore","character_origin":"imported-save","executable_sha256":"bad","observation":"memory-tool"},"cases":[]}`
	if _, err := analyze(strings.NewReader(input)); err == nil {
		t.Fatal("non-target/community/server/imported-save capture was accepted")
	}
}

func TestAnalyzeRejectsContradictoryTimelineAndStashMutation(t *testing.T) {
	input := `{"schema":"d2legacy.player_death_probe/v1","target":"diablo-ii-lod-1.14d-expansion","source":"owned-runtime",` + ownedRuntime + `,
  "cases":[{"id":"bad","scenario":"single_no_recovery","difficulty":"normal","class":"amazon","level":1,"killer_kind":"monster","observations":[
    {"phase":"before_death","death_index":1,"frame":10,"area":"field","controlled":true,"health":10,"max_health":10,"experience":0,"carried_gold":0,"stashed_gold":1,"ground_gold":0,"corpse_count":0,"equipment":[],"inventory":[]},
    {"phase":"death_started","death_index":1,"frame":9,"area":"field","controlled":true,"health":10,"max_health":10,"experience":0,"carried_gold":0,"stashed_gold":2,"ground_gold":0,"corpse_count":0,"equipment":[],"inventory":[]}
  ]}]}`
	if _, err := analyze(strings.NewReader(input)); err == nil {
		t.Fatal("contradictory timeline was accepted")
	}
}

func TestValidateAcceptsDeadSaveExitWithoutInventingRespawn(t *testing.T) {
	observed := func(phase string, frame int, controlled bool, health int64) observation {
		return observation{
			Phase: phase, DeathIndex: 1, Frame: frame, Area: "blood_moor", Controlled: controlled,
			Health: health, MaxHealth: 10, Equipment: []slotItem{}, Inventory: []slotItem{},
		}
	}
	captured := capture{
		Schema: probeSchema, Target: probeTarget, Source: "owned-runtime",
		Runtime: runtime{
			Patch: "1.14d", Mode: "expansion", Session: "single-player", CharacterMode: "softcore",
			CharacterOrigin: "probe-created", ExecutableSHA256: strings.Repeat("a", 64), Observation: "manual-frame-log",
		},
		Cases: []probeCase{{
			ID: "dead-save-exit", Scenario: "save_exit_dead", Difficulty: "normal", Class: "amazon", Level: 1,
			KillerKind: "monster",
			Observations: []observation{
				observed("before_death", 1, true, 10),
				observed("death_started", 2, false, 0),
				observed("death_animation_complete", 3, false, 0),
				observed("save_exit", 4, false, 0),
				observed("rejoined", 5, false, 0),
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
