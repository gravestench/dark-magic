package interaction

import (
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

func testAuthority(t *testing.T) *Authority {
	t.Helper()
	authority, err := NewAuthority(Target{ID: "act1-akara", NPC: "Akara", Vendor: "Akara", Categories: []string{"misc", "weap", "misc"}, Services: []string{"identify"}, X: 10, Y: 10, Radius: 5})
	if err != nil {
		t.Fatal(err)
	}
	if err := authority.RegisterOwner("alice", ""); err != nil {
		t.Fatal(err)
	}
	return authority
}

func TestInteractionCommandsOpenCloseAndReplayState(t *testing.T) {
	authority := testAuthority(t)
	engine := gameecs.New()
	materializeControlledPosition(t, engine, "alice", 13, 14)
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := RegisterCommands(session, authority); err != nil {
		t.Fatal(err)
	}
	open, _ := Command(OpenCommand, Payload{Target: "act1-akara"}, "alice", 1, 1, simulation.AuthorityPlayer)
	if err := session.Submit(open); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	context, err := authority.Snapshot("alice")
	if err != nil || context.Vendor != "Akara" || len(context.Categories) != 2 {
		t.Fatalf("context = %#v, %v", context, err)
	}
	if !authority.CanTrade("alice", "akara") || !authority.CanService("alice", "identify") {
		t.Fatal("active context did not authorize declared actions")
	}
	replay, err := session.Replay()
	if err != nil || len(replay.InitialParticipants) != 1 || len(replay.Checkpoints) != 1 {
		t.Fatalf("replay = %#v, %v", replay, err)
	}
	closeCommand, _ := Command(CloseCommand, Payload{}, "alice", 2, 2, simulation.AuthorityPlayer)
	if err := session.Submit(closeCommand); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	if authority.CanTrade("alice", "Akara") {
		t.Fatal("closed interaction still authorized trade")
	}
}

func TestSessionStateRestoresContextAndRejectsOtherConfiguration(t *testing.T) {
	authority := testAuthority(t)
	if err := authority.open("alice", "act1-akara"); err != nil {
		t.Fatal(err)
	}
	encoded, err := authority.SnapshotState()
	if err != nil {
		t.Fatal(err)
	}
	restored := testAuthority(t)
	if err := restored.RestoreState(encoded); err != nil {
		t.Fatal(err)
	}
	if !restored.CanTrade("alice", "Akara") {
		t.Fatal("restored context lost vendor")
	}
	other, err := NewAuthority(Target{ID: "act1-charsi", NPC: "Charsi", Vendor: "Charsi", Radius: 5})
	if err != nil {
		t.Fatal(err)
	}
	if err := other.RestoreState(encoded); err == nil {
		t.Fatal("mismatched target configuration was accepted")
	}
}

func TestOpenRejectsMissingOrOutOfRangeControlledPlayer(t *testing.T) {
	authority := testAuthority(t)
	if err := authority.openSpatial(gameecs.New(), "alice", "act1-akara"); err == nil {
		t.Fatal("interaction without a controlled player was accepted")
	}
	engine := gameecs.New()
	materializeControlledPosition(t, engine, "alice", 100, 100)
	if err := authority.openSpatial(engine, "alice", "act1-akara"); err == nil {
		t.Fatal("out-of-range interaction was accepted")
	}
	context, err := authority.Snapshot("alice")
	if err != nil || context.TargetID != "" {
		t.Fatalf("rejected interaction mutated context: %#v, %v", context, err)
	}
}

func TestActiveCommercePermissionExpiresWhenPlayerWalksAway(t *testing.T) {
	authority := testAuthority(t)
	engine := gameecs.New()
	materializeControlledPosition(t, engine, "alice", 10, 10)
	if err := authority.openSpatial(engine, "alice", "act1-akara"); err != nil {
		t.Fatal(err)
	}
	if !authority.CanTradeAt(engine, "alice", "Akara") || !authority.CanServiceAt(engine, "alice", "identify") {
		t.Fatal("in-range active interaction did not admit declared actions")
	}
	positions, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.position")
	position, _ := positions.Get(positions.Entities()[0])
	if err := position.Set("x", float64(100)); err != nil {
		t.Fatal(err)
	}
	if authority.CanTradeAt(engine, "alice", "Akara") || authority.CanServiceAt(engine, "alice", "identify") {
		t.Fatal("walking out of range retained commerce or service permission")
	}
}

func materializeControlledPosition(t *testing.T, engine *gameecs.Engine, player string, x, y float64) {
	t.Helper()
	controls, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2legacy.world.player_control", Version: 1, Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}})
	if err != nil {
		t.Fatal(err)
	}
	positions, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2legacy.world.position", Version: 1, Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}})
	if err != nil {
		t.Fatal(err)
	}
	entity, err := engine.World().CreateEntity()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := controls.Set(entity, map[string]any{"player": player}); err != nil {
		t.Fatal(err)
	}
	if _, err := positions.Set(entity, map[string]any{"x": x, "y": y}); err != nil {
		t.Fatal(err)
	}
}

func TestControllerPreservesInteractionIntentOrder(t *testing.T) {
	controller := &Controller{}
	if err := controller.Open("act1-akara"); err != nil {
		t.Fatal(err)
	}
	controller.Close()
	source, err := NewSource(controller, "alice")
	if err != nil {
		t.Fatal(err)
	}
	commands := source.Commands(7)
	if len(commands) != 2 || commands[0].Kind != OpenCommand || commands[1].Kind != CloseCommand || commands[0].Sequence != 1 || commands[1].Sequence != 2 {
		t.Fatalf("commands = %#v", commands)
	}
}

func TestCoordinateInteractionIsSelectedAndCheckedByAuthority(t *testing.T) {
	authority := testAuthority(t)
	engine := gameecs.New()
	materializeControlledPosition(t, engine, "alice", 10, 10)
	if err := authority.openSpatialAt(engine, "alice", 10.5, 10); err != nil {
		t.Fatal(err)
	}
	context, err := authority.Snapshot("alice")
	if err != nil || context.TargetID != "act1-akara" {
		t.Fatalf("context = %#v, %v", context, err)
	}
	if err := authority.openSpatialAt(engine, "alice", 40, 40); err == nil {
		t.Fatal("empty coordinate selected a target")
	}
}

func TestControllerPreservesCoordinateIntent(t *testing.T) {
	controller := &Controller{}
	if err := controller.OpenAt(12.5, 9.25); err != nil {
		t.Fatal(err)
	}
	source, err := NewSource(controller, "alice")
	if err != nil {
		t.Fatal(err)
	}
	commands := source.Commands(3)
	if len(commands) != 1 {
		t.Fatalf("commands = %#v", commands)
	}
	payload, err := decode(commands[0], true)
	if err != nil || !payload.At || payload.X != 12.5 || payload.Y != 9.25 {
		t.Fatalf("payload = %#v, %v", payload, err)
	}
}

func TestPlayerCannotOpenInteractionForAnotherOwner(t *testing.T) {
	command, _ := Command(OpenCommand, Payload{Owner: "bob", Target: "act1-akara"}, "alice", 1, 1, simulation.AuthorityPlayer)
	if _, err := decode(command, true); err == nil {
		t.Fatal("cross-owner interaction was accepted")
	}
}

func TestConfigureWorldExcludesTargetsFromInactiveZone(t *testing.T) {
	const targetID = "ds1-object:1:42:0"
	authority, err := NewAuthority(Target{ID: targetID, NPC: "Town NPC", X: 10, Y: 10, Radius: 5})
	if err != nil {
		t.Fatal(err)
	}
	town := &gameworld.Map{Objects: []gameworld.Object{{Type: 1, ID: 42, X: 10, Y: 10, Resolved: true, Description: "Town NPC"}}}
	authority.ConfigureWorld(town)
	if _, err := authority.targetAt(10, 10); err != nil {
		t.Fatalf("town target missing: %v", err)
	}
	authority.ConfigureWorld(&gameworld.Map{})
	if _, err := authority.targetAt(10, 10); err == nil {
		t.Fatal("target from inactive town remained pointer-selectable")
	}
}
