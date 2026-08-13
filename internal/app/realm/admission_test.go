package realm

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

func TestAdmissionsLeasesValidatesEntersAndIssuesTicket(t *testing.T) {
	manager, host, identity := admissionFixture(t, func(simulation.Command) error { return nil })
	record := CharacterRecord{AccountID: "account", Revision: 4, Character: d2save.Character{ID: "character", Name: "Hero", Class: "Amazon", Level: 1, Stats: &d2save.Stats{Health: 20, MaxHealth: 20}}, Compatibility: host.Allocation.Durable("character")}
	repository, err := NewMemoryCharacters(record)
	if err != nil {
		t.Fatal(err)
	}
	authority, err := gameserver.NewTicketAuthority([]byte("0123456789abcdef0123456789abcdef"), "game")
	if err != nil {
		t.Fatal(err)
	}
	admissions, err := NewAdmissions(manager, repository, time.Minute, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := admissions.RegisterGame("game", authority, GameEndpoint{Address: "game.example:4433", TLSFingerprint: "sha256:cert"}); err != nil {
		t.Fatal(err)
	}
	destination, err := playeradapter.NewDestination(10, 20, 100, 100, 1, 40)
	if err != nil {
		t.Fatal(err)
	}
	assignment, err := admissions.Join(context.Background(), JoinRequest{AccountID: "account", CharacterID: "character", PlayerID: "player", GameID: "game", Destination: destination})
	if err != nil {
		t.Fatal(err)
	}
	if assignment.Endpoint.Address != "game.example:4433" || assignment.Runtime.ModID != identity.ModID || assignment.Lease.Revision != 4 {
		t.Fatalf("assignment = %#v", assignment)
	}
	principal, err := authority.Authenticate(context.Background(), assignment.Ticket)
	if err != nil || principal.PlayerID != "player" || principal.CharacterRevision != 4 || principal.RuntimeIdentityHash != host.Allocation.IdentityHash {
		t.Fatalf("principal=%#v error=%v", principal, err)
	}
	if _, _, err := repository.Acquire(context.Background(), "account", "character", "other", time.Minute); !errors.Is(err, ErrCharacterLeased) {
		t.Fatalf("lease error = %v", err)
	}
	if _, err := admissions.RenewMembership(context.Background(), "game", "player"); err != nil {
		t.Fatal(err)
	}
	if err := host.Session.Step(); err != nil {
		t.Fatal(err)
	}
	replay, err := host.Session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	if len(replay.Commands) != 1 || replay.Commands[0].Kind != playeradapter.EnterCommand || replay.Commands[0].Authority != simulation.AuthoritySystem {
		t.Fatalf("commands = %#v", replay.Commands)
	}
}

func TestAdmissionsRollsBackLeaseAndTicketWhenEntryRejected(t *testing.T) {
	manager, host, _ := admissionFixture(t, func(simulation.Command) error { return errors.New("rejected") })
	repository, err := NewMemoryCharacters(CharacterRecord{AccountID: "account", Revision: 1, Character: d2save.Character{ID: "character", Name: "Hero", Class: "Amazon"}, Compatibility: host.Allocation.Durable("character")})
	if err != nil {
		t.Fatal(err)
	}
	authority, err := gameserver.NewTicketAuthority([]byte("0123456789abcdef0123456789abcdef"), "game")
	if err != nil {
		t.Fatal(err)
	}
	admissions, err := NewAdmissions(manager, repository, time.Minute, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := admissions.RegisterGame("game", authority, GameEndpoint{Address: "server", TLSFingerprint: "cert"}); err != nil {
		t.Fatal(err)
	}
	destination, _ := playeradapter.NewDestination(1, 1, 10, 10, 1, 1)
	if _, err := admissions.Join(context.Background(), JoinRequest{AccountID: "account", CharacterID: "character", PlayerID: "player", GameID: "game", Destination: destination}); !errors.Is(err, ErrAdmission) {
		t.Fatalf("join error = %v", err)
	}
	if _, _, err := repository.Acquire(context.Background(), "account", "character", "other", time.Minute); err != nil {
		t.Fatalf("lease was not released: %v", err)
	}
}

func admissionFixture(t *testing.T, validate simulation.CommandValidator) (*Manager, *gameserver.Host, simulation.RuntimeIdentity) {
	t.Helper()
	manager, err := NewManager(fstest.MapFS{"boot.lua": {}}, fixtureRecords{})
	if err != nil {
		t.Fatal(err)
	}
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close(); _ = engine.Close() })
	if err := session.Register(playeradapter.EnterCommand, gamesession.CommandHandler{Validate: validate, Apply: func(*gameecs.Engine, simulation.Command) error { return nil }, Allowed: []simulation.Authority{simulation.AuthoritySystem}}); err != nil {
		t.Fatal(err)
	}
	identity := simulation.RuntimeIdentity{ModID: "d2legacy", ContractVersion: "v1", PackageHash: "package", AuthoritativeHash: "rules", ConfigurationHash: "config"}
	allocation, err := gamesession.Allocate("game", identity, gamesession.PredictionLimited)
	if err != nil {
		t.Fatal(err)
	}
	host := &gameserver.Host{Engine: engine, Session: session, Allocation: allocation}
	manager.hosts["game"] = host
	return manager, host, identity
}
