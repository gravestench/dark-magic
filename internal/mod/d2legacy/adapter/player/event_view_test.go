package player

import (
	"math"
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func TestProjectEventViewReturnsBoundedNearbySemanticTail(t *testing.T) {
	view, err := ProjectEventView("alice", semanticCheckpoint())
	if err != nil {
		t.Fatal(err)
	}
	if view.FromTick != 37 || view.Tick != 100 || view.Truncated || len(view.Events) != 3 {
		t.Fatalf("event window = %+v", view)
	}
	cast := view.Events[0]
	if cast.ID != 20 || cast.Type != "cast" || cast.Tick != 98 || cast.Position != (HUDPosition{X: 12, Y: 10}) ||
		cast.Direction != 3 || cast.Cast == nil || cast.Cast.Kind != "cast_effect" || cast.Cast.SkillID != 47 {
		t.Fatalf("cast event = %+v", cast)
	}
	state := view.Events[1]
	if state.ID != 21 || state.Type != "state" || state.OverlayHeight != 4 || state.State == nil || state.State.StateID != "frozenarmor" {
		t.Fatalf("state event = %+v", state)
	}
	death := view.Events[2]
	if death.ID != 23 || death.Type != "monster_death" || death.Position != (HUDPosition{X: 13, Y: 10}) ||
		death.MonsterDeath == nil || death.MonsterDeath.Kind != "monster_death_presented" || death.MonsterDeath.MonsterID != "fallen-a" {
		t.Fatalf("monster death event = %+v", death)
	}
}

func TestValidateEventViewRejectsUntrustedUnionAndOrdering(t *testing.T) {
	valid := EventView{Version: EventViewVersion, Tick: 10, Events: []SemanticEvent{
		{ID: 1, Type: "cast", Tick: 9, Position: HUDPosition{}, Act: 1, LevelID: 1, Cast: &SemanticCastCue{Kind: "cast_started", Player: "alice", SkillID: 1}},
		{ID: 2, Type: "state", Tick: 10, Position: HUDPosition{}, Act: 1, LevelID: 1, State: &SemanticStateCue{Kind: "applied", StateID: "armor"}},
	}}
	if err := validateEventView(valid, 10); err != nil {
		t.Fatalf("valid event view: %v", err)
	}
	for name, mutate := range map[string]func(*EventView){
		"duplicate id":       func(view *EventView) { view.Events[1].ID = 1 },
		"wrong union":        func(view *EventView) { view.Events[0].State = &SemanticStateCue{Kind: "applied", StateID: "armor"} },
		"non-finite target":  func(view *EventView) { view.Events[0].Cast.Target.X = math.NaN() },
		"unknown type":       func(view *EventView) { view.Events[0].Type = "payload" },
		"wrong tail window":  func(view *EventView) { view.FromTick = 1 },
		"bad overlay height": func(view *EventView) { view.Events[1].OverlayHeight = 5 },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := valid
			candidate.Events = append([]SemanticEvent(nil), valid.Events...)
			cast := *valid.Events[0].Cast
			candidate.Events[0].Cast = &cast
			mutate(&candidate)
			if err := validateEventView(candidate, 10); err == nil {
				t.Fatal("invalid event projection accepted")
			}
		})
	}
}

func TestProjectEventViewCapsTheNewestSemanticEvents(t *testing.T) {
	checkpoint := semanticCheckpoint()
	for index := range checkpoint.Snapshot.Components {
		component := &checkpoint.Snapshot.Components[index]
		if component.Name != "d2legacy.skill.cast_cue" {
			continue
		}
		component.Instances = nil
		for offset := range 300 {
			kind, player, targetID := "cast_effect", "alice", ""
			tick, effectTick, skillID := int64(98), int64(98), int64(47)
			caster := uint64(2)
			targetX, targetY := math.Float64bits(20), math.Float64bits(21)
			component.Instances = append(component.Instances, gameecs.InstanceSnapshot{Entity: uint64(100 + offset), Values: []gameecs.ValueSnapshot{
				{String: &kind}, {Int: &tick}, {Int: &effectTick}, {Entity: &caster}, {String: &player}, {Int: &skillID},
				{Float: &targetX}, {Float: &targetY}, {String: &targetID},
			}})
		}
	}
	view, err := ProjectEventView("alice", checkpoint)
	if err != nil {
		t.Fatal(err)
	}
	if !view.Truncated || len(view.Events) != MaxEventViewEvents || view.Events[0].ID != 146 ||
		view.Events[len(view.Events)-3].ID != 399 || view.Events[len(view.Events)-2].Type != "state" || view.Events[len(view.Events)-1].Type != "monster_death" {
		t.Fatalf("bounded semantic tail first=%d penultimate=%d last=%+v count=%d truncated=%v", view.Events[0].ID, view.Events[len(view.Events)-2].ID, view.Events[len(view.Events)-1], len(view.Events), view.Truncated)
	}
}

