package monster

import (
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/game/targeting"
)

func TestScheduledAIAcquiresChasesAndRequestsAttack(t *testing.T) {
	stats, graphics, level := ordinaryFixture()
	definition, err := JoinDefinition(stats, graphics, &level, Normal)
	if err != nil {
		t.Fatal(err)
	}
	engine := gameecs.New()
	defer engine.Close()
	if err := RegisterAI(engine, nil); err != nil {
		t.Fatal(err)
	}
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := Register(session); err != nil {
		t.Fatal(err)
	}
	spawn, _ := NewSpawn("fallen:1", definition, 7, 2, 2, 1, 2)
	command, _ := Command(spawn, "population", 1, 1, simulation.AuthoritySystem)
	addPlayerTarget(t, engine, "player:hero", 8, 2)
	if err := session.Submit(command); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	brains, _ := akara.GetDynamicStore(engine.World(), AIComponent)
	brain, _ := brains.Get(brains.Entities()[0])
	state, _ := brain.Get("state")
	target, _ := brain.Get("target_id")
	next, _ := brain.Get("next_think_tick")
	if state != AIChase || target != "player:hero" || next != int64(4) {
		t.Fatalf("first think = state %v target %v next %v", state, target, next)
	}
	velocities, _ := akara.GetDynamicStore(engine.World(), "d2.world.velocity")
	velocity, _ := velocities.Get(brains.Entities()[0])
	vx, _ := velocity.Get("x")
	if vx != float64(6) {
		t.Fatalf("chase velocity = %v", vx)
	}
	// Centers remain two subtiles apart, but the player's one-subtile footprint
	// is already inside the monster's one-subtile melee reach.
	movePlayerTarget(t, engine, "player:hero", 4, 2)
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	if _, found := akara.GetDynamicStore(engine.World(), BasicAttackComponent); !found {
		t.Fatal("attack request schema missing")
	}
	for engine.Tick() < 4 {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
	}
	state, _ = brain.Get("state")
	requests, _ := akara.GetDynamicStore(engine.World(), BasicAttackComponent)
	request, present := requests.Get(brains.Entities()[0])
	if state != AIAttack || !present {
		t.Fatalf("attack state = %v, request present = %v", state, present)
	}
	requestTarget, _ := request.Get("target_id")
	requestTick, _ := request.Get("request_tick")
	if requestTarget != "player:hero" || requestTick != int64(4) {
		t.Fatalf("request = target %v tick %v", requestTarget, requestTick)
	}
	snapshot, err := engine.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Components) == 0 {
		t.Fatal("AI state was not checkpointable")
	}
}

func addPlayerTarget(t *testing.T, engine *gameecs.Engine, id string, x, y float64) {
	t.Helper()
	positions, _ := akara.GetDynamicStore(engine.World(), "d2.world.position")
	locations, _ := akara.GetDynamicStore(engine.World(), "d2.world.location")
	selectables, _ := akara.GetDynamicStore(engine.World(), targeting.Component)
	entity := engine.World().MustCreateEntity()
	if _, err := positions.Set(entity, map[string]any{"x": x, "y": y}); err != nil {
		t.Fatal(err)
	}
	if _, err := locations.Set(entity, map[string]any{"act": int64(1), "level_id": int64(2)}); err != nil {
		t.Fatal(err)
	}
	if _, err := selectables.Set(entity, map[string]any{"id": id, "kind": targeting.KindPlayer, "label": "Hero", "owner": "player", "radius": 1.0, "priority": int64(10)}); err != nil {
		t.Fatal(err)
	}
}

func movePlayerTarget(t *testing.T, engine *gameecs.Engine, id string, x, y float64) {
	t.Helper()
	positions, _ := akara.GetDynamicStore(engine.World(), "d2.world.position")
	selectables, _ := akara.GetDynamicStore(engine.World(), targeting.Component)
	for _, entity := range selectables.Entities() {
		selectable, _ := selectables.Get(entity)
		value, _ := selectable.Get("id")
		if value == id {
			position, _ := positions.Get(entity)
			_ = position.Set("x", x)
			_ = position.Set("y", y)
			return
		}
	}
	t.Fatalf("target %q not found", id)
}
