package interaction

import (
	"testing"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func testAuthority(t *testing.T) *Authority {
	t.Helper()
	authority, err := NewAuthority(Target{ID: "act1-akara", NPC: "Akara", Vendor: "Akara", Categories: []string{"misc", "weap", "misc"}, Services: []string{"identify"}})
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
	session, err := gamesession.New(gameecs.New(), gamesession.Config{CheckpointInterval: 1})
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
	other, err := NewAuthority(Target{ID: "act1-charsi", NPC: "Charsi", Vendor: "Charsi"})
	if err != nil {
		t.Fatal(err)
	}
	if err := other.RestoreState(encoded); err == nil {
		t.Fatal("mismatched target configuration was accepted")
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

func TestPlayerCannotOpenInteractionForAnotherOwner(t *testing.T) {
	command, _ := Command(OpenCommand, Payload{Owner: "bob", Target: "act1-akara"}, "alice", 1, 1, simulation.AuthorityPlayer)
	if _, err := decode(command, true); err == nil {
		t.Fatal("cross-owner interaction was accepted")
	}
}
