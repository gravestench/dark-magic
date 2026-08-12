package monster

import (
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/game/targeting"
)

func TestSpawnMaterializesDeterministicHostile(t *testing.T) {
	stats, graphics, level := ordinaryFixture()
	definition, err := JoinDefinition(stats, graphics, &level, Normal)
	if err != nil {
		t.Fatal(err)
	}
	definition.DeathSound = "fallen_death"
	spawn, err := NewSpawn("blood-moor:1", definition, 42, 12, 13, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := Register(session); err != nil {
		t.Fatal(err)
	}
	command, err := Command(spawn, "population", 1, 1, simulation.AuthoritySystem)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Submit(command); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	identities, found := akara.GetDynamicStore(engine.World(), "d2legacy.monster.identity")
	if !found || identities.Len() != 1 {
		t.Fatalf("monster identities = %v, found=%v", identities, found)
	}
	statsStore, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.stats")
	component, _ := statsStore.Get(identities.Entities()[0])
	health, _ := component.Get("health")
	if health != rollLife(definition.LifeMin, definition.LifeMax, 42).Raw() {
		t.Fatalf("health = %v", health)
	}
	appearance, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.appearance")
	visual, _ := appearance.Get(identities.Entities()[0])
	components, _ := visual.Get("components")
	deathSound, _ := visual.Get("death_sound")
	if components != "HD=LIT,TR=MED" || deathSound != "fallen_death" {
		t.Fatalf("appearance recipe = %q, death sound = %q", components, deathSound)
	}
	hit, found := targeting.New(engine).HitAt(12, 13)
	if !found || hit.Kind != targeting.KindHostile || hit.ID != "monster:blood-moor:1" {
		t.Fatalf("target = %#v, found=%v", hit, found)
	}
	if err := session.Submit(simulation.Command{Tick: 2, Player: "client", Authority: simulation.AuthorityPlayer, Sequence: 2, Kind: SpawnCommand, Payload: command.Payload}); err == nil {
		t.Fatal("player authority admitted privileged monster spawn")
	}
}
