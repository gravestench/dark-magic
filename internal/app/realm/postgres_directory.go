package realm

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const (
	activeRealmGameState   = "active"
	drainingRealmGameState = "draining"
)

// PostgresGameDirectory persists the public directory and its private capacity
// reservations. Raw game passwords and reservation capabilities are never
// stored; only bcrypt and SHA-256 digests cross the repository boundary.
type PostgresGameDirectory struct {
	pool *pgxpool.Pool
	now  func() time.Time
}

type postgresGame struct {
	entry        GameDirectoryEntry
	request      CreateGameRequest
	passwordHash []byte
	players      []GamePlayer
}

// Create emits the canonical durable directory representation so persisted and transported values retain one stable
// shape.
func (directory *PostgresGameDirectory) Create(
	ctx context.Context,
	principal AuthenticatedPrincipal,
	request CreateGameRequest,
) (GameDetail, error) {
	request.Expansion = true

	if err := contextErr(ctx); err != nil {
		return GameDetail{}, err
	}

	if directory == nil || directory.pool == nil || !principal.valid() {
		return GameDetail{}, ErrGameDirectoryInput
	}

	displayName, normalizedName, err := normalizeGameName(request.Name)
	if err != nil || validateCreateGame(request) != nil {
		return GameDetail{}, ErrGameDirectoryInput
	}

	var passwordHash []byte
	if request.Password != "" {
		passwordHash, err = bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
		if err != nil {
			return GameDetail{}, err
		}
	}

	request.Name, request.Description, request.Password = displayName, strings.TrimSpace(request.Description), ""

	specification, err := json.Marshal(request)
	if err != nil {
		return GameDetail{}, fmt.Errorf("realm: encode game specification: %w", err)
	}

	createdAt := time.Now().UTC()
	if directory.now != nil {
		createdAt = directory.now().UTC()
	}

	gameID := uuid.New().String()

	_, err = directory.pool.Exec(ctx, `INSERT INTO realm_games
		(id, state, specification, revision, normalized_name, owner_account_id, created_by,`+
		` password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, 1, $4, $5, $6, $7, $8, $8)`, gameID, activeRealmGameState, specification,
		normalizedName, principal.accountID, principal.name, nullableBytes(passwordHash), createdAt)
	if postgresUniqueViolation(err) {
		return GameDetail{}, ErrGameExists
	}

	if err != nil {
		return GameDetail{}, fmt.Errorf("realm: create PostgreSQL game: %w", err)
	}

	return GameDetail{Entry: gameEntry(gameID, 1, request, principal.name, len(passwordHash) != 0, 0, createdAt)}, nil
}

