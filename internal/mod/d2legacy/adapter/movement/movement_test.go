package movement

import (
	"bytes"
	"errors"
	"math"
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/inputstate"
	lua "github.com/yuin/gopher-lua"
)

// scriptedPaths supplies one intermediate waypoint and rejects a sentinel target for route-retention tests.
type scriptedPaths struct{}

// FindPath makes successful and blocked route replacements deterministic without invoking production pathfinding.
func (scriptedPaths) FindPath(request gameworld.PathRequest) ([]gameworld.Point, error) {
	if request.Goal.X == 99 {
		return nil, errors.New("blocked target")
	}

	return []gameworld.Point{request.Start, {X: 11, Y: 10}, request.Goal}, nil
}

// countingPaths records collision searches so held-pointer movement can prove that sub-cell drift avoids replanning.
type countingPaths struct {
	calls int
	path  []gameworld.Point
}

// FindPath records each search and returns a direct route whose endpoint remains easy to assert.
func (paths *countingPaths) FindPath(request gameworld.PathRequest) ([]gameworld.Point, error) {
	paths.calls++
	if paths.path != nil {
		return paths.path, nil
	}
	return []gameworld.Point{request.Start, request.Goal}, nil
}

func TestMovementSourceRetainsFinalCellWhenHeldPointerDrifts(t *testing.T) {
	fixture := newMovementFixture(t)
	fixture.addPosition(0, 0)
	controller := &MovementController{}
	if err := controller.SetMoveTarget(5, 0); err != nil {
		t.Fatal(err)
	}
	source := fixture.source(controller)
	paths := &countingPaths{path: []gameworld.Point{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}, {5, 0}}}
	source.navigation = paths
	if first := commandPayload(t, source, 1); first.Target == nil || first.Target.X != 4 || first.Target.Y != 0 {
		t.Fatalf("straight route anchor = %#v", first.Target)
	}
	if err := controller.SetMoveTarget(5.25, 0.25); err != nil {
		t.Fatal(err)
	}
	if held := commandPayload(t, source, 2); held.Target == nil || held.Target.X != 4 || held.Target.Y != 0 {
		t.Fatalf("held-pointer drift redirected the long segment: %#v", held.Target)
	}
	positions, _ := akara.GetDynamicStore(fixture.engine.World(), "d2legacy.world.position")
	if _, err := positions.Set(fixture.entity, map[string]any{"x": 4.0, "y": 0.0}); err != nil {
		t.Fatal(err)
	}
	if final := commandPayload(t, source, 3); final.Target == nil || final.Target.X != 5.25 || final.Target.Y != 0.25 {
		t.Fatalf("last segment did not follow sub-cell drift: %#v", final.Target)
	}
	if paths.calls != 1 {
		t.Fatalf("sub-cell drift performed %d searches, want 1", paths.calls)
	}
}

// movementFixture owns the minimum ECS and input state needed by command-source tests.
type movementFixture struct {
	t      *testing.T
	engine *gameecs.Engine
	input  *inputstate.Store
	entity akara.Entity
}

