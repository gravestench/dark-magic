package realm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAllocations struct {
	pool *pgxpool.Pool
}

// Request executes request through the PostgreSQL allocation store so SQL encoding and database error translation
// remain centralized.
func (store *PostgresAllocations) Request(ctx context.Context, gameID, allocationID string) (AllocationRecord, error) {
	gameID, allocationID = strings.TrimSpace(gameID), strings.TrimSpace(allocationID)
	if store == nil || store.pool == nil || gameID == "" || allocationID == "" {
		return AllocationRecord{}, ErrAllocationRecord
	}

	command, err := store.pool.Exec(ctx, `INSERT INTO realm_allocations (game_id, allocator_id, state)
		VALUES ($1, $2, $3) ON CONFLICT (game_id) DO NOTHING`, gameID, allocationID, AllocationRequested)
	if postgresUniqueViolation(err) {
		return AllocationRecord{}, ErrGameExists
	}

	if err != nil {
		return AllocationRecord{}, fmt.Errorf("realm: request PostgreSQL allocation: %w", err)
	}

	record, err := store.get(ctx, gameID)
	if err != nil {
		return AllocationRecord{}, err
	}

	if command.RowsAffected() == 0 && record.AllocationID != allocationID {
		return AllocationRecord{}, ErrGameExists
	}

	return record, nil
}

// Ready decodes the durable allocation representation at one boundary so malformed data fails before it becomes shared
// state.
func (store *PostgresAllocations) Ready(
	ctx context.Context,
	gameID string,
	endpoint GameEndpoint,
	runtime simulation.RuntimeIdentity,
) (AllocationRecord, error) {
	if store == nil || store.pool == nil || !validGameEndpoint(endpoint) {
		return AllocationRecord{}, ErrAllocationRecord
	}

	if _, err := runtime.Digest(); err != nil {
		return AllocationRecord{}, ErrAllocationRecord
	}

	endpointJSON, err := json.Marshal(endpoint)
	if err != nil {
		return AllocationRecord{}, err
	}

	runtimeJSON, err := json.Marshal(runtime)
	if err != nil {
		return AllocationRecord{}, err
	}

	command, err := store.pool.Exec(ctx, `UPDATE realm_allocations SET state = $2, endpoint = $3,
		runtime_identity = $4, last_healthy_at = CURRENT_TIMESTAMP, last_error = NULL,
		updated_at = CURRENT_TIMESTAMP WHERE game_id = $1 AND state = $5`, strings.TrimSpace(gameID),
		AllocationReady, endpointJSON, runtimeJSON, AllocationRequested)
	if err != nil {
		return AllocationRecord{}, fmt.Errorf("realm: ready PostgreSQL allocation: %w", err)
	}

	if command.RowsAffected() != 1 {
		existing, getErr := store.get(ctx, gameID)
		if getErr == nil && existing.State == AllocationReady && allocationReadyMatches(existing, endpoint, runtime) {
			return existing, nil
		}

		return AllocationRecord{}, ErrAllocationRecord
	}

	return store.get(ctx, gameID)
}

// RestoreReady executes restore ready through the PostgreSQL allocation store so SQL encoding and database error
// translation remain centralized.
func (store *PostgresAllocations) RestoreReady(
	ctx context.Context,
	gameID, allocationID string,
	endpoint GameEndpoint,
	runtime simulation.RuntimeIdentity,
) (AllocationRecord, error) {
	if store == nil || store.pool == nil || !validGameEndpoint(endpoint) {
		return AllocationRecord{}, ErrAllocationRecord
	}

	if _, err := runtime.Digest(); err != nil {
		return AllocationRecord{}, ErrAllocationRecord
	}

	return store.restoreReadyTransaction(ctx, gameID, allocationID, endpoint, runtime)
}