// List executes list through the PostgreSQL directory store so SQL encoding and database error translation remain
// centralized.
func (directory *PostgresGameDirectory) List(ctx context.Context, filter GameFilter) ([]GameDirectoryEntry, error) {
	if err := contextErr(ctx); err != nil {
		return nil, err
	}

	if directory == nil || directory.pool == nil {
		return nil, ErrGameDirectoryInput
	}

	rows, err := directory.pool.Query(ctx, postgresActiveGamesQuery+` ORDER BY g.created_at, g.id`)
	if err != nil {
		return nil, fmt.Errorf("realm: list PostgreSQL games: %w", err)
	}
	defer rows.Close()

	result := make([]GameDirectoryEntry, 0)

	for rows.Next() {
		game, scanErr := scanPostgresGame(rows)
		if scanErr != nil {
			return nil, scanErr
		}

		entry := game.entry
		if entry.PasswordRequired ||
			filter.Difficulty != nil && entry.Difficulty != *filter.Difficulty ||
			filter.Expansion != nil && entry.Expansion != *filter.Expansion ||
			filter.Hardcore != nil && entry.Hardcore != *filter.Hardcore {
			continue
		}

		result = append(result, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("realm: list PostgreSQL games: %w", err)
	}

	return result, nil
}

// Detail executes detail through the PostgreSQL directory store so SQL encoding and database error translation remain
// centralized.
func (directory *PostgresGameDirectory) Detail(ctx context.Context, reference string) (GameDetail, error) {
	game, err := directory.resolve(ctx, reference, false)
	if err != nil {
		return GameDetail{}, err
	}

	if game.entry.PasswordRequired {
		return GameDetail{}, ErrGameNotFound
	}

	return GameDetail{Entry: game.entry, Players: game.players}, nil
}

// admissionDetail executes admission detail through the PostgreSQL directory store so SQL encoding and database error
// translation remain centralized.
func (directory *PostgresGameDirectory) admissionDetail(ctx context.Context, gameID string) (GameDetail, error) {
	game, err := directory.resolve(ctx, strings.TrimSpace(gameID), true)
	if err != nil {
		return GameDetail{}, err
	}

	return GameDetail{Entry: game.entry, Players: game.players}, nil
}

// ResolveJoin executes resolve join through the PostgreSQL directory store so SQL encoding and database error
// translation remain centralized.
func (directory *PostgresGameDirectory) ResolveJoin(ctx context.Context, reference, password string) (string, error) {
	if len(password) > maximumGamePasswordBytes {
		return "", ErrGameDirectoryInput
	}

	game, err := directory.resolve(ctx, reference, false)
	if err != nil {
		return "", err
	}

	if len(game.passwordHash) == 0 {
		if password != "" {
			return "", ErrGamePassword
		}
	} else if bcrypt.CompareHashAndPassword(game.passwordHash, []byte(password)) != nil {
		return "", ErrGamePassword
	}

	if game.entry.Players >= game.entry.MaximumPlayers {
		return "", ErrGameFull
	}

	var reservations int
	if err := directory.pool.QueryRow(ctx, `SELECT count(*) FROM realm_game_reservations WHERE game_id = $1`,
		game.entry.GameID).Scan(&reservations); err != nil {
		return "", fmt.Errorf("realm: count PostgreSQL game reservations: %w", err)
	}

	if game.entry.Players+reservations >= game.entry.MaximumPlayers {
		return "", ErrGameFull
	}

	return game.entry.GameID, nil
}

// ReservePlayer emits the canonical durable directory representation so persisted and transported values retain one
// stable shape.
func (directory *PostgresGameDirectory) ReservePlayer(
	ctx context.Context,
	gameID string,
	player GamePlayer,
) (GamePlayerReservation, error) {
	if directory == nil || directory.pool == nil {
		return GamePlayerReservation{}, ErrGameDirectoryInput
	}

	if _, err := validateGamePlayers([]GamePlayer{player}); err != nil {
		return GamePlayerReservation{}, err
	}

	token := uuid.New().String()
	digest := sha256.Sum256([]byte(token))

	payload, err := json.Marshal(player)
	if err != nil {
		return GamePlayerReservation{}, fmt.Errorf("realm: encode game player reservation: %w", err)
	}

	err = pgx.BeginFunc(ctx, directory.pool, func(tx pgx.Tx) error {
		var maximum, players, reservations int
		if err := tx.QueryRow(ctx, `SELECT (specification->>'maximum_players')::integer,
			(SELECT count(*) FROM realm_game_players WHERE game_id = g.id),
			(SELECT count(*) FROM realm_game_reservations WHERE game_id = g.id)
			FROM realm_games g WHERE id = $1 AND state = 'active' FOR UPDATE`, strings.TrimSpace(gameID)).
			Scan(&maximum, &players, &reservations); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrGameNotFound
			}

			return fmt.Errorf("realm: lock PostgreSQL game capacity: %w", err)
		}

		if players+reservations >= maximum {
			return ErrGameFull
		}

		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (
			SELECT 1 FROM realm_game_players WHERE game_id = $1 AND character_id = $2
		)`, gameID, player.CharacterID).Scan(&exists); err != nil {
			return fmt.Errorf("realm: inspect PostgreSQL game player: %w", err)
		}

		if exists {
			return ErrCharacterLeased
		}

		_, err := tx.Exec(ctx, `INSERT INTO realm_game_reservations
			(token_digest, game_id, character_id, player) VALUES ($1, $2, $3, $4)`,
			digest[:], gameID, player.CharacterID, payload)
		if postgresUniqueViolation(err) {
			return ErrCharacterLeased
		}

		if err != nil {
			return fmt.Errorf("realm: reserve PostgreSQL game capacity: %w", err)
		}

		return nil
	})
	if err != nil {
		return GamePlayerReservation{}, err
	}

	return GamePlayerReservation{GameID: strings.TrimSpace(gameID), Token: token}, nil
}

// CommitPlayer executes commit player through the PostgreSQL directory store so SQL encoding and database error
// translation remain centralized.
func (directory *PostgresGameDirectory) CommitPlayer(
	ctx context.Context,
	reservation GamePlayerReservation,
) (GameDetail, error) {
	if directory == nil || directory.pool == nil || strings.TrimSpace(reservation.Token) == "" {
		return GameDetail{}, ErrGameDirectoryInput
	}

	digest := sha256.Sum256([]byte(reservation.Token))

	err := pgx.BeginFunc(ctx, directory.pool, func(tx pgx.Tx) error {
		var (
			characterID string
			payload     []byte
		)
		if err := tx.QueryRow(ctx, `DELETE FROM realm_game_reservations
			WHERE game_id = $1 AND token_digest = $2 RETURNING character_id, player`,
			strings.TrimSpace(reservation.GameID), digest[:]).Scan(&characterID, &payload); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrGameDirectoryInput
			}

			return fmt.Errorf("realm: consume PostgreSQL game reservation: %w", err)
		}

		if _, err := tx.Exec(ctx, `INSERT INTO realm_game_players (game_id, character_id, player)
			VALUES ($1, $2, $3)`, reservation.GameID, characterID, payload); err != nil {
			if postgresUniqueViolation(err) {
				return ErrCharacterLeased
			}

			return fmt.Errorf("realm: commit PostgreSQL game player: %w", err)
		}

		command, err := tx.Exec(ctx, `UPDATE realm_games SET revision = revision + 1, updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND state = 'active'`, reservation.GameID)
		if err != nil {
			return fmt.Errorf("realm: revise PostgreSQL game: %w", err)
		}

		if command.RowsAffected() != 1 {
			return ErrGameNotFound
		}

		return nil
	})
	if err != nil {
		return GameDetail{}, err
	}

	return directory.admissionDetail(ctx, reservation.GameID)
}

// CancelPlayer executes cancel player through the PostgreSQL directory store so SQL encoding and database error
// translation remain centralized.
func (directory *PostgresGameDirectory) CancelPlayer(ctx context.Context, reservation GamePlayerReservation) error {
	if directory == nil || directory.pool == nil || strings.TrimSpace(reservation.Token) == "" {
		return ErrGameDirectoryInput
	}

	digest := sha256.Sum256([]byte(reservation.Token))

	command, err := directory.pool.Exec(
		ctx,
		`DELETE FROM realm_game_reservations WHERE game_id = $1 AND token_digest = $2`,
		strings.TrimSpace(reservation.GameID),
		digest[:],
	)
	if err != nil {
		return fmt.Errorf("realm: cancel PostgreSQL game reservation: %w", err)
	}

	if command.RowsAffected() != 1 {
		return ErrGameDirectoryInput
	}

	return nil
}

// SetPlayers emits the canonical durable directory representation so persisted and transported values retain one
// stable shape.
func (directory *PostgresGameDirectory) SetPlayers(ctx context.Context, gameID string, players []GamePlayer) error {
	if directory == nil || directory.pool == nil {
		return ErrGameDirectoryInput
	}

	cloned, err := validateGamePlayers(players)
	if err != nil {
		return err
	}

	return pgx.BeginFunc(ctx, directory.pool, func(tx pgx.Tx) error {
		var maximum int
		if err := tx.QueryRow(ctx, `SELECT (specification->>'maximum_players')::integer FROM realm_games
			WHERE id = $1 AND state <> 'completed' FOR UPDATE`, strings.TrimSpace(gameID)).Scan(&maximum); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrGameNotFound
			}

			return fmt.Errorf("realm: lock PostgreSQL game roster: %w", err)
		}

		if len(cloned) > maximum {
			return ErrGameFull
		}

		if _, err := tx.Exec(ctx, `DELETE FROM realm_game_players WHERE game_id = $1`, gameID); err != nil {
			return fmt.Errorf("realm: clear PostgreSQL game roster: %w", err)
		}

		for _, player := range cloned {
			payload, marshalErr := json.Marshal(player)
			if marshalErr != nil {
				return marshalErr
			}

			if _, err := tx.Exec(ctx, `INSERT INTO realm_game_players (game_id, character_id, player)
				VALUES ($1, $2, $3)`, gameID, player.CharacterID, payload); err != nil {
				return fmt.Errorf("realm: replace PostgreSQL game roster: %w", err)
			}
		}

		_, err := tx.Exec(
			ctx,
			`UPDATE realm_games SET revision = revision + 1, updated_at = CURRENT_TIMESTAMP WHERE id = $1`,
			gameID,
		)

		return err
	})
}

// BeginDrain executes begin drain through the PostgreSQL directory store so SQL encoding and database error
// translation remain centralized.
func (directory *PostgresGameDirectory) BeginDrain(ctx context.Context, gameID string) error {
	if directory == nil || directory.pool == nil {
		return ErrGameDirectoryInput
	}

	return pgx.BeginFunc(ctx, directory.pool, func(tx pgx.Tx) error {
		var state string
		if err := tx.QueryRow(ctx, `SELECT state FROM realm_games WHERE id = $1 FOR UPDATE`, strings.TrimSpace(gameID)).
			Scan(&state); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrGameNotFound
			}

			return err
		}

		if state == drainingRealmGameState {
			return nil
		}

		if state != activeRealmGameState {
			return ErrGameNotFound
		}

		if _, err := tx.Exec(ctx, `UPDATE realm_games SET state = 'draining', revision = revision + 1,
			updated_at = CURRENT_TIMESTAMP WHERE id = $1`, gameID); err != nil {
			return fmt.Errorf("realm: begin PostgreSQL game drain: %w", err)
		}

		_, err := tx.Exec(ctx, `DELETE FROM realm_game_reservations WHERE game_id = $1`, gameID)

		return err
	})
}

// RemovePlayer executes remove player through the PostgreSQL directory store so SQL encoding and database error
// translation remain centralized.
func (directory *PostgresGameDirectory) RemovePlayer(
	ctx context.Context,
	gameID, characterID string,
) (GameDetail, error) {
	if directory == nil || directory.pool == nil || strings.TrimSpace(characterID) == "" {
		return GameDetail{}, ErrGameDirectoryInput
	}

	err := pgx.BeginFunc(ctx, directory.pool, func(tx pgx.Tx) error {
		command, err := tx.Exec(ctx, `DELETE FROM realm_game_players WHERE game_id = $1 AND character_id = $2`,
			strings.TrimSpace(gameID), strings.TrimSpace(characterID))
		if err != nil {
			return fmt.Errorf("realm: remove PostgreSQL game player: %w", err)
		}

		if command.RowsAffected() != 1 {
			var exists bool
			if queryErr := tx.QueryRow(
				ctx,
				`SELECT EXISTS (SELECT 1 FROM realm_games WHERE id = $1 AND state <> 'completed')`,
				gameID,
			).
				Scan(&exists); queryErr != nil {
				return queryErr
			}

			if !exists {
				return ErrGameNotFound
			}

			return ErrCharacterNotFound
		}

		_, err = tx.Exec(
			ctx,
			`UPDATE realm_games SET revision = revision + 1, updated_at = CURRENT_TIMESTAMP WHERE id = $1`,
			gameID,
		)

		return err
	})
	if err != nil {
		return GameDetail{}, err
	}

	return directory.nonterminalDetail(ctx, gameID)
}

// Remove executes remove through the PostgreSQL directory store so SQL encoding and database error translation remain
// centralized.
func (directory *PostgresGameDirectory) Remove(ctx context.Context, gameID string) error {
	if directory == nil || directory.pool == nil {
		return ErrGameDirectoryInput
	}

	return pgx.BeginFunc(ctx, directory.pool, func(tx pgx.Tx) error {
		command, err := tx.Exec(ctx, `UPDATE realm_games SET state = 'completed', revision = revision + 1,
			updated_at = CURRENT_TIMESTAMP WHERE id = $1 AND state <> 'completed'`, strings.TrimSpace(gameID))
		if err != nil {
			return fmt.Errorf("realm: complete PostgreSQL game: %w", err)
		}

		if command.RowsAffected() != 1 {
			return ErrGameNotFound
		}

		if _, err := tx.Exec(ctx, `DELETE FROM realm_game_reservations WHERE game_id = $1`, gameID); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `DELETE FROM realm_game_players WHERE game_id = $1`, gameID); err != nil {
			return err
		}

		return nil
	})
}

// gameIDs executes game ids through the PostgreSQL directory store so SQL encoding and database error translation
// remain centralized.
func (directory *PostgresGameDirectory) gameIDs(ctx context.Context) ([]string, error) {
	if directory == nil || directory.pool == nil {
		return nil, ErrGameDirectoryInput
	}

	rows, err := directory.pool.Query(ctx, `SELECT id FROM realm_games WHERE state <> 'completed' ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("realm: list PostgreSQL game IDs: %w", err)
	}
	defer rows.Close()

	var result []string

	for rows.Next() {
		var gameID string
		if err := rows.Scan(&gameID); err != nil {
			return nil, err
		}

		result = append(result, gameID)
	}

	return result, rows.Err()
}

