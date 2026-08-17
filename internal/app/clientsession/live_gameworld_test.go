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
		"links": []map[string]any{},
		"rooms": []map[string]any{{"id": "blood-moor-network", "populate": true,
			"x": 0, "y": 0, "width": 20, "height": 20,
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
		Stats: &d2save.Stats{Dexterity: 20, Vitality: 20, Health: 50, MaxHealth: 50, Mana: 20, MaxMana: 20}}
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
	if connected.HUD.Player.PlayerID != "alice-1" || connected.HUD.Player.CharacterID != "hero" || connected.HUD.Player.Name != "Hero" || connected.World.Tick == 0 {
		t.Fatalf("live initial view = HUD %#v world tick %d", connected.HUD.Player, connected.World.Tick)
	}
	if !containsWorldEntity(connected.World.Entities, "monster:level:2:room:blood-moor-network:monster:1", "hostile") {
		t.Fatalf("live generated hostile missing from world view: %#v", connected.World.Entities)
	}
	barbarian := d2save.Character{ID: "barbarian", Name: "Conan", Class: "Barbarian", Level: 1, Expansion: true,
		Stats: &d2save.Stats{Dexterity: 20, Vitality: 25, Health: 60, MaxHealth: 60, Mana: 10, MaxMana: 10}}
	barbarianProfile := d2save.New(barbarian)
	if err := barbarianProfile.Select(barbarian.ID); err != nil {
		t.Fatal(err)
	}
	second, err := ConnectSelfHosted(ctx, SelfHostedAssignment{
		GameID: "live-gameworld", Endpoint: realm.GameEndpoint{Address: server.Addr(), TLSFingerprint: fingerprint},
		Runtime: host.Authority.Identity, ProfileCredential: "profile-secret",
	}, clientTLS, barbarianProfile)
	if err != nil {
		t.Fatal(err)
	}
	if second.HUD.Player.PlayerID != "alice-2" || second.HUD.Player.CharacterID != "barbarian" || second.HUD.Player.Class != "Barbarian" {
		t.Fatalf("second client HUD identity = %#v", second.HUD.Player)
	}
	// The invite is a UI-visible relationship action. Make the inviter observe
	// the newly entered target before deriving its command tick; otherwise a
	// stale client timeline can legally replay the invite before target entry.
	waitForProjectedPlayer(t, ctx, connected, "player:alice-2")
	invitePayload, _ := json.Marshal(map[string]any{"target": "alice-2"})
	if err := connected.Submit(ctx, gameserver.CommandIntent{TargetTick: connected.NextInputTick(time.Now()), Sequence: 1,
		Kind: "party.invite", Payload: invitePayload}); err != nil {
		t.Fatal(err)
	}
	waitForPartyInvite(t, ctx, host.Authority.State, "alice-1", "alice-2")
	// Authority has committed the invite, but the invitee's local clock may
	// still predate that tick under race-detector scheduling. Refresh before
	// deriving acceptance so deterministic command order cannot put accept
	// ahead of its prerequisite invite.
	if _, err := second.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	acceptPayload, _ := json.Marshal(map[string]any{"inviter": "alice-1"})
	if err := second.Submit(ctx, gameserver.CommandIntent{TargetTick: second.NextInputTick(time.Now()), Sequence: 1,
		Kind: "party.accept", Payload: acceptPayload}); err != nil {
		t.Fatal(err)
	}
	partyID := waitForPartyMembership(t, ctx, host.Authority.State, "alice-1", "alice-2")
	if err := second.Reconnect(ctx); err != nil {
		t.Fatal(err)
	}
	if reconnectedParty := waitForPartyMembership(t, ctx, host.Authority.State, "alice-1", "alice-2"); reconnectedParty != partyID {
		t.Fatalf("reconnect changed party identity from %q to %q", partyID, reconnectedParty)
	}
	if hostPeer, found := findWorldEntity(second.World.Entities, "player:alice-1", "player"); !found {
		t.Fatalf("host player absent from second client's initial projection: %#v", second.World.Entities)
	} else if hostPeer.Owner != "alice-1" || hostPeer.Class != "Amazon" || hostPeer.Token != "AM" || hostPeer.Position.X != 10 || hostPeer.Position.Y != 10 {
		t.Fatalf("host player projection = %#v", hostPeer)
	}
	t.Cleanup(func() {
		closeContext, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = second.Close(closeContext)
	})
	for attempt := 0; attempt < 4; attempt++ {
		if _, err := connected.Refresh(ctx); err != nil {
			t.Fatal(err)
		}
		_, firstWorld := connected.View()
		if peer, found := findWorldEntity(firstWorld.Entities, "player:alice-2", "player"); found {
			if peer.Owner != "alice-2" || peer.Class != "Barbarian" || peer.Token != "BA" || peer.Position.X != 18 || peer.Position.Y != 10 {
				t.Fatalf("second player projection = %#v", peer)
			}
			break
		}
		if attempt == 3 {
			t.Fatalf("second player absent from first client's projection: %#v", firstWorld.Entities)
		}
		time.Sleep(550 * time.Millisecond)
	}
	movePayload, err := json.Marshal(movement.MovePayload{X: 1})
	if err != nil {
		t.Fatal(err)
	}
	// The joining membership must advance under its own input while the first
	// (hosting) membership stays idle. This guards against accidentally using a
	// process-global command clock, acknowledgement, or wakeup path.
	secondWatchContext, stopSecondWatch := context.WithCancel(ctx)
	secondDeltas, secondWatchErrors, err := second.Watch(secondWatchContext)
	if err != nil {
		t.Fatal(err)
	}
	defer stopSecondWatch()
	secondHUD, secondWorld := second.View()
	secondInitialTick, secondInitialX := secondWorld.Tick, secondWorld.Origin.X
	secondCommandTick := second.NextInputTick(time.Now())
	if err := second.Submit(ctx, gameserver.CommandIntent{TargetTick: secondCommandTick, Sequence: 2,
		Kind: movement.MoveCommand, Payload: movePayload}); err != nil {
		t.Fatal(err)
	}
	for secondWorld.Tick <= secondInitialTick || secondWorld.Origin.X <= secondInitialX {
		select {
		case _, open := <-secondDeltas:
			if !open {
				t.Fatal("second correction stream closed before independent movement")
			}
			secondHUD, secondWorld = second.View()
		case err := <-secondWatchErrors:
			t.Fatalf("second correction stream error = %v", err)
		case <-ctx.Done():
			replay, _ := host.Session.Replay()
			t.Fatalf("second-client movement timed out: %v; HUD=%#v world=%#v commands=%#v",
				ctx.Err(), secondHUD, secondWorld, replay.Commands)
		}
	}
	// The second membership's movement may advance authority well beyond the
	// first client's last correction. Refresh that independent connection before
	// deriving a tick from its synchronized network clock.
	if _, err := connected.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	commandTick := connected.NextInputTick(time.Now())
	watchContext, stopWatch := context.WithCancel(ctx)
	deltas, watchErrors, err := connected.Watch(watchContext)
	if err != nil {
		t.Fatal(err)
	}
	defer stopWatch()
	if err := connected.Submit(ctx, gameserver.CommandIntent{ObservedServerTick: commandTick - 2, TargetTick: commandTick, Sequence: 2,
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
	waitForTransformDatagram(t, ctx, connected)
	if err := connected.Close(ctx); err != nil {
		t.Fatal(err)
	}

	stopServe()
	stopRun()
	assertCanceledLoop(t, ctx, runErrors, "session")
	assertCanceledLoop(t, ctx, serveErrors, "QUIC")
}

type livePartyState struct {
	Membership map[string]string                     `json:"membership"`
	Invites    map[string]map[string]json.RawMessage `json:"invites"`
}

func readLivePartyState(t *testing.T, store *simulation.StateStore) livePartyState {
	t.Helper()
	registered, found := store.Read("d2legacy.party")
	if !found {
		t.Fatal("live party authority state is missing")
	}
	var state livePartyState
	if err := json.Unmarshal(registered.Data, &state); err != nil {
		t.Fatal(err)
	}
	return state
}

func waitForPartyInvite(t *testing.T, ctx context.Context, store *simulation.StateStore, inviter, target string) {
	t.Helper()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		if state := readLivePartyState(t, store); state.Invites[target][inviter] != nil {
			return
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			t.Fatalf("party invitation %s -> %s timed out: %v", inviter, target, ctx.Err())
		}
	}
}

func waitForProjectedPlayer(t *testing.T, ctx context.Context, session *Session, selectableID string) {
	t.Helper()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		if _, err := session.Refresh(ctx); err != nil {
			t.Fatal(err)
		}
		_, world := session.View()
		if _, found := findWorldEntity(world.Entities, selectableID, "player"); found {
			return
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			t.Fatalf("projected player %s timed out: %v", selectableID, ctx.Err())
		}
	}
}

