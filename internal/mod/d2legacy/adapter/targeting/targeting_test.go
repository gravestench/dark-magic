package targeting

import (
	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"testing"
)

func TestResolverUsesSpawnedEntityKindsAndPriority(t *testing.T) {
	engine := gameecs.New()
	positions, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2legacy.world.position", Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}})
	if err != nil {
		t.Fatal(err)
	}
	selectables, err := akara.RegisterSchema(engine.World(), Schema())
	if err != nil {
		t.Fatal(err)
	}
	spawn := func(id, kind string, priority int64) {
		entity := engine.World().MustCreateEntity()
		if _, err := positions.Set(entity, map[string]any{"x": 10.0, "y": 10.0}); err != nil {
			t.Fatal(err)
		}
		if _, err := selectables.Set(entity, map[string]any{"id": id, "kind": kind, "label": id, "owner": "", "radius": 2.0, "priority": priority}); err != nil {
			t.Fatal(err)
		}
	}
	spawn("npc:1", "npc", 0)
	spawn("item:1", "item", 5)
	hit, found := New(engine).HitAt(10, 10)
	if !found || hit.ID != "item:1" || hit.Kind != "item" {
		t.Fatalf("hit=%#v,%v", hit, found)
	}
}
