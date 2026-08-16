package realm

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const (
	GameCheckpointVersion      = "RealmGameCheckpoint/v1"
	maximumGameCheckpointBytes = 32 << 20
)

var ErrGameCheckpoint = errors.New("realm: invalid game checkpoint")

type GameCheckpoint struct {
	Version      string                         `json:"version"`
	GameID       string                         `json:"game_id"`
	AllocationID string                         `json:"allocation_id"`
	IdentityHash string                         `json:"identity_hash"`
	Tick         uint64                         `json:"tick"`
	Checksum     string                         `json:"checksum"`
	Checkpoint   gamesession.RecoveryCheckpoint `json:"checkpoint"`
	CreatedAt    time.Time                      `json:"created_at"`
	UpdatedAt    time.Time                      `json:"updated_at"`
}

type CheckpointRepository interface {
	Save(context.Context, GameCheckpoint) (GameCheckpoint, error)
	Latest(context.Context, string) (GameCheckpoint, error)
	Remove(context.Context, string) error
}

type MemoryCheckpoints struct {
	mu      sync.Mutex
	now     func() time.Time
	records map[string]GameCheckpoint
}

func NewMemoryCheckpoints() *MemoryCheckpoints {
	return &MemoryCheckpoints{now: time.Now, records: make(map[string]GameCheckpoint)}
}

func (store *MemoryCheckpoints) Save(ctx context.Context, record GameCheckpoint) (GameCheckpoint, error) {
	if err := contextErr(ctx); err != nil {
		return GameCheckpoint{}, err
	}
	validated, _, err := validateGameCheckpoint(record)
	if err != nil {
		return GameCheckpoint{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if existing, found := store.records[validated.GameID]; found {
		if existing.AllocationID != validated.AllocationID || validated.Tick < existing.Tick ||
			(validated.Tick == existing.Tick && validated.Checksum != existing.Checksum) {
			return GameCheckpoint{}, ErrGameCheckpoint
		}
		if validated.Tick == existing.Tick {
			return cloneGameCheckpoint(existing)
		}
		validated.CreatedAt = existing.CreatedAt
	} else {
		validated.CreatedAt = store.now().UTC()
	}
	validated.UpdatedAt = store.now().UTC()
	store.records[validated.GameID] = validated
	return cloneGameCheckpoint(validated)
}

func (store *MemoryCheckpoints) Latest(ctx context.Context, gameID string) (GameCheckpoint, error) {
	if err := contextErr(ctx); err != nil {
		return GameCheckpoint{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	record, found := store.records[strings.TrimSpace(gameID)]
	if !found {
		return GameCheckpoint{}, ErrGameCheckpoint
	}
	return cloneGameCheckpoint(record)
}

func (store *MemoryCheckpoints) Remove(ctx context.Context, gameID string) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	store.mu.Lock()
	delete(store.records, strings.TrimSpace(gameID))
	store.mu.Unlock()
	return nil
}

func NewGameCheckpoint(gameID, allocationID, identityHash string, checkpoint gamesession.RecoveryCheckpoint) (GameCheckpoint, error) {
	record := GameCheckpoint{Version: GameCheckpointVersion, GameID: strings.TrimSpace(gameID),
		AllocationID: strings.TrimSpace(allocationID), IdentityHash: strings.TrimSpace(identityHash),
		Tick: checkpoint.State.Tick, Checksum: checkpoint.Checksum, Checkpoint: checkpoint}
	validated, _, err := validateGameCheckpoint(record)
	return validated, err
}

func validateGameCheckpoint(record GameCheckpoint) (GameCheckpoint, []byte, error) {
	if record.Version != GameCheckpointVersion || strings.TrimSpace(record.GameID) == "" ||
		strings.TrimSpace(record.AllocationID) == "" || strings.TrimSpace(record.IdentityHash) == "" ||
		record.Checkpoint.State.Snapshot == nil || record.Tick != record.Checkpoint.State.Tick ||
		record.Checksum == "" || record.Checksum != record.Checkpoint.Checksum ||
		record.Checkpoint.State.Snapshot.Tick != record.Tick {
		return GameCheckpoint{}, nil, ErrGameCheckpoint
	}
	if err := gamesession.ValidateRecoveryCheckpoint(record.Checkpoint); err != nil {
		return GameCheckpoint{}, nil, ErrGameCheckpoint
	}
	engine, err := gameecs.RestoreSnapshot(*record.Checkpoint.State.Snapshot)
	if err != nil {
		return GameCheckpoint{}, nil, ErrGameCheckpoint
	}
	if err := engine.Close(); err != nil {
		return GameCheckpoint{}, nil, ErrGameCheckpoint
	}
	participantIDs := make(map[string]struct{}, len(record.Checkpoint.State.Participants))
	for _, participant := range record.Checkpoint.State.Participants {
		if strings.TrimSpace(participant.ID) == "" {
			return GameCheckpoint{}, nil, ErrGameCheckpoint
		}
		if _, duplicate := participantIDs[participant.ID]; duplicate {
			return GameCheckpoint{}, nil, ErrGameCheckpoint
		}
		participantIDs[participant.ID] = struct{}{}
	}
	identity, err := simulation.RuntimeIdentityFromParticipants(record.Checkpoint.State.Participants)
	if err != nil {
		return GameCheckpoint{}, nil, ErrGameCheckpoint
	}
	digest, err := identity.Digest()
	if err != nil || digest != record.IdentityHash {
		return GameCheckpoint{}, nil, ErrGameCheckpoint
	}
	payload, err := json.Marshal(record.Checkpoint)
	if err != nil || len(payload) == 0 || len(payload) > maximumGameCheckpointBytes {
		return GameCheckpoint{}, nil, ErrGameCheckpoint
	}
	checkpoint, err := decodeGameCheckpointPayload(payload)
	if err != nil {
		return GameCheckpoint{}, nil, err
	}
	record.Checkpoint = checkpoint
	return record, payload, nil
}

func decodeGameCheckpointPayload(payload []byte) (gamesession.RecoveryCheckpoint, error) {
	if len(payload) == 0 || len(payload) > maximumGameCheckpointBytes {
		return gamesession.RecoveryCheckpoint{}, ErrGameCheckpoint
	}
	var checkpoint gamesession.RecoveryCheckpoint
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&checkpoint); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return gamesession.RecoveryCheckpoint{}, ErrGameCheckpoint
	}
	return checkpoint, nil
}

func cloneGameCheckpoint(record GameCheckpoint) (GameCheckpoint, error) {
	payload, err := json.Marshal(record.Checkpoint)
	if err != nil {
		return GameCheckpoint{}, fmt.Errorf("%w: %v", ErrGameCheckpoint, err)
	}
	checkpoint, err := decodeGameCheckpointPayload(payload)
	if err != nil {
		return GameCheckpoint{}, err
	}
	record.Checkpoint = checkpoint
	return record, nil
}

func checkpointPayloadDigest(payload []byte) []byte {
	digest := sha256.Sum256(payload)
	return digest[:]
}

var _ CheckpointRepository = (*MemoryCheckpoints)(nil)
