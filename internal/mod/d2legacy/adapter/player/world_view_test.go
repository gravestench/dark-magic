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
	selectable := registerProjectionStore(t, engine, "d2legacy.world.selectable", []akara.Field{{Name: "id", Kind: akara.FieldString}, {Name: "kind", Kind: akara.FieldString}, {Name: "label", Kind: akara.FieldString}, {Name: "owner", Kind: akara.FieldString}, {Name: "radius", Kind: akara.FieldFloat64}, {Name: "priority", Kind: akara.FieldInt64}})
	inactive := registerProjectionStore(t, engine, "d2legacy.world.inactive", nil)
	monster := registerProjectionStore(t, engine, "d2legacy.monster.stats", []akara.Field{{Name: "health", Kind: akara.FieldInt64}, {Name: "max_health", Kind: akara.FieldInt64}, {Name: "hidden_damage", Kind: akara.FieldInt64}})
	secret := registerProjectionStore(t, engine, "d2legacy.monster.ai", []akara.Field{{Name: "target_id", Kind: akara.FieldString}})
	player := engine.World().MustCreateEntity()
	nearB := engine.World().MustCreateEntity()
	nearA := engine.World().MustCreateEntity()
	far := engine.World().MustCreateEntity()
	_, _ = identity.Set(player, map[string]any{"character_id": "character", "player": "alice", "name": "Alice", "class": "Amazon"})
	_, _ = position.Set(player, map[string]any{"x": 10.0, "y": 10.0})
	_, _ = location.Set(player, map[string]any{"act": int64(1), "level_id": int64(2)})
	setPublic := func(entity akara.Entity, id string, x, y float64) {
		_, _ = position.Set(entity, map[string]any{"x": x, "y": y})
		_, _ = location.Set(entity, map[string]any{"act": int64(1), "level_id": int64(2)})
		_, _ = selectable.Set(entity, map[string]any{"id": id, "kind": "monster", "label": "Fallen", "owner": "", "radius": 0.75, "priority": int64(2)})
		_, _ = monster.Set(entity, map[string]any{"health": int64(8), "max_health": int64(10), "hidden_damage": int64(9999)})
		_, _ = secret.Set(entity, map[string]any{"target_id": "alice-secret"})
	}
	setPublic(nearB, "monster:b", 12, 10)
	_, _ = inactive.Set(nearB, nil)
	setPublic(nearA, "monster:a", 8, 10)
	setPublic(far, "monster:far", 500, 500)
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
	if len(view.Entities) != 1 || view.Entities[0].ID != "monster:a" || view.Entities[0].Health == nil || *view.Entities[0].Health != 8 {
		t.Fatalf("world view = %#v", view)
	}
	if strings.Contains(string(payload), "9999") || strings.Contains(string(payload), "alice-secret") || strings.Contains(string(payload), "monster:b") || strings.Contains(string(payload), "monster:far") {
		t.Fatalf("world view leaked hidden/far state: %s", payload)
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