func semanticCheckpoint() simulation.Checkpoint {
	stringValue := func(value string) gameecs.ValueSnapshot { return gameecs.ValueSnapshot{String: &value} }
	intValue := func(value int64) gameecs.ValueSnapshot { return gameecs.ValueSnapshot{Int: &value} }
	entityValue := func(value uint64) gameecs.ValueSnapshot { return gameecs.ValueSnapshot{Entity: &value} }
	floatValue := func(value float64) gameecs.ValueSnapshot {
		bits := math.Float64bits(value)
		return gameecs.ValueSnapshot{Float: &bits}
	}
	component := func(name string, fields []string, instances ...gameecs.InstanceSnapshot) gameecs.ComponentSnapshot {
		result := gameecs.ComponentSnapshot{Name: name, Version: 1, Instances: instances}
		for _, name := range fields {
			result.Fields = append(result.Fields, gameecs.FieldSnapshot{Name: name, Kind: akara.FieldString})
		}
		return result
	}
	player, caster, target := uint64(1), uint64(2), uint64(3)
	return simulation.Checkpoint{Tick: 100, Snapshot: &gameecs.Snapshot{Version: gameecs.SnapshotVersion, Tick: 100, Entities: []uint64{player, caster, target, 20, 21, 22, 23}, Components: []gameecs.ComponentSnapshot{
		component("d2legacy.player.identity", []string{"player"}, gameecs.InstanceSnapshot{Entity: player, Values: []gameecs.ValueSnapshot{stringValue("alice")}}),
		component("d2legacy.world.position", []string{"x", "y"},
			gameecs.InstanceSnapshot{Entity: player, Values: []gameecs.ValueSnapshot{floatValue(10), floatValue(10)}},
			gameecs.InstanceSnapshot{Entity: caster, Values: []gameecs.ValueSnapshot{floatValue(12), floatValue(10)}},
			gameecs.InstanceSnapshot{Entity: target, Values: []gameecs.ValueSnapshot{floatValue(13), floatValue(10)}}),
		component("d2legacy.world.location", []string{"act", "level_id"},
			gameecs.InstanceSnapshot{Entity: player, Values: []gameecs.ValueSnapshot{intValue(1), intValue(2)}},
			gameecs.InstanceSnapshot{Entity: caster, Values: []gameecs.ValueSnapshot{intValue(1), intValue(2)}},
			gameecs.InstanceSnapshot{Entity: target, Values: []gameecs.ValueSnapshot{intValue(1), intValue(2)}}),
		component("d2legacy.world.facing", []string{"direction"}, gameecs.InstanceSnapshot{Entity: caster, Values: []gameecs.ValueSnapshot{intValue(3)}}),
		component("d2legacy.monster.appearance", []string{"overlay_height"}, gameecs.InstanceSnapshot{Entity: target, Values: []gameecs.ValueSnapshot{intValue(4)}}),
		component("d2legacy.monster.identity", []string{"spawn_id"}, gameecs.InstanceSnapshot{Entity: target, Values: []gameecs.ValueSnapshot{stringValue("fallen-a")}}),
		component("d2legacy.skill.cast_cue", []string{"kind", "tick", "effect_tick", "caster", "player", "skill_id", "target_x", "target_y", "target_id"},
			gameecs.InstanceSnapshot{Entity: 20, Values: []gameecs.ValueSnapshot{stringValue("cast_effect"), intValue(98), intValue(98), entityValue(caster), stringValue("alice"), intValue(47), floatValue(20), floatValue(21), stringValue("")}},
			gameecs.InstanceSnapshot{Entity: 22, Values: []gameecs.ValueSnapshot{stringValue("cast_started"), intValue(2), intValue(5), entityValue(caster), stringValue("alice"), intValue(47), floatValue(20), floatValue(21), stringValue("")}}),
		component("d2legacy.state.event", []string{"kind", "tick", "target", "state_id", "source_id", "expires_tick", "reason"},
			gameecs.InstanceSnapshot{Entity: 21, Values: []gameecs.ValueSnapshot{stringValue("applied"), intValue(99), entityValue(target), stringValue("frozenarmor"), stringValue("skill:40"), intValue(200), stringValue("")}}),
		component("d2legacy.monster.death_event", []string{"kind", "tick", "monster_id"},
			gameecs.InstanceSnapshot{Entity: 23, Values: []gameecs.ValueSnapshot{stringValue("monster_death_presented"), intValue(100), stringValue("fallen-a")}}),
	}}}
}
