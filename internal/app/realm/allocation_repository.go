package realm

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gravestench/dark-magic/internal/game/simulation"
)

type AllocationState string

const (
	AllocationRequested AllocationState = "requested"
	AllocationReady     AllocationState = "ready"
	AllocationFailed    AllocationState = "failed"
	AllocationCompleted AllocationState = "completed"
)

var ErrAllocationRecord = errors.New("realm: invalid allocation record")

type AllocationRecord struct {
	GameID        string
	AllocationID  string
	State         AllocationState
	Endpoint      GameEndpoint
	Runtime       simulation.RuntimeIdentity
	LastHealthyAt *time.Time
	LastError     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// AllocationRepository is the durable half of external worker orchestration.
// It records intent before an allocator call and terminal state after cleanup,
// allowing restart reconciliation to distinguish interrupted transitions from
// healthy games. It never stores worker bearer credentials or ticket keys.
type AllocationRepository interface {
	Request(context.Context, string, string) (AllocationRecord, error)
	Ready(context.Context, string, GameEndpoint, simulation.RuntimeIdentity) (AllocationRecord, error)
	RestoreReady(context.Context, string, string, GameEndpoint, simulation.RuntimeIdentity) (AllocationRecord, error)
	Healthy(context.Context, string) error
	Fail(context.Context, string, error) error
	Complete(context.Context, string) error
	Get(context.Context, string) (AllocationRecord, error)
	Active(context.Context) ([]AllocationRecord, error)
}

func (store *MemoryAllocations) RestoreReady(ctx context.Context, gameID, allocationID string, endpoint GameEndpoint, runtime simulation.RuntimeIdentity) (AllocationRecord, error) {
	if err := contextErr(ctx); err != nil {
		return AllocationRecord{}, err
	}
	if store == nil || !validGameEndpoint(endpoint) {
		return AllocationRecord{}, ErrAllocationRecord
	}
	if _, err := runtime.Digest(); err != nil {
		return AllocationRecord{}, ErrAllocationRecord
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	record, found := store.records[strings.TrimSpace(gameID)]
	if !found || record.State != AllocationReady || record.AllocationID != strings.TrimSpace(allocationID) {
		return AllocationRecord{}, ErrAllocationRecord
	}
	want, wantErr := record.Runtime.Digest()
	got, gotErr := runtime.Digest()
	if wantErr != nil || gotErr != nil || want != got {
		return AllocationRecord{}, ErrAllocationRecord
	}
	now := store.now().UTC()
	record.Endpoint, record.Runtime, record.LastHealthyAt, record.UpdatedAt = endpoint, cloneRuntimeIdentity(runtime), &now, now
	record.LastError = ""
	store.records[record.GameID] = record
	return cloneAllocationRecord(record), nil
}

type MemoryAllocations struct {
	mu      sync.Mutex
	now     func() time.Time
	records map[string]AllocationRecord
}

func NewMemoryAllocations() *MemoryAllocations {
	return &MemoryAllocations{now: time.Now, records: make(map[string]AllocationRecord)}
}

func (store *MemoryAllocations) Request(ctx context.Context, gameID, allocationID string) (AllocationRecord, error) {
	if err := contextErr(ctx); err != nil {
		return AllocationRecord{}, err
	}
	gameID, allocationID = strings.TrimSpace(gameID), strings.TrimSpace(allocationID)
	if store == nil || gameID == "" || allocationID == "" {
		return AllocationRecord{}, ErrAllocationRecord
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if existing, exists := store.records[gameID]; exists {
		if existing.AllocationID == allocationID {
			return cloneAllocationRecord(existing), nil
		}
		return AllocationRecord{}, ErrGameExists
	}
	now := store.now().UTC()
	record := AllocationRecord{GameID: gameID, AllocationID: allocationID, State: AllocationRequested,
		CreatedAt: now, UpdatedAt: now}
	store.records[gameID] = record
	return cloneAllocationRecord(record), nil
}

func (store *MemoryAllocations) Ready(ctx context.Context, gameID string, endpoint GameEndpoint, runtime simulation.RuntimeIdentity) (AllocationRecord, error) {
	if err := contextErr(ctx); err != nil {
		return AllocationRecord{}, err
	}
	if store == nil || !validGameEndpoint(endpoint) {
		return AllocationRecord{}, ErrAllocationRecord
	}
	if _, err := runtime.Digest(); err != nil {
		return AllocationRecord{}, ErrAllocationRecord
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	record, found := store.records[strings.TrimSpace(gameID)]
	if !found {
		return AllocationRecord{}, ErrAllocationRecord
	}
	if record.State == AllocationReady {
		if allocationReadyMatches(record, endpoint, runtime) {
			return cloneAllocationRecord(record), nil
		}
		return AllocationRecord{}, ErrAllocationRecord
	}
	if record.State != AllocationRequested {
		return AllocationRecord{}, ErrAllocationRecord
	}
	now := store.now().UTC()
	record.State, record.Endpoint, record.Runtime = AllocationReady, endpoint, cloneRuntimeIdentity(runtime)
	record.LastHealthyAt, record.UpdatedAt = &now, now
	store.records[record.GameID] = record
	return cloneAllocationRecord(record), nil
}

func (store *MemoryAllocations) Healthy(ctx context.Context, gameID string) error {
	return store.transition(ctx, gameID, AllocationReady, "")
}

func (store *MemoryAllocations) Fail(ctx context.Context, gameID string, cause error) error {
	message := ""
	if cause != nil {
		message = cause.Error()
	}
	return store.transition(ctx, gameID, AllocationFailed, message)
}

func (store *MemoryAllocations) Complete(ctx context.Context, gameID string) error {
	return store.transition(ctx, gameID, AllocationCompleted, "")
}

func (store *MemoryAllocations) transition(ctx context.Context, gameID string, state AllocationState, message string) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	if store == nil {
		return ErrAllocationRecord
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	record, found := store.records[strings.TrimSpace(gameID)]
	if !found {
		return ErrAllocationRecord
	}
	if state == AllocationReady && record.State != AllocationReady {
		return ErrAllocationRecord
	}
	if state != AllocationReady && record.State == state {
		return nil
	}
	if state != AllocationReady && record.State != AllocationRequested && record.State != AllocationReady {
		return ErrAllocationRecord
	}
	now := store.now().UTC()
	record.State, record.LastError, record.UpdatedAt = state, message, now
	if state == AllocationReady {
		record.LastHealthyAt = &now
	}
	store.records[record.GameID] = record
	return nil
}

func (store *MemoryAllocations) Active(ctx context.Context) ([]AllocationRecord, error) {
	if err := contextErr(ctx); err != nil {
		return nil, err
	}
	if store == nil {
		return nil, ErrAllocationRecord
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	result := make([]AllocationRecord, 0)
	for _, record := range store.records {
		if record.State == AllocationRequested || record.State == AllocationReady {
			result = append(result, cloneAllocationRecord(record))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].GameID < result[j].GameID })
	return result, nil
}

func (store *MemoryAllocations) Get(ctx context.Context, gameID string) (AllocationRecord, error) {
	if err := contextErr(ctx); err != nil {
		return AllocationRecord{}, err
	}
	if store == nil {
		return AllocationRecord{}, ErrAllocationRecord
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	record, found := store.records[strings.TrimSpace(gameID)]
	if !found {
		return AllocationRecord{}, ErrAllocationRecord
	}
	return cloneAllocationRecord(record), nil
}

func validGameEndpoint(endpoint GameEndpoint) bool {
	return strings.TrimSpace(endpoint.Address) != "" && strings.TrimSpace(endpoint.TLSFingerprint) != ""
}

func allocationReadyMatches(record AllocationRecord, endpoint GameEndpoint, runtime simulation.RuntimeIdentity) bool {
	if record.Endpoint != endpoint {
		return false
	}
	want, wantErr := record.Runtime.Digest()
	got, gotErr := runtime.Digest()
	return wantErr == nil && gotErr == nil && want == got
}

func cloneAllocationRecord(record AllocationRecord) AllocationRecord {
	record.Runtime = cloneRuntimeIdentity(record.Runtime)
	if record.LastHealthyAt != nil {
		value := *record.LastHealthyAt
		record.LastHealthyAt = &value
	}
	return record
}

var _ AllocationRepository = (*MemoryAllocations)(nil)
