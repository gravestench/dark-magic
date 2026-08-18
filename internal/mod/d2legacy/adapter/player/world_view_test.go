package player

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func TestProjectWorldViewFiltersOrdersAndBoundsPublicState(t *testing.T) {
	engine := gameecs.New()
	t.Cleanup(func() { _ = engine.Close() })
	identity := registerProjectionStore(t, engine, "d2legacy.player.identity", []akara.Field{{Name: "character_id", Kind: akara.FieldString}, {Name: "player", Kind: akara.FieldString}, {Name: "name", Kind: akara.FieldString}, {Name: "class", Kind: akara.FieldString}})
	position := registerProjectionStore(t, engine, "d2legacy.world.position", []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}})
	location := registerProjectionStore(t, engine, "d2legacy.world.location", []akara.Field{{Name: "act", Kind: akara.FieldInt64}, {Name: "level_id", Kind: akara.FieldInt64}})
	velocity := registerProjectionStore(t, engine, "d2legacy.world.velocity", []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}})
	selectable := registerProjectionStore(t, engine, "d2legacy.world.selectable", []akara.Field{{Name: "id", Kind: akara.FieldString}, {Name: "kind", Kind: akara.FieldString}, {Name: "label", Kind: akara.FieldString}, {Name: "owner", Kind: akara.FieldString}, {Name: "radius", Kind: akara.FieldFloat64}, {Name: "priority", Kind: akara.FieldInt64}})
	inactive := registerProjectionStore(t, engine, "d2legacy.world.inactive", nil)
	monster := registerProjectionStore(t, engine, "d2legacy.monster.stats", []akara.Field{{Name: "health", Kind: akara.FieldInt64}, {Name: "max_health", Kind: akara.FieldInt64}, {Name: "hidden_damage", Kind: akara.FieldInt64}})
	monsterIdentity := registerProjectionStore(t, engine, "d2legacy.monster.identity", []akara.Field{{Name: "spawn_id", Kind: akara.FieldString}, {Name: "definition_id", Kind: akara.FieldString}})
	monsterAppearance := registerProjectionStore(t, engine, "d2legacy.monster.appearance", []akara.Field{
		{Name: "token", Kind: akara.FieldString}, {Name: "mode", Kind: akara.FieldString},
		{Name: "weapon_class", Kind: akara.FieldString}, {Name: "components", Kind: akara.FieldString},
		{Name: "name_key", Kind: akara.FieldString}, {Name: "death_sound", Kind: akara.FieldString},
		{Name: "overlay_height", Kind: akara.FieldInt64},
	})
	monsterDeath := registerProjectionStore(t, engine, "d2legacy.monster.death", []akara.Field{{Name: "tick", Kind: akara.FieldInt64}, {Name: "drops", Kind: akara.FieldString}})
	monsterAI := registerProjectionStore(t, engine, "d2legacy.monster.ai", []akara.Field{{Name: "target_id", Kind: akara.FieldString}, {Name: "state", Kind: akara.FieldString}})
	auraEffect := registerProjectionStore(t, engine, "d2legacy.skill.aura_effect", []akara.Field{
		{Name: "emitter", Kind: akara.FieldEntity}, {Name: "target", Kind: akara.FieldEntity},
		{Name: "source_id", Kind: akara.FieldString}, {Name: "skill_id", Kind: akara.FieldInt64},
		{Name: "skill_level", Kind: akara.FieldInt64}, {Name: "state_id", Kind: akara.FieldString},
		{Name: "refresh_delay", Kind: akara.FieldInt64},
	})
	player := engine.World().MustCreateEntity()
	nearB := engine.World().MustCreateEntity()
	nearA := engine.World().MustCreateEntity()
	far := engine.World().MustCreateEntity()
	corpse := engine.World().MustCreateEntity()
	_, _ = identity.Set(player, map[string]any{"character_id": "character", "player": "alice", "name": "Alice", "class": "Amazon"})
	_, _ = position.Set(player, map[string]any{"x": 10.0, "y": 10.0})
	_, _ = location.Set(player, map[string]any{"act": int64(1), "level_id": int64(2)})
	effect := engine.World().MustCreateEntity()
	_, _ = auraEffect.Set(effect, map[string]any{
		"emitter": player, "target": player, "source_id": "server-secret-source",
		"skill_id": int64(98), "skill_level": int64(20), "state_id": "might", "refresh_delay": int64(50),
	})
	secondEffect := engine.World().MustCreateEntity()
	_, _ = auraEffect.Set(secondEffect, map[string]any{
		"emitter": nearA, "target": player, "source_id": "second-secret-source",
		"skill_id": int64(99), "skill_level": int64(9), "state_id": "prayer", "refresh_delay": int64(75),
	})
	setPublic := func(entity akara.Entity, id string, x, y float64) {
		_, _ = position.Set(entity, map[string]any{"x": x, "y": y})
		_, _ = location.Set(entity, map[string]any{"act": int64(1), "level_id": int64(2)})
		_, _ = selectable.Set(entity, map[string]any{"id": id, "kind": "monster", "label": "Fallen", "owner": "", "radius": 0.75, "priority": int64(2)})
		_, _ = monster.Set(entity, map[string]any{"health": int64(8), "max_health": int64(10), "hidden_damage": int64(9999)})
		_, _ = monsterAI.Set(entity, map[string]any{"target_id": "alice-secret"})
	}
	setPublic(nearB, "monster:b", 12, 10)
	_, _ = inactive.Set(nearB, nil)
	setPublic(nearA, "monster:a", 8, 10)
	_, _ = velocity.Set(nearA, map[string]any{"x": 1.0, "y": 0.0})
	_, _ = monsterAI.Set(nearA, map[string]any{"target_id": "alice-secret", "state": "attack"})
	_, _ = monsterIdentity.Set(nearA, map[string]any{"spawn_id": "a", "definition_id": "fallen1"})
	_, _ = monsterAppearance.Set(nearA, map[string]any{
		"token": "FA", "mode": "NU", "weapon_class": "HTH", "components": "HD=LIT",
		"name_key": "Fallen", "death_sound": "fallen_death", "overlay_height": int64(3),
	})
	setPublic(far, "monster:far", 500, 500)
	_, _ = position.Set(corpse, map[string]any{"x": 9.0, "y": 12.0})
	_, _ = location.Set(corpse, map[string]any{"act": int64(1), "level_id": int64(2)})
	_, _ = monster.Set(corpse, map[string]any{"health": int64(0), "max_health": int64(10), "hidden_damage": int64(7777)})
	_, _ = monsterIdentity.Set(corpse, map[string]any{"spawn_id": "corpse", "definition_id": "fallen1"})
	_, _ = monsterAppearance.Set(corpse, map[string]any{
		"token": "FA", "mode": "DT", "weapon_class": "HTH", "components": "HD=LIT",
		"name_key": "Fallen", "death_sound": "fallen_death", "overlay_height": int64(3),
	})
	_, _ = monsterDeath.Set(corpse, map[string]any{"tick": int64(7), "drops": "server-only-drop"})
	snapshot, err := engine.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	payload, err := ProjectWorldView("alice", simulation.Checkpoint{Tick: snapshot.Tick, Snapshot: &snapshot})
	if err != nil {
		t.Fatal(err)
	}
	var view WorldView
	if err := json.Unmarshal(payload, &view); err != nil {
		t.Fatal(err)
	}
	if len(view.Entities) != 2 || view.Entities[0].ID != "monster:a" || view.Entities[0].Health == nil || *view.Entities[0].Health != 8 {
		t.Fatalf("world view = %#v", view)
	}
	living, dead := view.Entities[0], view.Entities[1]
	if living.Token != "FA" || living.Mode != "A1" || living.DeathSound != "fallen_death" || living.OverlayHeight != 3 ||
		dead.ID != "monster:corpse" || dead.Kind != "corpse" || dead.Label != "Fallen" || dead.Mode != "DT" || dead.SpawnID != "corpse" {
		t.Fatalf("living=%#v corpse=%#v", living, dead)
	}
	if len(view.States) != 2 ||
		view.States[0] != (WorldState{TargetID: "player:alice", StateID: "might", PeriodTicks: 50}) ||
		view.States[1] != (WorldState{TargetID: "player:alice", StateID: "prayer", PeriodTicks: 75}) {
		t.Fatalf("world states = %#v", view.States)
	}
	if strings.Contains(string(payload), "9999") || strings.Contains(string(payload), "7777") || strings.Contains(string(payload), "server-only-drop") ||
		strings.Contains(string(payload), "alice-secret") || strings.Contains(string(payload), "attack") ||
		strings.Contains(string(payload), "server-secret-source") || strings.Contains(string(payload), "second-secret-source") || strings.Contains(string(payload), "source_id") ||
		strings.Contains(string(payload), "skill_id") || strings.Contains(string(payload), "skill_level") ||
		strings.Contains(string(payload), "monster:b") || strings.Contains(string(payload), "monster:far") {
		t.Fatalf("world view leaked hidden/far state: %s", payload)
	}
}

