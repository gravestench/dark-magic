package session

import (
	"encoding/json"
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func TestMovementCommandOnlyMutatesOwnedPlayerEntity(t *testing.T) {
	engine := gameecs.New()
	controls, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "dm.world.player_control", Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}})
	if err != nil {
		t.Fatal(err)
	}
	velocities, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "dm.world.velocity", Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}})
	if err != nil {
		t.Fatal(err)
	}
	entity := engine.World().MustCreateEntity()
	if _, err := controls.Set(entity, map[string]any{"player": "alpha"}); err != nil {
		t.Fatal(err)
	}
	velocity, err := velocities.Set(entity, nil)
	if err != nil {
		t.Fatal(err)
	}
	session, err := New(engine, Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := RegisterMovement(session); err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(MovePayload{X: 1})
	if err := session.Submit(simulation.Command{Tick: 1, Player: "bravo", Sequence: 1, Kind: MoveCommand, Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	if x, _ := velocity.Get("x"); x != float64(0) {
		t.Fatalf("foreign player moved entity: x=%v", x)
	}
	if err := session.Submit(simulation.Command{Tick: 2, Player: "alpha", Sequence: 1, Kind: MoveCommand, Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	if x, _ := velocity.Get("x"); x != float64(10) {
		t.Fatalf("owner did not move entity: x=%v", x)
	}
}
