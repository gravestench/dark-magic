package realm

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gravestench/akara"
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
	if assignment.Endpoint.Address != "game.example:4433" || assignment.Runtime.Recipe.Packages.Base.ID != identity.Recipe.Packages.Base.ID ||
		len(assignment.Runtime.Recipe.Packages.Extensions) != 1 || assignment.Runtime.Recipe.Packages.Extensions[0].ID != "realm_extension" || assignment.CharacterRevision != 4 {
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
	installCanonicalCharacter(t, host.Engine, "character", "player")
	committed, err := admissions.CommitCanonicalMembership(context.Background(), "game", "player")
	if err != nil {
		t.Fatal(err)
	}
	if committed.Revision != 5 || committed.Character.Name != "Saved Hero" || committed.Character.Level != 2 || committed.Character.Stats.Health != 18 {
		t.Fatalf("committed character = %#v", committed)
	}
	if _, err := admissions.CommitMembership(context.Background(), "game", "player", committed.Character); !errors.Is(err, ErrLease) {
		t.Fatalf("replayed membership commit error = %v", err)
	}
}

func installCanonicalCharacter(t *testing.T, engine *gameecs.Engine, characterID, playerID string) {
	t.Helper()
	register := func(name string, fields []akara.Field) *akara.DynamicStore {
		store, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: name, Version: 1, Fields: fields})
		if err != nil {
			t.Fatal(err)
		}
		return store
	}
	identity := register("d2legacy.player.identity", []akara.Field{{Name: "character_id", Kind: akara.FieldString}, {Name: "player", Kind: akara.FieldString}, {Name: "name", Kind: akara.FieldString}, {Name: "class", Kind: akara.FieldString}})
	vitals := register("d2legacy.player.vitals", []akara.Field{{Name: "health", Kind: akara.FieldInt64}, {Name: "max_health", Kind: akara.FieldInt64}, {Name: "mana", Kind: akara.FieldInt64}, {Name: "max_mana", Kind: akara.FieldInt64}})
	progress := register("d2legacy.player.progress", []akara.Field{{Name: "level", Kind: akara.FieldInt64}, {Name: "experience", Kind: akara.FieldInt64}, {Name: "unspent_skill_points", Kind: akara.FieldInt64}})
	combat := register("d2legacy.player.combat_stats", []akara.Field{{Name: "attack_rating", Kind: akara.FieldInt64}, {Name: "defense", Kind: akara.FieldInt64}})
	position := register("d2legacy.world.position", []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}})
	location := register("d2legacy.world.location", []akara.Field{{Name: "act", Kind: akara.FieldInt64}, {Name: "level_id", Kind: akara.FieldInt64}})
	entity := engine.World().MustCreateEntity()
	for store, values := range map[*akara.DynamicStore]map[string]any{
		identity: {"character_id": characterID, "player": playerID, "name": "Saved Hero", "class": "Amazon"},
		vitals:   {"health": int64(18), "max_health": int64(20), "mana": int64(8), "max_mana": int64(10)},
		progress: {"level": int64(2), "experience": int64(100), "unspent_skill_points": int64(0)},
		combat:   {"attack_rating": int64(12), "defense": int64(4)}, position: {"x": 1.0, "y": 1.0},
		location: {"act": int64(1), "level_id": int64(1)},
	} {
		if _, err := store.Set(entity, values); err != nil {
			t.Fatal(err)
		}
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
	identity := simulation.RuntimeIdentity{Recipe: simulation.RuntimeRecipe{
		Schema: simulation.RuntimeRecipeSchema, EngineAPI: "v1", NetworkProtocol: "test/v1",
		Packages:          simulation.RuntimePackageSet{Base: simulation.RuntimePackage{ID: "d2legacy", Version: "1.0.0", Digest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Size: 1, Redistributable: true}},
		AuthoritativeHash: "rules", ConfigurationHash: "config",
	}}
	identity.Recipe.Packages.Extensions = []simulation.RuntimePackage{{ID: "realm_extension", Version: "1.0.0", Digest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Size: 2, Redistributable: true}}
	allocation, err := gamesession.Allocate("game", identity, gamesession.PredictionLimited)
	if err != nil {
		t.Fatal(err)
	}
	host := &gameserver.Host{Engine: engine, Session: session, Allocation: allocation}
	manager.hosts["game"] = host
	return manager, host, identity
}
