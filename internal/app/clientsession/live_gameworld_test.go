package clientsession

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/app/serverapp"
	"github.com/gravestench/dark-magic/internal/content"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

// TestConnectSelfHostedEntersLiveGeneratedGameworld crosses the complete
// production boundary: Lua population, ticking ECS, profile admission, real
// QUIC, and canonical owner-private plus nearby-public projections.
func TestConnectSelfHostedEntersLiveGeneratedGameworld(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	host, err := gameserver.Start(ctx, content.D2Legacy(), liveGameworldRecords{}, gameserver.Config{
		Mode: gameserver.ModeStandalone, SessionID: "live-gameworld", Seed: 314,
		Prediction: gamesession.PredictionLimited,
		Session:    gamesession.Config{Step: 40 * time.Millisecond, CheckpointInterval: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = host.Close(context.Background()) })

	population, err := json.Marshal(map[string]any{
		"act": 1, "level_id": 2, "difficulty": 0,
		"rooms": []map[string]any{{"id": "blood-moor-network", "populate": true,
			"points": []map[string]any{{"x": 14, "y": 10}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := host.Session.Submit(simulation.Command{Tick: 1, Player: "population", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.population.bootstrap", Payload: population}); err != nil {
		t.Fatal(err)
	}
	runContext, stopRun := context.WithCancel(ctx)
	runErrors := make(chan error, 1)
	go func() { runErrors <- host.Session.Run(runContext) }()

	tickets, err := gameserver.NewTicketAuthority([]byte("0123456789abcdef0123456789abcdef"), "live-gameworld")
	if err != nil {
		t.Fatal(err)
	}
	endpoint, err := gameserver.NewEndpoint(host, tickets, playeradapter.ProjectClientView)
	if err != nil {
		t.Fatal(err)
	}
	endpoint.SetSnapshotPending(func(err error) bool { return errors.Is(err, playeradapter.ErrHUDPlayer) })
	serverTLS, clientTLS, fingerprint := connectTLS(t)
	server, err := sessionquic.Listen("127.0.0.1:0", serverTLS, endpoint)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })
	destination, err := playeradapter.NewDestination(10, 10, 100, 100, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	profiles, err := serverapp.NewRemoteProfileAdmissions(host, tickets, serverapp.RemoteProfileConfig{
		Credential: "profile-secret", PrincipalID: "local-account", PlayerID: "alice",
		Destination: destination, Lifetime: time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	server.SetProfileAdmissions(profiles)
	serveContext, stopServe := context.WithCancel(ctx)
	serveErrors := make(chan error, 1)
	go func() { serveErrors <- server.Serve(serveContext) }()

	character := d2save.Character{ID: "hero", Name: "Hero", Class: "Amazon", Level: 1, Expansion: true,
		Stats: &d2save.Stats{Dexterity: 20, Health: 50, MaxHealth: 50, Mana: 20, MaxMana: 20}}
	profile := d2save.New(character)
	if err := profile.Select(character.ID); err != nil {
		t.Fatal(err)
	}
	connected, err := ConnectSelfHosted(ctx, SelfHostedAssignment{
		GameID: "live-gameworld", Endpoint: realm.GameEndpoint{Address: server.Addr(), TLSFingerprint: fingerprint},
		Runtime: host.Authority.Identity, ProfileCredential: "profile-secret",
	}, clientTLS, profile)
	if err != nil {
		t.Fatal(err)
	}
	if connected.HUD.Player.CharacterID != "hero" || connected.HUD.Player.Name != "Hero" || connected.World.Tick == 0 {
		t.Fatalf("live initial view = HUD %#v world tick %d", connected.HUD.Player, connected.World.Tick)
	}
	if !containsWorldEntity(connected.World.Entities, "monster:level:2:room:blood-moor-network:monster:1", "hostile") {
		t.Fatalf("live generated hostile missing from world view: %#v", connected.World.Entities)
	}
	movePayload, err := json.Marshal(movement.MovePayload{X: 1})
	if err != nil {
		t.Fatal(err)
	}
	commandTick := connected.World.Tick + 2
	watchContext, stopWatch := context.WithCancel(ctx)
	deltas, watchErrors, err := connected.Watch(watchContext)
	if err != nil {
		t.Fatal(err)
	}
	defer stopWatch()
	if err := connected.Submit(ctx, gameserver.CommandIntent{Tick: commandTick, Sequence: 1,
		Kind: movement.MoveCommand, Payload: movePayload}); err != nil {
		t.Fatal(err)
	}
	_, currentWorld := connected.View()
	initialTick, initialX := currentWorld.Tick, currentWorld.Origin.X
	for currentWorld.Tick <= initialTick || currentWorld.Origin.X <= initialX {
		select {
		case _, open := <-deltas:
			if !open {
				t.Fatal("correction stream closed before movement")
			}
			_, currentWorld = connected.View()
		case err := <-watchErrors:
			t.Fatalf("correction stream error = %v", err)
		case <-ctx.Done():
			replay, _ := host.Session.Replay()
			t.Fatalf("movement correction timed out: %v; world=%#v commands=%#v", ctx.Err(), currentWorld, replay.Commands)
		}
	}
	currentHUD, currentWorld := connected.View()
	if currentHUD.Tick != currentWorld.Tick {
		t.Fatalf("correction HUD/world ticks differ: %d/%d", currentHUD.Tick, currentWorld.Tick)
	}
	if err := connected.Close(ctx); err != nil {
		t.Fatal(err)
	}

	stopServe()
	stopRun()
	assertCanceledLoop(t, ctx, runErrors, "session")
	assertCanceledLoop(t, ctx, serveErrors, "QUIC")
}

func assertCanceledLoop(t *testing.T, ctx context.Context, result <-chan error, name string) {
	t.Helper()
	select {
	case err := <-result:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatalf("%s loop error = %v", name, err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}

func containsWorldEntity(entities []playeradapter.WorldEntity, id, kind string) bool {
	for _, entity := range entities {
		if entity.ID == id && entity.Kind == kind {
			return true
		}
	}
	return false
}

// liveGameworldRecords are synthetic authored facts, not alternate gameplay
// implementations. Production Lua consumes them through the ordinary Records
// capability, keeping CI independent from copyrighted MPQ installations.
type liveGameworldRecords struct{}

func (liveGameworldRecords) Invalidate(string)  {}
func (liveGameworldRecords) Loaded(string) bool { return true }
func (liveGameworldRecords) Load(path string) ([]map[string]string, error) {
	switch path {
	case "data/global/excel/charstats.txt":
		return []map[string]string{{"class": "Amazon", "StartSkill": "Fire Bolt"}}, nil
	case "data/global/excel/levels.txt":
		return []map[string]string{{"Id": "2", "MonDen": "100000", "NumMon": "1", "mon1": "fallen"}}, nil
	case "data/global/excel/monstats.txt":
		return []map[string]string{{"Id": "fallen", "BaseId": "fallen", "NameStr": "Fallen", "AI": "fallen", "Code": "FA",
			"enabled": "1", "isSpawn": "1", "npc": "0", "noRatio": "1", "Level": "1", "minHP": "3", "maxHP": "3",
			"AC": "0", "A1TH": "0", "A1MinD": "1", "A1MaxD": "1", "Exp": "5", "Velocity": "0", "aidel": "1",
			"aidist": "20", "MinGrp": "1", "MaxGrp": "1", "Rarity": "1", "TreasureClass1": "fallen-drop"}}, nil
	case "data/global/excel/monstats2.txt":
		return []map[string]string{{"Id": "fallen", "BaseW": "HTH", "SizeX": "1", "SizeY": "1", "MeleeRng": "1"}}, nil
	case "data/global/excel/monlvl.txt":
		return []map[string]string{{"Level": "1"}}, nil
	case "data/global/excel/skills.txt":
		return []map[string]string{{"Id": "36", "skill": "Fire Bolt", "skilldesc": "firebolt", "leftskill": "1", "general": "0",
			"passive": "0", "srvmissile": "firebolt", "etype": "fire", "interrupt": "1", "mana": "5", "manashift": "7",
			"srvstfunc": "", "srvdofunc": "", "emin": "3", "emax": "6", "HitShift": "8"}}, nil
	case "data/global/excel/skilldesc.txt":
		return []map[string]string{{"skilldesc": "firebolt", "ListRow": "0", "IconCel": "0"}}, nil
	case "data/global/excel/Missiles.txt":
		return []map[string]string{{"Missile": "firebolt", "Skill": "Fire Bolt", "pSrvDoFunc": "1", "CollideType": "3",
			"CollideKill": "1", "Vel": "20", "Range": "40", "Size": "2", "CelFile": "firebolt"}}, nil
	default:
		return nil, nil
	}
}
