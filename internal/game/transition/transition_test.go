package transition_test

import (
	"testing"
	"time"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameplayer "github.com/gravestench/dark-magic/internal/game/player"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/game/transition"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/persistence"
)

func TestSourceTransitionsPlayerAcrossVerifiedSeamWithoutBounce(t *testing.T) {
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := gameplayer.Register(session); err != nil {
		t.Fatal(err)
	}
	seam := gameworld.Seam{
		Town:       gameworld.SeamEndpoint{LevelID: 1, X: 10, Y: 10, ArrivalX: 5, ArrivalY: 5, Width: 100, Height: 80, Direction: "east"},
		Wilderness: gameworld.SeamEndpoint{LevelID: 2, X: 0, Y: 40, ArrivalX: 6, ArrivalY: 40, Width: 400, Height: 400, Direction: "west"},
	}
	authority, err := transition.NewAuthority(seam)
	if err != nil {
		t.Fatal(err)
	}
	if err := transition.Register(session, authority); err != nil {
		t.Fatal(err)
	}
	entry := gameplayer.EntryFromCharacter(persistence.Character{ID: "hero", Name: "Hero", Class: "Amazon", Level: 1}, "player", 10, 10, 100, 80)
	command, _ := gameplayer.Command(entry, "system", 1, 1, simulation.AuthoritySystem)
	if err := session.Submit(command); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	source, err := transition.NewSource(engine, "player", authority)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := session.AdvanceWithSource(time.Second/25, source.Commands); err != nil {
		t.Fatal(err)
	}
	locations, _ := akara.GetDynamicStore(engine.World(), "d2.world.location")
	entity := locations.Entities()[0]
	location, _ := locations.Get(entity)
	level, _ := location.Get("level_id")
	if level != int64(2) {
		t.Fatalf("level = %v", level)
	}
	positions, _ := akara.GetDynamicStore(engine.World(), "d2.world.position")
	position, _ := positions.Get(entity)
	x, _ := position.Get("x")
	y, _ := position.Get("y")
	if x != float64(6) || y != float64(40) {
		t.Fatalf("arrival = %v,%v", x, y)
	}
	if commands := source.Commands(3); len(commands) != 0 {
		t.Fatalf("arrival immediately bounced: %#v", commands)
	}
}
