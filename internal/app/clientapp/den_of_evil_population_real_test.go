package clientapp

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	entryworld "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/entryworld"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

type denResident struct {
	entity         akara.Entity
	x, y, radius   float64
	health         int64
	dead, inactive bool
}

// Read a canonical snapshot so the same assertions cover local and hosted authority.
func denResidents(t *testing.T, checkpoint simulation.Checkpoint) map[string]denResident {
	t.Helper()

	engine, err := gameecs.RestoreSnapshot(*checkpoint.Snapshot)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = engine.Close() }()

	world := engine.World()
	identities, _ := akara.GetDynamicStore(world, "d2legacy.monster.identity")
	positions, _ := akara.GetDynamicStore(world, "d2legacy.world.position")
	colliders, _ := akara.GetDynamicStore(world, "d2legacy.world.collider")
	stats, _ := akara.GetDynamicStore(world, "d2legacy.monster.stats")
	deaths, _ := akara.GetDynamicStore(world, "d2legacy.monster.death")

	inactive, _ := akara.GetDynamicStore(world, "d2legacy.world.inactive")
	if identities == nil || positions == nil || colliders == nil || stats == nil || deaths == nil || inactive == nil {
		t.Fatal("checkpoint lacks Den resident component stores")
	}

	result := map[string]denResident{}

	for _, entity := range identities.Entities() {
		identity, _ := identities.Get(entity)
		idValue, _ := identity.Get("spawn_id")

		id, ok := idValue.(string)
		if !ok || !strings.HasPrefix(id, "level:8:room:") {
			continue
		}

		position, hasPosition := positions.Get(entity)
		collider, hasCollider := colliders.Get(entity)
		values, hasStats := stats.Get(entity)
		_, isInactive := inactive.Get(entity)
		death, hasDeath := deaths.Get(entity)
		dead := false

		if hasDeath {
			active, _ := death.Get("active")
			dead = active == false
		}

		if !hasPosition || !hasStats || (!hasCollider && !isInactive && !dead) {
			t.Fatalf("Den resident %s has incomplete state: position=%t collider=%t stats=%t inactive=%t dead=%t",
				id, hasPosition, hasCollider, hasStats, isInactive, dead)
		}

		x, _ := position.Get("x")
		y, _ := position.Get("y")
		health, _ := values.Get("health")
		resident := denResident{entity: entity, x: x.(float64), y: y.(float64),
			health: health.(int64), dead: dead, inactive: isInactive}

		if hasCollider {
			radius, _ := collider.Get("radius")
			resident.radius = radius.(float64)
		}

		if _, duplicate := result[id]; duplicate {
			t.Fatalf("duplicate Den resident ID %s", id)
		}

		result[id] = resident
	}

	return result
}

// The ordinary clicked cave journey must activate a real level-8 resident in
// authority and project that same identity to the player inside the Den.
func TestDenOfEvilPopulationEntryReal(t *testing.T) {
	app, input, _, _ := newDenPointerClient(t, false, realD2LegacyFixtureConfig{fixtureWorldLevel: 2})
	denFrames(t, app, 10)
	pair := *app.caveWarp
	denMoveTo(t, app, pair.First.X, pair.First.Y, 0.1)
	denClick(t, app, input, pair.First)
	denAssertArrival(t, app, pair.First, 8)
	denFrames(t, app, 2)
	checkpoint := denCheckpoint(t, app)
	ids := checkpointPopulationIDs(t, checkpoint)
	view := coldPopulationWorldView(t, "local-player", checkpoint)
	visible := 0

	for _, entity := range view.Entities {
		if entity.Kind != "hostile" || entity.LevelID != 8 {
			continue
		}

		if !strings.HasPrefix(entity.SpawnID, "level:8:room:") {
			t.Fatalf("Den hostile has unqualified ID %q", entity.SpawnID)
		}

		if _, ok := ids[entity.SpawnID]; !ok {
			t.Fatalf("public Den hostile %q is missing from authority", entity.SpawnID)
		}

		visible++
	}

	if visible == 0 {
		t.Fatal("clicked Den entry has no visible authoritative hostile")
	}
}

