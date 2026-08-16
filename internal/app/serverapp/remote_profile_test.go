package serverapp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

func TestRemoteProfileAdmissionAuthenticatesQueuesAndIssuesOneUseTicket(t *testing.T) {
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close(); _ = engine.Close() })
	if err := session.Register(playeradapter.EnterCommand, gamesession.CommandHandler{Validate: func(simulation.Command) error { return nil }, Apply: func(*gameecs.Engine, simulation.Command) error { return nil }, Allowed: []simulation.Authority{simulation.AuthoritySystem}}); err != nil {
		t.Fatal(err)
	}
	identity := remoteProfileIdentity()
	allocation, err := gamesession.Allocate("game", identity, gamesession.PredictionLimited)
	if err != nil {
		t.Fatal(err)
	}
	host := &gameserver.Host{Mode: gameserver.ModeStandalone, Engine: engine, Session: session, Allocation: allocation}
	tickets, err := gameserver.NewTicketAuthority([]byte("0123456789abcdef0123456789abcdef"), "game")
	if err != nil {
		t.Fatal(err)
	}
	destination, _ := playeradapter.NewDestination(10, 20, 100, 100, 1, 40)
	admissions, err := NewRemoteProfileAdmissions(host, tickets, RemoteProfileConfig{Credential: "host-password", PrincipalID: "self-host-user", PlayerID: "player", Destination: destination, Lifetime: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	offer, err := d2save.EncodeCharacterOffer(d2save.Character{ID: "hero", Name: "Hero", Class: "Amazon"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := admissions.Admit(context.Background(), "wrong", offer); !errors.Is(err, ErrRemoteProfileAdmission) {
		t.Fatalf("credential error = %v", err)
	}
	ticket, err := admissions.Admit(context.Background(), "host-password", offer)
	if err != nil {
		t.Fatal(err)
	}
	principal, err := tickets.Authenticate(context.Background(), ticket)
	if err != nil || principal.CharacterID != "hero" || principal.PlayerID != "player-1" || principal.RuntimeIdentityHash != allocation.IdentityHash {
		t.Fatalf("principal=%#v error=%v", principal, err)
	}
	if _, err := admissions.Admit(context.Background(), "host-password", offer); err != nil {
		t.Fatalf("second player admission: %v", err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	replay, err := session.Replay()
	if err != nil || len(replay.Commands) != 2 || replay.Commands[0].Player != "self-host:remote-profile" {
		t.Fatalf("replay=%#v error=%v", replay, err)
	}
	var first, second playeradapter.Entry
	if err := json.Unmarshal(replay.Commands[0].Payload, &first); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(replay.Commands[1].Payload, &second); err != nil {
		t.Fatal(err)
	}
	if first.X != 10 || second.X != 18 || first.Y != second.Y {
		t.Fatalf("direct player spawns = (%v,%v), (%v,%v)", first.X, first.Y, second.X, second.Y)
	}
}

func TestRemoteProfileAdmissionIsUnavailableToRealmHosts(t *testing.T) {
	tickets, _ := gameserver.NewTicketAuthority([]byte("0123456789abcdef0123456789abcdef"), "game")
	_, err := NewRemoteProfileAdmissions(&gameserver.Host{Mode: gameserver.ModeRealm, Session: &gamesession.Session{}}, tickets, RemoteProfileConfig{Credential: "secret", PrincipalID: "user", PlayerID: "player", Lifetime: time.Minute})
	if !errors.Is(err, ErrRemoteProfileAdmission) {
		t.Fatalf("realm error = %v", err)
	}
}

func TestRemoteProfileAdmissionThrottleIsPerClientAndRefills(t *testing.T) {
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close(); _ = engine.Close() })
	if err := session.Register(playeradapter.EnterCommand, gamesession.CommandHandler{Validate: func(simulation.Command) error { return nil }, Apply: func(*gameecs.Engine, simulation.Command) error { return nil }, Allowed: []simulation.Authority{simulation.AuthoritySystem}}); err != nil {
		t.Fatal(err)
	}
	identity := remoteProfileIdentity()
	allocation, _ := gamesession.Allocate("game", identity, gamesession.PredictionLimited)
	host := &gameserver.Host{Mode: gameserver.ModeStandalone, Engine: engine, Session: session, Allocation: allocation}
	tickets, _ := gameserver.NewTicketAuthority([]byte("0123456789abcdef0123456789abcdef"), "game")
	destination, _ := playeradapter.NewDestination(10, 20, 1000, 100, 1, 40)
	admissions, err := NewRemoteProfileAdmissions(host, tickets, RemoteProfileConfig{Credential: "secret", PrincipalID: "principal", PlayerID: "player", Destination: destination, Lifetime: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(100, 0)
	admissions.now = func() time.Time { return now }
	offer, _ := d2save.EncodeCharacterOffer(d2save.Character{ID: "hero", Name: "Hero", Class: "Amazon"})
	clientA := WithProfileAdmissionClient(context.Background(), "192.0.2.1:6112")
	for range int(remoteProfileBurst) {
		if _, err := admissions.Admit(clientA, "wrong", offer); !errors.Is(err, ErrRemoteProfileAdmission) {
			t.Fatalf("wrong credential error = %v", err)
		}
	}
	if _, err := admissions.Admit(clientA, "secret", offer); !errors.Is(err, ErrRemoteProfileAdmission) {
		t.Fatalf("exhausted client error = %v", err)
	}
	clientB := WithProfileAdmissionClient(context.Background(), "192.0.2.2:6112")
	if _, err := admissions.Admit(clientB, "secret", offer); err != nil {
		t.Fatalf("independent client was throttled: %v", err)
	}
	now = now.Add(time.Second)
	if _, err := admissions.Admit(clientA, "secret", offer); err != nil {
		t.Fatalf("refilled client error = %v", err)
	}
}

func remoteProfileIdentity() simulation.RuntimeIdentity {
	const packageDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	return simulation.RuntimeIdentity{Recipe: simulation.RuntimeRecipe{
		Schema: simulation.RuntimeRecipeSchema, EngineAPI: "v1", NetworkProtocol: "test/v1", AssetSetID: simulation.EmptyAssetSetID,
		Packages:          simulation.RuntimePackageSet{Base: simulation.RuntimePackage{ID: "d2legacy", Version: "1.0.0", Digest: packageDigest, Size: 1, Redistributable: true}},
		AuthoritativeHash: "rules", ConfigurationHash: "config",
	}}
}
