package movement

import (
	"bytes"
	"errors"
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/inputstate"
)

type scriptedPaths struct{}

func (scriptedPaths) FindPath(request gameworld.PathRequest) ([]gameworld.Point, error) {
	if request.Goal.X == 99 {
		return nil, errors.New("blocked target")
	}
	return []gameworld.Point{request.Start, {X: 11, Y: 10}, request.Goal}, nil
}

func TestMovementSourceHonorsGameplayInputRouting(t *testing.T) {
	engine := gameecs.New()
	controls, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2legacy.world.player_control", Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}})
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

func TestMovementSourceAdmitsQueuedAndKeyboardRunIntent(t *testing.T) {
	engine := gameecs.New()
	controls, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2legacy.world.player_control", Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}})
	if err != nil {
		t.Fatal(err)
	}
	entity := engine.World().MustCreateEntity()
	if _, err := controls.Set(entity, map[string]any{"player": "alpha"}); err != nil {
		t.Fatal(err)
	}
	var input inputstate.Store
	controller := &MovementController{}
	controller.SetRunning(true)
	source, err := NewMovementSource(engine, &input, "alpha", "game_world", controller)
	if err != nil {
		t.Fatal(err)
	}
	input.Publish(inputstate.Frame{Owner: inputstate.FocusOwner{Domain: inputstate.FocusScene, ID: "game_world"}})
	payload, err := decodeMove(source.Commands(1)[0].Payload)
	if err != nil || !payload.Running {
		t.Fatalf("queued run payload = %#v, %v", payload, err)
	}
	input.Publish(inputstate.Frame{Actions: map[string]inputstate.ActionState{"toggle_run": {Pressed: true}}, Owner: inputstate.FocusOwner{Domain: inputstate.FocusScene, ID: "game_world"}})
	payload, err = decodeMove(source.Commands(2)[0].Payload)
	if err != nil || payload.Running {
		t.Fatalf("keyboard toggle payload = %#v, %v", payload, err)
	}
}

func TestMovementSourceEmitsPointerWorldTarget(t *testing.T) {
	engine := gameecs.New()
	controls, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2legacy.world.player_control", Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}})
	if err != nil {
		t.Fatal(err)
	}
	entity := engine.World().MustCreateEntity()
	if _, err := controls.Set(entity, map[string]any{"player": "alpha"}); err != nil {
		t.Fatal(err)
	}
	var input inputstate.Store
	input.Publish(inputstate.Frame{Owner: inputstate.FocusOwner{Domain: inputstate.FocusScene, ID: "game_world"}})
	controller := &MovementController{}
	if err := controller.SetMoveTarget(12.5, 44.25); err != nil {
		t.Fatal(err)
	}
	source, err := NewMovementSource(engine, &input, "alpha", "game_world", controller)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := decodeMove(source.Commands(1)[0].Payload)
	if err != nil {
		t.Fatal(err)
	}
	if payload.Target == nil || payload.Target.X != 12.5 || payload.Target.Y != 44.25 {
		t.Fatalf("pointer target payload = %#v", payload)
	}
	encoded := source.Commands(2)[0].Payload
	if !bytes.Contains(encoded, []byte(`"target":{"x":12.5,"y":44.25,"stop_radius":0}`)) {
		t.Fatalf("pointer target wire schema = %s", encoded)
	}
}

func TestMovementSourceKeepsAcceptedRouteWhenReplacementIsBlocked(t *testing.T) {
	engine := gameecs.New()
	controls, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2legacy.world.player_control", Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}})
	if err != nil {
		t.Fatal(err)
	}
	positions, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2legacy.world.position", Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}})
	if err != nil {
		t.Fatal(err)
	}
	entity := engine.World().MustCreateEntity()
	if _, err := controls.Set(entity, map[string]any{"player": "alpha"}); err != nil {
		t.Fatal(err)
	}
	if _, err := positions.Set(entity, map[string]any{"x": 10.0, "y": 10.0}); err != nil {
		t.Fatal(err)
	}
	var input inputstate.Store
	input.Publish(inputstate.Frame{Owner: inputstate.FocusOwner{Domain: inputstate.FocusScene, ID: "game_world"}})
	controller := &MovementController{}
	if err := controller.SetMoveTarget(20, 10); err != nil {
		t.Fatal(err)
	}
	source, err := NewMovementSource(engine, &input, "alpha", "game_world", controller)
	if err != nil {
		t.Fatal(err)
	}
	source.navigation = scriptedPaths{}
	first, err := decodeMove(source.Commands(1)[0].Payload)
	if err != nil || first.Target == nil || first.Target.X != 11 || first.Target.Y != 10 {
		t.Fatalf("accepted route waypoint = %#v, %v", first.Target, err)
	}
	if err := controller.SetMoveTarget(99, 10); err != nil {
		t.Fatal(err)
	}
	second, err := decodeMove(source.Commands(2)[0].Payload)
	if err != nil || second.Target == nil || second.Target.X != 11 || second.Target.Y != 10 {
		t.Fatalf("blocked replacement discarded accepted route: %#v, %v", second.Target, err)
	}
	retained := controller.moveTarget()
	if retained == nil || retained.X != 20 || retained.Y != 10 {
		t.Fatalf("controller retained target = %#v, want original target", retained)
	}
}