// newMovementFixture installs an active player-control component so command generation is admitted.
func newMovementFixture(t *testing.T) movementFixture {
	t.Helper()

	engine := gameecs.New()

	controls, err := akara.RegisterSchema(engine.World(), akara.Schema{
		Name: "d2legacy.world.player_control",
		Fields: []akara.Field{
			{Name: "player", Kind: akara.FieldString},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	entity := engine.World().MustCreateEntity()
	if _, err := controls.Set(entity, map[string]any{"player": "alpha"}); err != nil {
		t.Fatal(err)
	}

	return movementFixture{
		t:      t,
		engine: engine,
		input:  &inputstate.Store{},
		entity: entity,
	}
}

// addPosition gives pathfinding a stable world origin for the controlled entity.
func (fixture movementFixture) addPosition(x, y float64) {
	fixture.t.Helper()

	positions, err := akara.RegisterSchema(fixture.engine.World(), akara.Schema{
		Name: "d2legacy.world.position",
		Fields: []akara.Field{
			{Name: "x", Kind: akara.FieldFloat64},
			{Name: "y", Kind: akara.FieldFloat64},
		},
	})
	if err != nil {
		fixture.t.Fatal(err)
	}

	if _, err := positions.Set(fixture.entity, map[string]any{"x": x, "y": y}); err != nil {
		fixture.t.Fatal(err)
	}
}

// source constructs the production command source with the fixture's stable player and focus identifiers.
func (fixture movementFixture) source(controller *MovementController) *MovementSource {
	fixture.t.Helper()

	source, err := NewMovementSource(fixture.engine, fixture.input, "alpha", "game_world", controller)
	if err != nil {
		fixture.t.Fatal(err)
	}

	return source
}

// publishInput replaces the entire input snapshot, matching how production frame ownership is sampled.
func (fixture movementFixture) publishInput(
	ownerID string,
	gameplay bool,
	actions map[string]inputstate.ActionState,
) {
	fixture.input.Publish(inputstate.Frame{
		Actions:  actions,
		Owner:    inputstate.FocusOwner{Domain: inputstate.FocusScene, ID: ownerID},
		Gameplay: gameplay,
	})
}

// commandPayload decodes the single command expected from an active fixture and keeps assertions focused on policy.
func commandPayload(t *testing.T, source *MovementSource, tick uint64) MovePayload {
	t.Helper()

	commands := source.Commands(tick)
	if len(commands) != 1 {
		t.Fatalf("commands = %d, want one", len(commands))
	}

	payload, err := decodeMove(commands[0].Payload)
	if err != nil {
		t.Fatal(err)
	}

	return payload
}

// TestMovementSourceHonorsGameplayInputRouting protects overlays from leaking keys unless they opt into passthrough.
func TestMovementSourceHonorsGameplayInputRouting(t *testing.T) {
	fixture := newMovementFixture(t)
	source := fixture.source(nil)
	right := map[string]inputstate.ActionState{"right": {Down: true}}

	fixture.publishInput("inventory", false, right)

	payload := commandPayload(t, source, 1)
	if payload.X != 0 || payload.Y != 0 {
		t.Fatalf("overlay focus leaked movement: %#v", payload)
	}

	fixture.publishInput("inventory", true, right)

	payload = commandPayload(t, source, 2)
	if payload.X != 1 || payload.Y != 0 {
		t.Fatalf("passthrough overlay blocked movement: %#v", payload)
	}

	fixture.publishInput("game_world", false, right)

	payload = commandPayload(t, source, 3)
	if payload.X != 1 || payload.Y != 0 {
		t.Fatalf("gameplay focus did not admit movement: %#v", payload)
	}
}

// TestMovementSourceAdmitsQueuedAndKeyboardRunIntent verifies mailbox state precedes the next keyboard toggle.
func TestMovementSourceAdmitsQueuedAndKeyboardRunIntent(t *testing.T) {
	fixture := newMovementFixture(t)
	controller := &MovementController{}
	controller.SetRunning(true)
	source := fixture.source(controller)

	fixture.publishInput("game_world", false, nil)

	if payload := commandPayload(t, source, 1); !payload.Running {
		t.Fatalf("queued run payload = %#v", payload)
	}

	fixture.publishInput("game_world", false, map[string]inputstate.ActionState{
		"toggle_run": {Pressed: true},
	})

	if payload := commandPayload(t, source, 2); payload.Running {
		t.Fatalf("keyboard toggle payload = %#v", payload)
	}
}

// TestMovementSourceEmitsPointerWorldTarget protects exact coordinate and JSON field compatibility with Lua.
func TestMovementSourceEmitsPointerWorldTarget(t *testing.T) {
	fixture := newMovementFixture(t)
	fixture.publishInput("game_world", false, nil)

	controller := &MovementController{}
	if err := controller.SetMoveTarget(12.5, 44.25); err != nil {
		t.Fatal(err)
	}

	source := fixture.source(controller)

	payload := commandPayload(t, source, 1)
	if payload.Target == nil || payload.Target.X != 12.5 || payload.Target.Y != 44.25 {
		t.Fatalf("pointer target payload = %#v", payload)
	}

	encoded := source.Commands(2)[0].Payload
	if !bytes.Contains(encoded, []byte(`"target":{"x":12.5,"y":44.25,"stop_radius":0}`)) {
		t.Fatalf("pointer target wire schema = %s", encoded)
	}
}

// TestMovementSourceKeepsAcceptedRouteWhenReplacementIsBlocked prevents an unreachable click from stopping motion.
func TestMovementSourceKeepsAcceptedRouteWhenReplacementIsBlocked(t *testing.T) {
	fixture := newMovementFixture(t)
	fixture.addPosition(10, 10)
	fixture.publishInput("game_world", false, nil)

	controller := &MovementController{}
	if err := controller.SetMoveTarget(20, 11); err != nil {
		t.Fatal(err)
	}

	source := fixture.source(controller)
	source.navigation = scriptedPaths{}

	first := commandPayload(t, source, 1)
	if first.Target == nil || first.Target.X != 11 || first.Target.Y != 10 {
		t.Fatalf("accepted route waypoint = %#v", first.Target)
	}

	if err := controller.SetMoveTarget(99, 10); err != nil {
		t.Fatal(err)
	}

	second := commandPayload(t, source, 2)
	if second.Target == nil || second.Target.X != 11 || second.Target.Y != 10 {
		t.Fatalf("blocked replacement discarded accepted route: %#v", second.Target)
	}

	retained := controller.moveTarget()
	if retained == nil || retained.X != 20 || retained.Y != 11 {
		t.Fatalf("controller retained target = %#v, want original target", retained)
	}
	controller.ClearMoveTarget()
	if stopped := commandPayload(t, source, 3); stopped.Target != nil || source.path != nil {
		t.Fatalf("superseded route resumed: payload=%#v path=%v", stopped, source.path)
	}
	if err := controller.SetMoveTarget(20, 11); err != nil {
		t.Fatal(err)
	}
	if next := commandPayload(t, source, 4); next.Target == nil {
		t.Fatalf("new ground route did not start: %#v", next)
	}
	fixture.publishInput("game_world", false, map[string]inputstate.ActionState{"right": {Down: true}})
	if keyboard := commandPayload(t, source, 5); keyboard.Target != nil || keyboard.X != 1 || controller.HasMoveTarget() {
		t.Fatalf("keyboard movement did not replace the ground route: %#v", keyboard)
	}
	fixture.publishInput("game_world", false, nil)
	if idle := commandPayload(t, source, 6); idle.Target != nil || idle.X != 0 {
		t.Fatalf("released keyboard input resumed the ground route: %#v", idle)
	}
}

// TestMovementSourceDoesNotReplanHeldPointerInsideCollisionCell protects camera-follow pointer stability.
func TestMovementSourceDoesNotReplanHeldPointerInsideCollisionCell(t *testing.T) {
	fixture := newMovementFixture(t)
	fixture.addPosition(10, 10)

	controller := &MovementController{}
	if err := controller.SetMoveTarget(20.1, 10.1); err != nil {
		t.Fatal(err)
	}

	source := fixture.source(controller)
	paths := &countingPaths{}
	source.navigation = paths

	first := commandPayload(t, source, 1)
	if first.Target == nil {
		t.Fatalf("first waypoint = %#v", first.Target)
	}

	if err := controller.SetMoveTarget(20.2, 10.2); err != nil {
		t.Fatal(err)
	}

	second := commandPayload(t, source, 2)
	if second.Target == nil || second.Target.X != 20.2 || second.Target.Y != 10.2 {
		t.Fatalf("updated sub-cell waypoint = %#v", second.Target)
	}

	if paths.calls != 1 {
		t.Fatalf("path searches = %d, want one within a collision cell", paths.calls)
	}
}

// TestMovementSourceInvalidatesWorldRelativeRouteWhenNavigationChanges prevents coordinates crossing map instances.
func TestMovementSourceInvalidatesWorldRelativeRouteWhenNavigationChanges(t *testing.T) {
	fixture := newMovementFixture(t)

	controller := &MovementController{}
	if err := controller.SetMoveTarget(20, 10); err != nil {
		t.Fatal(err)
	}

	source := fixture.source(controller)
	source.path = []gameworld.Point{{X: 10, Y: 10}, {X: 11, Y: 10}}
	source.pathTarget = &MoveTarget{X: 20, Y: 10}

	source.SetNavigation(&gameworld.Map{})

	if controller.HasMoveTarget() || source.path != nil || source.pathTarget != nil {
		t.Fatalf(
			"world replacement retained route state: controller=%t path=%v target=%v",
			controller.HasMoveTarget(),
			source.path,
			source.pathTarget,
		)
	}
}

// TestMovementSourceKeepsRouteWhenSameNavigationIsReinstalled protects correction-driven map rebinding.
func TestMovementSourceKeepsRouteWhenSameNavigationIsReinstalled(t *testing.T) {
	fixture := newMovementFixture(t)
	controller := &MovementController{}
	source := fixture.source(controller)
	navigation := &gameworld.Map{}
	source.SetNavigation(navigation)

	if err := controller.SetMoveTarget(20, 10); err != nil {
		t.Fatal(err)
	}

	source.path = []gameworld.Point{{X: 10, Y: 10}, {X: 11, Y: 10}, {X: 20, Y: 10}}
	source.pathTarget = &MoveTarget{X: 20, Y: 10}

	// Connected correction projection reactivates the authoritative level even when the level did not change.
	source.SetNavigation(navigation)

	if !controller.HasMoveTarget() || len(source.path) != 3 || source.pathTarget == nil {
		t.Fatalf(
			"same-map correction discarded route: target=%t path=%v route_target=%#v",
			controller.HasMoveTarget(),
			source.path,
			source.pathTarget,
		)
	}
}

func TestStraightRouteMatchesDirectMovementCadence(t *testing.T) {
	for _, rates := range []ClassRates{{Walk: 6, Run: 9}, {Walk: 40, Run: 60}} {
		for _, running := range []bool{false, true} {
			state := lua.NewState()
			defer state.Close()
			bridge := movementRulesBridge{catalog: Catalog{byClass: map[string]ClassRates{"amazon": rates}}}
			velocity := func(position gameworld.Point, payload MovePayload) gameworld.Point {
				target := state.NewTable()
				target.RawSetString("x", lua.LNumber(payload.Target.X))
				target.RawSetString("y", lua.LNumber(payload.Target.Y))
				intent := state.NewTable()
				intent.RawSetString("target", target)
				intent.RawSetString("running", lua.LBool(payload.Running))
				err := state.CallByParam(lua.P{Fn: state.NewFunction(bridge.velocity), NRet: 3, Protect: true},
					lua.LNumber(position.X), lua.LNumber(position.Y), lua.LString("Amazon"), intent,
					lua.LNumber(0), lua.LNumber(0), lua.LNumber(.04))
				if err != nil {
					t.Fatal(err)
				}
				result := gameworld.Point{X: float64(state.Get(-3).(lua.LNumber)), Y: float64(state.Get(-2).(lua.LNumber))}
				state.Pop(3)
				return result
			}
			fixture := newMovementFixture(t)
			fixture.addPosition(0, 0)
			controller := &MovementController{}
			controller.SetRunning(running)
			if err := controller.SetMoveTarget(100, 0); err != nil {
				t.Fatal(err)
			}
			source := fixture.source(controller)
			path := make([]gameworld.Point, 101)
			for i := range path {
				path[i].X = float64(i)
			}
			source.navigation = &countingPaths{path: path}
			positions, _ := akara.GetDynamicStore(fixture.engine.World(), "d2legacy.world.position")
			planned, direct := gameworld.Point{}, gameworld.Point{}
			for tick := uint64(1); tick <= 6; tick++ {
				payload := commandPayload(t, source, tick)
				plannedVelocity := velocity(planned, payload)
				directVelocity := velocity(direct, MovePayload{Running: running, Target: &MoveTarget{X: 100}})
				planned = gameworld.IntegrateVelocity(nil, planned, plannedVelocity, gameworld.Point{X: 100, Y: 100}, 0, .04)
				direct = gameworld.IntegrateVelocity(nil, direct, directVelocity, gameworld.Point{X: 100, Y: 100}, 0, .04)
				if math.Abs(planned.X-direct.X) > 1e-8 || planned.Y != direct.Y {
					t.Fatalf("rates=%+v running=%t tick=%d: planned=%v direct=%v", rates, running, tick, planned, direct)
				}
				if _, err := positions.Set(fixture.entity, map[string]any{"x": planned.X, "y": planned.Y}); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}
