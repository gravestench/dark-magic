package main

import (
	"strings"
	"testing"
)

func TestAnalyzeNormalizesOwned114dExplosionObservation(t *testing.T) {
	input := `{
  "schema":"d2legacy.fire_golem_death_probe/v1",
  "target":"diablo-ii-lod-1.14d-expansion",
  "source":"owned-runtime",
  "runtime":{"patch":"1.14d","mode":"expansion","session":"single-player",
    "character_origin":"probe-created",
    "executable_sha256":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
    "observation":"debugger-stat-log-plus-video","executable_unmodified":true},
  "records":{"skill_id":94,"skill_name_key":"Fire Golem","localized_skill_name":"Fire Golem",
    "locale":"enUS","monster_id":"firegolem","death_damage_enabled":true,
    "extracted_records_sha256":"abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"},
  "cases":[{
    "id":"replacement-normal-level-1","trigger":"replacement","difficulty":"normal",
    "skill_level":1,"player_level":30,"map_seed":42,
    "event_frames":{"old_golem_removed":105,"explosion_started":104,"new_golem_created":106},
    "explosion_center":{"x_subtiles":10,"y_subtiles":10},
    "targets":[
      {"id":"near-hostile","kind":"hostile_monster","hostile_to_owner":true,
       "position":{"x_subtiles":12,"y_subtiles":10},"distance_millisubtiles":2000,
       "health_before_raw":10000,"health_after_raw":8000,
       "fire_resistance_percent":0,"physical_resistance_percent":0,
	   "no_absorb_or_flat_damage_reduction":true,
       "pre_mitigation_channels_raw":{"physical":0,"fire":2000,"cold":0,"lightning":0,"poison":0,"magic":0},
	   "damage_event":true,"hit_reaction":true,"died":false},
      {"id":"far-hostile","kind":"hostile_monster","hostile_to_owner":true,
       "position":{"x_subtiles":15,"y_subtiles":10},"distance_millisubtiles":5000,
       "health_before_raw":10000,"health_after_raw":10000,
       "fire_resistance_percent":0,"physical_resistance_percent":0,
	   "no_absorb_or_flat_damage_reduction":true,
       "pre_mitigation_channels_raw":{"physical":0,"fire":0,"cold":0,"lightning":0,"poison":0,"magic":0},
	   "damage_event":false,"hit_reaction":false,"died":false},
      {"id":"owner","kind":"owner","hostile_to_owner":false,
       "position":{"x_subtiles":11,"y_subtiles":11},"distance_millisubtiles":1414,
       "health_before_raw":10000,"health_after_raw":10000,
       "fire_resistance_percent":0,"physical_resistance_percent":0,
	   "no_absorb_or_flat_damage_reduction":true,
       "pre_mitigation_channels_raw":{"physical":0,"fire":0,"cold":0,"lightning":0,"poison":0,"magic":0},
	   "damage_event":false,"hit_reaction":false,"died":false}
    ]
  }]}`
	got, err := analyze(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Cases) != 1 || got.Cases[0].Targets[0].HealthDeltaRaw != 2000 {
		t.Fatalf("report = %#v", got)
	}
	current := got.Cases[0]
	if len(current.RadiusBrackets) != 1 || !current.RadiusBrackets[0].BoundaryBracketed ||
		current.RadiusBrackets[0].FarthestAffectedMilli != 2000 || current.RadiusBrackets[0].NearestUnaffectedMilli != 5000 {
		// The owner is intentionally not a radius boundary candidate because it
		// belongs to a different target class.
		t.Fatalf("radius report = %#v", current)
	}
	if len(current.OrderedEvents) != 3 || current.OrderedEvents[0].Name != "explosion_started" {
		t.Fatalf("ordered events = %#v", current.OrderedEvents)
	}
}

func TestAnalyzeRejectsNonTargetAndContradictorySamples(t *testing.T) {
	input := `{
  "schema":"d2legacy.fire_golem_death_probe/v1","target":"classic","source":"community-tool",
  "runtime":{"patch":"1.10f","mode":"classic","session":"vanilla-server","character_origin":"imported-save",
    "executable_sha256":"bad","observation":"memory-dump","executable_unmodified":false},
  "records":{"skill_id":0,"skill_name_key":"","localized_skill_name":"","locale":"",
    "monster_id":"","death_damage_enabled":false,"extracted_records_sha256":"bad"},
  "cases":[{"id":"bad","trigger":"replacement","difficulty":"normal","skill_level":0,"player_level":1,
    "map_seed":0,"event_frames":{"old_golem_removed":null,"explosion_started":null},
    "explosion_center":{"x_subtiles":0,"y_subtiles":0},"targets":[]}]}`
	if _, err := analyze(strings.NewReader(input)); err == nil {
		t.Fatal("non-target and contradictory capture was accepted")
	}
}
