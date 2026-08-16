package realm

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresCheckpoints struct {
	pool *pgxpool.Pool
}

func (store *PostgresCheckpoints) Save(ctx context.Context, record GameCheckpoint) (GameCheckpoint, error) {
	if store == nil || store.pool == nil {
		return GameCheckpoint{}, ErrGameCheckpoint
	}
	validated, payload, err := validateGameCheckpoint(record)
	if err != nil {
		return GameCheckpoint{}, err
	}
	err = pgx.BeginFunc(ctx, store.pool, func(tx pgx.Tx) error {
		var allocationID string
		var runtimeJSON []byte
		if err := tx.QueryRow(ctx, `SELECT allocator_id, runtime_identity FROM realm_allocations
			WHERE game_id = $1 AND state = $2 FOR UPDATE`, validated.GameID, AllocationReady).
			Scan(&allocationID, &runtimeJSON); err != nil {
			return ErrGameCheckpoint
		}
		var runtime simulation.RuntimeIdentity
		if json.Unmarshal(runtimeJSON, &runtime) != nil {
			return ErrGameCheckpoint
		}
		identityHash, digestErr := runtime.Digest()
		if digestErr != nil || allocationID != validated.AllocationID || identityHash != validated.IdentityHash {
			return ErrGameCheckpoint
		}
		var existingAllocationID, existingChecksum string
		var existingTick uint64
		scanErr := tx.QueryRow(ctx, `SELECT allocation_id, tick, checksum FROM realm_game_checkpoints
			WHERE game_id = $1 FOR UPDATE`, validated.GameID).
			Scan(&existingAllocationID, &existingTick, &existingChecksum)
		if scanErr != nil && !errors.Is(scanErr, pgx.ErrNoRows) {
			return scanErr
		}
		if scanErr == nil {
			if existingAllocationID != validated.AllocationID || validated.Tick < existingTick ||
				(validated.Tick == existingTick && validated.Checksum != existingChecksum) {
				return ErrGameCheckpoint
			}
			if validated.Tick == existingTick {
				return nil
			}
		}
		_, err := tx.Exec(ctx, `INSERT INTO realm_game_checkpoints
			(game_id, allocation_id, identity_hash, tick, checksum, payload, payload_digest)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (game_id) DO UPDATE SET allocation_id = EXCLUDED.allocation_id,
			identity_hash = EXCLUDED.identity_hash, tick = EXCLUDED.tick, checksum = EXCLUDED.checksum,
			payload = EXCLUDED.payload, payload_digest = EXCLUDED.payload_digest,
			updated_at = CURRENT_TIMESTAMP`, validated.GameID, validated.AllocationID, validated.IdentityHash,
			validated.Tick, validated.Checksum, payload, checkpointPayloadDigest(payload))
		return err
	})
	if err != nil {
		return GameCheckpoint{}, fmt.Errorf("realm: save PostgreSQL game checkpoint: %w", err)
	}
	return store.Latest(ctx, validated.GameID)
}

func (store *PostgresCheckpoints) Latest(ctx context.Context, gameID string) (GameCheckpoint, error) {
	if store == nil || store.pool == nil {
		return GameCheckpoint{}, ErrGameCheckpoint
	}
	var record GameCheckpoint
	var payload, payloadDigest []byte
	err := store.pool.QueryRow(ctx, `SELECT allocation_id, identity_hash, tick, checksum, payload,
		payload_digest, created_at, updated_at FROM realm_game_checkpoints WHERE game_id = $1`,
		strings.TrimSpace(gameID)).Scan(&record.AllocationID, &record.IdentityHash, &record.Tick, &record.Checksum,
		&payload, &payloadDigest, &record.CreatedAt, &record.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return GameCheckpoint{}, ErrGameCheckpoint
	}
	if err != nil {
		return GameCheckpoint{}, fmt.Errorf("realm: read PostgreSQL game checkpoint: %w", err)
	}
	digest := sha256.Sum256(payload)
	if !equalBytes(digest[:], payloadDigest) {
		return GameCheckpoint{}, ErrGameCheckpoint
	}
	checkpoint, err := decodeGameCheckpointPayload(payload)
	if err != nil {
		return GameCheckpoint{}, err
	}
	record.Version, record.GameID, record.Checkpoint = GameCheckpointVersion, strings.TrimSpace(gameID), checkpoint
	validated, _, err := validateGameCheckpoint(record)
	if err != nil {
		return GameCheckpoint{}, err
	}
	validated.CreatedAt, validated.UpdatedAt = record.CreatedAt, record.UpdatedAt
	return validated, nil
}

func (store *PostgresCheckpoints) Remove(ctx context.Context, gameID string) error {
	if store == nil || store.pool == nil {
		return ErrGameCheckpoint
	}
	_, err := store.pool.Exec(ctx, `DELETE FROM realm_game_checkpoints WHERE game_id = $1`, strings.TrimSpace(gameID))
	if err != nil {
		return fmt.Errorf("realm: remove PostgreSQL game checkpoint: %w", err)
	}
	return nil
}

func equalBytes(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	var difference byte
	for index := range left {
		difference |= left[index] ^ right[index]
	}
	return difference == 0
}

var _ CheckpointRepository = (*PostgresCheckpoints)(nil)
