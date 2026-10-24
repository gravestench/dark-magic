package clientapp

import (
	"context"
	"encoding/json"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/app/networktrust"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

type coldPopulationPlan struct {
	Schema string               `json:"schema"`
	Zones  []coldPopulationZone `json:"zones"`
}

type coldPopulationZone struct {
	Level   int `json:"level_id"`
	Created int `json:"created"`
	Rooms   []struct {
		ID        string `json:"id"`
		Activated bool   `json:"activated"`
		Active    bool   `json:"active"`
	} `json:"rooms"`
}

// populationPlan decodes the production plan so activation assertions inspect authoritative state.
func populationPlan(t *testing.T, fixture *realD2LegacyFixture) coldPopulationPlan {
	t.Helper()

	value, ok := fixture.app.authoritativeState.Read("d2legacy.population.plan")
	if !ok {
		t.Fatal("population plan is absent")
	}

	var plan coldPopulationPlan
	if err := json.Unmarshal(value.Data, &plan); err != nil {
		t.Fatal(err)
	}

	if plan.Schema != "d2legacy.population.plan/v6" {
		t.Fatalf("population schema = %q", plan.Schema)
	}

	return plan
}

// newColdPopulationFixture starts the owned-data fixture with the population-specific initial rules.
func newColdPopulationFixture(t *testing.T) *realD2LegacyFixture {
	t.Helper()
	return newRealD2LegacyFixture(t, coldPopulationConfig())
}

// Keep the initial player and topology recipe identical in offline, recovery, and listen cases.
func coldPopulationConfig() realD2LegacyFixtureConfig {
	return realD2LegacyFixtureConfig{
		startScene: "game_world", fixtureCharacters: 1, fixtureWorldLevel: 2,
		fixtureWorldSpawn: "cold_plains_seam", applySceneDefaults: true,
	}
}

// Reach Cold Plains only through the production movement seam and preserve both levels on return.
func TestLiveColdPlainsOrdinaryPopulation(t *testing.T) {
	f := newColdPopulationFixture(t)
	if err := f.app.network.StartSelected(); err != nil {
		t.Fatal(err)
	}

	f.advanceGame(t, 10)

	before := populationPlan(t, f)
	if len(before.Zones) != 4 {
		t.Fatalf("initial population zones=%d, want 4", len(before.Zones))
	}

	if before.Zones[0].Created == 0 ||
		before.Zones[1].Created != 0 || before.Zones[2].Created != 0 || before.Zones[3].Created != 0 {
		t.Fatalf("initial population created counts=%v",
			[]int{before.Zones[0].Created, before.Zones[1].Created, before.Zones[2].Created, before.Zones[3].Created})
	}

	bloodIDs := populationResidentIDs(t, f, 2, false)
	seam := f.app.transitionSeams[1]
	moveAcrossColdPlainsSeam(t, f, 2, 3, seam.First, seam.Second)
	f.advanceGame(t, 2) // Allow the next population phase after the committed arrival.

	coldIDs := populationResidentIDs(t, f, 3, true)
	if len(coldIDs) == 0 {
		t.Fatal("Cold Plains created no ordinary monsters")
	}

	checkpoint, err := f.app.offlineSession.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	assertColdPopulationVisible(t, coldPopulationWorldView(t, "local-player", checkpoint))
	assertPopulationInactive(t, f, bloodIDs, true)
	assertPopulationActiveLevel(t, populationPlan(t, f), 3)
	coldPlan := populationPlan(t, f).Zones[1]
	moveAcrossColdPlainsSeam(t, f, 3, 2, seam.Second, seam.First)
	f.advanceGame(t, 2)
	assertPopulationInactive(t, f, coldIDs, true)

	if got := populationPlan(t, f).Zones[0].Created; got != before.Zones[0].Created {
		t.Fatalf("Blood Moor created=%d, want %d", got, before.Zones[0].Created)
	}

	if got := populationResidentIDs(t, f, 2, false); !reflect.DeepEqual(got, bloodIDs) {
		t.Fatal("Blood Moor resident IDs changed")
	}

	assertPopulationInactive(t, f, bloodIDs, false)
	returned := populationPlan(t, f)
	assertPopulationActiveLevel(t, returned, 2)

	if !reflect.DeepEqual(returned.Zones[0], before.Zones[0]) {
		t.Fatal("return changed the Blood Moor room plan")
	}

	moveAcrossColdPlainsSeam(t, f, 2, 3, seam.First, seam.Second)
	f.advanceGame(t, 2)

	if got := populationResidentIDs(t, f, 3, false); !reflect.DeepEqual(got, coldIDs) {
		t.Fatal("Cold Plains return recreated residents")
	}

	if got := populationPlan(t, f).Zones[1].Created; got != coldPlan.Created {
		t.Fatalf("Cold Plains created=%d, want %d", got, coldPlan.Created)
	}
}

// Inspect real ECS residents, including their full circular terrain footprints at first activation.
func populationResidentIDs(
	t *testing.T, f *realD2LegacyFixture, level int, checkPlacement bool,
) map[string]akara.Entity {
	t.Helper()
	monsters := requireRealStore(t, f, "d2legacy.monster.identity", "missing monsters")
	residents := requireRealStore(t, f, "d2legacy.world.room_resident", "missing residents")
	positions := requireRealStore(t, f, "d2legacy.world.position", "missing positions")
	colliders := requireRealStore(t, f, "d2legacy.world.collider", "missing colliders")
	result := map[string]akara.Entity{}

	type footprint struct{ x, y, r float64 }

	var points []footprint

	for _, entity := range monsters.Entities() {
		resident, ok := residents.Get(entity)
		if !ok {
			continue
		}

		residentLevel, _ := resident.Get("level_id")
		if residentLevel != int64(level) {
			continue
		}

		identity, _ := monsters.Get(entity)
		id, _ := identity.Get("spawn_id")
		definition, _ := identity.Get("definition_id")

		if _, duplicate := result[id.(string)]; duplicate {
			t.Fatalf("duplicate spawn ID %v", id)
		}

		result[id.(string)] = entity

		if level == 3 && definition != "brute1" && definition != "corruptrogue1" && definition != "fallenshaman1" {
			t.Fatalf("unexpected ordinary family %v", definition)
		}

		if !checkPlacement {
			continue
		}

		x, y := dynamicPosition(positions, entity)

		collider, ok := colliders.Get(entity)
		if !ok {
			t.Fatal("ordinary monster has no collider")
		}

		radius, _ := collider.Get("radius")
		r := radius.(float64)

		point := gameworld.Point{X: x, Y: y}
		if !f.app.gameWorlds[level].WalkableStep(point, point, r) {
			t.Fatalf("blocked footprint %s at %.2f,%.2f radius %.2f", id, x, y, r)
		}

		for _, p := range points {
			if math.Hypot(x-p.x, y-p.y) < r+p.r {
				t.Fatal("ordinary monster footprints overlap")
			}
		}

		points = append(points, footprint{x, y, r})
	}

	return result
}

// The occupied level activates rooms; the other wilderness has no active room window.
func assertPopulationActiveLevel(t *testing.T, plan coldPopulationPlan, level int) {
	t.Helper()

	for _, zone := range plan.Zones {
		count := 0

		for _, room := range zone.Rooms {
			if room.Active {
				count++
			}
		}

		if (count > 0) != (zone.Level == level) {
			t.Fatalf("level %d active rooms=%d, occupied level=%d", zone.Level, count, level)
		}
	}
}

// assertPopulationInactive checks that retained resident entities carry the expected activation tag.
func assertPopulationInactive(t *testing.T, f *realD2LegacyFixture, ids map[string]akara.Entity, want bool) {
	t.Helper()

	inactive := requireRealStore(t, f, "d2legacy.world.inactive", "missing inactive store")
	for id, entity := range ids {
		_, got := inactive.Get(entity)
		if got != want {
			t.Fatalf("resident %s inactive=%v, want %v", id, got, want)
		}
	}
}

// assertColdPopulationVisible requires the public view to expose ordinary Cold Plains monsters.
func assertColdPopulationVisible(t *testing.T, view playeradapter.WorldView) {
	t.Helper()

	count := 0

	for _, entity := range view.Entities {
		if entity.Kind == "hostile" && entity.LevelID == 3 {
			count++

			if entity.SpawnID == "" || entity.DefinitionID == "" {
				t.Fatal("visible monster lost identity")
			}
		}
	}

	if count == 0 {
		t.Fatal("Cold Plains ordinary population is absent from public projection")
	}
}

// coldPopulationWorldView projects a checkpoint through the public player adapter for visibility assertions.
func coldPopulationWorldView(t *testing.T, player string, checkpoint simulation.Checkpoint) playeradapter.WorldView {
	t.Helper()

	payload, err := playeradapter.ProjectWorldView(player, checkpoint)
	if err != nil {
		t.Fatal(err)
	}

	var view playeradapter.WorldView
	if err := json.Unmarshal(payload, &view); err != nil {
		t.Fatal(err)
	}

	return view
}

// Both authorities continue from a Cold Plains checkpoint through 3→2→3 with identical future inputs.
func TestColdPlainsPopulationRecoveryMatchesContinuation(t *testing.T) {
	source := newColdPopulationFixture(t)
	if err := source.app.network.StartSelected(); err != nil {
		t.Fatal(err)
	}

	source.advanceGame(t, 10)
	seam := source.app.transitionSeams[1]
	moveAcrossColdPlainsSeam(t, source, 2, 3, seam.First, seam.Second)
	source.advanceGame(t, 2)

	if len(populationResidentIDs(t, source, 3, false)) == 0 {
		t.Fatal("checkpoint lacks Cold Plains population")
	}

	bundle, err := source.app.entryWorldRecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	config := coldPopulationConfig()
	config.skipPopulationBootstrap = true
	config.recovery = &bundle

	target := newRealD2LegacyFixture(t, config)
	if err := target.app.network.StartSelected(); err != nil {
		t.Fatal(err)
	}

	assertColdRecoveryRejectsOldMovement(t, target)

	for _, crossing := range []struct {
		to   int
		x, y float64
	}{
		{2, seam.Second.X, seam.Second.Y},
		{3, seam.First.X, seam.First.Y},
	} {
		if err := source.app.playerControl.SetMoveTargetWithRadius(crossing.x, crossing.y, 0.5); err != nil {
			t.Fatal(err)
		}

		crossed := false

		for range 1500 {
			advancePopulationRecoveryTick(t, source, target)

			if hud := recoveryPlayerHUD(t, source); int(hud.Location.LevelID) == crossing.to {
				crossed = true
				break
			}
		}

		if !crossed {
			t.Fatalf("recovery continuation never crossed to level %d", crossing.to)
		}

		advancePopulationRecoveryTick(t, source, target)
	}

	continuation, err := source.app.offlineSession.Replay()
	if err != nil {
		t.Fatal(err)
	}

	finalTick := continuation.Checkpoints[len(continuation.Checkpoints)-1].Tick

	var captured uint64

	for _, checkpoint := range continuation.Checkpoints {
		if checkpoint.Tick > bundle.Session.State.Tick {
			captured++
		}
	}

	if want := finalTick - bundle.Session.State.Tick; captured != want {
		t.Fatalf("continuation records %d checkpoints for %d completed ticks", captured, want)
	}

	a, err := source.app.offlineSession.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	b, err := target.app.offlineSession.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(a, b) {
		t.Fatalf("recovered checkpoint differs: tick/checksum %d/%s vs %d/%s", a.Tick, a.Checksum, b.Tick, b.Checksum)
	}

	ae, err := playeradapter.ProjectEventView("local-player", a)
	if err != nil {
		t.Fatal(err)
	}

	be, err := playeradapter.ProjectEventView("local-player", b)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(ae, be) {
		t.Fatal("recovered semantic events differ")
	}

	ar, err := source.app.offlineSession.Replay()
	if err != nil {
		t.Fatal(err)
	}

	br, err := target.app.offlineSession.Replay()
	if err != nil {
		t.Fatal(err)
	}

	var commands []simulation.Command

	for _, c := range ar.Commands {
		if c.Tick > bundle.Session.State.Tick {
			commands = append(commands, c)
		}
	}

	if !reflect.DeepEqual(commands, br.Commands) {
		t.Fatal("recovered continuation commands differ")
	}

	if ar.Checkpoints[len(ar.Checkpoints)-1].Checksum != br.Checkpoints[len(br.Checkpoints)-1].Checksum {
		t.Fatal("replay final checksums differ")
	}

	verifyProductionPopulationReplay(t, source, ar)
	verifyProductionPopulationReplay(t, source, br)
}

// Compare the complete state and visible semantic events before a transient divergence can leave the event window.
func advancePopulationRecoveryTick(t *testing.T, source, target *realD2LegacyFixture) {
	t.Helper()
	source.advanceGame(t, 1)

	recorded, err := source.app.offlineSession.Replay()
	if err != nil {
		t.Fatal(err)
	}

	want := recorded.Checkpoints[len(recorded.Checkpoints)-1]
	for _, command := range recorded.Commands {
		if command.Tick == want.Tick {
			if _, err := target.app.offlineSession.SubmitNetwork(command); err != nil {
				t.Fatal(err)
			}
		}
	}

	if _, err := target.app.offlineSession.Advance(time.Second / 25); err != nil {
		t.Fatal(err)
	}

	target.app.syncActiveWorldFromPlayer()

	got, err := target.app.offlineSession.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("recovered continuation differs at tick %d: checksum %s, want %s", want.Tick, got.Checksum, want.Checksum)
	}

	wantEvents, err := playeradapter.ProjectEventView("local-player", want)
	if err != nil {
		t.Fatal(err)
	}

	gotEvents, err := playeradapter.ProjectEventView("local-player", got)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(gotEvents, wantEvents) {
		t.Fatalf("recovered semantic events differ at continuation tick %d", want.Tick)
	}
}