// resolve executes resolve through the PostgreSQL directory store so SQL encoding and database error translation
// remain centralized.
func (directory *PostgresGameDirectory) resolve(
	ctx context.Context,
	reference string,
	idOnly bool,
) (postgresGame, error) {
	return directory.resolveQuery(ctx, reference, idOnly, postgresActiveGamesQuery)
}

// nonterminalDetail executes nonterminal detail through the PostgreSQL directory store so SQL encoding and database
// error translation remain centralized.
func (directory *PostgresGameDirectory) nonterminalDetail(ctx context.Context, gameID string) (GameDetail, error) {
	game, err := directory.resolveQuery(ctx, strings.TrimSpace(gameID), true, postgresNonterminalGamesQuery)
	if err != nil {
		return GameDetail{}, err
	}

	return GameDetail{Entry: game.entry, Players: game.players}, nil
}

// resolveQuery executes resolve query through the PostgreSQL directory store so SQL encoding and database error
// translation remain centralized.
func (directory *PostgresGameDirectory) resolveQuery(
	ctx context.Context,
	reference string,
	idOnly bool,
	baseQuery string,
) (postgresGame, error) {
	if err := contextErr(ctx); err != nil {
		return postgresGame{}, err
	}

	if directory == nil || directory.pool == nil {
		return postgresGame{}, ErrGameDirectoryInput
	}

	reference = strings.TrimSpace(reference)
	if reference == "" {
		return postgresGame{}, ErrGameNotFound
	}

	normalized := strings.ToLower(strings.Join(strings.Fields(reference), " "))
	query := baseQuery + ` AND (g.id = $1 OR (NOT $3 AND g.normalized_name = $2))
		ORDER BY (g.id = $1) DESC LIMIT 1`

	game, err := scanPostgresGame(directory.pool.QueryRow(ctx, query, reference, normalized, idOnly))
	if errors.Is(err, pgx.ErrNoRows) {
		return postgresGame{}, ErrGameNotFound
	}

	if err != nil {
		return postgresGame{}, err
	}

	players, err := directory.players(ctx, game.entry.GameID)
	if err != nil {
		return postgresGame{}, err
	}

	game.players = players

	return game, nil
}

