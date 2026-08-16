package realm

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

//go:embed schema.sql
var realmSchema string

const realmSchemaLock int64 = 0x444d5245414c4d // "DMREALM"

// Postgres owns the production Realm repository pool. Accounts and Characters
// expose the same semantic interfaces as the deterministic in-memory fixtures.
type Postgres struct {
	Pool        *pgxpool.Pool
	Accounts    *PostgresAccounts
	Characters  *PostgresCharacters
	Games       GameRepository
	Allocations AllocationRepository
	Memberships MembershipRepository
	Checkpoints CheckpointRepository
	Audit       AuditSink
	Mail        MailOutbox
}

func OpenPostgres(ctx context.Context, connectionString string, sessionLifetime time.Duration) (*Postgres, error) {
	if ctx == nil || strings.TrimSpace(connectionString) == "" {
		return nil, errors.New("realm: PostgreSQL connection string is required")
	}
	if sessionLifetime == 0 {
		sessionLifetime = defaultSessionLifetime
	}
	if sessionLifetime < time.Minute {
		return nil, ErrAccountInput
	}
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("realm: parse PostgreSQL configuration: %w", err)
	}
	config.ConnConfig.ConnectTimeout = 5 * time.Second
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("realm: open PostgreSQL: %w", err)
	}
	closeOnFailure := true
	defer func() {
		if closeOnFailure {
			pool.Close()
		}
	}()
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("realm: ping PostgreSQL: %w", err)
	}
	if err := initializeRealmPostgres(ctx, pool); err != nil {
		return nil, err
	}
	dummyHash, err := bcrypt.GenerateFromPassword([]byte("dark-magic-invalid-password"), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	closeOnFailure = false
	accounts := &PostgresAccounts{pool: pool, sessionLifetime: sessionLifetime, dummyHash: dummyHash,
		verificationLifetime: defaultVerificationLifetime, recoveryLifetime: defaultRecoveryLifetime,
		accountBaseURL: "https://accounts.dark-magic.test"}
	return &Postgres{Pool: pool, Accounts: accounts,
		Characters: &PostgresCharacters{pool: pool}, Games: &PostgresGameDirectory{pool: pool},
		Allocations: &PostgresAllocations{pool: pool}, Memberships: &PostgresMemberships{pool: pool},
		Checkpoints: &PostgresCheckpoints{pool: pool}, Audit: postgresAuditSink{pool: pool},
		Mail: &PostgresMailOutbox{pool: pool}}, nil
}

func (postgres *Postgres) Close() {
	if postgres != nil && postgres.Pool != nil {
		postgres.Pool.Close()
	}
}

