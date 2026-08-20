package realm_test

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"flag"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/app/serverapp"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

const processHelperEnvironment = "DARK_MAGIC_PROCESS_ALLOCATOR_HELPER"

// TestProcessAllocatorStartsControlsAndGracefullyReleasesWorker verifies process allocator starts controls and
// gracefully releases worker. The scenario keeps the process allocator contract visible to maintainers.
func TestProcessAllocatorStartsControlsAndGracefullyReleasesWorker(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	stateDirectory := t.TempDir()

	allocator, err := realm.NewProcessAllocator(realm.ProcessAllocatorConfig{
		Executable: executable, Arguments: []string{"-test.run=TestProcessAllocatorHelper", "--"},
		Environment: []string{processHelperEnvironment + "=1"}, StateDirectory: stateDirectory,
		StartupTimeout: 5 * time.Second, ShutdownTimeout: 3 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = allocator.Close(context.Background()) })

	allocation, err := allocator.Allocate(
		t.Context(),
		realm.ProcessGameSpec{GameID: "game-1", AllocationID: "allocation-1"},
	)
	if err != nil {
		t.Fatal(err)
	}

	if allocation.GameID != "game-1" || allocation.Endpoint.Address == "" || allocation.Worker == nil ||
		allocation.Tickets == nil {
		t.Fatalf("allocation = %#v", allocation)
	}

	description, err := allocation.Worker.Describe(t.Context())
	if err != nil || description.IdentityHash == "" || description.EntryDestination.LevelID != 1 {
		t.Fatalf("description=%#v error=%v", description, err)
	}

	ticket, err := allocation.Tickets.Issue(
		t.Context(),
		realm.AdmissionPrincipal{AccountID: "account", CharacterID: "character",
			PlayerID: "player", CharacterRevision: 1, RuntimeIdentityHash: description.IdentityHash},
		10*time.Second,
	)
	if err != nil || ticket != "helper-ticket" {
		t.Fatalf("ticket=%q error=%v", ticket, err)
	}

	if err := allocation.Tickets.Revoke(t.Context(), ticket); err != nil {
		t.Fatal(err)
	}

	repository, err := realm.NewMemoryCharacters(realm.CharacterRecord{AccountID: "account", Revision: 1,
		Character: d2save.Character{ID: "character", Name: "Hero", Class: "Amazon"}})
	if err != nil {
		t.Fatal(err)
	}

	admissions, err := realm.NewAdmissions(allocator, repository, time.Minute, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if err := admissions.RegisterGame(allocation.GameID, allocation.Tickets, allocation.Endpoint); err != nil {
		t.Fatal(err)
	}

	assignment, err := admissions.Join(t.Context(), realm.JoinRequest{AccountID: "account", CharacterID: "character",
		PlayerID: "player", GameID: allocation.GameID})
	if err != nil {
		t.Fatal(err)
	}

	if assignment.Endpoint != allocation.Endpoint || assignment.Ticket != "helper-ticket" ||
		assignment.CharacterRevision != 1 {
		t.Fatalf("assignment = %#v", assignment)
	}

	committed, err := admissions.CommitCanonicalMembership(t.Context(), allocation.GameID, "player")
	if err != nil || committed.Revision != 2 || committed.Character.ID != "character" {
		t.Fatalf("committed=%#v error=%v", committed, err)
	}

	if _, err := allocator.Allocate(
		t.Context(),
		realm.ProcessGameSpec{GameID: "game-1", AllocationID: "allocation-2"},
	); err == nil {
		t.Fatal("duplicate process allocation succeeded")
	}

	if err := allocator.Release(t.Context(), "game-1"); err != nil {
		t.Fatal(err)
	}

	if _, found := allocator.Game("game-1"); found {
		t.Fatal("released process worker remains registered")
	}

	recovery, err := processRecovery(processHelperDescription().Runtime, 5)
	if err != nil {
		t.Fatal(err)
	}

	handoff, err := realm.NewGameRecovery(recovery, []string{"player"})
	if err != nil {
		t.Fatal(err)
	}

	restored, err := allocator.Restore(
		t.Context(),
		realm.ProcessGameSpec{GameID: "game-restored", AllocationID: "allocation-restored"},
		handoff,
	)
	if err != nil {
		t.Fatal(err)
	}

	restoredCheckpoint, err := restored.Worker.Checkpoint(t.Context())
	if err != nil || restoredCheckpoint.State.Tick != 5 || restoredCheckpoint.Checksum != recovery.Checksum {
		t.Fatalf("restored checkpoint=%#v error=%v", restoredCheckpoint, err)
	}

	if err := allocator.Release(t.Context(), "game-restored"); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(stateDirectory)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 0 {
		t.Fatalf("worker secret directory remains after release: %#v", entries)
	}

	if err := allocator.Close(t.Context()); err != nil {
		t.Fatal(err)
	}

	if _, err := allocator.Allocate(
		t.Context(),
		realm.ProcessGameSpec{GameID: "after-close", AllocationID: "allocation-after-close"},
	); err == nil {
		t.Fatal("closed allocator accepted a new worker")
	}
}

// TestProcessAllocatorCleansWorkerThatExitsBeforeReadiness verifies process allocator cleans worker that exits before
// readiness. The scenario keeps the process allocator contract visible to maintainers.
func TestProcessAllocatorCleansWorkerThatExitsBeforeReadiness(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	stateDirectory := t.TempDir()

	allocator, err := realm.NewProcessAllocator(realm.ProcessAllocatorConfig{
		Executable: executable, Arguments: []string{"-test.run=TestProcessAllocatorHelper", "--"},
		Environment: []string{processHelperEnvironment + "=exit"}, StateDirectory: stateDirectory,
		StartupTimeout: time.Second, ShutdownTimeout: time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := allocator.Allocate(
		t.Context(),
		realm.ProcessGameSpec{GameID: "failed", AllocationID: "allocation-failed"},
	); err == nil {
		t.Fatal("worker that exited before readiness was accepted")
	}

	entries, err := os.ReadDir(stateDirectory)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 0 {
		t.Fatalf("failed worker secret directory remains: %#v", entries)
	}
}

// TestProcessAllocatorRejectsWorkerWithDifferentAssetSet verifies process allocator rejects worker with different
// asset set. The scenario keeps the process allocator contract visible to maintainers.
func TestProcessAllocatorRejectsWorkerWithDifferentAssetSet(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	stateDirectory := t.TempDir()

	allocator, err := realm.NewProcessAllocator(realm.ProcessAllocatorConfig{
		Executable: executable, Arguments: []string{"-test.run=TestProcessAllocatorHelper", "--"},
		Environment: []string{processHelperEnvironment + "=1"}, StateDirectory: stateDirectory,
		ExpectedAssetSetID: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		StartupTimeout:     5 * time.Second, ShutdownTimeout: 3 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := allocator.Allocate(
		t.Context(),
		realm.ProcessGameSpec{GameID: "wrong-assets", AllocationID: "allocation"},
	); err == nil {
		t.Fatal("worker with a different asset set became allocatable")
	}

	entries, err := os.ReadDir(stateDirectory)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 0 {
		t.Fatalf("rejected worker secret directory remains: %#v", entries)
	}
}

// TestProcessAllocatorFencesSurvivingWorkerBeforeRestartRestore verifies process allocator fences surviving worker
// before restart restore. The scenario keeps the process allocator contract visible to maintainers.
func TestProcessAllocatorFencesSurvivingWorkerBeforeRestartRestore(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	stateDirectory := t.TempDir()
	configuration := realm.ProcessAllocatorConfig{Executable: executable,
		Arguments:   []string{"-test.run=TestProcessAllocatorHelper", "--"},
		Environment: []string{processHelperEnvironment + "=1"}, StateDirectory: stateDirectory,
		StartupTimeout: 5 * time.Second, ShutdownTimeout: 3 * time.Second}

	first, err := realm.NewProcessAllocator(configuration)
	if err != nil {
		t.Fatal(err)
	}

	spec := realm.GameSpec{GameID: "restart-game", AllocationID: "restart-allocation"}
	if _, err := first.Allocate(t.Context(), spec); err != nil {
		t.Fatal(err)
	}

	restarted, err := realm.NewProcessAllocator(configuration)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = restarted.Close(context.Background()) })

	if err := restarted.Fence(t.Context(), realm.GameSpec{GameID: spec.GameID, AllocationID: "wrong"}); err == nil {
		t.Fatal("mismatched allocation generation fenced a worker")
	}

	if err := restarted.Fence(t.Context(), spec); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, found := first.Game(spec.GameID); !found {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	if _, found := first.Game(spec.GameID); found {
		t.Fatal("fenced worker remains alive")
	}

	checkpoint, err := processRecovery(processHelperDescription().Runtime, 8)
	if err != nil {
		t.Fatal(err)
	}

	recovery, err := realm.NewGameRecovery(checkpoint, []string{"player"})
	if err != nil {
		t.Fatal(err)
	}

	replacement, err := restarted.Restore(t.Context(), spec, recovery)
	if err != nil {
		t.Fatal(err)
	}

	description, err := replacement.Worker.Describe(t.Context())
	if err != nil || description.GameID != spec.GameID {
		t.Fatalf("replacement description=%#v error=%v", description, err)
	}
}

// TestProcessAllocatorHelper verifies process allocator helper. The scenario keeps the process allocator contract
// visible to maintainers.
func TestProcessAllocatorHelper(t *testing.T) {
	mode := os.Getenv(processHelperEnvironment)
	if mode == "exit" {
		return
	}

	if mode != "1" {
		return
	}

	arguments := os.Args
	separator := -1

	for index, argument := range arguments {
		if argument == "--" {
			separator = index
			break
		}
	}

	if separator < 0 {
		t.Fatal("helper arguments are missing separator")
	}

	flags := flag.NewFlagSet("process-worker-helper", flag.ContinueOnError)
	_ = flags.Bool("realm-worker", false, "")
	sessionID := flags.String("session-id", "", "")
	allocationID := flags.String("allocation-id", "", "")
	gameAddress := flags.String("quic-listen", "", "")
	certificatePath := flags.String("tls-cert", "", "")
	privateKeyPath := flags.String("tls-key", "", "")
	_ = flags.String("admission-key", "", "")
	controlAddress := flags.String("worker-control-listen", "", "")
	controlTokenPath := flags.String("worker-control-token", "", "")
	readyPath := flags.String("worker-ready-file", "", "")
	restorePath := flags.String("restore-checkpoint", "", "")
	_ = flags.String("game-difficulty", "normal", "")
	_ = flags.Int("game-maximum-players", 8, "")
	_ = flags.Bool("game-hardcore", false, "")

	_ = flags.Bool("game-ladder", false, "")
	if err := flags.Parse(arguments[separator+1:]); err != nil {
		t.Fatal(err)
	}

	if *sessionID == "" {
		t.Fatal("helper session ID is empty")
	}

	certificate, err := tls.LoadX509KeyPair(*certificatePath, *privateKeyPath)
	if err != nil || len(certificate.Certificate) == 0 {
		t.Fatalf("load helper certificate: %v", err)
	}

	token, err := os.ReadFile(*controlTokenPath)
	if err != nil {
		t.Fatal(err)
	}

	drain := make(chan struct{}, 1)
	worker := &processHelperWorker{description: processHelperDescription()}
	worker.description.GameID = *sessionID

	if *restorePath != "" {
		recovery, recoveryErr := serverapp.ReadGameRecovery(*restorePath)

		err = recoveryErr
		if err != nil {
			t.Fatal(err)
		}

		if len(recovery.PlayerIDs) != 1 || recovery.PlayerIDs[0] != "player" {
			t.Fatalf("helper recovered player IDs = %#v", recovery.PlayerIDs)
		}

		worker.checkpoint = recovery.Checkpoint
	}

	handler, err := realm.NewWorkerHTTPHandler(
		worker,
		processHelperTickets{},
		string(token),
		func() { drain <- struct{}{} },
	)
	if err != nil {
		t.Fatal(err)
	}

	listener, err := net.Listen("tcp", *controlAddress)
	if err != nil {
		t.Fatal(err)
	}

	server := &http.Server{Handler: handler}

	serverDone := make(chan error, 1)
	go func() {
		serverDone <- server.Serve(tls.NewListener(listener, &tls.Config{
			Certificates: []tls.Certificate{certificate},
			MinVersion:   tls.VersionTLS13,
		}))
	}()

	packet, err := net.ListenPacket("udp", *gameAddress)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = packet.Close() }()

	digest := sha256.Sum256(certificate.Certificate[0])
	if err := realm.WriteWorkerProcessReady(*readyPath, realm.WorkerProcessReady{
		GameID: *sessionID, AllocationID: *allocationID, ProcessID: os.Getpid(),
		ControlAddress: listener.Addr().String(), GameEndpoint: realm.GameEndpoint{
			Address: packet.LocalAddr().String(), TLSFingerprint: "sha256:" + hex.EncodeToString(digest[:]),
		},
	}); err != nil {
		t.Fatal(err)
	}

	select {
	case <-drain:
	case <-time.After(10 * time.Second):
		t.Fatal("helper did not receive drain")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	_ = server.Shutdown(shutdownCtx)

	cancel()
	<-serverDone
}

type processHelperWorker struct {
	description realm.WorkerDescription
	checkpoint  gamesession.RecoveryCheckpoint
}

// Describe supplies the test double's describe behavior, keeping the scenario deterministic and independent of
// external services.
func (worker *processHelperWorker) Describe(context.Context) (realm.WorkerDescription, error) {
	return worker.description, nil
}

// Status supplies the test double's status behavior, keeping the scenario deterministic and independent of external
// services.
func (*processHelperWorker) Status(context.Context) (realm.WorkerStatus, error) {
	return realm.WorkerStatus{Ready: true}, nil
}

// Checkpoint supplies the test double's checkpoint behavior, keeping the scenario deterministic and independent of
// external services.
func (worker *processHelperWorker) Checkpoint(context.Context) (gamesession.RecoveryCheckpoint, error) {
	if worker.checkpoint.Version != "" {
		return worker.checkpoint, nil
	}

	return processRecovery(worker.description.Runtime, 0)
}

// processRecovery supplies the test double's process recovery behavior, keeping the scenario deterministic and
// independent of external services.
func processRecovery(identity simulation.RuntimeIdentity, tick uint64) (gamesession.RecoveryCheckpoint, error) {
	participant, err := simulation.NewIdentityParticipant(identity)
	if err != nil {
		return gamesession.RecoveryCheckpoint{}, err
	}

	data, err := participant.SnapshotState()
	if err != nil {
		return gamesession.RecoveryCheckpoint{}, err
	}

	snapshot := gameecs.Snapshot{Version: gameecs.SnapshotVersion, Tick: tick}
	participants := []simulation.ParticipantState{{ID: participant.StateID(), Data: data}}

	checksum, err := simulation.CompositeChecksum(snapshot, participants)
	if err != nil {
		return gamesession.RecoveryCheckpoint{}, err
	}

	state := simulation.Checkpoint{Tick: tick, Checksum: checksum, Snapshot: &snapshot, Participants: participants}

	return gamesession.NewRecoveryCheckpoint(state, nil, nil, nil)
}

// AdmitCharacter supplies the test double's admit character behavior, keeping the scenario deterministic and
// independent of external services.
func (*processHelperWorker) AdmitCharacter(context.Context, realm.WorkerAdmission) error { return nil }

// RemoveCharacter supplies the test double's remove character behavior, keeping the scenario deterministic and
// independent of external services.
func (*processHelperWorker) RemoveCharacter(context.Context, string) error { return nil }

// ProjectCharacter supplies the test double's project character behavior, keeping the scenario deterministic and
// independent of external services.
func (*processHelperWorker) ProjectCharacter(
	_ context.Context,
	_ string,
	baseline d2save.Character,
) (d2save.Character, error) {
	return baseline, nil
}

// Close supplies the test double's close behavior, keeping the scenario deterministic and independent of external
// services.
func (*processHelperWorker) Close(context.Context) error { return nil }

type processHelperTickets struct{}

// Issue supplies the test double's issue behavior, keeping the scenario deterministic and independent of external
// services.
func (processHelperTickets) Issue(context.Context, realm.AdmissionPrincipal, time.Duration) (string, error) {
	return "helper-ticket", nil
}

// Revoke supplies the test double's revoke behavior, keeping the scenario deterministic and independent of external
// services.
func (processHelperTickets) Revoke(context.Context, string) error { return nil }

// processHelperDescription supplies the test double's process helper description behavior, keeping the scenario
// deterministic and independent of external services.
func processHelperDescription() realm.WorkerDescription {
	identity := simulation.RuntimeIdentity{Recipe: simulation.RuntimeRecipe{
		Schema:               simulation.RuntimeRecipeSchema,
		EngineAPI:            "v1",
		NetworkProtocol:      "test/v1",
		AssetSetID:           simulation.EmptyAssetSetID,
		GameDataGenerationID: simulation.GameDataGenerationIDForAssetSet(simulation.EmptyAssetSetID),
		Packages: simulation.RuntimePackageSet{Base: simulation.RuntimePackage{ID: "d2legacy", Version: "1.0.0",
			Digest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Size: 1}},
		AuthoritativeHash: "rules",
		ConfigurationHash: "config",
	}}
	digest, _ := identity.Digest()
	destination, _ := playeradapter.NewDestination(10, 10, 100, 100, 1, 1)

	return realm.WorkerDescription{Runtime: identity, IdentityHash: digest, EntryDestination: destination}
}
