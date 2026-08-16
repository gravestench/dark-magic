package realm

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresMemberships struct {
	pool *pgxpool.Pool
}

func (store *PostgresMemberships) Admit(ctx context.Context, record MembershipRecord) error {
	if store == nil || store.pool == nil {
		return ErrMembership
	}
	if err := validateActiveMembership(record); err != nil {
		return err
	}
	return pgx.BeginFunc(ctx, store.pool, func(tx pgx.Tx) error {
		canonical, err := (&PostgresCharacters{pool: store.pool}).lockLeasedCharacter(ctx, tx, record.Lease)
		if err != nil || canonical.AccountID != record.AccountID || canonical.Character.ID != record.Baseline.Character.ID {
			return ErrMembership
		}
		if _, err := tx.Exec(ctx, `DELETE FROM realm_memberships
			WHERE game_id = $1 AND character_id = $2 AND state <> $3`, record.GameID,
			record.Baseline.Character.ID, MembershipActive); err != nil {
			return fmt.Errorf("realm: clear PostgreSQL departure receipt for rejoin: %w", err)
		}
		baseline, err := json.Marshal(canonical)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO realm_memberships
			(game_id, player_id, account_id, character_id, state, baseline, departure_receipt)
			VALUES ($1, $2, $3, $4, $5, $6, NULL)`, record.GameID, record.PlayerID,
			record.AccountID, canonical.Character.ID, MembershipActive, baseline)
		if postgresUniqueViolation(err) {
			return ErrCharacterLeased
		}
		if err != nil {
			return fmt.Errorf("realm: admit PostgreSQL membership: %w", err)
		}
		return nil
	})
}

func (store *PostgresMemberships) Cancel(ctx context.Context, gameID, playerID string) error {
	if store == nil || store.pool == nil {
		return ErrMembership
	}
	_, err := store.pool.Exec(ctx, `DELETE FROM realm_memberships
		WHERE game_id = $1 AND player_id = $2 AND state = $3`, strings.TrimSpace(gameID),
		strings.TrimSpace(playerID), MembershipActive)
	if err != nil {
		return fmt.Errorf("realm: cancel PostgreSQL membership: %w", err)
	}
	return nil
}

func (store *PostgresMemberships) Depart(ctx context.Context, wanted MembershipRecord, character d2save.Character) (departureReceipt, error) {
	if store == nil || store.pool == nil || validateActiveMembership(wanted) != nil ||
		strings.TrimSpace(character.ID) == "" || character.ID != wanted.Baseline.Character.ID {
		return departureReceipt{}, ErrMembership
	}
	var receipt departureReceipt
	err := pgx.BeginFunc(ctx, store.pool, func(tx pgx.Tx) error {
		persisted, err := scanPostgresMembership(tx.QueryRow(ctx, postgresMembershipSelect+
			` WHERE game_id = $1 AND player_id = $2 FOR UPDATE`, wanted.GameID, wanted.PlayerID))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrMembership
		}
		if err != nil {
			return err
		}
		if persisted.State == MembershipDeparted && persisted.Departure != nil {
			receipt = cloneDepartureReceipt(*persisted.Departure)
			return nil
		}
		if persisted.State != MembershipActive || persisted.AccountID != wanted.AccountID ||
			persisted.Baseline.Character.ID != wanted.Baseline.Character.ID {
			return ErrMembership
		}
		record, err := (&PostgresCharacters{pool: store.pool}).lockLeasedCharacter(ctx, tx, wanted.Lease)
		if err != nil || record.AccountID != wanted.AccountID || record.Character.ID != character.ID {
			return ErrCharacterCommit
		}
		encoded, err := json.Marshal(character)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE realm_characters SET character = $1, revision = revision + 1,
			updated_at = CURRENT_TIMESTAMP WHERE id = $2`, encoded, wanted.Lease.CharacterID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `DELETE FROM realm_character_leases WHERE character_id = $1`, wanted.Lease.CharacterID); err != nil {
			return err
		}
		record.Character = cloneCharacter(character)
		record.Revision++
		receipt = departureReceipt{Record: record, PlayerID: wanted.PlayerID}
		encodedReceipt, err := json.Marshal(receipt)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE realm_memberships SET state = $3, departure_receipt = $4,
			updated_at = CURRENT_TIMESTAMP WHERE game_id = $1 AND player_id = $2`, wanted.GameID,
			wanted.PlayerID, MembershipDeparted, encodedReceipt); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return departureReceipt{}, fmt.Errorf("realm: commit PostgreSQL membership departure: %w", err)
	}
	return cloneDepartureReceipt(receipt), nil
}

func (store *PostgresMemberships) MarkWorkerRemoved(ctx context.Context, gameID, playerID string) (departureReceipt, error) {
	if store == nil || store.pool == nil {
		return departureReceipt{}, ErrMembership
	}
	var receipt departureReceipt
	err := pgx.BeginFunc(ctx, store.pool, func(tx pgx.Tx) error {
		record, err := scanPostgresMembership(tx.QueryRow(ctx, postgresMembershipSelect+
			` WHERE game_id = $1 AND player_id = $2 FOR UPDATE`, strings.TrimSpace(gameID), strings.TrimSpace(playerID)))
		if err != nil || record.State != MembershipDeparted || record.Departure == nil {
			return ErrMembership
		}
		receipt = cloneDepartureReceipt(*record.Departure)
		if receipt.WorkerRemoved {
			return nil
		}
		receipt.WorkerRemoved = true
		encoded, err := json.Marshal(receipt)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE realm_memberships SET departure_receipt = $3,
			updated_at = CURRENT_TIMESTAMP WHERE game_id = $1 AND player_id = $2`, record.GameID, record.PlayerID, encoded)
		return err
	})
	if err != nil {
		return departureReceipt{}, fmt.Errorf("realm: mark PostgreSQL membership worker removed: %w", err)
	}
	return receipt, nil
}