func TestMonsterPresentationModePublishesOutcomeWithoutAIPolicy(t *testing.T) {
	for name, test := range map[string]struct {
		authored, ai string
		velocityX    float64
		want         string
	}{
		"authored idle": {authored: "NU", want: "NU"},
		"movement":      {authored: "NU", ai: "chase", velocityX: 1, want: "WL"},
		"attack":        {authored: "NU", ai: "attack", velocityX: 1, want: "A1"},
		"death wins":    {authored: "DT", ai: "attack", velocityX: 1, want: "DT"},
	} {
		t.Run(name, func(t *testing.T) {
			if got := monsterPresentationMode(test.authored, test.ai, test.velocityX, 0); got != test.want {
				t.Fatalf("presentation mode = %q, want %q", got, test.want)
			}
		})
	}
}

func TestProjectWorldViewRejectsDuplicatePublicIDs(t *testing.T) {
	engine := gameecs.New()
	t.Cleanup(func() { _ = engine.Close() })
	identity := registerProjectionStore(t, engine, "d2legacy.player.identity", []akara.Field{{Name: "player", Kind: akara.FieldString}})
	position := registerProjectionStore(t, engine, "d2legacy.world.position", []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}})
	location := registerProjectionStore(t, engine, "d2legacy.world.location", []akara.Field{{Name: "act", Kind: akara.FieldInt64}, {Name: "level_id", Kind: akara.FieldInt64}})
	selectable := registerProjectionStore(t, engine, "d2legacy.world.selectable", []akara.Field{{Name: "id", Kind: akara.FieldString}, {Name: "kind", Kind: akara.FieldString}, {Name: "label", Kind: akara.FieldString}, {Name: "owner", Kind: akara.FieldString}, {Name: "radius", Kind: akara.FieldFloat64}, {Name: "priority", Kind: akara.FieldInt64}})
	player := engine.World().MustCreateEntity()
	_, _ = identity.Set(player, map[string]any{"player": "alice"})
	_, _ = position.Set(player, map[string]any{"x": 0.0, "y": 0.0})
	_, _ = location.Set(player, map[string]any{"act": int64(1), "level_id": int64(1)})
	for range 2 {
		entity := engine.World().MustCreateEntity()
		_, _ = position.Set(entity, map[string]any{"x": 1.0, "y": 1.0})
		_, _ = location.Set(entity, map[string]any{"act": int64(1), "level_id": int64(1)})
		_, _ = selectable.Set(entity, map[string]any{"id": "duplicate", "kind": "object", "label": "", "owner": "", "radius": 1.0, "priority": int64(1)})
	}
	snapshot, err := engine.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ProjectWorldView("alice", simulation.Checkpoint{Snapshot: &snapshot}); err == nil {
		t.Fatal("duplicate public IDs were accepted")
	}
}

