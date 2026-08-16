package realm

import (
	"context"
	"fmt"
	"sync"
	"time"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

type orchestrationAllocator struct {
	mu          sync.Mutex
	workers     map[string]*orchestrationWorker
	allocations int
	releases    int
	fail        error
	admitErr    error
}

type orchestrationWorker struct {
	description WorkerDescription
	admitted    []WorkerAdmission
	removed     []string
	admitErr    error
	healthErr   error
	removeErr   error
	expired     []string
	checkpoints int
}

func newOrchestrationAllocator() *orchestrationAllocator {
	return &orchestrationAllocator{workers: make(map[string]*orchestrationWorker)}
}

func (allocator *orchestrationAllocator) Allocate(_ context.Context, spec GameSpec) (WorkerAllocation, error) {
	allocator.mu.Lock()
	defer allocator.mu.Unlock()
	if allocator.fail != nil {
		return WorkerAllocation{}, allocator.fail
	}
	if allocator.workers[spec.GameID] != nil {
		return WorkerAllocation{}, ErrGameExists
	}
	identity := orchestrationIdentity()
	digest, _ := identity.Digest()
	destination, _ := playeradapter.NewDestination(25, 30, 200, 150, 1, 1)
	worker := &orchestrationWorker{description: WorkerDescription{GameID: spec.GameID, Runtime: identity, IdentityHash: digest,
		EntryDestination: destination}, admitErr: allocator.admitErr}
	allocator.workers[spec.GameID] = worker
	allocator.allocations++
	return WorkerAllocation{GameID: spec.GameID, AllocationID: spec.AllocationID, Worker: worker, Tickets: orchestrationTickets{},
		Endpoint: GameEndpoint{Address: "127.0.0.1:4433", TLSFingerprint: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}, nil
}

func (allocator *orchestrationAllocator) Game(gameID string) (WorkerClient, bool) {
	allocator.mu.Lock()
	defer allocator.mu.Unlock()
	worker := allocator.workers[gameID]
	return worker, worker != nil
}

func (allocator *orchestrationAllocator) Release(_ context.Context, gameID string) error {
	allocator.mu.Lock()
	defer allocator.mu.Unlock()
	if allocator.workers[gameID] == nil {
		return ErrGameNotFound
	}
	delete(allocator.workers, gameID)
	allocator.releases++
	return nil
}

func (worker *orchestrationWorker) Describe(context.Context) (WorkerDescription, error) {
	return worker.description, nil
}
func (worker *orchestrationWorker) Status(context.Context) (WorkerStatus, error) {
	if worker.healthErr != nil {
		return WorkerStatus{}, worker.healthErr
	}
	return WorkerStatus{Ready: true, ActivePlayers: len(worker.admitted) - len(worker.removed),
		ExpiredPlayers: append([]string(nil), worker.expired...)}, nil
}
func (worker *orchestrationWorker) Checkpoint(context.Context) (gamesession.RecoveryCheckpoint, error) {
	worker.checkpoints++
	return testCheckpoint(worker.description.Runtime, 0)
}
func (worker *orchestrationWorker) AdmitCharacter(_ context.Context, admission WorkerAdmission) error {
	if worker.admitErr != nil {
		return worker.admitErr
	}
	worker.admitted = append(worker.admitted, admission)
	return nil
}
func (worker *orchestrationWorker) RemoveCharacter(_ context.Context, playerID string) error {
	worker.removed = append(worker.removed, playerID)
	if worker.removeErr == nil {
		worker.expired = nil
	}
	return worker.removeErr
}
func (*orchestrationWorker) ProjectCharacter(_ context.Context, _ string, baseline d2save.Character) (d2save.Character, error) {
	return baseline, nil
}
func (*orchestrationWorker) Close(context.Context) error { return nil }

type orchestrationTickets struct{}

func (orchestrationTickets) Issue(_ context.Context, principal AdmissionPrincipal, _ time.Duration) (string, error) {
	return fmt.Sprintf("ticket:%s", principal.PlayerID), nil
}
func (orchestrationTickets) Revoke(context.Context, string) error { return nil }

func orchestratedControlConfig(audit AuditSink) ControlPlaneConfig {
	destination, _ := playeradapter.NewDestination(10, 10, 100, 100, 1, 1)
	return ControlPlaneConfig{Allocator: newOrchestrationAllocator(), EntryDestination: destination, Audit: audit}
}

func orchestrationIdentity() simulation.RuntimeIdentity {
	return simulation.RuntimeIdentity{Recipe: simulation.RuntimeRecipe{Schema: simulation.RuntimeRecipeSchema,
		EngineAPI: "v1", NetworkProtocol: "test/v1", AssetSetID: simulation.EmptyAssetSetID,
		GameDataGenerationID: simulation.GameDataGenerationIDForAssetSet(simulation.EmptyAssetSetID), Packages: simulation.RuntimePackageSet{Base: simulation.RuntimePackage{
			ID: "d2legacy", Version: "1.0.0", Digest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Size: 1}},
		AuthoritativeHash: "rules", ConfigurationHash: "config"}}
}

func testCheckpoint(identity simulation.RuntimeIdentity, tick uint64) (gamesession.RecoveryCheckpoint, error) {
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
