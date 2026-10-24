package clientapp

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/app/networktrust"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
	d2mapgen "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/mapgen"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
)

// A fourth map must participate in ordinary movement and survive a checkpoint before the return trip.
func TestStonyFieldOfflinePopulationRecoveryRoundTrip(t *testing.T) {
	config := realD2LegacyFixtureConfig{startScene: "game_world", fixtureCharacters: 1,
		fixtureWorldLevel: 3, applySceneDefaults: true}

	source := newRealD2LegacyFixture(t, config)
	if err := source.app.network.StartSelected(); err != nil {
		t.Fatal(err)
	}

	source.advanceGame(t, 10)
	// Production seed 1 has eligible Stony Field rooms; pin that fixture evidence.
	if seed := source.app.entryWorldTopology.Zones[4].Request().Seed; seed != 1 {
		t.Fatalf("population fixture seed = %d, want 1", seed)
	}

	seam := source.app.transitionSeams[2]
	moveAcrossColdPlainsSeam(t, source, 3, 4, seam.First, seam.Second)
	assertStonyFieldIdle(t, source, 4, 1, seam.Second.ArrivalX, seam.Second.ArrivalY)

	ids := populationResidentIDs(t, source, 4, false)
	if len(ids) == 0 {
		t.Fatal("Stony Field created no ordinary residents")
	}
	assertStonyPopulationProjection(t, source)

	bundle, err := source.app.entryWorldRecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	config.skipPopulationBootstrap, config.recovery = true, &bundle

	restored := newRealD2LegacyFixture(t, config)
	if err := restored.app.network.StartSelected(); err != nil {
		t.Fatal(err)
	}

	if restored.app.entryWorldTopology.Version != 3 || len(restored.app.gameWorlds) != 4 ||
		restored.app.entryWorldChecksum != bundle.TopologyChecksum || restored.app.activeWorldLevel != 4 {
		t.Fatal("recovery lost the four-map topology or active Stony Field map")
	}

	got, err := restored.app.offlineSession.RecoveryCheckpoint()
	if err != nil || got.Checksum != bundle.Session.Checksum {
		t.Fatalf("recovery checksum differs: %v", err)
	}

	for _, crossing := range []struct {
		to    int
		point gametransition.SeamEndpoint
	}{
		{3, seam.Second}, {4, seam.First},
	} {
		if err := source.app.playerControl.SetMoveTargetWithRadius(
			crossing.point.X, crossing.point.Y, 0.5,
		); err != nil {
			t.Fatal(err)
		}

		crossed := false

		for range 1500 {
			advancePopulationRecoveryTick(t, source, restored)
			if int(recoveryPlayerHUD(t, source).Location.LevelID) == crossing.to {
				crossed = true
				break
			}
		}

		if !crossed {
			t.Fatalf("continuation never crossed to %d", crossing.to)
		}

		advancePopulationRecoveryTick(t, source, restored)
		if crossing.to == 3 {
			assertPopulationInactive(t, source, ids, true)
		}
	}

	if !reflect.DeepEqual(ids, populationResidentIDs(t, source, 4, false)) ||
		!reflect.DeepEqual(ids, populationResidentIDs(t, restored, 4, false)) {
		t.Fatal("leave/return recreated Stony Field residents")
	}

	assertStonyPopulationProjection(t, source)
	assertStonyPopulationProjection(t, restored)

	a, err := source.app.offlineSession.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	b, err := restored.app.offlineSession.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	if a.Tick != b.Tick || a.Checksum != b.Checksum {
		t.Fatal("recovered reverse crossing diverged from uninterrupted authority")
	}
}

// Waiting beyond arrival catches a second trigger or a retained old-map movement target.
func assertStonyFieldIdle(t *testing.T, f *realD2LegacyFixture, level int, sequence int64, x, y float64) {
	t.Helper()

	for range 12 {
		f.advanceGame(t, 1)

		hud := recoveryPlayerHUD(t, f)
		if int(hud.Location.LevelID) != level || hud.Location.TransitionSequence != sequence ||
			math.Hypot(hud.Position.X-x, hud.Position.Y-y) > 0.01 {
			t.Fatalf("arrival retriggered or drifted: level %d position %#v", hud.Location.LevelID, hud.Position)
		}
	}
}

