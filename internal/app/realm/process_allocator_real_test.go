package realm_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/app/clientsession"
	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
	"github.com/gravestench/dark-magic/internal/app/networktrust"
	"github.com/gravestench/dark-magic/internal/app/realm"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// This opt-in acceptance crosses the real process composition root and real
// d2legacy data. Ordinary CI uses the deterministic helper process because it
// cannot redistribute the required game assets.
func TestProcessAllocatorStartsRealPreparedRealmWorker(t *testing.T) {
	executable := os.Getenv("DARK_MAGIC_REALM_WORKER_ACCEPTANCE")
	if executable == "" || os.Getenv("MPQ_DIRECTORY") == "" {
		t.Skip("DARK_MAGIC_REALM_WORKER_ACCEPTANCE and MPQ_DIRECTORY are required")
	}

	allocator, err := realm.NewProcessAllocator(realm.ProcessAllocatorConfig{
		Executable: executable, Arguments: []string{"--log-level", "error"}, StateDirectory: t.TempDir(),
		ControlListenAddress: "127.0.0.1:0", GameListenAddress: "127.0.0.1:0",
		StartupTimeout: 30 * time.Second, ShutdownTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := allocator.Close(ctx); err != nil {
			t.Errorf("close allocator: %v", err)
		}
	})

	allocation, err := allocator.Allocate(
		t.Context(),
		realm.GameSpec{GameID: "real-worker-acceptance", AllocationID: "real-worker-allocation"},
	)
	if err != nil {
		t.Fatal(err)
	}

	description, err := allocation.Worker.Describe(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	if err := description.Runtime.Validate(); err != nil {
		t.Fatalf("worker runtime identity: %v", err)
	}

	if _, err := playeradapter.NewDestination(description.EntryDestination.X, description.EntryDestination.Y,
		description.EntryDestination.Width, description.EntryDestination.Height, description.EntryDestination.Act,
		description.EntryDestination.LevelID); err != nil {
		t.Fatalf("worker entry destination = %#v: %v", description.EntryDestination, err)
	}

	if allocation.Endpoint.Address == "" || allocation.Endpoint.TLSFingerprint == "" {
		t.Fatalf("worker endpoint = %#v", allocation.Endpoint)
	}

	status, err := allocation.Worker.Status(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	if !status.Ready || status.ActivePlayers != 0 {
		t.Fatalf("worker status = %#v", status)
	}
}

// TestProcessAllocatorRealWorkerAdmitsTwoCharactersFromSameAccount verifies process allocator real worker admits two
// characters from same account. The scenario keeps the process allocator real contract visible to maintainers.
func TestProcessAllocatorRealWorkerAdmitsTwoCharactersFromSameAccount(t *testing.T) {
	executable := os.Getenv("DARK_MAGIC_REALM_WORKER_ACCEPTANCE")
	if executable == "" || os.Getenv("MPQ_DIRECTORY") == "" {
		t.Skip("DARK_MAGIC_REALM_WORKER_ACCEPTANCE and MPQ_DIRECTORY are required")
	}

	allocator, err := realm.NewProcessAllocator(realm.ProcessAllocatorConfig{
		Executable: executable, Arguments: []string{"--log-level", "debug"}, StateDirectory: t.TempDir(),
		ControlListenAddress: "127.0.0.1:0", GameListenAddress: "127.0.0.1:0",
		StartupTimeout: 30 * time.Second, ShutdownTimeout: 5 * time.Second, LogWriter: os.Stderr,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_ = allocator.Close(ctx)
	})

	control, err := realm.NewControlPlane(realm.ControlPlaneConfig{Allocator: allocator})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := control.CreateAccount(t.Context(), "MuleAccount", "long enough password"); err != nil {
		t.Fatal(err)
	}

	creator, err := control.Authenticate(t.Context(), "MuleAccount", "long enough password")
	if err != nil {
		t.Fatal(err)
	}

	joiner, err := control.Authenticate(t.Context(), "MuleAccount", "long enough password")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := control.CreateCharacter(t.Context(), creator.Token, realm.CreateCharacterRequest{
		Name: "Mulehost", Class: "Necromancer", Expansion: true,
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := control.CreateCharacter(t.Context(), joiner.Token, realm.CreateCharacterRequest{
		Name: "Mulejoin", Class: "Amazon", Expansion: true,
	}); err != nil {
		t.Fatal(err)
	}

	created, err := control.CreateGame(t.Context(), creator.Token, realm.CreateGameRequest{
		Name: "Mule Game", Difficulty: realm.DifficultyNormal, Maximum: 8, Expansion: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	joined, err := control.JoinGame(t.Context(), joiner.Token, created.Game.Entry.GameID, "")
	if err != nil {
		t.Fatal(err)
	}

	if joined.Game.Entry.Players != 2 || len(joined.Game.Players) != 2 {
		t.Fatalf("joined game = %#v", joined.Game)
	}
}

// TestProcessAllocatorRealWorkerRestoresLiveClientSession verifies process allocator real worker restores live client
// session. The scenario keeps the process allocator real contract visible to maintainers.
func TestProcessAllocatorRealWorkerRestoresLiveClientSession(t *testing.T) {
	executable := os.Getenv("DARK_MAGIC_REALM_WORKER_ACCEPTANCE")
	if executable == "" || os.Getenv("MPQ_DIRECTORY") == "" {
		t.Skip("DARK_MAGIC_REALM_WORKER_ACCEPTANCE and MPQ_DIRECTORY are required")
	}

	allocator, err := realm.NewProcessAllocator(realm.ProcessAllocatorConfig{
		Executable: executable, Arguments: []string{"--log-level", "error"}, StateDirectory: t.TempDir(),
		ControlListenAddress: "127.0.0.1:0", GameListenAddress: "127.0.0.1:0",
		StartupTimeout: 30 * time.Second, ShutdownTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_ = allocator.Close(ctx)
	})

	control, err := realm.NewControlPlane(
		realm.ControlPlaneConfig{Allocator: allocator, CheckpointInterval: time.Nanosecond},
	)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := control.CreateAccount(t.Context(), "RecoveryAlice", "long enough password"); err != nil {
		t.Fatal(err)
	}

	session, err := control.Authenticate(t.Context(), "RecoveryAlice", "long enough password")
	if err != nil {
		t.Fatal(err)
	}

	created, err := control.CreateCharacter(t.Context(), session.Token,
		realm.CreateCharacterRequest{Name: "RecoveryHero", Class: "Amazon"})
	if err != nil {
		t.Fatal(err)
	}

	handoff, err := control.CreateGame(t.Context(), session.Token, realm.CreateGameRequest{
		Name: "Live Recovery", Difficulty: realm.DifficultyNormal, Maximum: 8})
	if err != nil {
		t.Fatal(err)
	}

	tlsConfig, err := networktrust.PinnedTLSFingerprint(handoff.Assignment.Endpoint.TLSFingerprint)
	if err != nil {
		t.Fatal(err)
	}

	connected, err := clientsession.Connect(t.Context(), handoff.Assignment, tlsConfig)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = connected.Close(context.Background()) })

	if _, err := control.ReconcileGames(t.Context()); err != nil {
		t.Fatal(err)
	}

	gameID := handoff.Game.Entry.GameID
	if err := allocator.Release(t.Context(), gameID); err != nil {
		t.Fatal(err)
	}

	for attempt := 1; attempt <= 3; attempt++ {
		reconciled, err := control.ReconcileGames(t.Context())
		if err != nil {
			t.Fatalf("replacement attempt %d: %v", attempt, err)
		}

		if attempt < 3 && reconciled != 0 || attempt == 3 && reconciled != 1 {
			t.Fatalf("replacement attempt %d reconciled=%d", attempt, reconciled)
		}
	}

	reassignment, err := control.ReconnectGame(t.Context(), session.Token, gameID)
	if err != nil {
		t.Fatal(err)
	}

	replacementTLS, err := networktrust.PinnedTLSFingerprint(reassignment.Assignment.Endpoint.TLSFingerprint)
	if err != nil {
		t.Fatal(err)
	}

	if err := connected.Reassign(t.Context(), reassignment.Assignment, replacementTLS); err != nil {
		t.Fatal(err)
	}

	hud, _ := connected.View()
	if hud.Player.CharacterID != created.Character.ID || hud.Player.PlayerID == "" {
		t.Fatalf("reassigned HUD = %#v", hud)
	}

	worker, found := allocator.Game(gameID)
	if !found {
		t.Fatal("replacement worker is not registered")
	}

	status, err := worker.Status(t.Context())
	if err != nil || !status.Ready || status.ActivePlayers != 1 {
		t.Fatalf("replacement worker status=%#v error=%v", status, err)
	}
}

// TestProcessAllocatorRealWorkerCommitsExpiredTransportMembership verifies process allocator real worker commits
// expired transport membership. The scenario keeps the process allocator real contract visible to maintainers.
func TestProcessAllocatorRealWorkerCommitsExpiredTransportMembership(t *testing.T) {
	executable := os.Getenv("DARK_MAGIC_REALM_WORKER_ACCEPTANCE")
	if executable == "" || os.Getenv("MPQ_DIRECTORY") == "" {
		t.Skip("DARK_MAGIC_REALM_WORKER_ACCEPTANCE and MPQ_DIRECTORY are required")
	}

	allocator, err := realm.NewProcessAllocator(realm.ProcessAllocatorConfig{
		Executable: executable, Arguments: []string{"--log-level", "error"}, StateDirectory: t.TempDir(),
		ControlListenAddress: "127.0.0.1:0", GameListenAddress: "127.0.0.1:0",
		StartupTimeout: 30 * time.Second, ShutdownTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := allocator.Close(ctx); err != nil {
			t.Errorf("close allocator: %v", err)
		}
	})

	control, err := realm.NewControlPlane(realm.ControlPlaneConfig{Allocator: allocator})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := control.CreateAccount(t.Context(), "Alice", "long enough password"); err != nil {
		t.Fatal(err)
	}

	session, err := control.Authenticate(t.Context(), "Alice", "long enough password")
	if err != nil {
		t.Fatal(err)
	}

	createdCharacter, err := control.CreateCharacter(t.Context(), session.Token,
		realm.CreateCharacterRequest{Name: "CrashHero", Class: "Amazon"})
	if err != nil {
		t.Fatal(err)
	}

	handoff, err := control.CreateGame(t.Context(), session.Token, realm.CreateGameRequest{
		Name: "Crash Recovery", Difficulty: realm.DifficultyNormal, Maximum: 8,
	})
	if err != nil {
		t.Fatal(err)
	}

	tlsConfig, err := networktrust.PinnedTLSFingerprint(handoff.Assignment.Endpoint.TLSFingerprint)
	if err != nil {
		t.Fatal(err)
	}

	transport, err := sessionquic.Dial(t.Context(), handoff.Assignment.Endpoint.Address, tlsConfig)
	if err != nil {
		t.Fatal(err)
	}

	joined, err := transport.Join(t.Context(), gameserver.JoinRequest{Version: gameserver.SessionProtocolVersion,
		Credential: handoff.Assignment.Ticket, Identity: handoff.Assignment.Runtime})
	if err != nil {
		_ = transport.Close()

		t.Fatal(err)
	}

	if joined.Admission.CharacterID != createdCharacter.Character.ID {
		_ = transport.Close()

		t.Fatalf("joined admission = %#v", joined.Admission)
	}
	// Close the QUIC connection without the reliable Leave request, matching a
	// killed client. The worker holds the entity through reconnect grace.
	if err := transport.Close(); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := control.ReconcileGames(t.Context()); err != nil {
			t.Fatal(err)
		}

		committed, err := control.SelectedCharacter(t.Context(), session.Token)
		if err == nil && committed.Revision == createdCharacter.Revision+1 {
			if games, listErr := control.ListGames(
				t.Context(),
				session.Token,
				realm.GameFilter{},
			); listErr != nil ||
				len(games) != 0 {
				t.Fatalf("games after reconnect expiry = %#v, %v", games, listErr)
			}

			return
		}

		time.Sleep(100 * time.Millisecond)
	}

	committed, _ := control.SelectedCharacter(t.Context(), session.Token)
	t.Fatalf("reconnect expiry did not commit canonical character: %#v", committed)
}