func (store *PostgresMemberships) ByAccount(ctx context.Context, gameID, accountID string) (MembershipRecord, error) {
	return store.get(ctx, `game_id = $1 AND account_id = $2`, strings.TrimSpace(gameID), strings.TrimSpace(accountID))
}

func (store *PostgresMemberships) ByCharacter(ctx context.Context, gameID, characterID string) (MembershipRecord, error) {
	return store.get(ctx, `game_id = $1 AND character_id = $2`, strings.TrimSpace(gameID), strings.TrimSpace(characterID))
}

func (store *PostgresMemberships) ByPlayer(ctx context.Context, gameID, playerID string) (MembershipRecord, error) {
	return store.get(ctx, `game_id = $1 AND player_id = $2`, strings.TrimSpace(gameID), strings.TrimSpace(playerID))
}

func (store *PostgresMemberships) ActivePlayerIDs(ctx context.Context, gameID string) ([]string, error) {
	return store.playerIDs(ctx, gameID, false)
}

func (store *PostgresMemberships) DrainPlayerIDs(ctx context.Context, gameID string) ([]string, error) {
	return store.playerIDs(ctx, gameID, true)
}

func (store *PostgresMemberships) playerIDs(ctx context.Context, gameID string, includeDeparted bool) ([]string, error) {
	if store == nil || store.pool == nil || strings.TrimSpace(gameID) == "" {
		return nil, ErrMembership
	}
	states := []string{string(MembershipActive)}
	if includeDeparted {
		states = append(states, string(MembershipDeparted))
	}
	rows, err := store.pool.Query(ctx, `SELECT player_id FROM realm_memberships
		WHERE game_id = $1 AND state = ANY($2) ORDER BY player_id`, strings.TrimSpace(gameID), states)
	if err != nil {
		return nil, fmt.Errorf("realm: list PostgreSQL memberships: %w", err)
	}
	defer rows.Close()
	result := make([]string, 0)
	for rows.Next() {
		var playerID string
		if err := rows.Scan(&playerID); err != nil {
			return nil, err
		}
		result = append(result, playerID)
	}
	return result, rows.Err()
}