// Recompose each replay baseline independently and verify every completed checkpoint.
func verifyProductionPopulationReplay(t *testing.T, source *realD2LegacyFixture, replay simulation.Replay) {
	t.Helper()

	bundle, err := source.app.entryWorldRecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	checksum, err := simulation.CompositeChecksum(replay.Initial, replay.InitialParticipants)
	if err != nil {
		t.Fatal(err)
	}

	initial := simulation.Checkpoint{
		Tick: replay.Initial.Tick, Checksum: checksum,
		Snapshot: &replay.Initial, Participants: replay.InitialParticipants,
	}

	bundle.Session, err = gamesession.NewRecoveryCheckpoint(initial, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	config := coldPopulationConfig()

	config.skipPopulationBootstrap = true
	if replay.Initial.Tick != 0 {
		config.recovery = &bundle
	} else {
		config.topologySnapshot = bundle.Topology
	}

	f := newRealD2LegacyFixture(t, config)
	index := 0

	tick := replay.Initial.Tick
	for _, want := range replay.Checkpoints {
		for tick < want.Tick {
			tick++
			for index < len(replay.Commands) && replay.Commands[index].Tick == tick {
				if _, err := f.app.offlineSession.SubmitNetwork(replay.Commands[index]); err != nil {
					t.Fatal(err)
				}

				index++
			}

			if _, err := f.app.offlineSession.Advance(time.Duration(replay.StepNanos)); err != nil {
				t.Fatal(err)
			}
		}

		got, err := f.app.offlineSession.CanonicalCheckpoint()
		if err != nil {
			t.Fatal(err)
		}

		if got.Checksum != want.Checksum {
			t.Fatalf("production replay diverged at %d: %s; participants_equal=%v",
				tick, gameecs.FirstDifference(*want.Snapshot, *got.Snapshot),
				reflect.DeepEqual(got.Participants, want.Participants))
		}
	}

	if index != len(replay.Commands) {
		t.Fatal("replay did not consume every admitted command")
	}

	actual, err := f.app.offlineSession.Replay()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(actual.Commands, replay.Commands) {
		t.Fatal("offline replay changed admitted command identities or ticks")
	}
}

// Reliable content owns identity while datagrams update transforms; compare their shared population facts.
type coldPopulationPublicMonster struct {
	SpawnID, DefinitionID string
	Position              playeradapter.HUDPosition
	LevelID               int64
}

// publicColdPopulation extracts stable public monster facts for authority and client comparisons.
func publicColdPopulation(view playeradapter.WorldView) map[string]coldPopulationPublicMonster {
	result := make(map[string]coldPopulationPublicMonster)

	for _, entity := range view.Entities {
		if entity.Kind == "hostile" {
			result[entity.ID] = coldPopulationPublicMonster{
				entity.SpawnID, entity.DefinitionID, entity.Position, entity.LevelID,
			}
		}
	}

	return result
}

// Compare the QUIC client's ordinary population with its own acknowledged host tick.
func TestConnectedColdPlainsPopulationMatchesAuthority(t *testing.T) {
	f := newColdPopulationFixture(t)
	app := f.app
	app.ctx = t.Context()

	app.components = host.NewManager()
	if err := app.components.Register(host.ManagedDefinition{
		ID:  "d2legacy.boot",
		New: func(context.Context) (host.Component, error) { return headlessClientBoot{}, nil },
	}); err != nil {
		t.Fatal(err)
	}

	var err error

	app.networkTrust, err = networktrust.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = app.network.Close() })

	if err := app.network.Host(); err != nil {
		t.Fatal(err)
	}

	if err := app.network.StartSelected(); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(30 * time.Second)
	for !app.network.Connected() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if !app.network.Connected() {
		t.Fatal("listen connection timed out")
	}

	f.advanceGame(t, 10)

	seam := app.transitionSeams[1]
	moveConnectedAcrossColdPlains(t, f, 3, seam.First, seam.Second)

	for range 100 {
		time.Sleep(40 * time.Millisecond)
		f.advanceGame(t, 1)

		presentation := app.network.client.PresentationSnapshot()
		hud, view := presentation.HUD, presentation.World
		// Events retain the reliable correction tick. A newer datagram cannot acknowledge monster lifecycle.
		if presentation.Events.Tick != hud.Tick || view.Tick != hud.Tick {
			continue
		}

		replay, err := app.network.host.Session.Replay()
		if err != nil {
			t.Fatal(err)
		}

		for _, checkpoint := range replay.Checkpoints {
			if checkpoint.Tick != hud.Tick || view.Tick != hud.Tick {
				continue
			}

			want := coldPopulationWorldView(t, app.presentationPlayerID(), checkpoint)
			if !reflect.DeepEqual(publicColdPopulation(view), publicColdPopulation(want)) {
				t.Fatalf("client public population differs from authority at acknowledged tick %d", hud.Tick)
			}

			assertColdPopulationVisible(t, want)

			initialIDs := checkpointPopulationIDs(t, checkpoint)
			moveConnectedAcrossColdPlains(t, f, 2, seam.Second, seam.First)
			moveConnectedAcrossColdPlains(t, f, 3, seam.First, seam.Second)

			final, err := app.network.host.Session.CanonicalCheckpoint()
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(checkpointPopulationIDs(t, final), initialIDs) {
				t.Fatal("connected return recreated wilderness residents")
			}

			// Offline replay can outlast the live event tail; finish connected frames before verifying it.
			verifyProductionPopulationReplay(t, f, replay)

			return
		}
	}

	t.Fatal("no acknowledged host checkpoint reached")
}

// Restore an immutable host snapshot locally so a running server never exposes mutable ECS stores to the test.
func checkpointPopulationIDs(t *testing.T, checkpoint simulation.Checkpoint) map[string]akara.Entity {
	t.Helper()

	engine, err := gameecs.RestoreSnapshot(*checkpoint.Snapshot)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = engine.Close() }()

	identities, ok := akara.GetDynamicStore(engine.World(), "d2legacy.monster.identity")
	if !ok {
		t.Fatal("host snapshot has no monster identities")
	}

	result := make(map[string]akara.Entity)

	for _, entity := range identities.Entities() {
		identity, _ := identities.Get(entity)
		id, _ := identity.Get("spawn_id")
		result[id.(string)] = entity
	}

	return result
}