// players executes players through the PostgreSQL directory store so SQL encoding and database error translation
// remain centralized.
func (directory *PostgresGameDirectory) players(ctx context.Context, gameID string) ([]GamePlayer, error) {
	rows, err := directory.pool.Query(
		ctx,
		`SELECT player FROM realm_game_players WHERE game_id = $1 ORDER BY joined_at, character_id`,
		gameID,
	)
	if err != nil {
		return nil, fmt.Errorf("realm: list PostgreSQL game players: %w", err)
	}
	defer rows.Close()

	result := make([]GamePlayer, 0)

	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}

		var player GamePlayer
		if err := json.Unmarshal(payload, &player); err != nil {
			return nil, fmt.Errorf("realm: decode PostgreSQL game player: %w", err)
		}

		result = append(result, player)
	}

	return result, rows.Err()
}

const postgresActiveGamesQuery = `SELECT g.id, g.revision, g.specification, g.created_by,
	COALESCE(g.password_hash, ''::bytea), g.created_at,
	(SELECT count(*) FROM realm_game_players p WHERE p.game_id = g.id)
	FROM realm_games g WHERE g.state = 'active'`

const postgresNonterminalGamesQuery = `SELECT g.id, g.revision, g.specification, g.created_by,
	COALESCE(g.password_hash, ''::bytea), g.created_at,
	(SELECT count(*) FROM realm_game_players p WHERE p.game_id = g.id)
	FROM realm_games g WHERE g.state <> 'completed'`