// Ordinary pointer journeys must retain the same cave residents across travel and reconnection.
func TestDenOfEvilPopulationJourneyReal(t *testing.T) {
	for difficulty := range 3 {
		for _, hosted := range []bool{false, true} {
			t.Run(fmt.Sprintf("difficulty_%d/connected_%t", difficulty, hosted), func(t *testing.T) {
				app, input, _, reconnect := newDenPointerClient(t, hosted,
					realD2LegacyFixtureConfig{fixtureWorldLevel: 2, gameDifficulty: difficulty})
				denFrames(t, app, 10)
				pair := *app.caveWarp
				denMoveTo(t, app, pair.First.X, pair.First.Y, 0.1)
				denClick(t, app, input, pair.First)
				denAssertArrival(t, app, pair.First, 8)
				denFrames(t, app, 2)
				before := denResidents(t, denCheckpoint(t, app))
				assertDenResidentPlacement(t, app, pair, before)
				assertDenProjection(t, app, before, true)

				if hosted {
					reconnect()
					denFrames(t, app, 8)
					assertDenResidentState(t, before, denResidents(t, denCheckpoint(t, app)))
				}

				denMoveTo(t, app, pair.Second.X, pair.Second.Y, 0.1)
				denClick(t, app, input, pair.Second)
				denAssertArrival(t, app, pair.Second, 2)
				assertDenProjection(t, app, before, false)
				denMoveTo(t, app, pair.First.X, pair.First.Y, 0.1)
				denClick(t, app, input, pair.First)

				if location := denAuthorityHUD(t, app).Location; location.LevelID != 8 ||
					location.TransitionSequence != 3 {
					t.Fatalf("return to Den relocated to %+v, want level 8 sequence 3", location)
				}

				assertDenResidentState(t, before, denResidents(t, denCheckpoint(t, app)))
				assertDenProjection(t, app, before, true)
			})
		}
	}
}

// assertDenProjection reads the actual connected view when hosted, and the same
// player adapter over the local checkpoint when offline.
func assertDenProjection(t *testing.T, app *application, residents map[string]denResident, inside bool) {
	t.Helper()

	var view playeradapter.WorldView
	if app.network.Connected() {
		_, view = app.network.client.View()
	} else {
		view = coldPopulationWorldView(t, app.presentationPlayerID(), denCheckpoint(t, app))
	}

	visible := 0

	for _, entity := range view.Entities {
		if entity.Kind != "hostile" || entity.LevelID != 8 {
			continue
		}

		if _, ok := residents[entity.SpawnID]; !ok {
			t.Fatalf("player view has unknown Den hostile %s", entity.SpawnID)
		}

		visible++
	}

	if (visible > 0) != inside {
		t.Fatalf("player view has %d Den hostiles, inside=%t", visible, inside)
	}
}

// assertDenResidentPlacement checks the full active collider and reachability,
// while retaining inactive room residents for identity comparisons.
func assertDenResidentPlacement(t *testing.T, app *application, pair entryworld.CaveWarpPair,
	residents map[string]denResident,
) {
	t.Helper()

	if len(residents) == 0 {
		t.Fatal("entered Den with no ordinary residents")
	}

	active := 0

	for id, resident := range residents {
		if resident.health <= 0 || resident.dead {
			t.Fatalf("Den resident %s began with health %d and dead=%t", id, resident.health, resident.dead)
		}

		if resident.inactive {
			continue
		}

		active++

		point := gameworld.Point{X: resident.x, Y: resident.y}
		if !app.gameWorlds[8].WalkableStep(point, point, resident.radius) {
			t.Fatalf("Den resident %s has blocked footprint", id)
		}

		if _, err := app.gameWorlds[8].FindPath(gameworld.PathRequest{
			Start: gameworld.Point{X: pair.First.ArrivalX, Y: pair.First.ArrivalY},
			Goal:  point, Radius: 1,
		}); err != nil {
			t.Fatalf("Den resident %s is unreachable from cave arrival: %v", id, err)
		}

		if math.Hypot(resident.x-pair.First.ArrivalX, resident.y-pair.First.ArrivalY) < 1+resident.radius ||
			math.Hypot(resident.x-pair.Second.X, resident.y-pair.Second.Y) < 3.5+resident.radius {
			t.Fatalf("Den resident %s overlaps a cave endpoint", id)
		}
	}

	if active == 0 {
		t.Fatalf("entered Den without an active ordinary resident: %+v", residents)
	}
}

// assertDenResidentState reports the exact identity or combat field that changed after a journey.
func assertDenResidentState(t *testing.T, left, right map[string]denResident) {
	t.Helper()

	if len(left) != len(right) {
		t.Fatalf("Den resident count changed from %d to %d", len(left), len(right))
	}

	for id, resident := range left {
		other, ok := right[id]
		if !ok {
			t.Fatalf("Den resident %s was lost", id)
		}

		if other.entity != resident.entity || other.dead != resident.dead || other.health != resident.health {
			t.Fatalf("Den resident %s changed: before=%+v after=%+v", id, resident, other)
		}
	}
}

