package transition_test

import (
	"encoding/json"
	"testing"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameplayer "github.com/gravestench/dark-magic/internal/game/player"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/game/transition"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/persistence"
)

func TestSourceTransitionsPlayerAcrossVerifiedSeamWithoutBounce(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	seam := gameworld.Seam{
		Town:       gameworld.SeamEndpoint{LevelID: 1, X: 10, Y: 10, ArrivalX: 5, ArrivalY: 5, Width: 100, Height: 80, Direction: "east"},
		Wilderness: gameworld.SeamEndpoint{LevelID: 2, X: 0, Y: 40, ArrivalX: 6, ArrivalY: 40, Width: 400, Height: 400, Direction: "west"},
	}
	authority, err := transition.NewAuthority(seam)
	if err != nil {
		t.Fatal(err)
	}
	entry := gameplayer.EntryFromCharacter(persistence.Character{ID: "hero", Name: "Hero", Class: "Amazon", Level: 1}, "player", 10, 10, 100, 80)
	command, _ := gameplayer.Command(entry, "system", 1, 1, simulation.AuthoritySystem)
	if err := gameplayer.ApplyEntryCommand(engine, command); err != nil {
		t.Fatal(err)
	}
	source, err := transition.NewSource(engine, "player", authority)
	if err != nil {
		t.Fatal(err)
	}
	commands := source.Commands(3)
	if len(commands) != 1 {
		t.Fatalf("commands = %#v", commands)
	}
	var payload transition.Payload
	if err := json.Unmarshal(commands[0].Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.SourceLevel != 1 || payload.DestinationLevel != 2 || payload.ArrivalX != 6 || payload.ArrivalY != 40 || payload.WorldWidth != 400 {
		t.Fatalf("transition recipe = %#v", payload)
	}
}
