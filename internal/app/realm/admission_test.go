package realm

import (
	"context"
	"errors"
	"fmt"
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

// TestAdmissionsLeasesValidatesEntersAndIssuesTicket verifies admissions leases validates enters and issues ticket.
// The scenario keeps the admission contract visible to maintainers.
func TestAdmissionsLeasesValidatesEntersAndIssuesTicket(t *testing.T) {
	manager, host, identity := admissionFixture(t, func(simulation.Command) error { return nil })
	record := CharacterRecord{
		AccountID: "account",
		Revision:  4,
		Character: d2save.Character{
			ID:    "character",
			Name:  "Hero",
			Class: "Amazon",
			Level: 1,
			Stats: &d2save.Stats{Health: 20, MaxHealth: 20},
		},
		Compatibility: host.Allocation.Durable("character"),
	}

	repository, err := NewMemoryCharacters(record)
	if err != nil {
		t.Fatal(err)
	}

	authority, err := gameserver.NewTicketAuthority([]byte("0123456789abcdef0123456789abcdef"), "game")
	if err != nil {
		t.Fatal(err)
	}

	tickets, err := newLocalTicketIssuer(authority)
	if err != nil {
		t.Fatal(err)
	}

	admissions, err := NewAdmissions(manager, repository, time.Minute, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if err := admissions.RegisterGame(
		"game",
		tickets,
		GameEndpoint{Address: "game.example:4433", TLSFingerprint: "sha256:cert"},
	); err != nil {
		t.Fatal(err)
	}

	destination, err := playeradapter.NewDestination(10, 20, 100, 100, 1, 40)
	if err != nil {
		t.Fatal(err)
	}

	assignment, err := admissions.Join(
		context.Background(),
		JoinRequest{
			AccountID:   "account",
			CharacterID: "character",
			PlayerID:    "player",
			GameID:      "game",
			Destination: destination,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if assignment.Endpoint.Address != "game.example:4433" ||
		assignment.Runtime.Recipe.Packages.Base.ID != identity.Recipe.Packages.Base.ID ||
		len(assignment.Runtime.Recipe.Packages.Extensions) != 1 ||
		assignment.Runtime.Recipe.Packages.Extensions[0].ID != "realm_extension" ||
		assignment.CharacterRevision != 4 {
		t.Fatalf("assignment = %#v", assignment)
	}

	principal, err := authority.Authenticate(context.Background(), assignment.Ticket)
	if err != nil || principal.PlayerID != "player" || principal.CharacterRevision != 4 ||
		principal.RuntimeIdentityHash != host.Allocation.IdentityHash {
		t.Fatalf("principal=%#v error=%v", principal, err)
	}

	if _, _, err := repository.Acquire(
		context.Background(),
		"account",
		"character",
		"other",
		time.Minute,
	); !errors.Is(
		err,
		ErrCharacterLeased,
	) {
		t.Fatalf("lease error = %v", err)
	}

	renewed, err := admissions.RenewMembership(context.Background(), "game", "player")
	if err != nil {
		t.Fatal(err)
	}

	now := renewed.ExpiresAt.Add(-admissions.leaseLifetime/2 + time.Millisecond)
	admissions.now = func() time.Time { return now }
	repository.now = func() time.Time { return now }

	if count, err := admissions.RenewGameMemberships(context.Background(), "game"); err != nil || count != 1 {
		t.Fatalf("renew game memberships = %d, %v", count, err)
	}

	admissions.mu.RLock()
	periodicRenewal := admissions.memberships["game\x00player"].lease
	admissions.mu.RUnlock()

	if !periodicRenewal.ExpiresAt.After(renewed.ExpiresAt) {
		t.Fatalf("periodic renewal expiration = %s, want after %s", periodicRenewal.ExpiresAt, renewed.ExpiresAt)
	}

	if err := host.Session.Step(); err != nil {
		t.Fatal(err)
	}

	replay, err := host.Session.Replay()
	if err != nil {
		t.Fatal(err)
	}

	if len(replay.Commands) != 1 || replay.Commands[0].Kind != playeradapter.EnterCommand ||
		replay.Commands[0].Authority != simulation.AuthoritySystem {
		t.Fatalf("commands = %#v", replay.Commands)
	}

	installCanonicalCharacter(t, host.Engine, "character", "player")

	committed, err := admissions.CommitCanonicalMembership(context.Background(), "game", "player")
	if err != nil {
		t.Fatal(err)
	}

	if committed.Revision != 5 || committed.Character.Name != "Saved Hero" || committed.Character.Level != 2 ||
		committed.Character.Stats.Health != 18 {
		t.Fatalf("committed character = %#v", committed)
	}

	if _, err := admissions.CommitMembership(
		context.Background(),
		"game",
		"player",
		committed.Character,
	); !errors.Is(
		err,
		ErrLease,
	) {
		t.Fatalf("replayed membership commit error = %v", err)
	}
}

// TestAdmissionsResumeGameRehydratesDurableMemberships verifies admissions resume game rehydrates durable memberships.
// The scenario keeps the admission contract visible to maintainers.
func TestAdmissionsResumeGameRehydratesDurableMemberships(t *testing.T) {
	manager, host, _ := admissionFixture(t, func(simulation.Command) error { return nil })
	record := CharacterRecord{AccountID: "account", Revision: 4,
		Character:     d2save.Character{ID: "character", Name: "Hero", Class: "Amazon", Level: 1},
		Compatibility: host.Allocation.Durable("character")}

	characters, err := NewMemoryCharacters(record)
	if err != nil {
		t.Fatal(err)
	}

	baseline, lease, err := characters.Acquire(t.Context(), "account", "character", "game", time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	memberships, err := NewMemoryMemberships(characters)
	if err != nil {
		t.Fatal(err)
	}

	if err := memberships.Admit(t.Context(), MembershipRecord{GameID: "game", PlayerID: "player",
		AccountID: "account", Baseline: baseline, Lease: lease, State: MembershipActive}); err != nil {
		t.Fatal(err)
	}

	authority, err := gameserver.NewTicketAuthority([]byte("0123456789abcdef0123456789abcdef"), "game")
	if err != nil {
		t.Fatal(err)
	}

	tickets, err := newLocalTicketIssuer(authority)
	if err != nil {
		t.Fatal(err)
	}

	admissions, err := NewAdmissionsWithMemberships(manager, characters, memberships, 2*time.Minute, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	resumed, err := admissions.ResumeGame(t.Context(), "game", tickets,
		GameEndpoint{Address: "game.example:4433", TLSFingerprint: "sha256:cert"})
	if err != nil {
		t.Fatal(err)
	}

	if len(resumed) != 1 || resumed[0].Lease.Token == "" || !resumed[0].Lease.ExpiresAt.After(lease.ExpiresAt) {
		t.Fatalf("resumed memberships = %#v", resumed)
	}

	playerID, recovered, err := admissions.CharacterMembership("game", "account", "character")
	if err != nil || playerID != "player" || recovered.Character.ID != "character" {
		t.Fatalf("recovered membership player=%q record=%#v error=%v", playerID, recovered, err)
	}

	if err := admissions.RegisterGame("game", tickets,
		GameEndpoint{Address: "game.example:4433", TLSFingerprint: "sha256:cert"}); !errors.Is(err, ErrGameExists) {
		t.Fatalf("duplicate resumed game error = %v", err)
	}

	assignment, err := admissions.ReconnectAssignment(t.Context(), "game", "account", "character")
	if err != nil {
		t.Fatal(err)
	}

	principal, err := authority.Authenticate(t.Context(), assignment.Ticket)
	if err != nil || assignment.Endpoint.Address != "game.example:4433" || principal.PlayerID != "player" ||
		principal.CharacterID != "character" || principal.CharacterRevision != 4 {
		t.Fatalf("reconnect assignment=%#v principal=%#v error=%v", assignment, principal, err)
	}
}

// TestAdmissionsReconnectSameAccountCharactersIndependently verifies admissions reconnect same account characters
// independently. The scenario keeps the admission contract visible to maintainers.
func TestAdmissionsReconnectSameAccountCharactersIndependently(t *testing.T) {
	manager, host, _ := admissionFixture(t, func(simulation.Command) error { return nil })
	records := []CharacterRecord{
		{
			AccountID:     "account",
			Revision:      2,
			Character:     d2save.Character{ID: "first-character", Name: "First", Class: "Amazon", Level: 1},
			Compatibility: host.Allocation.Durable("first-character"),
		},
		{
			AccountID:     "account",
			Revision:      3,
			Character:     d2save.Character{ID: "second-character", Name: "Second", Class: "Barbarian", Level: 1},
			Compatibility: host.Allocation.Durable("second-character"),
		},
	}

	characters, err := NewMemoryCharacters(records...)
	if err != nil {
		t.Fatal(err)
	}

	memberships, err := NewMemoryMemberships(characters)
	if err != nil {
		t.Fatal(err)
	}

	for index, record := range records {
		baseline, lease, err := characters.Acquire(t.Context(), record.AccountID, record.Character.ID, "game", time.Minute)
		if err != nil {
			t.Fatal(err)
		}

		if err := memberships.Admit(t.Context(), MembershipRecord{
			GameID: "game", PlayerID: fmt.Sprintf("player-%d", index+1), AccountID: record.AccountID,
			Baseline: baseline, Lease: lease, State: MembershipActive,
		}); err != nil {
			t.Fatal(err)
		}
	}

	authority, err := gameserver.NewTicketAuthority([]byte("0123456789abcdef0123456789abcdef"), "game")
	if err != nil {
		t.Fatal(err)
	}

	tickets, err := newLocalTicketIssuer(authority)
	if err != nil {
		t.Fatal(err)
	}

	admissions, err := NewAdmissionsWithMemberships(manager, characters, memberships, 2*time.Minute, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := admissions.ResumeGame(t.Context(), "game", tickets,
		GameEndpoint{Address: "game.example:4433", TLSFingerprint: "sha256:cert"}); err != nil {
		t.Fatal(err)
	}

	for index, record := range records {
		assignment, err := admissions.ReconnectAssignment(t.Context(), "game", record.AccountID, record.Character.ID)
		if err != nil {
			t.Fatal(err)
		}

		principal, err := authority.Authenticate(t.Context(), assignment.Ticket)
		if err != nil || principal.PlayerID != fmt.Sprintf("player-%d", index+1) ||
			principal.CharacterID != record.Character.ID || principal.CharacterRevision != record.Revision {
			t.Fatalf("character %q reconnect assignment=%#v principal=%#v error=%v",
				record.Character.ID, assignment, principal, err)
		}
	}
}

// installCanonicalCharacter builds the shared install canonical character fixture so each test starts from the same
// explicit ownership and compatibility state.
func installCanonicalCharacter(t *testing.T, engine *gameecs.Engine, characterID, playerID string) {
	t.Helper()

	register := func(name string, fields []akara.Field) *akara.DynamicStore {
		store, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: name, Version: 1, Fields: fields})
		if err != nil {
			t.Fatal(err)
		}

		return store
	}
	identity := register(
		"d2legacy.player.identity",
		[]akara.Field{
			{Name: "character_id", Kind: akara.FieldString},
			{Name: "player", Kind: akara.FieldString},
			{Name: "name", Kind: akara.FieldString},
			{Name: "class", Kind: akara.FieldString},
		},
	)
	vitals := register(
		"d2legacy.player.vitals",
		[]akara.Field{
			{Name: "health", Kind: akara.FieldInt64},
			{Name: "max_health", Kind: akara.FieldInt64},
			{Name: "mana", Kind: akara.FieldInt64},
			{Name: "max_mana", Kind: akara.FieldInt64},
		},
	)
	progress := register(
		"d2legacy.player.progress",
		[]akara.Field{
			{Name: "level", Kind: akara.FieldInt64},
			{Name: "experience", Kind: akara.FieldInt64},
			{Name: "unspent_skill_points", Kind: akara.FieldInt64},
		},
	)
	combat := register(
		"d2legacy.player.combat_stats",
		[]akara.Field{{Name: "attack_rating", Kind: akara.FieldInt64}, {Name: "defense", Kind: akara.FieldInt64}},
	)
	position := register(
		"d2legacy.world.position",
		[]akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}},
	)
	location := register(
		"d2legacy.world.location",
		[]akara.Field{{Name: "act", Kind: akara.FieldInt64}, {Name: "level_id", Kind: akara.FieldInt64}},
	)

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

// TestAdmissionsRollsBackLeaseAndTicketWhenEntryRejected verifies admissions rolls back lease and ticket when entry
// rejected. The scenario keeps the admission contract visible to maintainers.
func TestAdmissionsRollsBackLeaseAndTicketWhenEntryRejected(t *testing.T) {
	manager, host, _ := admissionFixture(t, func(simulation.Command) error { return errors.New("rejected") })

	repository, err := NewMemoryCharacters(
		CharacterRecord{
			AccountID:     "account",
			Revision:      1,
			Character:     d2save.Character{ID: "character", Name: "Hero", Class: "Amazon"},
			Compatibility: host.Allocation.Durable("character"),
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	authority, err := gameserver.NewTicketAuthority([]byte("0123456789abcdef0123456789abcdef"), "game")
	if err != nil {
		t.Fatal(err)
	}

	tickets, err := newLocalTicketIssuer(authority)
	if err != nil {
		t.Fatal(err)
	}

	admissions, err := NewAdmissions(manager, repository, time.Minute, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if err := admissions.RegisterGame(
		"game",
		tickets,
		GameEndpoint{Address: "server", TLSFingerprint: "cert"},
	); err != nil {
		t.Fatal(err)
	}

	destination, _ := playeradapter.NewDestination(1, 1, 10, 10, 1, 1)
	if _, err := admissions.Join(
		context.Background(),
		JoinRequest{
			AccountID:   "account",
			CharacterID: "character",
			PlayerID:    "player",
			GameID:      "game",
			Destination: destination,
		},
	); !errors.Is(
		err,
		ErrAdmission,
	) {
		t.Fatalf("join error = %v", err)
	}

	if _, _, err := repository.Acquire(context.Background(), "account", "character", "other", time.Minute); err != nil {
		t.Fatalf("lease was not released: %v", err)
	}
}

// TestAdmissionsRollBackWorkerAndLeaseWhenMembershipPersistenceFails verifies admissions roll back worker and lease
// when membership persistence fails. The scenario keeps the admission contract visible to maintainers.
func TestAdmissionsRollBackWorkerAndLeaseWhenMembershipPersistenceFails(t *testing.T) {
	manager, host, _ := admissionFixture(t, func(simulation.Command) error { return nil })
	record := CharacterRecord{AccountID: "account", Revision: 1,
		Character:     d2save.Character{ID: "character", Name: "Hero", Class: "Amazon"},
		Compatibility: host.Allocation.Durable("character")}

	characters, err := NewMemoryCharacters(record)
	if err != nil {
		t.Fatal(err)
	}

	memberships, err := NewMemoryMemberships(characters)
	if err != nil {
		t.Fatal(err)
	}

	store := &rejectingMembershipStore{MembershipRepository: memberships, err: errors.New("database unavailable")}

	authority, err := gameserver.NewTicketAuthority([]byte("0123456789abcdef0123456789abcdef"), "game")
	if err != nil {
		t.Fatal(err)
	}

	tickets, err := newLocalTicketIssuer(authority)
	if err != nil {
		t.Fatal(err)
	}

	admissions, err := NewAdmissionsWithMemberships(manager, characters, store, time.Minute, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if err := admissions.RegisterGame("game", tickets,
		GameEndpoint{Address: "game.example:4433", TLSFingerprint: "sha256:cert"}); err != nil {
		t.Fatal(err)
	}

	if _, err := admissions.Join(t.Context(), JoinRequest{AccountID: "account", CharacterID: "character",
		PlayerID: "player", GameID: "game"}); err == nil {
		t.Fatal("membership persistence failure admitted a player")
	}

	if _, err := memberships.ByPlayer(t.Context(), "game", "player"); !errors.Is(err, ErrMembership) {
		t.Fatalf("failed admission persisted membership: %v", err)
	}

	_, lease, err := characters.Acquire(t.Context(), "account", "character", "replacement", time.Minute)
	if err != nil {
		t.Fatalf("failed admission leaked character lease: %v", err)
	}

	if err := characters.Release(t.Context(), lease); err != nil {
		t.Fatal(err)
	}
}

type rejectingMembershipStore struct {
	MembershipRepository
	err error
}

// Admit supplies the test double's admit behavior, keeping the scenario deterministic and independent of external
// services.
func (store *rejectingMembershipStore) Admit(context.Context, MembershipRecord) error {
	return store.err
}

// TestAdmissionsRejectsInconsistentWorkerIdentityAndReleasesLease verifies admissions rejects inconsistent worker
// identity and releases lease. The scenario keeps the admission contract visible to maintainers.
func TestAdmissionsRejectsInconsistentWorkerIdentityAndReleasesLease(t *testing.T) {
	manager, host, identity := admissionFixture(t, func(simulation.Command) error { return nil })

	worker, found := manager.Game("game")
	if !found {
		t.Fatal("fixture worker is missing")
	}

	manager.workers["game"] = descriptionWorker{WorkerClient: worker,
		description: WorkerDescription{Runtime: identity, IdentityHash: "inconsistent"}}

	repository, err := NewMemoryCharacters(CharacterRecord{AccountID: "account", Revision: 1,
		Character:     d2save.Character{ID: "character", Name: "Hero", Class: "Amazon"},
		Compatibility: host.Allocation.Durable("character")})
	if err != nil {
		t.Fatal(err)
	}

	authority, err := gameserver.NewTicketAuthority([]byte("0123456789abcdef0123456789abcdef"), "game")
	if err != nil {
		t.Fatal(err)
	}

	tickets, err := newLocalTicketIssuer(authority)
	if err != nil {
		t.Fatal(err)
	}

	admissions, err := NewAdmissions(manager, repository, time.Minute, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if err := admissions.RegisterGame(
		"game",
		tickets,
		GameEndpoint{Address: "server", TLSFingerprint: "cert"},
	); err != nil {
		t.Fatal(err)
	}

	destination, _ := playeradapter.NewDestination(1, 1, 10, 10, 1, 1)
	if _, err := admissions.Join(t.Context(), JoinRequest{AccountID: "account", CharacterID: "character",
		PlayerID: "player", GameID: "game", Destination: destination}); !errors.Is(err, ErrAdmission) {
		t.Fatalf("join error = %v", err)
	}

	if _, _, err := repository.Acquire(t.Context(), "account", "character", "other", time.Minute); err != nil {
		t.Fatalf("lease was not released: %v", err)
	}
}

// TestAdmissionsCancellationStillReleasesLease verifies admissions cancellation still releases lease. The scenario
// keeps the admission contract visible to maintainers.
func TestAdmissionsCancellationStillReleasesLease(t *testing.T) {
	manager, host, _ := admissionFixture(t, func(simulation.Command) error { return nil })

	repository, err := NewMemoryCharacters(CharacterRecord{AccountID: "account", Revision: 1,
		Character:     d2save.Character{ID: "character", Name: "Hero", Class: "Amazon"},
		Compatibility: host.Allocation.Durable("character")})
	if err != nil {
		t.Fatal(err)
	}

	authority, err := gameserver.NewTicketAuthority([]byte("0123456789abcdef0123456789abcdef"), "game")
	if err != nil {
		t.Fatal(err)
	}

	tickets, err := newLocalTicketIssuer(authority)
	if err != nil {
		t.Fatal(err)
	}

	admissions, err := NewAdmissions(manager, repository, time.Minute, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if err := admissions.RegisterGame(
		"game",
		tickets,
		GameEndpoint{Address: "server", TLSFingerprint: "cert"},
	); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	destination, _ := playeradapter.NewDestination(1, 1, 10, 10, 1, 1)
	if _, err := admissions.Join(ctx, JoinRequest{AccountID: "account", CharacterID: "character",
		PlayerID: "player", GameID: "game", Destination: destination}); !errors.Is(err, ErrAdmission) {
		t.Fatalf("join error = %v", err)
	}

	if _, _, err := repository.Acquire(t.Context(), "account", "character", "other", time.Minute); err != nil {
		t.Fatalf("canceled join leaked its lease: %v", err)
	}
}

type descriptionWorker struct {
	WorkerClient
	description WorkerDescription
}

// Describe supplies the test double's describe behavior, keeping the scenario deterministic and independent of
// external services.
func (worker descriptionWorker) Describe(context.Context) (WorkerDescription, error) {
	return worker.description, nil
}

// admissionFixture builds the shared admission fixture fixture so each test starts from the same explicit ownership
// and compatibility state.
func admissionFixture(
	t *testing.T,
	validate simulation.CommandValidator,
) (*Manager, *gameserver.Host, simulation.RuntimeIdentity) {
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

	if err := session.Register(
		playeradapter.EnterCommand,
		gamesession.CommandHandler{
			Validate: validate,
			Apply:    func(*gameecs.Engine, simulation.Command) error { return nil },
			Allowed:  []simulation.Authority{simulation.AuthoritySystem},
		},
	); err != nil {
		t.Fatal(err)
	}

	identity := simulation.RuntimeIdentity{Recipe: simulation.RuntimeRecipe{
		Schema:               simulation.RuntimeRecipeSchema,
		EngineAPI:            "v1",
		NetworkProtocol:      "test/v1",
		AssetSetID:           simulation.EmptyAssetSetID,
		GameDataGenerationID: simulation.GameDataGenerationIDForAssetSet(simulation.EmptyAssetSetID),
		Packages: simulation.RuntimePackageSet{
			Base: simulation.RuntimePackage{
				ID:              "d2legacy",
				Version:         "1.0.0",
				Digest:          "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Size:            1,
				Redistributable: true,
			},
		},
		AuthoritativeHash: "rules",
		ConfigurationHash: "config",
	}}
	identity.Recipe.Packages.Extensions = []simulation.RuntimePackage{
		{
			ID:              "realm_extension",
			Version:         "1.0.0",
			Digest:          "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			Size:            2,
			Redistributable: true,
		},
	}

	allocation, err := gamesession.Allocate("game", identity, gamesession.PredictionLimited)
	if err != nil {
		t.Fatal(err)
	}

	host := &gameserver.Host{Engine: engine, Session: session, Allocation: allocation}

	worker, err := newInProcessWorker(host)
	if err != nil {
		t.Fatal(err)
	}

	manager.workers["game"] = worker

	return manager, host, identity
}
