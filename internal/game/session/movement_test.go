package session

import (
	"encoding/json"
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/inputstate"
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

func TestMovementSourceHonorsGameplayInputRouting(t *testing.T) {
	engine := gameecs.New()
	controls, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "dm.world.player_control", Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}})
	if err != nil {
		t.Fatal(err)
	}
	entity := engine.World().MustCreateEntity()
	if _, err := controls.Set(entity, map[string]any{"player": "alpha"}); err != nil {
		t.Fatal(err)
	}
	var input inputstate.Store
	source, err := NewMovementSource(engine, &input, "alpha", "game_world")
	if err != nil {
		t.Fatal(err)
	}
	input.Publish(inputstate.Frame{Actions: map[string]inputstate.ActionState{"right": {Down: true}}, Owner: inputstate.FocusOwner{Domain: inputstate.FocusScene, ID: "inventory"}})
	commands := source.Commands(1)
	if len(commands) != 1 {
		t.Fatalf("commands = %d", len(commands))
	}
	payload, err := decodeMove(commands[0].Payload)
	if err != nil {
		t.Fatal(err)
	}
	if payload.X != 0 || payload.Y != 0 {
		t.Fatalf("overlay focus leaked movement: %#v", payload)
	}
	input.Publish(inputstate.Frame{Actions: map[string]inputstate.ActionState{"right": {Down: true}}, Owner: inputstate.FocusOwner{Domain: inputstate.FocusScene, ID: "inventory"}, Gameplay: true})
	payload, err = decodeMove(source.Commands(2)[0].Payload)
	if err != nil {
		t.Fatal(err)
	}
	if payload.X != 1 || payload.Y != 0 {
		t.Fatalf("passthrough overlay blocked movement: %#v", payload)
	}
	input.Publish(inputstate.Frame{Actions: map[string]inputstate.ActionState{"right": {Down: true}}, Owner: inputstate.FocusOwner{Domain: inputstate.FocusScene, ID: "game_world"}})
	payload, err = decodeMove(source.Commands(3)[0].Payload)
	if err != nil {
		t.Fatal(err)
	}
	if payload.X != 1 || payload.Y != 0 {
		t.Fatalf("gameplay focus did not admit movement: %#v", payload)
	}
}
