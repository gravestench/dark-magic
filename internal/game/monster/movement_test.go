package monster

import (
	"testing"
	"time"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
)

func TestMonsterMovementAppliesFixedStepVelocity(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	if err := RegisterMovement(engine, nil); err != nil {
		t.Fatal(err)
	}
	ai, _, positions, velocities, _, _, err := registerAIStores(engine)
	if err != nil {
		t.Fatal(err)
	}
	colliders, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.collider")
	entity := engine.World().MustCreateEntity()
	set := func(store *akara.DynamicStore, values map[string]any) {
		if _, err := store.Set(entity, values); err != nil {
			t.Fatal(err)
		}
	}
	set(ai, map[string]any{"behavior": "test", "state": AIChase, "target_id": "player:hero", "next_think_tick": int64(1), "think_interval": int64(1), "aggro_radius": 10.0, "attack_range": 1.0, "speed": 10.0})
	set(positions, map[string]any{"x": 2.0, "y": 3.0})
	set(velocities, map[string]any{"x": 10.0, "y": -5.0})
	set(colliders, map[string]any{"radius": 0.5})
	if err := engine.Update(40 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	position, _ := positions.Get(entity)
	x, _ := position.Get("x")
	y, _ := position.Get("y")
	if x != 2.4 || y != 2.8 {
		t.Fatalf("position = %v,%v", x, y)
	}
}
