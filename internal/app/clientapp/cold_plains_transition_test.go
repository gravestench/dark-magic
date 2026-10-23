package clientapp

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/app/clientsession"
	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/app/networktrust"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/content"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	d2mapgen "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/mapgen"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
)

// TestLiveColdPlainsRoundTrip follows both generated seams through player movement and checks authority and
// presentation after each committed crossing against the same prepared topology.
func TestLiveColdPlainsRoundTrip(t *testing.T) {
	fixture := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{
		startScene: "game_world", fixtureCharacters: 1, fixtureWorldSpawn: "seam", applySceneDefaults: true,
	})
	app := fixture.app
	if err := app.network.StartSelected(); err != nil {
		t.Fatal(err)
	}
	fixture.advanceGame(t, 10)
	if len(app.transitionSeams) != 2 {
		t.Fatalf("prepared seams = %d, want two", len(app.transitionSeams))
	}
	if checksum, err := app.entryWorldTopology.Checksum(); err != nil || checksum != app.entryWorldChecksum {
		t.Fatalf("published topology checksum = %s, want %s: %v", app.entryWorldChecksum, checksum, err)
	}
	if got := app.sessionInitialData()["d2legacy.entry_world_checksum"]; got != app.entryWorldChecksum {
		t.Fatalf("session identity topology = %v, want %s", got, app.entryWorldChecksum)
	}

	first, second := app.transitionSeams[0], app.transitionSeams[1]
	for _, crossing := range []struct {
		from, to int
		point    gametransition.SeamEndpoint
		arrival  gametransition.SeamEndpoint
	}{
		{1, 2, first.First, first.Second},
		{2, 3, second.First, second.Second},
		{3, 2, second.Second, second.First},
		{2, 1, first.Second, first.First},
	} {
		moveAcrossColdPlainsSeam(t, fixture, crossing.from, crossing.to, crossing.point, crossing.arrival)
	}
}

// A valid but different level recipe must fail the real QUIC join after a matching control joins successfully.
func TestConnectedColdPlainsRejectsDifferentTopologyAtHandshake(t *testing.T) {
	fixture := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{
		startScene: "game_world", fixtureCharacters: 1, applySceneDefaults: true,
	})
	app := fixture.app
	app.ctx = t.Context()
	var err error
	app.networkTrust, err = networktrust.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	source, err := app.modSource("d2legacy")
	if err != nil {
		t.Fatal(err)
	}
	other, err := d2mapgen.GenerateEntryWorld(t.Context(), content.D2Legacy(), app.records, 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	otherChecksum, err := other.Checksum()
	if err != nil || otherChecksum == app.entryWorldChecksum {
		t.Fatalf("different valid topology checksum = %q, original %q: %v",
			otherChecksum, app.entryWorldChecksum, err)
	}
	originalChecksum := app.entryWorldChecksum
	app.entryWorldChecksum = otherChecksum
	mismatched, err := app.network.runtimeIdentity(source, app.options.Packages)
	app.entryWorldChecksum = originalChecksum
	if err != nil {
		t.Fatal(err)
	}
	runtime := &listenHostRuntime{}
	defer runtime.close()
	if err := app.network.prepareListenHost(t.Context(), runtime); err != nil {
		t.Fatal(err)
	}
	app.network.startListenServices(t.Context(), 0, runtime)
	endpoint := realm.GameEndpoint{Address: listenDialAddress, TLSFingerprint: runtime.fingerprint}
	control := clientsession.SelfHostedAssignment{
		GameID: listenGameID, Endpoint: endpoint,
		Runtime: runtime.host.Authority.Identity, ProfileCredential: runtime.profileCredential,
	}
	admitted, err := clientsession.ConnectSelfHosted(t.Context(), control, runtime.clientTLS, app.saves)
	if err != nil {
		t.Fatalf("matching topology failed the control handshake: %v", err)
	}
	defer func() { _ = admitted.Close(context.Background()) }()
	assignment := clientsession.SelfHostedAssignment{
		GameID:   listenGameID,
		Endpoint: endpoint,
		Runtime:  mismatched, ProfileCredential: runtime.profileCredential,
	}
	if client, err := clientsession.ConnectSelfHosted(t.Context(), assignment, runtime.clientTLS, app.saves); err == nil {
		_ = client.Close(context.Background())
		t.Fatal("handshake admitted a different valid entry topology")
	} else if !strings.Contains(err.Error(), gamesession.ErrCompatibility.Error()) {
		t.Fatalf("different topology failed for a reason other than runtime compatibility: %v", err)
	}
}