func TestProjectWorldViewPublishesPlayerMovementAnimationInputs(t *testing.T) {
	engine := gameecs.New()
	t.Cleanup(func() { _ = engine.Close() })
	identity := registerProjectionStore(t, engine, "d2legacy.player.identity", []akara.Field{{Name: "player", Kind: akara.FieldString}, {Name: "class", Kind: akara.FieldString}})
	position := registerProjectionStore(t, engine, "d2legacy.world.position", []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}})
	location := registerProjectionStore(t, engine, "d2legacy.world.location", []akara.Field{{Name: "act", Kind: akara.FieldInt64}, {Name: "level_id", Kind: akara.FieldInt64}})
	selectable := registerProjectionStore(t, engine, "d2legacy.world.selectable", []akara.Field{{Name: "id", Kind: akara.FieldString}, {Name: "kind", Kind: akara.FieldString}, {Name: "label", Kind: akara.FieldString}, {Name: "owner", Kind: akara.FieldString}, {Name: "radius", Kind: akara.FieldFloat64}, {Name: "priority", Kind: akara.FieldInt64}})
	appearance := registerProjectionStore(t, engine, "d2legacy.player.appearance", []akara.Field{{Name: "token", Kind: akara.FieldString}, {Name: "mode", Kind: akara.FieldString}})
	animation := registerProjectionStore(t, engine, "d2legacy.player.animation", []akara.Field{{Name: "mode", Kind: akara.FieldString}, {Name: "start_tick", Kind: akara.FieldInt64}})
	facing := registerProjectionStore(t, engine, "d2legacy.world.facing", []akara.Field{{Name: "direction", Kind: akara.FieldInt64}})
	movement := registerProjectionStore(t, engine, "d2legacy.player.movement_stats", []akara.Field{{Name: "velocitypercent", Kind: akara.FieldInt64}, {Name: "item_fastermovevelocity", Kind: akara.FieldInt64}})
	owner, peer := engine.World().MustCreateEntity(), engine.World().MustCreateEntity()
	for entity, player := range map[akara.Entity]string{owner: "owner", peer: "peer"} {
		_, _ = identity.Set(entity, map[string]any{"player": player, "class": "Amazon"})
		_, _ = position.Set(entity, map[string]any{"x": float64(0), "y": float64(0)})
		_, _ = location.Set(entity, map[string]any{"act": int64(1), "level_id": int64(1)})
	}
	_, _ = selectable.Set(peer, map[string]any{"id": "player:peer", "kind": "player", "label": "Peer", "owner": "peer", "radius": .75, "priority": int64(10)})
	_, _ = appearance.Set(peer, map[string]any{"token": "AM", "mode": "NU"})
	_, _ = animation.Set(peer, map[string]any{"mode": "WL", "start_tick": int64(7)})
	_, _ = facing.Set(peer, map[string]any{"direction": int64(3)})
	_, _ = movement.Set(peer, map[string]any{"velocitypercent": int64(-50), "item_fastermovevelocity": int64(100)})
	snapshot, _ := engine.Snapshot()
	payload, err := ProjectWorldView("owner", simulation.Checkpoint{Tick: snapshot.Tick, Snapshot: &snapshot})
	if err != nil {
		t.Fatal(err)
	}
	var view WorldView
	if err := json.Unmarshal(payload, &view); err != nil {
		t.Fatal(err)
	}
	if view.Version != WorldViewVersion || len(view.Entities) != 1 {
		t.Fatalf("world view = %#v", view)
	}
	got := view.Entities[0]
	if got.Class != "Amazon" || got.Token != "AM" || got.Mode != "WL" || got.Direction != 3 || got.AnimationStartTick != 7 || got.VelocityPercent != -50 || got.ItemFasterMoveVelocity != 100 {
		t.Fatalf("player projection = %#v", got)
	}
}