// Reconnect and an admitted late level-3 input must preserve the level-4 authority before ordinary return movement.
func TestStonyFieldConnectedPopulationReconnectRoundTrip(t *testing.T) {
	f := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{startScene: "game_world", fixtureCharacters: 1,
		fixtureWorldLevel: 3, applySceneDefaults: true})
	app := f.app
	app.ctx = t.Context()

	app.components = host.NewManager()
	if err := app.components.Register(host.ManagedDefinition{ID: "d2legacy.boot",
		New: func(context.Context) (host.Component, error) { return headlessClientBoot{}, nil }}); err != nil {
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
	// Own the watch lifetime so reconnect starts after the old credential's stream ends, as transport recovery does.
	app.network.mu.Lock()
	generation, ctx, _, _ := app.network.beginSelectedNetworkLocked()
	app.network.mu.Unlock()

	runtime := &listenHostRuntime{}
	if err := app.network.prepareListenHost(ctx, runtime); err != nil {
		runtime.close()
		t.Fatal(err)
	}

	app.network.startListenServices(ctx, generation, runtime)

	if err := app.network.connectListenPlayer(ctx, runtime); err != nil {
		runtime.close()
		t.Fatal(err)
	}

	if !app.network.commitListenHost(generation, runtime) {
		runtime.close()
		t.Fatal("listen host commit failed")
	}
	go app.network.send(ctx, runtime.client)

	watchContext, stopWatch := context.WithCancel(ctx)

	watchDone := make(chan struct{})
	go func() { defer close(watchDone); app.network.watch(watchContext, runtime.client) }()

	t.Cleanup(stopWatch)
	f.advanceGame(t, 10)

	source, err := app.modSource("d2legacy")
	if err != nil {
		t.Fatal(err)
	}

	identity, err := app.network.runtimeIdentity(source, app.options.Packages)
	if err != nil {
		t.Fatal(err)
	}

	want, err := identity.Digest()
	if err != nil {
		t.Fatal(err)
	}

	got, err := app.network.host.Authority.Identity.Digest()
	if err != nil || got != want {
		t.Fatalf("client/server topology-bound identities differ: %v", err)
	}

	seam := app.transitionSeams[2]
	moveConnectedAcrossStonyField(t, f, 4, seam.First, seam.Second)
	initialPopulation := assertConnectedStonyPopulation(t, f)

	targetTick := app.network.client.NextInputTick(time.Now())

	payload, err := json.Marshal(movement.MovePayload{X: 1, SourceLevel: 3})
	if err != nil {
		t.Fatal(err)
	}

	if err := app.network.submit(targetTick, movement.MoveCommand, payload); err != nil {
		t.Fatal(err)
	}

	for range 15 {
		time.Sleep(40 * time.Millisecond)
		f.advanceGame(t, 1)
	}

	replay, err := app.network.host.Session.Replay()
	if err != nil {
		t.Fatal(err)
	}

	admitted := false

	for _, command := range replay.Commands {
		if command.Tick == targetTick && string(command.Payload) == string(payload) {
			admitted = true
		}
	}

	if !admitted {
		t.Fatal("stale movement sample never reached authority")
	}

	assertStonyFieldConnectedArrival(t, app, app.clientWorld.lastHUD, 4, seam.Second)

	if app.clientWorld.lastHUD.Location.TransitionSequence != 1 {
		t.Fatal("3-to-4 crossing committed more than once")
	}

	stopWatch()

	select {
	case <-watchDone:
	case <-time.After(5 * time.Second):
		t.Fatal("old correction stream did not stop before reconnect")
	}

	epoch := app.network.currentConnectionEpoch()

	cause := errors.New("acceptance transport interruption")
	if err := app.network.recover(ctx, runtime.client, epoch, cause); err != nil {
		t.Fatal(err)
	}

	if app.network.currentConnectionEpoch() != epoch+1 {
		t.Fatal("reconnect did not commit a new connection epoch")
	}
	go app.network.watch(ctx, runtime.client)

	for range 5 {
		time.Sleep(40 * time.Millisecond)
		f.advanceGame(t, 1)
	}

	assertStonyFieldConnectedArrival(t, app, app.clientWorld.lastHUD, 4, seam.Second)

	if app.clientWorld.lastHUD.Location.TransitionSequence != 1 {
		t.Fatal("3-to-4 crossing committed more than once")
	}

	assertConnectedStonyPopulation(t, f)

	moveConnectedAcrossStonyField(t, f, 3, seam.Second, seam.First)

	for range 12 {
		time.Sleep(40 * time.Millisecond)
		f.advanceGame(t, 1)
		assertStonyFieldConnectedArrival(t, app, app.clientWorld.lastHUD, 3, seam.First)

		if app.clientWorld.lastHUD.Location.TransitionSequence != 2 {
			t.Fatal("4-to-3 return committed more than once")
		}
	}
	moveConnectedAcrossStonyField(t, f, 4, seam.First, seam.Second)
	finalPopulation := assertConnectedStonyPopulation(t, f)
	if !reflect.DeepEqual(initialPopulation, finalPopulation) {
		t.Fatal("connected return recreated residents")
	}
}