// A failed new preparation must leave the published maps, identity, and active navigation usable.
func TestColdPlainsSetupFailurePreservesExistingWorld(t *testing.T) {
	fixture := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{})
	app := fixture.app
	beforeWorld := app.gameWorlds[app.activeWorldLevel]
	beforeChecksum := app.entryWorldChecksum
	beforeTopology := app.entryWorldTopology
	app.options.GameDifficulty = 3
	if err := app.buildEntryWorld(); err == nil {
		t.Fatal("invalid three-zone preparation succeeded")
	}
	if app.gameWorlds[app.activeWorldLevel] != beforeWorld || app.entryWorldChecksum != beforeChecksum ||
		app.entryWorldTopology.Zones[3] != beforeTopology.Zones[3] || app.currentWorld().Map != beforeWorld {
		t.Fatal("failed preparation replaced the published topology")
	}
}

type headlessClientBoot struct{}

// The headless fixture has no renderer; the production QUIC, authority, replica, and movement paths still run.
func (headlessClientBoot) Start(context.Context) error { return nil }

// Releasing the headless boot has no renderer resources to reclaim.
func (headlessClientBoot) Stop(context.Context) error { return nil }

// TestConnectedColdPlainsMatchesAuthority joins the production listen authority over QUIC and follows its
// authenticated Cold Plains return projection with the same checksum-pinned topology as the client.
func TestConnectedColdPlainsMatchesAuthority(t *testing.T) {
	fixture := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{
		startScene: "game_world", fixtureCharacters: 1, fixtureWorldLevel: 2,
		fixtureWorldSpawn: "cold_plains_seam", applySceneDefaults: true,
	})
	app := fixture.app
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
		if status := app.network.Status(); status["phase"] == "failed" {
			t.Fatalf("listen connection failed: %v", status["error"])
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !app.network.Connected() {
		t.Fatalf("listen connection timed out: %#v", app.network.Status())
	}
	fixture.advanceGame(t, 10)
	if app.network.client == nil || app.network.host == nil || app.clientSimulation == nil {
		t.Fatal("connected session lacks authority or presentation replica")
	}
	if digest, err := app.network.host.Authority.Identity.Digest(); err != nil || digest == "" {
		t.Fatalf("listen authority identity digest = %q: %v", digest, err)
	}
	seam := app.transitionSeams[1]
	moveConnectedAcrossColdPlains(t, fixture, 3, seam.First, seam.Second)
	moveConnectedAcrossColdPlains(t, fixture, 2, seam.Second, seam.First)
}

// moveConnectedAcrossColdPlains waits for QUIC correction, then checks authority and replica against the arrival.
func moveConnectedAcrossColdPlains(t *testing.T, fixture *realD2LegacyFixture, to int,
	point, arrival gametransition.SeamEndpoint) {
	t.Helper()
	app := fixture.app
	if err := app.playerControl.SetMoveTargetWithRadius(point.X, point.Y, 0.5); err != nil {
		t.Fatal(err)
	}
	for range 400 {
		// The server ticks on a real timer; frame-only polling can outrun its authoritative movement.
		time.Sleep(40 * time.Millisecond)
		fixture.advanceGame(t, 1)
		hud, _ := app.network.client.View()
		if int(hud.Location.LevelID) != to {
			continue
		}
		assertConnectedColdPlainsArrival(t, app, hud, to, arrival)
		return
	}
	hud, _ := app.network.client.View()
	t.Fatalf("connected movement never crossed to %d; HUD level = %d", to, hud.Location.LevelID)
}

