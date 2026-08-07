package player

import (
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/persistence"
)

func TestEntryCommandMaterializesAuthoritativePlayerAtomically(t *testing.T) {
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := Register(session); err != nil {
		t.Fatal(err)
	}
	character := persistence.Character{ID: "amazon-hero", Name: "Hero", Class: "Amazon", Level: 3, Expansion: true, Stats: &persistence.Stats{Experience: 100, Health: 25, MaxHealth: 30, Mana: 12, MaxMana: 15}}
	entry := EntryFromCharacter(character, "player-1", 5, 7, 100, 80)
	command, err := Command(entry, "server", 1, 1, simulation.AuthoritySystem)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Submit(command); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	identity, found := akara.GetDynamicStore(engine.World(), "dm.player.identity")
	if !found || len(identity.Entities()) != 1 {
		t.Fatalf("identity store = %v, %v", identity, found)
	}
	entity := identity.Entities()[0]
	component, _ := identity.Get(entity)
	if player, _ := component.Get("player"); player != "player-1" {
		t.Fatalf("player = %v", player)
	}
	position, _ := akara.GetDynamicStore(engine.World(), "dm.world.position")
	transform, _ := position.Get(entity)
	if x, _ := transform.Get("x"); x != float64(5) {
		t.Fatalf("x = %v", x)
	}
	if audit := session.Audit(); len(audit) != 1 || audit[0].Kind != EnterCommand {
		t.Fatalf("audit = %#v", audit)
	}
}