// Invalid saved seams and zone identities must fail preflight before changing a live session or published maps.
func TestStonyFieldRecoveryRejectsCorruptTopology(t *testing.T) {
	f := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{startScene: "game_world", fixtureCharacters: 1,
		fixtureWorldLevel: 3, applySceneDefaults: true})
	if err := f.app.network.StartSelected(); err != nil {
		t.Fatal(err)
	}

	f.advanceGame(t, 10)

	bundle, err := f.app.entryWorldRecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	before, err := f.app.offlineSession.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	beforeMap, beforeChecksum := f.app.currentWorld().Map, f.app.entryWorldChecksum

	for _, field := range []string{"seam", "zone", "duplicate plains", "duplicate stony",
		"duplicate plains coordinates", "duplicate stony coordinates",
		"duplicate plains destination", "duplicate stony destination"} {
		t.Run(field, func(t *testing.T) {
			var snapshot struct {
				Version int                   `json:"version"`
				Zones   []worldgen.Definition `json:"zones"`
				Seams   []d2mapgen.SeamSpec   `json:"seams"`
			}
			if err := json.Unmarshal(bundle.Topology, &snapshot); err != nil {
				t.Fatal(err)
			}

			switch field {
			case "seam":
				snapshot.Seams[2].SecondWarpRole = "blood-moor-entry"
			case "zone":
				snapshot.Zones[2], snapshot.Zones[3] = snapshot.Zones[3], snapshot.Zones[2]
			default:
				zone, role := &snapshot.Zones[2], "stony-field-exit"
				if strings.Contains(field, "stony") {
					zone, role = &snapshot.Zones[3], "cold-plains-entry"
				}

				found := false

				for _, warp := range zone.Warps {
					if warp.Role != role {
						continue
					}

					found = true

					warp.ID = 99
					if strings.Contains(field, "coordinates") {
						warp.Y++
					} else if strings.Contains(field, "destination") {
						warp.DestinationLevel = 99
					}

					zone.Warps = append(zone.Warps, warp)

					break
				}

				if !found {
					t.Fatal("fixture lacks bound endpoint")
				}
			}

			corrupt := bundle

			corrupt.Topology, err = json.Marshal(snapshot)
			if err != nil {
				t.Fatal(err)
			}

			if err := f.app.restoreOfflineEntryWorldRecovery(corrupt); err == nil ||
				!strings.Contains(err.Error(), "invalid topology") {
				t.Fatalf("corrupt v3 topology must reject before checksum or session recovery: %v", err)
			}

			after, err := f.app.offlineSession.CanonicalCheckpoint()
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(before, after) || f.app.currentWorld().Map != beforeMap ||
				f.app.entryWorldChecksum != beforeChecksum || f.app.activeWorldLevel != 3 ||
				f.app.pendingEntryRecovery != nil {
				t.Fatal("rejected topology mutated session or composition")
			}
		})
	}
}

// Legacy topology versions keep their original population set under the current binary.
func TestStonyFieldLegacyV1PopulationRecovery(t *testing.T) {
	assertStonyLegacyPopulationRecovery(t, 1)
}

// Version 2 recovery retains its three-map topology without introducing level-4 population.
func TestStonyFieldLegacyV2RecoveryKeepsThreeMaps(t *testing.T) {
	assertStonyLegacyPopulationRecovery(t, 2)
}