// restoreReadyTransaction emits the canonical durable allocation representation so persisted and transported values
// retain one stable shape.
func (store *PostgresAllocations) restoreReadyTransaction(
	ctx context.Context,
	gameID, allocationID string,
	endpoint GameEndpoint,
	runtime simulation.RuntimeIdentity,
) (AllocationRecord, error) {
	err := pgx.BeginFunc(ctx, store.pool, func(tx pgx.Tx) error {
		record, err := scanPostgresAllocation(tx.QueryRow(ctx, postgresAllocationSelect+
			` WHERE game_id = $1 FOR UPDATE`, strings.TrimSpace(gameID)))
		if err != nil || record.State != AllocationReady || record.AllocationID != strings.TrimSpace(allocationID) {
			return ErrAllocationRecord
		}

		want, wantErr := record.Runtime.Digest()

		got, gotErr := runtime.Digest()
		if wantErr != nil || gotErr != nil || want != got {
			return ErrAllocationRecord
		}

		endpointJSON, _ := json.Marshal(endpoint)

		runtimeJSON, _ := json.Marshal(runtime)
		if _, err := tx.Exec(ctx, `UPDATE realm_allocations SET endpoint = $2, runtime_identity = $3,
			last_healthy_at = CURRENT_TIMESTAMP, last_error = NULL, updated_at = CURRENT_TIMESTAMP
			WHERE game_id = $1`, record.GameID, endpointJSON, runtimeJSON); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return AllocationRecord{}, err
	}

	return store.get(ctx, gameID)
}

// Healthy executes healthy through the PostgreSQL allocation store so SQL encoding and database error translation
// remain centralized.
func (store *PostgresAllocations) Healthy(ctx context.Context, gameID string) error {
	if store == nil || store.pool == nil {
		return ErrAllocationRecord
	}

	command, err := store.pool.Exec(ctx, `UPDATE realm_allocations SET last_healthy_at = CURRENT_TIMESTAMP,
		updated_at = CURRENT_TIMESTAMP WHERE game_id = $1 AND state = $2`, strings.TrimSpace(gameID), AllocationReady)
	if err != nil {
		return fmt.Errorf("realm: mark PostgreSQL allocation healthy: %w", err)
	}

	if command.RowsAffected() != 1 {
		return ErrAllocationRecord
	}

	return nil
}

// Fail executes fail through the PostgreSQL allocation store so SQL encoding and database error translation remain
// centralized.
func (store *PostgresAllocations) Fail(ctx context.Context, gameID string, cause error) error {
	message := ""
	if cause != nil {
		message = boundedAllocationError(cause.Error())
	}

	return store.terminal(ctx, gameID, AllocationFailed, message)
}

// Complete executes complete through the PostgreSQL allocation store so SQL encoding and database error translation
// remain centralized.
func (store *PostgresAllocations) Complete(ctx context.Context, gameID string) error {
	return store.terminal(ctx, gameID, AllocationCompleted, "")
}

// terminal executes terminal through the PostgreSQL allocation store so SQL encoding and database error translation
// remain centralized.
func (store *PostgresAllocations) terminal(
	ctx context.Context,
	gameID string,
	state AllocationState,
	message string,
) error {
	if store == nil || store.pool == nil {
		return ErrAllocationRecord
	}

	command, err := store.pool.Exec(ctx, `UPDATE realm_allocations SET state = $2, last_error = NULLIF($3, ''),
		updated_at = CURRENT_TIMESTAMP WHERE game_id = $1 AND state IN ($2, $4, $5)`, strings.TrimSpace(gameID),
		state, message, AllocationRequested, AllocationReady)
	if err != nil {
		return fmt.Errorf("realm: finish PostgreSQL allocation: %w", err)
	}

	if command.RowsAffected() != 1 {
		return ErrAllocationRecord
	}

	return nil
}

// Active executes active through the PostgreSQL allocation store so SQL encoding and database error translation remain
// centralized.
func (store *PostgresAllocations) Active(ctx context.Context) ([]AllocationRecord, error) {
	if store == nil || store.pool == nil {
		return nil, ErrAllocationRecord
	}

	rows, err := store.pool.Query(ctx, postgresAllocationSelect+` WHERE state IN ($1, $2) ORDER BY game_id`,
		AllocationRequested, AllocationReady)
	if err != nil {
		return nil, fmt.Errorf("realm: list PostgreSQL allocations: %w", err)
	}
	defer rows.Close()

	result := make([]AllocationRecord, 0)

	for rows.Next() {
		record, scanErr := scanPostgresAllocation(rows)
		if scanErr != nil {
			return nil, scanErr
		}

		result = append(result, record)
	}

	return result, rows.Err()
}

// get executes get through the PostgreSQL allocation store so SQL encoding and database error translation remain
// centralized.
func (store *PostgresAllocations) get(ctx context.Context, gameID string) (AllocationRecord, error) {
	record, err := scanPostgresAllocation(store.pool.QueryRow(ctx, postgresAllocationSelect+` WHERE game_id = $1`,
		strings.TrimSpace(gameID)))
	if errors.Is(err, pgx.ErrNoRows) {
		return AllocationRecord{}, ErrAllocationRecord
	}

	return record, err
}

// Get executes get through the PostgreSQL allocation store so SQL encoding and database error translation remain
// centralized.
func (store *PostgresAllocations) Get(ctx context.Context, gameID string) (AllocationRecord, error) {
	if store == nil || store.pool == nil {
		return AllocationRecord{}, ErrAllocationRecord
	}

	return store.get(ctx, gameID)
}

const postgresAllocationSelect = `SELECT game_id, allocator_id, state, endpoint, runtime_identity,
	last_healthy_at, COALESCE(last_error, ''), created_at, updated_at FROM realm_allocations`

// scanPostgresAllocation decodes the durable allocation representation at one boundary so malformed data fails before
// it becomes shared state.
func scanPostgresAllocation(row postgresRowScanner) (AllocationRecord, error) {
	var (
		record                    AllocationRecord
		endpointJSON, runtimeJSON []byte
	)
	if err := row.Scan(&record.GameID, &record.AllocationID, &record.State, &endpointJSON, &runtimeJSON,
		&record.LastHealthyAt, &record.LastError, &record.CreatedAt, &record.UpdatedAt); err != nil {
		return AllocationRecord{}, err
	}

	if len(endpointJSON) != 0 {
		if err := json.Unmarshal(endpointJSON, &record.Endpoint); err != nil {
			return AllocationRecord{}, fmt.Errorf("realm: decode PostgreSQL allocation endpoint: %w", err)
		}
	}

	if len(runtimeJSON) != 0 {
		if err := json.Unmarshal(runtimeJSON, &record.Runtime); err != nil {
			return AllocationRecord{}, fmt.Errorf("realm: decode PostgreSQL allocation runtime: %w", err)
		}
	}

	return cloneAllocationRecord(record), nil
}

// boundedAllocationError executes bounded allocation error through the PostgreSQL allocation store so SQL encoding and
// database error translation remain centralized.
func boundedAllocationError(message string) string {
	message = strings.ToValidUTF8(strings.TrimSpace(message), "?")
	if utf8.RuneCountInString(message) <= 512 {
		return message
	}

	return string([]rune(message)[:512])
}

var _ AllocationRepository = (*PostgresAllocations)(nil)