func initializeRealmPostgres(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return fmt.Errorf("realm: begin PostgreSQL schema initialization: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, realmSchemaLock); err != nil {
		return fmt.Errorf("realm: lock PostgreSQL schema initialization: %w", err)
	}
	if _, err := tx.Exec(ctx, realmSchema); err != nil {
		return fmt.Errorf("realm: initialize PostgreSQL schema (run make realm-fresh-install after incompatible pre-production changes): %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("realm: commit PostgreSQL schema initialization: %w", err)
	}
	return nil
}

type PostgresAccounts struct {
	pool                 *pgxpool.Pool
	sessionLifetime      time.Duration
	verificationLifetime time.Duration
	recoveryLifetime     time.Duration
	accountBaseURL       string
	dummyHash            []byte
}

// SetAccountBaseURL configures the public HTTPS origin embedded in account
// verification and recovery mail. It rejects HTTP even for local operation;
// the local profile must use a developer-trusted certificate.
func (accounts *PostgresAccounts) SetAccountBaseURL(value string) error {
	if _, err := accountActionURL(value, "/health", "validation"); err != nil {
		return err
	}
	accounts.accountBaseURL = strings.TrimRight(strings.TrimSpace(value), "/")
	return nil
}

type postgresAuditSink struct{ pool *pgxpool.Pool }

func (sink postgresAuditSink) Record(ctx context.Context, event AuditEvent) {
	if sink.pool == nil {
		return
	}
	payload, err := json.Marshal(event)
	if err != nil {
		slog.Error("encoding Realm PostgreSQL audit event", "error", err)
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	auditCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer cancel()
	if _, err := sink.pool.Exec(auditCtx, `INSERT INTO realm_audit_events (event) VALUES ($1)`, payload); err != nil {
		slog.Error("persisting Realm PostgreSQL audit event", "operation", event.Operation, "error", err)
	}
}

func (accounts *PostgresAccounts) Create(ctx context.Context, name, password string) (Account, error) {
	displayName, normalizedName, err := validateAccountName(name)
	if err != nil || len(password) < minimumPasswordBytes || len(password) > maximumPasswordBytes {
		return Account{}, ErrAccountInput
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return Account{}, err
	}
	account := Account{ID: uuid.New().String(), Name: displayName, EmailVerified: true, CreatedAt: time.Now().UTC()}
	_, err = accounts.pool.Exec(ctx, `INSERT INTO realm_accounts
		(id, name, normalized_name, password_hash, email_verified_at, created_at) VALUES ($1, $2, $3, $4, $5, $5)`,
		account.ID, account.Name, normalizedName, passwordHash, account.CreatedAt)
	if postgresUniqueViolation(err) {
		return Account{}, ErrAccountExists
	}
	if err != nil {
		return Account{}, fmt.Errorf("realm: create PostgreSQL account: %w", err)
	}
	return account, nil
}

func (accounts *PostgresAccounts) Authenticate(ctx context.Context, name, password string) (RealmSession, error) {
	if len(password) > maximumPasswordBytes {
		return RealmSession{}, ErrAccountCredentials
	}
	_, normalizedName, nameErr := validateAccountName(name)
	var account Account
	var passwordHash []byte
	var verifiedAt *time.Time
	err := accounts.pool.QueryRow(ctx, `SELECT id, name, created_at, password_hash, email_verified_at
        FROM realm_accounts WHERE normalized_name = $1`, normalizedName).
		Scan(&account.ID, &account.Name, &account.CreatedAt, &passwordHash, &verifiedAt)
	missing := errors.Is(err, pgx.ErrNoRows)
	if err != nil && !missing {
		return RealmSession{}, fmt.Errorf("realm: read PostgreSQL account: %w", err)
	}
	hash := accounts.dummyHash
	if !missing {
		hash = passwordHash
	}
	passwordErr := bcrypt.CompareHashAndPassword(hash, []byte(password))
	if nameErr != nil || passwordErr != nil || missing {
		return RealmSession{}, ErrAccountCredentials
	}
	if verifiedAt == nil {
		return RealmSession{}, ErrAccountUnverified
	}
	account.EmailVerified = true
	token, digest, err := newRealmSessionToken()
	if err != nil {
		return RealmSession{}, err
	}
	session := RealmSession{ID: uuid.New().String(), Account: account, Token: token,
		ExpiresAt: time.Now().Add(accounts.sessionLifetime).UTC()}
	_, err = accounts.pool.Exec(ctx, `INSERT INTO realm_sessions
        (id, account_id, token_digest, expires_at) VALUES ($1, $2, $3, $4)`,
		session.ID, account.ID, digest[:], session.ExpiresAt)
	if err != nil {
		return RealmSession{}, fmt.Errorf("realm: create PostgreSQL session: %w", err)
	}
	return session, nil
}

func (accounts *PostgresAccounts) Authorize(ctx context.Context, token string) (AuthenticatedPrincipal, error) {
	if strings.TrimSpace(token) == "" {
		return AuthenticatedPrincipal{}, ErrRealmSession
	}
	digest := sha256.Sum256([]byte(token))
	var principal AuthenticatedPrincipal
	err := accounts.pool.QueryRow(ctx, `SELECT a.id, a.name, s.id
        FROM realm_sessions s JOIN realm_accounts a ON a.id = s.account_id
        WHERE s.token_digest = $1 AND s.expires_at > CURRENT_TIMESTAMP`, digest[:]).
		Scan(&principal.accountID, &principal.name, &principal.sessionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return AuthenticatedPrincipal{}, ErrRealmSession
	}
	if err != nil {
		return AuthenticatedPrincipal{}, fmt.Errorf("realm: authorize PostgreSQL session: %w", err)
	}
	return principal, nil
}

func (accounts *PostgresAccounts) SelectCharacter(ctx context.Context, token, characterID string) error {
	if strings.TrimSpace(token) == "" || strings.TrimSpace(characterID) == "" {
		return ErrRealmSession
	}
	digest := sha256.Sum256([]byte(token))
	result, err := accounts.pool.Exec(ctx, `UPDATE realm_sessions SET selected_character_id = $1
        WHERE token_digest = $2 AND expires_at > CURRENT_TIMESTAMP`, characterID, digest[:])
	if err != nil {
		return fmt.Errorf("realm: select PostgreSQL character: %w", err)
	}
	if result.RowsAffected() != 1 {
		return ErrRealmSession
	}
	return nil
}

func (accounts *PostgresAccounts) SelectedCharacter(ctx context.Context, token string) (string, error) {
	if strings.TrimSpace(token) == "" {
		return "", ErrRealmSession
	}
	digest := sha256.Sum256([]byte(token))
	var characterID *string
	err := accounts.pool.QueryRow(ctx, `SELECT selected_character_id FROM realm_sessions
        WHERE token_digest = $1 AND expires_at > CURRENT_TIMESTAMP`, digest[:]).Scan(&characterID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrRealmSession
	}
	if err != nil {
		return "", fmt.Errorf("realm: read selected PostgreSQL character: %w", err)
	}
	if characterID == nil || *characterID == "" {
		return "", ErrCharacterNotFound
	}
	return *characterID, nil
}

func (accounts *PostgresAccounts) Logout(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return ErrRealmSession
	}
	digest := sha256.Sum256([]byte(token))
	result, err := accounts.pool.Exec(ctx, `DELETE FROM realm_sessions
        WHERE token_digest = $1 AND expires_at > CURRENT_TIMESTAMP`, digest[:])
	if err != nil {
		return fmt.Errorf("realm: delete PostgreSQL session: %w", err)
	}
	if result.RowsAffected() != 1 {
		return ErrRealmSession
	}
	return nil
}

func (accounts *PostgresAccounts) PruneExpired(ctx context.Context) ([]AuthenticatedPrincipal, error) {
	rows, err := accounts.pool.Query(ctx, `WITH expired AS (
        DELETE FROM realm_sessions WHERE expires_at <= CURRENT_TIMESTAMP RETURNING id, account_id
    ) SELECT a.id, a.name, expired.id FROM expired JOIN realm_accounts a ON a.id = expired.account_id
       ORDER BY expired.id`)
	if err != nil {
		return nil, fmt.Errorf("realm: prune PostgreSQL sessions: %w", err)
	}
	defer rows.Close()
	var principals []AuthenticatedPrincipal
	for rows.Next() {
		var principal AuthenticatedPrincipal
		if err := rows.Scan(&principal.accountID, &principal.name, &principal.sessionID); err != nil {
			return nil, fmt.Errorf("realm: scan expired PostgreSQL session: %w", err)
		}
		principals = append(principals, principal)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("realm: read expired PostgreSQL sessions: %w", err)
	}
	return principals, nil
}

type PostgresCharacters struct{ pool *pgxpool.Pool }

func (repository *PostgresCharacters) Create(ctx context.Context, record CharacterRecord) error {
	if strings.TrimSpace(record.AccountID) == "" || strings.TrimSpace(record.Character.ID) == "" || record.Revision == 0 {
		return errors.New("realm: character record requires account, character, and revision")
	}
	character, compatibility, err := encodePostgresCharacter(record)
	if err != nil {
		return err
	}
	_, err = repository.pool.Exec(ctx, `INSERT INTO realm_characters
        (id, account_id, revision, character, compatibility) VALUES ($1, $2, $3, $4, $5)`,
		record.Character.ID, record.AccountID, record.Revision, character, compatibility)
	if postgresUniqueViolation(err) {
		return fmt.Errorf("realm: duplicate character %q", record.Character.ID)
	}
	if err != nil {
		return fmt.Errorf("realm: create PostgreSQL character: %w", err)
	}
	return nil
}

func (repository *PostgresCharacters) Delete(ctx context.Context, accountID, characterID string) error {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var owner string
	if err := tx.QueryRow(ctx, `SELECT account_id FROM realm_characters WHERE id = $1 FOR UPDATE`, characterID).Scan(&owner); errors.Is(err, pgx.ErrNoRows) {
		return ErrCharacterNotFound
	} else if err != nil {
		return fmt.Errorf("realm: lock PostgreSQL character: %w", err)
	}
	if owner != accountID {
		return ErrCharacterOwner
	}
	var leased bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM realm_character_leases
        WHERE character_id = $1 AND expires_at > CURRENT_TIMESTAMP)`, characterID).Scan(&leased); err != nil {
		return err
	}
	if leased {
		return ErrCharacterLeased
	}
	if _, err := tx.Exec(ctx, `DELETE FROM realm_characters WHERE id = $1`, characterID); err != nil {
		return fmt.Errorf("realm: delete PostgreSQL character: %w", err)
	}
	return tx.Commit(ctx)
}

func (repository *PostgresCharacters) Get(ctx context.Context, accountID, characterID string) (CharacterRecord, error) {
	record, err := scanPostgresCharacter(repository.pool.QueryRow(ctx, `SELECT account_id, revision, character, compatibility
        FROM realm_characters WHERE id = $1`, characterID))
	if errors.Is(err, pgx.ErrNoRows) {
		return CharacterRecord{}, ErrCharacterNotFound
	}
	if err != nil {
		return CharacterRecord{}, fmt.Errorf("realm: read PostgreSQL character: %w", err)
	}
	if record.AccountID != accountID {
		return CharacterRecord{}, ErrCharacterOwner
	}
	return record, nil
}

func (repository *PostgresCharacters) List(ctx context.Context, accountID string) ([]CharacterRecord, error) {
	if strings.TrimSpace(accountID) == "" {
		return nil, ErrCharacterOwner
	}
	rows, err := repository.pool.Query(ctx, `SELECT account_id, revision, character, compatibility
        FROM realm_characters WHERE account_id = $1 ORDER BY id`, accountID)
	if err != nil {
		return nil, fmt.Errorf("realm: list PostgreSQL characters: %w", err)
	}
	defer rows.Close()
	var records []CharacterRecord
	for rows.Next() {
		record, err := scanPostgresCharacter(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (repository *PostgresCharacters) Acquire(ctx context.Context, accountID, characterID, gameID string, lifetime time.Duration) (CharacterRecord, CharacterLease, error) {
	if strings.TrimSpace(accountID) == "" || strings.TrimSpace(characterID) == "" || strings.TrimSpace(gameID) == "" || lifetime <= 0 {
		return CharacterRecord{}, CharacterLease{}, ErrLease
	}
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return CharacterRecord{}, CharacterLease{}, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	record, err := scanPostgresCharacter(tx.QueryRow(ctx, `SELECT account_id, revision, character, compatibility
        FROM realm_characters WHERE id = $1 FOR UPDATE`, characterID))
	if errors.Is(err, pgx.ErrNoRows) {
		return CharacterRecord{}, CharacterLease{}, ErrCharacterNotFound
	}
	if err != nil {
		return CharacterRecord{}, CharacterLease{}, err
	}
	if record.AccountID != accountID {
		return CharacterRecord{}, CharacterLease{}, ErrCharacterOwner
	}
	if _, err := tx.Exec(ctx, `DELETE FROM realm_character_leases
        WHERE character_id = $1 AND expires_at <= CURRENT_TIMESTAMP`, characterID); err != nil {
		return CharacterRecord{}, CharacterLease{}, err
	}
	lease, err := newCharacterLease(characterID, record.Revision, gameID, time.Now().Add(lifetime).UTC())
	if err != nil {
		return CharacterRecord{}, CharacterLease{}, err
	}
	digest := sha256.Sum256([]byte(lease.Token))
	result, err := tx.Exec(ctx, `INSERT INTO realm_character_leases
        (character_id, token_digest, revision, game_id, expires_at) VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (character_id) DO NOTHING`, characterID, digest[:], lease.Revision, lease.GameID, lease.ExpiresAt)
	if err != nil {
		return CharacterRecord{}, CharacterLease{}, err
	}
	if result.RowsAffected() != 1 {
		return CharacterRecord{}, CharacterLease{}, ErrCharacterLeased
	}
	if err := tx.Commit(ctx); err != nil {
		return CharacterRecord{}, CharacterLease{}, err
	}
	return record, lease, nil
}

func (repository *PostgresCharacters) BindCompatibility(ctx context.Context, lease CharacterLease, compatibility gamesession.DurableCompatibility) (CharacterRecord, error) {
	if compatibility.CharacterID != lease.CharacterID || strings.TrimSpace(compatibility.ModID) == "" ||
		strings.TrimSpace(compatibility.ContractVersion) == "" || strings.TrimSpace(compatibility.IdentityHash) == "" {
		return CharacterRecord{}, ErrCharacterCommit
	}
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return CharacterRecord{}, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	record, err := repository.lockLeasedCharacter(ctx, tx, lease)
	if err != nil {
		return CharacterRecord{}, ErrCharacterCommit
	}
	if !emptyCompatibility(record.Compatibility) && record.Compatibility != compatibility {
		return CharacterRecord{}, ErrCharacterCommit
	}
	encoded, err := json.Marshal(compatibility)
	if err != nil {
		return CharacterRecord{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE realm_characters SET compatibility = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, encoded, lease.CharacterID); err != nil {
		return CharacterRecord{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CharacterRecord{}, err
	}
	record.Compatibility = compatibility
	return record, nil
}

func (repository *PostgresCharacters) Renew(ctx context.Context, lease CharacterLease, lifetime time.Duration) (CharacterLease, error) {
	if lifetime <= 0 {
		return CharacterLease{}, ErrLease
	}
	digest := sha256.Sum256([]byte(lease.Token))
	err := repository.pool.QueryRow(ctx, `UPDATE realm_character_leases
        SET expires_at = CURRENT_TIMESTAMP + ($1 * INTERVAL '1 microsecond')
        WHERE character_id = $2 AND token_digest = $3 AND revision = $4 AND game_id = $5
          AND expires_at > CURRENT_TIMESTAMP RETURNING expires_at`,
		lifetime.Microseconds(), lease.CharacterID, digest[:], lease.Revision, lease.GameID).Scan(&lease.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return CharacterLease{}, ErrLease
	}
	if err != nil {
		return CharacterLease{}, fmt.Errorf("realm: renew PostgreSQL lease: %w", err)
	}
	return lease, nil
}

func (repository *PostgresCharacters) Release(ctx context.Context, lease CharacterLease) error {
	digest := sha256.Sum256([]byte(lease.Token))
	result, err := repository.pool.Exec(ctx, `DELETE FROM realm_character_leases
        WHERE character_id = $1 AND token_digest = $2 AND revision = $3 AND game_id = $4`,
		lease.CharacterID, digest[:], lease.Revision, lease.GameID)
	if err != nil {
		return fmt.Errorf("realm: release PostgreSQL lease: %w", err)
	}
	if result.RowsAffected() != 1 {
		return ErrLease
	}
	return nil
}

func (repository *PostgresCharacters) ReleaseGame(ctx context.Context, gameID string) (int, error) {
	gameID = strings.TrimSpace(gameID)
	if repository == nil || repository.pool == nil || gameID == "" {
		return 0, ErrLease
	}
	result, err := repository.pool.Exec(ctx, `DELETE FROM realm_character_leases WHERE game_id = $1`, gameID)
	if err != nil {
		return 0, fmt.Errorf("realm: release PostgreSQL game leases: %w", err)
	}
	return int(result.RowsAffected()), nil
}

func (repository *PostgresCharacters) Commit(ctx context.Context, lease CharacterLease, character d2save.Character) (CharacterRecord, error) {
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return CharacterRecord{}, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	record, err := repository.lockLeasedCharacter(ctx, tx, lease)
	if err != nil || strings.TrimSpace(character.ID) == "" || character.ID != record.Character.ID {
		return CharacterRecord{}, ErrCharacterCommit
	}
	encoded, err := json.Marshal(character)
	if err != nil {
		return CharacterRecord{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE realm_characters SET character = $1, revision = revision + 1,
        updated_at = CURRENT_TIMESTAMP WHERE id = $2`, encoded, lease.CharacterID); err != nil {
		return CharacterRecord{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM realm_character_leases WHERE character_id = $1`, lease.CharacterID); err != nil {
		return CharacterRecord{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CharacterRecord{}, err
	}
	record.Character = cloneCharacter(character)
	record.Revision++
	return record, nil
}

func (repository *PostgresCharacters) lockLeasedCharacter(ctx context.Context, tx pgx.Tx, lease CharacterLease) (CharacterRecord, error) {
	digest := sha256.Sum256([]byte(lease.Token))
	return scanPostgresCharacter(tx.QueryRow(ctx, `SELECT c.account_id, c.revision, c.character, c.compatibility
        FROM realm_characters c JOIN realm_character_leases l ON l.character_id = c.id
        WHERE c.id = $1 AND l.token_digest = $2 AND l.revision = $3 AND l.game_id = $4
          AND l.expires_at > CURRENT_TIMESTAMP FOR UPDATE OF c, l`,
		lease.CharacterID, digest[:], lease.Revision, lease.GameID))
}

type postgresRow interface{ Scan(...any) error }

func scanPostgresCharacter(row postgresRow) (CharacterRecord, error) {
	var record CharacterRecord
	var character, compatibility []byte
	if err := row.Scan(&record.AccountID, &record.Revision, &character, &compatibility); err != nil {
		return CharacterRecord{}, err
	}
	if record.Revision == 0 || json.Unmarshal(character, &record.Character) != nil ||
		json.Unmarshal(compatibility, &record.Compatibility) != nil || strings.TrimSpace(record.Character.ID) == "" {
		return CharacterRecord{}, errors.New("realm: malformed PostgreSQL character")
	}
	return cloneCharacterRecord(record), nil
}

func encodePostgresCharacter(record CharacterRecord) ([]byte, []byte, error) {
	character, err := json.Marshal(record.Character)
	if err != nil {
		return nil, nil, err
	}
	compatibility, err := json.Marshal(record.Compatibility)
	if err != nil {
		return nil, nil, err
	}
	return character, compatibility, nil
}

func postgresUniqueViolation(err error) bool {
	var databaseError *pgconn.PgError
	return errors.As(err, &databaseError) && databaseError.Code == "23505"
}
