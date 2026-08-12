package session

import (
	"encoding/json"
	"errors"
	"math"
	"testing"

	"github.com/gravestench/akara"
	gameaction "github.com/gravestench/dark-magic/internal/game/action"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
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

func TestMovementCommandOnlyMutatesOwnedPlayerEntity(t *testing.T) {
	engine := gameecs.New()
	controls, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2.world.player_control", Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}})
	if err != nil {
		t.Fatal(err)
	}
	velocities, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2.world.velocity", Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}})
	if err != nil {
		t.Fatal(err)
	}
	modes, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2.player.movement_mode", Fields: []akara.Field{{Name: "running", Kind: akara.FieldBool}}})
	if err != nil {
		t.Fatal(err)
	}
	animations, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2.player.animation", Fields: []akara.Field{{Name: "direction", Kind: akara.FieldInt64}, {Name: "mode", Kind: akara.FieldString}}})
	if err != nil {
		t.Fatal(err)
	}
	positions, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2.world.position", Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}})
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
	mode, err := modes.Set(entity, nil)
	if err != nil {
		t.Fatal(err)
	}
	animationState, err := animations.Set(entity, map[string]any{"direction": int64(2), "mode": "NU"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := positions.Set(entity, map[string]any{"x": 10.0, "y": 20.0}); err != nil {
		t.Fatal(err)
	}
	pendingActions, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: gameaction.AttackApproachComponent, Fields: []akara.Field{{Name: "target_id", Kind: akara.FieldString}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pendingActions.Set(entity, map[string]any{"target_id": "monster:fallen"}); err != nil {
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
	payload, _ := json.Marshal(MovePayload{X: 1, Running: true})
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
	if x, _ := velocity.Get("x"); x != float64(15) {
		t.Fatalf("owner did not move entity: x=%v", x)
	}
	if running, _ := mode.Get("running"); running != true {
		t.Fatalf("movement mode = %v, want running", running)
	}
	if animation, _ := animationState.Get("mode"); animation != "RN" {
		t.Fatalf("animation mode = %v, want RN", animation)
	}
	if direction, _ := animationState.Get("direction"); direction != int64(3) {
		t.Fatalf("animation direction = %v, want east/3", direction)
	}
	if pendingActions.Has(entity) {
		t.Fatal("explicit movement did not cancel pending attack approach")
	}
	payload, _ = json.Marshal(MovePayload{X: 1, Y: 1})
	if err := session.Submit(simulation.Command{Tick: 3, Player: "alpha", Sequence: 2, Kind: MoveCommand, Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	x, _ := velocity.Get("x")
	y, _ := velocity.Get("y")
	if magnitude := math.Hypot(x.(float64), y.(float64)); math.Abs(magnitude-10) > 1e-9 {
		t.Fatalf("diagonal speed = %v, want normalized walk speed 10", magnitude)
	}
	payload, _ = json.Marshal(MovePayload{Target: &MoveTarget{X: 20, Y: 20}})
	if err := session.Submit(simulation.Command{Tick: 4, Player: "alpha", Sequence: 3, Kind: MoveCommand, Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	x, _ = velocity.Get("x")
	y, _ = velocity.Get("y")
	if x != float64(10) || y != float64(0) {
		t.Fatalf("target velocity = %v,%v", x, y)
	}
	if _, err := pendingActions.Set(entity, map[string]any{"target_id": "monster:new-target"}); err != nil {
		t.Fatal(err)
	}
	idlePayload, _ := json.Marshal(MovePayload{})
	if err := session.Submit(simulation.Command{Tick: 5, Player: "alpha", Sequence: 4, Kind: MoveCommand, Payload: idlePayload}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	if !pendingActions.Has(entity) {
		t.Fatal("idle movement snapshot canceled pending attack approach")
	}
	if x, _ := velocity.Get("x"); x != float64(10) {
		t.Fatalf("idle movement snapshot overwrote attack-owned velocity: x=%v", x)
	}
	if animation, _ := animationState.Get("mode"); animation != "WL" {
		t.Fatalf("idle movement snapshot overwrote attack-owned animation: mode=%v", animation)
	}
}

func TestMovementDirectionMatchesLegacyEightWayEncoding(t *testing.T) {
	tests := []struct {
		x, y int
		want int64
	}{
		{0, 1, 0}, {-1, 0, 1}, {0, -1, 2}, {1, 0, 3},
		{1, 1, 4}, {-1, 1, 5}, {-1, -1, 6}, {1, -1, 7},
	}
	for _, test := range tests {
		if got := movementDirection(test.x, test.y); got != test.want {
			t.Fatalf("direction(%d,%d) = %d, want %d", test.x, test.y, got, test.want)
		}
	}
}

func TestMovementSourceHonorsGameplayInputRouting(t *testing.T) {
	engine := gameecs.New()
	controls, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2.world.player_control", Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}})
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
	controls, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2.world.player_control", Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}})
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
	controls, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2.world.player_control", Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}})
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
}

func TestSkillIntentStopsAnOlderPointerRoute(t *testing.T) {
	controller := &MovementController{}
	if err := controller.SetMoveTarget(20, 10); err != nil {
		t.Fatal(err)
	}
	if err := controller.UseSkill("left", 12, 10, ""); err != nil {
		t.Fatal(err)
	}
	if controller.HasMoveTarget() {
		t.Fatal("stand-still skill retained older movement route")
	}
}

func TestMovementSourceKeepsAcceptedRouteWhenReplacementIsBlocked(t *testing.T) {
	engine := gameecs.New()
	controls, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2.world.player_control", Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}})
	if err != nil {
		t.Fatal(err)
	}
	positions, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2.world.position", Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}})
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