// Separate fresh sessions with identical rules must publish the same cave plan and IDs.
func TestDenOfEvilPopulationDeterministicReal(t *testing.T) {
	var wantPlan []byte

	var wantIDs map[string]denResident

	for attempt := range 2 {
		app, input, _, _ := newDenPointerClient(t, false, realD2LegacyFixtureConfig{fixtureWorldLevel: 2})
		denFrames(t, app, 10)
		pair := *app.caveWarp
		denMoveTo(t, app, pair.First.X, pair.First.Y, 0.1)
		denClick(t, app, input, pair.First)
		denFrames(t, app, 2)

		value, ok := app.authoritativeState.Read("d2legacy.population.plan")
		if !ok {
			t.Fatal("Den population plan is absent")
		}

		ids := denResidents(t, denCheckpoint(t, app))
		if attempt == 0 {
			wantPlan, wantIDs = append([]byte(nil), value.Data...), ids
		} else {
			if !reflect.DeepEqual(value.Data, wantPlan) {
				t.Fatal("identical sessions generated different Den plans")
			}

			assertDenResidentState(t, wantIDs, ids)
		}
	}
}

// The Den resident uses the ordinary production attack and corpse lifecycle.
func TestDenOfEvilPopulationCombatPersistenceReal(t *testing.T) {
	app, input, _, _ := newDenPointerClient(t, false, realD2LegacyFixtureConfig{fixtureWorldLevel: 2})
	denFrames(t, app, 10)
	pair := *app.caveWarp
	denMoveTo(t, app, pair.First.X, pair.First.Y, 0.1)
	denClick(t, app, input, pair.First)
	denAssertArrival(t, app, pair.First, 8)
	denFrames(t, app, 2)
	residents := denResidents(t, denCheckpoint(t, app))

	var targetID string

	nearest := math.Inf(1)

	for id, resident := range residents {
		if resident.inactive || resident.dead {
			continue
		}

		distance := math.Hypot(resident.x-pair.First.ArrivalX, resident.y-pair.First.ArrivalY)
		if distance < nearest {
			targetID, nearest = id, distance
		}
	}

	if targetID == "" {
		t.Fatal("Den has no combat target")
	}

	initial := residents[targetID]
	damaged := false

	for range 80 {
		current, ok := denResidents(t, denCheckpoint(t, app))[targetID]
		if !ok {
			t.Fatalf("Den combat target %s lost its persistent identity", targetID)
		}

		if current.health < initial.health {
			damaged = true
		}

		if current.dead {
			break
		}

		if err := app.commandIntents.Submit("player.use_skill", map[string]any{
			"side": "left", "target_x": current.x, "target_y": current.y,
			"target_id": "monster:" + targetID,
		}); err != nil {
			t.Fatal(err)
		}

		denFrames(t, app, 20)
	}

	dead := denResidents(t, denCheckpoint(t, app))[targetID]
	if !damaged || !dead.dead {
		t.Fatalf("production attacks left Den resident %s at health %d, dead=%t, damaged=%t",
			targetID, dead.health, dead.dead, damaged)
	}

	denMoveTo(t, app, pair.Second.X, pair.Second.Y, 0.1)
	denClick(t, app, input, pair.Second)
	denAssertArrival(t, app, pair.Second, 2)
	denMoveTo(t, app, pair.First.X, pair.First.Y, 0.1)
	denClick(t, app, input, pair.First)
	denAssertPopulationReturn(t, app, pair.First)

	returned := denResidents(t, denCheckpoint(t, app))[targetID]
	if returned.entity != dead.entity || returned.health != dead.health || !returned.dead {
		t.Fatalf("Den corpse %s changed after round trip: before=%+v after=%+v", targetID, dead, returned)
	}

	if err := app.scenes.Close(t.Context()); err != nil {
		t.Fatal(err)
	}

	bundle, err := app.entryWorldRecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	restored, _ := newPointerWorldClient(t, false,
		realD2LegacyFixtureConfig{fixtureWorldLevel: 2, recovery: &bundle},
		func(restored *application) error {
			restored.options.StartScene = ""
			err := restored.loadScriptComponents()
			restored.options.StartScene = "game_world"

			if err != nil {
				return err
			}

			if err := restored.network.StartSelected(); err != nil {
				return err
			}

			return restored.openInitialScene()
		})
	denFrames(t, restored, 10)
	denAssertPopulationReturn(t, restored, pair.First)

	continued := denResidents(t, denCheckpoint(t, restored))[targetID]
	if continued.entity != dead.entity || continued.health != dead.health || !continued.dead {
		t.Fatalf("restored Den corpse %s changed: before=%+v after=%+v", targetID, dead, continued)
	}
}

// denAssertPopulationReturn checks the third relocation used by a full cave round trip.
func denAssertPopulationReturn(t *testing.T, app *application, entrance entryworld.CaveWarpEndpoint) {
	t.Helper()
	location := denAuthorityHUD(t, app).Location

	position := denAuthorityHUD(t, app).Position
	if location.LevelID != 8 || location.TransitionSequence != 3 || app.activeWorldLevel != 8 ||
		math.Hypot(position.X-entrance.ArrivalX, position.Y-entrance.ArrivalY) > .01 {
		t.Fatalf("return to Den has unexpected location %+v position %+v", location, position)
	}
}