func waitForPartyMembership(t *testing.T, ctx context.Context, store *simulation.StateStore, first, second string) string {
	t.Helper()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		state := readLivePartyState(t, store)
		if partyID := state.Membership[first]; partyID != "" && state.Membership[second] == partyID {
			return partyID
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			t.Fatalf("party membership for %s/%s timed out: %v", first, second, ctx.Err())
		}
	}
}

func waitForTransformDatagram(t *testing.T, ctx context.Context, session *Session) {
	t.Helper()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		stats := session.transport.NetworkStats()
		if stats.TransformsReceived > 0 {
			return
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			t.Fatalf("movement used no compact transform datagrams: %v; stats=%#v", ctx.Err(), stats)
		}
	}
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

func findWorldEntity(entities []playeradapter.WorldEntity, id, kind string) (playeradapter.WorldEntity, bool) {
	for _, entity := range entities {
		if entity.ID == id && entity.Kind == kind {
			return entity, true
		}
	}
	return playeradapter.WorldEntity{}, false
}

func containsWorldEntity(entities []playeradapter.WorldEntity, id, kind string) bool {
	_, found := findWorldEntity(entities, id, kind)
	return found
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
		return []map[string]string{
			{"class": "Amazon", "vit": "20", "StartSkill": "Fire Bolt", "WalkVelocity": "6", "RunVelocity": "9", "stamina": "84", "RunDrain": "20", "StaminaPerLevel": "4", "StaminaPerVitality": "4"},
			{"class": "Barbarian", "vit": "25", "StartSkill": "Fire Bolt", "WalkVelocity": "6", "RunVelocity": "9", "stamina": "92", "RunDrain": "20", "StaminaPerLevel": "4", "StaminaPerVitality": "4"},
		}, nil
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