func TestProjectWorldViewExcludesEntitiesInAnotherLevel(t *testing.T) {
	engine := gameecs.New()
	t.Cleanup(func() { _ = engine.Close() })
	identity := registerProjectionStore(t, engine, "d2legacy.player.identity", []akara.Field{{Name: "player", Kind: akara.FieldString}})
	position := registerProjectionStore(t, engine, "d2legacy.world.position", []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}})
	location := registerProjectionStore(t, engine, "d2legacy.world.location", []akara.Field{{Name: "act", Kind: akara.FieldInt64}, {Name: "level_id", Kind: akara.FieldInt64}})
	selectable := registerProjectionStore(t, engine, "d2legacy.world.selectable", []akara.Field{{Name: "id", Kind: akara.FieldString}, {Name: "kind", Kind: akara.FieldString}, {Name: "label", Kind: akara.FieldString}, {Name: "owner", Kind: akara.FieldString}, {Name: "radius", Kind: akara.FieldFloat64}, {Name: "priority", Kind: akara.FieldInt64}})
	player := engine.World().MustCreateEntity()
	other := engine.World().MustCreateEntity()
	_, _ = identity.Set(player, map[string]any{"player": "alice"})
	_, _ = position.Set(player, map[string]any{"x": 10.0, "y": 10.0})
	_, _ = location.Set(player, map[string]any{"act": int64(1), "level_id": int64(1)})
	_, _ = position.Set(other, map[string]any{"x": 10.0, "y": 10.0})
	_, _ = location.Set(other, map[string]any{"act": int64(1), "level_id": int64(2)})
	_, _ = selectable.Set(other, map[string]any{"id": "other-level", "kind": "monster", "radius": 1.0})
	snapshot, _ := engine.Snapshot()
	payload, err := ProjectWorldView("alice", simulation.Checkpoint{Snapshot: &snapshot})
	if err != nil {
		t.Fatal(err)
	}
	var view WorldView
	if err := json.Unmarshal(payload, &view); err != nil {
		t.Fatal(err)
	}
	if len(view.Entities) != 0 {
		t.Fatalf("cross-level entities = %#v", view.Entities)
	}
}

func registerProjectionStore(t *testing.T, engine *gameecs.Engine, name string, fields []akara.Field) *akara.DynamicStore {
	t.Helper()
	store, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: name, Version: 1, Fields: fields})
	if err != nil {
		t.Fatal(err)
	}
	return store
}