// ResumeGame atomically replaces every active membership's lost raw lease
// token. Membership rows intentionally contain no bearer credential; a Realm
// process can therefore recover authority without making those credentials
// durable or accepting client-supplied lease material.
func (store *PostgresMemberships) ResumeGame(ctx context.Context, gameID string, lifetime time.Duration) ([]MembershipRecord, error) {
	gameID = strings.TrimSpace(gameID)
	if store == nil || store.pool == nil || gameID == "" || lifetime <= 0 {
		return nil, ErrMembership
	}
	result := make([]MembershipRecord, 0)
	err := pgx.BeginFunc(ctx, store.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `SELECT m.game_id, m.player_id, m.account_id, m.baseline, m.character_id
			FROM realm_memberships m
			WHERE m.game_id = $1 AND m.state = $2
			ORDER BY m.player_id
			FOR UPDATE OF m`, gameID, MembershipActive)
		if err != nil {
			return err
		}
		defer rows.Close()
		type resumable struct {
			record      MembershipRecord
			characterID string
		}
		locked := make([]resumable, 0)
		for rows.Next() {
			var record MembershipRecord
			var baseline []byte
			var characterID string
			if err := rows.Scan(&record.GameID, &record.PlayerID, &record.AccountID, &baseline, &characterID); err != nil {
				return err
			}
			if err := json.Unmarshal(baseline, &record.Baseline); err != nil || record.Baseline.Character.ID != characterID {
				return ErrMembership
			}
			locked = append(locked, resumable{record: record, characterID: characterID})
		}
		if err := rows.Err(); err != nil {
			return err
		}
		rows.Close()
		for _, candidate := range locked {
			var revision uint64
			var leaseGameID string
			if err := tx.QueryRow(ctx, `SELECT revision, game_id FROM realm_character_leases
				WHERE character_id = $1 FOR UPDATE`, candidate.characterID).Scan(&revision, &leaseGameID); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return ErrLease
				}
				return err
			}
			if leaseGameID != gameID || candidate.record.Baseline.Revision != revision {
				return ErrLease
			}
			lease, err := newCharacterLease(candidate.characterID, revision, gameID, time.Now().UTC().Add(lifetime))
			if err != nil {
				return err
			}
			digest := sha256.Sum256([]byte(lease.Token))
			command, err := tx.Exec(ctx, `UPDATE realm_character_leases
				SET token_digest = $1, expires_at = $2
				WHERE character_id = $3 AND revision = $4 AND game_id = $5`,
				digest[:], lease.ExpiresAt, lease.CharacterID, lease.Revision, lease.GameID)
			if err != nil {
				return err
			}
			if command.RowsAffected() != 1 {
				return ErrLease
			}
			candidate.record.Lease, candidate.record.State = lease, MembershipActive
			result = append(result, cloneMembershipRecord(candidate.record))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("realm: resume PostgreSQL game memberships: %w", err)
	}
	return result, nil
}

func (store *PostgresMemberships) get(ctx context.Context, predicate string, values ...any) (MembershipRecord, error) {
	if store == nil || store.pool == nil {
		return MembershipRecord{}, ErrMembership
	}
	record, err := scanPostgresMembership(store.pool.QueryRow(ctx, postgresMembershipSelect+` WHERE `+predicate, values...))
	if errors.Is(err, pgx.ErrNoRows) {
		return MembershipRecord{}, ErrMembership
	}
	if err != nil {
		return MembershipRecord{}, fmt.Errorf("realm: read PostgreSQL membership: %w", err)
	}
	return record, nil
}

func (store *PostgresMemberships) AbandonGame(ctx context.Context, gameID string) error {
	if store == nil || store.pool == nil {
		return ErrMembership
	}
	err := pgx.BeginFunc(ctx, store.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `UPDATE realm_memberships SET state = $2,
			updated_at = CURRENT_TIMESTAMP WHERE game_id = $1 AND state = $3`, strings.TrimSpace(gameID),
			MembershipAbandoned, MembershipActive); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `UPDATE realm_memberships
			SET departure_receipt = jsonb_set(departure_receipt, '{worker_removed}', 'true'::jsonb, true),
			updated_at = CURRENT_TIMESTAMP
			WHERE game_id = $1 AND state = $2 AND departure_receipt IS NOT NULL`, strings.TrimSpace(gameID), MembershipDeparted)
		return err
	})
	if err != nil {
		return fmt.Errorf("realm: abandon PostgreSQL game memberships: %w", err)
	}
	return nil
}

const postgresMembershipSelect = `SELECT game_id, player_id, account_id, state, baseline, departure_receipt
	FROM realm_memberships`

func scanPostgresMembership(row postgresRowScanner) (MembershipRecord, error) {
	var record MembershipRecord
	var baseline, receipt []byte
	if err := row.Scan(&record.GameID, &record.PlayerID, &record.AccountID, &record.State, &baseline, &receipt); err != nil {
		return MembershipRecord{}, err
	}
	if err := json.Unmarshal(baseline, &record.Baseline); err != nil || record.Baseline.Character.ID == "" {
		return MembershipRecord{}, ErrMembership
	}
	if len(receipt) != 0 {
		var departure departureReceipt
		if err := json.Unmarshal(receipt, &departure); err != nil || departure.Record.Character.ID == "" {
			return MembershipRecord{}, ErrMembership
		}
		record.Departure = &departure
	}
	return cloneMembershipRecord(record), nil
}

var _ MembershipRepository = (*PostgresMemberships)(nil)