// assertConnectedColdPlainsArrival compares the replica, authoritative checkpoint, map, and route at a committed seam.
func assertConnectedColdPlainsArrival(t *testing.T, app *application, hud playeradapter.HUD, to int,
	arrival gametransition.SeamEndpoint) {
	t.Helper()
	if app.activeWorldLevel != to || app.currentWorld().Map != app.gameWorlds[to] {
		t.Fatalf("connected map = %d, want %d", app.activeWorldLevel, to)
	}
	level, found := controlledPlayerLevel(app.clientSimulation.World(), app.presentationPlayerID())
	if !found || level != to {
		t.Fatalf("connected player projection = %d/%v, want %d", level, found, to)
	}
	checkpoint, err := app.network.host.Session.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	payload, err := playeradapter.ProjectHUD(app.presentationPlayerID(), checkpoint)
	if err != nil {
		t.Fatal(err)
	}
	var authority playeradapter.HUD
	if err := json.Unmarshal(payload, &authority); err != nil {
		t.Fatal(err)
	}
	if math.Hypot(hud.Position.X-arrival.ArrivalX, hud.Position.Y-arrival.ArrivalY) > 0.01 {
		t.Fatalf("client arrival = (%v,%v) tick %d; authority = (%v,%v) tick %d; want (%v,%v); target=%v",
			hud.Position.X, hud.Position.Y, hud.Tick, authority.Position.X, authority.Position.Y,
			authority.Tick, arrival.ArrivalX, arrival.ArrivalY, app.playerControl.HasMoveTarget())
	}
	if int(authority.Location.LevelID) != to ||
		math.Hypot(authority.Position.X-arrival.ArrivalX, authority.Position.Y-arrival.ArrivalY) > 0.01 {
		t.Fatalf("authority arrived at level %d (%v,%v), want %d (%v,%v)",
			authority.Location.LevelID, authority.Position.X, authority.Position.Y,
			to, arrival.ArrivalX, arrival.ArrivalY)
	}
	worldMap := app.gameWorlds[to]
	goalX, goalY, open := worldMap.OpenPointNearSubtile(hud.Position.X+3, hud.Position.Y+3)
	if !open {
		t.Fatal("connected arrival has no nearby navigation point")
	}
	if _, err := worldMap.FindPath(gameworld.PathRequest{
		Start: gameworld.Point{X: hud.Position.X, Y: hud.Position.Y},
		Goal:  gameworld.Point{X: goalX, Y: goalY},
	}); err != nil {
		t.Fatalf("connected arrival has no navigation route: %v", err)
	}
}

// moveAcrossColdPlainsSeam uses the production point movement controller, then observes only committed player state.
func moveAcrossColdPlainsSeam(t *testing.T, fixture *realD2LegacyFixture, from, to int,
	point, arrival gametransition.SeamEndpoint) {
	t.Helper()
	app := fixture.app
	player := controlledColdPlainsPlayer(t, fixture)
	locations := requireRealStore(t, fixture, "d2legacy.world.location", "missing player location")
	positions := requireRealStore(t, fixture, "d2legacy.world.position", "missing player position")
	location, present := locations.Get(player)
	if !present {
		t.Fatal("controlled player has no location")
	}
	if level, _ := location.Get("level_id"); level != int64(from) {
		t.Fatalf("before crossing level = %v, want %d", level, from)
	}
	if err := app.playerControl.SetMoveTargetWithRadius(point.X, point.Y, 0.5); err != nil {
		t.Fatal(err)
	}
	for range 1500 {
		fixture.advanceGame(t, 1)
		if level, _ := location.Get("level_id"); level == int64(to) {
			x, y := dynamicPosition(positions, player)
			if math.Hypot(x-arrival.ArrivalX, y-arrival.ArrivalY) > 0.01 {
				t.Fatalf("arrival at (%v,%v), want (%v,%v)", x, y, arrival.ArrivalX, arrival.ArrivalY)
			}
			if app.activeWorldLevel != to || app.currentWorld().Map != app.gameWorlds[to] {
				t.Fatalf("active presentation map = %d, want %d", app.activeWorldLevel, to)
			}
			if app.playerControl.HasMoveTarget() {
				t.Fatal("crossing retained an old-world movement target")
			}
			worldMap := app.gameWorlds[to]
			goalX, goalY, found := worldMap.OpenPointNearSubtile(x+3, y+3)
			if !found {
				t.Fatal("arrival has no nearby walkable navigation point")
			}
			if _, err := worldMap.FindPath(gameworld.PathRequest{
				Start: gameworld.Point{X: x, Y: y}, Goal: gameworld.Point{X: goalX, Y: goalY},
			}); err != nil {
				t.Fatalf("arrival has no active navigation route: %v", err)
			}
			return
		}
	}
	t.Fatalf("ordinary movement never crossed %d to %d", from, to)
}

// controlledColdPlainsPlayer selects the authority-owned character rather than a projected or remote entity.
func controlledColdPlainsPlayer(t *testing.T, fixture *realD2LegacyFixture) akara.Entity {
	t.Helper()
	controls := requireRealStore(t, fixture, "d2legacy.world.player_control", "missing player control")
	if controls.Len() != 1 {
		t.Fatalf("controlled players = %d, want one", controls.Len())
	}
	return controls.Entities()[0]
}
