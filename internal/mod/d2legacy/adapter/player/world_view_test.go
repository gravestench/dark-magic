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
	selectable := registerProjectionStore(t, engine, "d2legacy.world.selectable", []akara.Field{{Name: "id", Kind: akara.FieldString}, {Name: "kind", Kind: akara.FieldString}, {Name: "label", Kind: akara.FieldString}, {Name: "owner", Kind: akara.FieldString}, {Name: "radius", Kind: akara.FieldFloat64}, {Name: "priority", Kind: akara.FieldInt64}})
	monster := registerProjectionStore(t, engine, "d2legacy.monster.stats", []akara.Field{{Name: "health", Kind: akara.FieldInt64}, {Name: "max_health", Kind: akara.FieldInt64}, {Name: "hidden_damage", Kind: akara.FieldInt64}})
	secret := registerProjectionStore(t, engine, "d2legacy.monster.ai", []akara.Field{{Name: "target_id", Kind: akara.FieldString}})
	player := engine.World().MustCreateEntity()
	nearB := engine.World().MustCreateEntity()
	nearA := engine.World().MustCreateEntity()
	far := engine.World().MustCreateEntity()
	_, _ = identity.Set(player, map[string]any{"character_id": "character", "player": "alice", "name": "Alice", "class": "Amazon"})
	_, _ = position.Set(player, map[string]any{"x": 10.0, "y": 10.0})
	setPublic := func(entity akara.Entity, id string, x, y float64) {
		_, _ = position.Set(entity, map[string]any{"x": x, "y": y})
		_, _ = selectable.Set(entity, map[string]any{"id": id, "kind": "monster", "label": "Fallen", "owner": "", "radius": 0.75, "priority": int64(2)})
		_, _ = monster.Set(entity, map[string]any{"health": int64(8), "max_health": int64(10), "hidden_damage": int64(9999)})
		_, _ = secret.Set(entity, map[string]any{"target_id": "alice-secret"})
	}
	setPublic(nearB, "monster:b", 12, 10)
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
	if len(view.Entities) != 2 || view.Entities[0].ID != "monster:a" || view.Entities[1].ID != "monster:b" || view.Entities[0].Health == nil || *view.Entities[0].Health != 8 {
		t.Fatalf("world view = %#v", view)
	}
	if strings.Contains(string(payload), "9999") || strings.Contains(string(payload), "alice-secret") || strings.Contains(string(payload), "monster:far") {
		t.Fatalf("world view leaked hidden/far state: %s", payload)
	}
}

func TestProjectWorldViewRejectsDuplicatePublicIDs(t *testing.T) {
	engine := gameecs.New()
	t.Cleanup(func() { _ = engine.Close() })
	identity := registerProjectionStore(t, engine, "d2legacy.player.identity", []akara.Field{{Name: "player", Kind: akara.FieldString}})
	position := registerProjectionStore(t, engine, "d2legacy.world.position", []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}})
	selectable := registerProjectionStore(t, engine, "d2legacy.world.selectable", []akara.Field{{Name: "id", Kind: akara.FieldString}, {Name: "kind", Kind: akara.FieldString}, {Name: "label", Kind: akara.FieldString}, {Name: "owner", Kind: akara.FieldString}, {Name: "radius", Kind: akara.FieldFloat64}, {Name: "priority", Kind: akara.FieldInt64}})
	player := engine.World().MustCreateEntity()
	_, _ = identity.Set(player, map[string]any{"player": "alice"})
	_, _ = position.Set(player, map[string]any{"x": 0.0, "y": 0.0})
	for range 2 {
		entity := engine.World().MustCreateEntity()
		_, _ = position.Set(entity, map[string]any{"x": 1.0, "y": 1.0})
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

func registerProjectionStore(t *testing.T, engine *gameecs.Engine, name string, fields []akara.Field) *akara.DynamicStore {
	t.Helper()
	store, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: name, Version: 1, Fields: fields})
	if err != nil {
		t.Fatal(err)
	}
	return store
}