type postgresRowScanner interface {
	Scan(...any) error
}

// scanPostgresGame decodes the durable directory representation at one boundary so malformed data fails before it
// becomes shared state.
func scanPostgresGame(row postgresRowScanner) (postgresGame, error) {
	var (
		game          postgresGame
		specification []byte
		revision      int64
		players       int
	)
	if err := row.Scan(&game.entry.GameID, &revision, &specification, &game.entry.CreatedBy,
		&game.passwordHash, &game.entry.CreatedAt, &players); err != nil {
		return postgresGame{}, err
	}

	if revision <= 0 {
		return postgresGame{}, ErrGameDirectoryInput
	}

	if err := json.Unmarshal(specification, &game.request); err != nil {
		return postgresGame{}, fmt.Errorf("realm: decode PostgreSQL game specification: %w", err)
	}

	game.entry = gameEntry(game.entry.GameID, uint64(revision), game.request, game.entry.CreatedBy,
		len(game.passwordHash) != 0, players, game.entry.CreatedAt)

	return game, nil
}

// gameEntry executes game entry through the PostgreSQL directory store so SQL encoding and database error translation
// remain centralized.
func gameEntry(
	gameID string,
	revision uint64,
	request CreateGameRequest,
	createdBy string,
	passwordRequired bool,
	players int,
	createdAt time.Time,
) GameDirectoryEntry {
	return GameDirectoryEntry{Version: GameDirectoryVersion, Revision: revision, GameID: gameID, Name: request.Name,
		Description: request.Description, CreatedBy: createdBy, Difficulty: request.Difficulty, Players: players,
		MaximumPlayers: request.Maximum, CharacterDifference: request.CharacterDifference,
		PasswordRequired: passwordRequired, Expansion: request.Expansion,
		Hardcore: request.Hardcore, CreatedAt: createdAt.UTC()}
}

// nullableBytes executes nullable bytes through the PostgreSQL directory store so SQL encoding and database error
// translation remain centralized.
func nullableBytes(value []byte) any {
	if len(value) == 0 {
		return nil
	}

	return value
}

var _ GameRepository = (*PostgresGameDirectory)(nil)
var _ GameRepository = (*GameDirectory)(nil)
