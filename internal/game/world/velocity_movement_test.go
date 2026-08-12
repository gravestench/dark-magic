package world

import (
	"testing"
	"time"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
)

type countingVelocityPaths struct {
	pathCalls int
	stepCalls int
	allowed   bool
}

func (paths *countingVelocityPaths) FindPath(request PathRequest) ([]Point, error) {
	paths.pathCalls++
	return []Point{request.Start, request.Goal}, nil
}

func (paths *countingVelocityPaths) WalkableStep(_, _ Point, _ float64) bool {
	paths.stepCalls++
	return paths.allowed
}

func TestVelocityMovementSkipsPathSearchInsideCollisionCell(t *testing.T) {
	engine := gameecs.NewWithClock(40*time.Millisecond, 1)
	defer engine.Close()
	paths := &countingVelocityPaths{allowed: true}
	components := VelocityComponents{
		Position: "test.position", Velocity: "test.velocity", Collider: "test.collider",
	}
	if err := RegisterVelocityMovement(engine, paths, components); err != nil {
		t.Fatal(err)
	}
	position, _ := akara.GetDynamicStore(engine.World(), components.Position)
	velocity, _ := akara.GetDynamicStore(engine.World(), components.Velocity)
	collider, _ := akara.GetDynamicStore(engine.World(), components.Collider)
	marker, _ := akara.GetDynamicStore(engine.World(), "engine.world.velocity_mover")
	entity := engine.World().MustCreateEntity()
	if _, err := position.Set(entity, map[string]any{"x": 10.0, "y": 10.0}); err != nil {
		t.Fatal(err)
	}
	if _, err := velocity.Set(entity, map[string]any{"x": 1.0, "y": 0.0}); err != nil {
		t.Fatal(err)
	}
	if _, err := collider.Set(entity, map[string]any{"radius": 0.5}); err != nil {
		t.Fatal(err)
	}
	if _, err := marker.Set(entity, map[string]any{}); err != nil {
		t.Fatal(err)
	}
	if err := engine.Update(40 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if paths.pathCalls != 0 || paths.stepCalls != 0 {
		t.Fatalf("same-cell velocity performed path/step checks %d/%d, want 0/0", paths.pathCalls, paths.stepCalls)
	}
	if _, err := velocity.Set(entity, map[string]any{"x": 20.0, "y": 0.0}); err != nil {
		t.Fatal(err)
	}
	if err := engine.Update(40 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if paths.pathCalls != 0 || paths.stepCalls != 1 {
		t.Fatalf("adjacent velocity performed path/step checks %d/%d, want 0/1", paths.pathCalls, paths.stepCalls)
	}
}