// Restore a full session checkpoint composed from the preserved topology policy.
func assertStonyLegacyPopulationRecovery(t *testing.T, version int) {
	t.Helper()
	seed := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{})
	generated := seed.app.entryWorldTopology

	runtime, err := d2mapgen.NewRuntime(t.Context(), content.D2Legacy(), seed.app.records)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = runtime.Close(context.Background()) }()

	request := generated.Zones[3].Request()

	plains, err := runtime.Generate(t.Context(), "cold_plains", float64(3), float64(request.Seed),
		generated.Seams[1].FirstDirection, float64(request.Difficulty))
	if err != nil {
		t.Fatal(err)
	}

	legacy := d2mapgen.EntryWorld{Version: 2, Zones: map[int]*worldgen.Zone{
		1: generated.Zones[1], 2: generated.Zones[2], 3: plains}, Seams: generated.Seams[:2]}

	if version == 1 {
		legacy.Version = 1
		delete(legacy.Zones, 3)
		legacy.Seams = legacy.Seams[:1]
	}

	snapshot, err := legacy.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	config := realD2LegacyFixtureConfig{startScene: "game_world", fixtureCharacters: 1,
		fixtureWorldLevel: version + 1, applySceneDefaults: true, topologySnapshot: snapshot}

	source := newRealD2LegacyFixture(t, config)
	if err := source.app.network.StartSelected(); err != nil {
		t.Fatal(err)
	}

	source.advanceGame(t, 10)

	bundle, err := source.app.entryWorldRecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	config.topologySnapshot, config.recovery, config.skipPopulationBootstrap = nil, &bundle, true

	restored := newRealD2LegacyFixture(t, config)
	if err := restored.app.network.StartSelected(); err != nil {
		t.Fatal(err)
	}

	if restored.app.entryWorldTopology.Version != version || len(restored.app.gameWorlds) != version+1 ||
		len(restored.app.transitionSeams) != version || restored.app.gameWorlds[4] != nil ||
		restored.app.entryWorldChecksum != bundle.TopologyChecksum {
		t.Fatal("legacy recovery gained a Stony Field map or seam")
	}

	got, err := restored.app.offlineSession.RecoveryCheckpoint()
	if err != nil || got.Checksum != bundle.Session.Checksum {
		t.Fatalf("legacy checkpoint changed: %v", err)
	}

	for _, zone := range populationPlan(t, restored).Zones {
		if zone.Level == 4 {
			t.Fatal("legacy recovery added level-4 population plan")
		}
	}

	if len(populationResidentIDs(t, restored, 4, false)) != 0 {
		t.Fatal("legacy recovery added level-4 residents")
	}

	if version == 2 {
		seam := restored.app.transitionSeams[1]
		moveAcrossColdPlainsSeam(t, restored, 3, 2, seam.Second, seam.First)
	}
}

// The level-3 entry is across the full map from its level-4 exit, so allow a full minute of real server ticks.
func moveConnectedAcrossStonyField(t *testing.T, f *realD2LegacyFixture, to int,
	point, arrival gametransition.SeamEndpoint) {
	t.Helper()

	if err := f.app.playerControl.SetMoveTargetWithRadius(point.X, point.Y, 0.5); err != nil {
		t.Fatal(err)
	}

	for range 1500 {
		time.Sleep(40 * time.Millisecond)
		f.advanceGame(t, 1)

		hud := f.app.clientWorld.lastHUD
		if int(hud.Location.LevelID) == to {
			assertStonyFieldConnectedArrival(t, f.app, hud, to, arrival)
			return
		}
	}

	t.Fatalf("connected movement never crossed to %d: %#v", to, f.app.clientWorld.lastHUD)
}

// Verify the rendered replica's hero transform too; matching HUD facts alone can hide a stale local projection.
func assertStonyFieldConnectedArrival(t *testing.T, app *application, hud playeradapter.HUD, to int,
	arrival gametransition.SeamEndpoint) {
	t.Helper()
	assertConnectedColdPlainsArrival(t, app, hud, to, arrival)

	controls, found := akara.GetDynamicStore(app.clientSimulation.World(), "d2legacy.world.player_control")
	if !found {
		t.Fatal("connected hero controls unavailable")
	}

	for _, entity := range controls.Entities() {
		control, _ := controls.Get(entity)

		owner, _ := control.Get("player")
		if owner != hud.Player.PlayerID {
			continue
		}

		position := currentPosition(app.clientSimulation.World(), entity)
		if math.Hypot(position.X-hud.Position.X, position.Y-hud.Position.Y) > 0.01 {
			t.Fatalf("projected hero position %#v differs from authoritative HUD %#v", position, hud.Position)
		}

		return
	}

	t.Fatal("connected projection lacks the authenticated hero")
}
