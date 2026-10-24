package clientapp

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/app/clientsession"
	"github.com/gravestench/dark-magic/internal/app/networkclock"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// TestConnectedReconciliationFeedsAuthenticatedRouteProgress proves the HUD reaches the command source
// while the replica retains smoothed lag. Removing that wiring strands waypoint 12.
func TestConnectedReconciliationFeedsAuthenticatedRouteProgress(t *testing.T) {
	engine := newRemotePresentationTestEngine(t)
	hero := engine.World().MustCreateEntity()

	for _, component := range []struct {
		name   string
		fields map[string]any
	}{
		{name: "d2legacy.world.player_control", fields: map[string]any{"player": "alpha"}},
		{name: "d2legacy.world.position", fields: map[string]any{"x": 10.0, "y": 10.0}},
		{name: "d2legacy.world.location", fields: map[string]any{"level_id": int64(1)}},
	} {
		store, found := akara.GetDynamicStore(engine.World(), component.name)
		if !found {
			t.Fatalf("replica component %s is missing", component.name)
		}

		if _, err := store.Set(hero, component.fields); err != nil {
			t.Fatal(err)
		}
	}

	controller := &movement.MovementController{}

	source, err := movement.NewMovementSource(engine, &inputstate.Store{}, "alpha", "game_world", controller)
	if err != nil {
		t.Fatal(err)
	}

	openMap, err := gameworld.NewOpenMap(20, 20)
	if err != nil {
		t.Fatal(err)
	}

	source.SetNavigation(openMap)

	if err := controller.SetMoveTarget(13, 10); err != nil {
		t.Fatal(err)
	}
	// The accepted route is planned while the player is at x=10, before presentation lags near waypoint 12.
	initial := source.Commands(1)
	if len(initial) != 1 {
		t.Fatalf("initial movement commands = %d, want one", len(initial))
	}

	var first movement.MovePayload
	if err := json.Unmarshal(initial[0].Payload, &first); err != nil {
		t.Fatal(err)
	}

	if first.Target == nil || first.Target.X != 12 {
		t.Fatalf("initial route waypoint = %#v, want x=12", first.Target)
	}

	positions, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.position")

	position, _ := positions.Get(hero)
	if err := position.Set("x", 11.667); err != nil {
		t.Fatal(err)
	}

	hud := playeradapter.HUD{
		Version:  playeradapter.HUDVersion,
		Tick:     1,
		Player:   playeradapter.HUDIdentity{PlayerID: "alpha"},
		Location: playeradapter.HUDLocation{Act: 1, LevelID: 1},
		Position: playeradapter.HUDPosition{X: 11.875, Y: 10},
	}
	client := newClientWorld()
	client.hasHUD = true
	client.lastHUD = hud
	client.local = localPresentation{
		initialized: true,
		playerID:    "alpha",
		error:       playeradapter.HUDPosition{X: -0.208},
	}

	app := &application{
		clientSimulation: engine,
		movementSource:   source,
		gameWorlds:       map[int]*gameworld.Map{1: openMap},
	}
	if err := client.presentLocalPrediction(app, &clientsession.Session{}, hud,
		networkclock.Timeline{}, 0, false); err != nil {
		t.Fatal(err)
	}

	x, _ := position.Get("x")
	if math.Abs(x.(float64)-11.667) > 1e-6 {
		t.Fatalf("replica position = %.3f, want lagging display 11.667", x)
	}

	commands := source.Commands(2)
	if len(commands) != 1 {
		t.Fatalf("movement commands = %d, want one", len(commands))
	}

	var payload movement.MovePayload
	if err := json.Unmarshal(commands[0].Payload, &payload); err != nil {
		t.Fatal(err)
	}

	if payload.Target == nil || payload.Target.X != 13 || payload.Target.Y != 10 {
		t.Fatalf("next connected waypoint = %#v, want (13,10)", payload.Target)
	}
}
