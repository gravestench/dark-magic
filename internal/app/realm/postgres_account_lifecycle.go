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

// Signup emits the canonical durable account lifecycle representation so persisted and transported values retain one
// stable shape.
func (accounts *PostgresAccounts) Signup(ctx context.Context, request SignupRequest) (Account, error) {
	displayName, normalizedName, err := validateAccountName(request.Name)
	if err != nil || len(request.Password) < minimumPasswordBytes || len(request.Password) > maximumPasswordBytes {
		return Account{}, ErrAccountInput
	}

	email, normalizedEmail, err := normalizeEmail(request.Email)
	if err != nil {
		return Account{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return Account{}, err
	}

	token, digest, err := newRealmSessionToken()
	if err != nil {
		return Account{}, err
	}

	verificationURL, err := accountActionURL(accounts.accountBaseURL, "/verify", token)
	if err != nil {
		return Account{}, err
	}

	now := time.Now().UTC()
	account := Account{ID: uuid.New().String(), Name: displayName, CreatedAt: now}

	payload, err := json.Marshal(map[string]any{"account_name": account.Name, "verification_url": verificationURL})
	if err != nil {
		return Account{}, err
	}

	tx, err := accounts.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return Account{}, fmt.Errorf("realm: begin account signup: %w", err)
	}

	defer func() { _ = tx.Rollback(context.Background()) }()

	_, err = tx.Exec(ctx, `INSERT INTO realm_accounts
		(id, name, normalized_name, email, normalized_email, password_hash, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`, account.ID, account.Name, normalizedName,
		email, normalizedEmail, passwordHash, account.CreatedAt)
	if postgresUniqueViolation(err) {
		return Account{}, ErrAccountExists
	}

	if err != nil {
		return Account{}, fmt.Errorf("realm: insert signup account: %w", err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO realm_account_challenges
		(id, account_id, kind, token_digest, expires_at) VALUES ($1, $2, 'verify_email', $3, $4)`,
		uuid.New().String(), account.ID, digest[:], now.Add(accounts.verificationLifetime)); err != nil {
		return Account{}, fmt.Errorf("realm: insert email verification challenge: %w", err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO realm_mail_outbox
		(id, kind, recipient, payload, state, available_at) VALUES ($1, 'verify_email', $2, $3, 'pending', $4)`,
		uuid.New().String(), email, payload, now); err != nil {
		return Account{}, fmt.Errorf("realm: insert verification mail: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		if postgresUniqueViolation(err) {
			return Account{}, ErrAccountExists
		}

		return Account{}, fmt.Errorf("realm: commit account signup: %w", err)
	}

	return account, nil
}

// VerifyEmail executes verify email through the PostgreSQL account lifecycle store so SQL encoding and database error
// translation remain centralized.
func (accounts *PostgresAccounts) VerifyEmail(ctx context.Context, token string) (Account, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return Account{}, ErrAccountChallenge
	}

	digest := sha256.Sum256([]byte(token))

	tx, err := accounts.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return Account{}, fmt.Errorf("realm: begin email verification: %w", err)
	}

	defer func() { _ = tx.Rollback(context.Background()) }()

	var (
		account     Account
		challengeID string
		expiresAt   time.Time
		consumedAt  *time.Time
	)

	err = tx.QueryRow(ctx, `SELECT c.id, c.expires_at, c.consumed_at, a.id, a.name, a.created_at
		FROM realm_account_challenges c JOIN realm_accounts a ON a.id = c.account_id
		WHERE c.kind = 'verify_email' AND c.token_digest = $1 FOR UPDATE OF c, a`, digest[:]).
		Scan(&challengeID, &expiresAt, &consumedAt, &account.ID, &account.Name, &account.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrAccountChallenge
	}

	if err != nil {
		return Account{}, fmt.Errorf("realm: read email verification challenge: %w", err)
	}

	if consumedAt != nil || !expiresAt.After(time.Now()) {
		return Account{}, ErrAccountChallenge
	}

	now := time.Now().UTC()
	if _, err := tx.Exec(
		ctx,
		`UPDATE realm_accounts SET email_verified_at = COALESCE(email_verified_at, $1) WHERE id = $2`,
		now,
		account.ID,
	); err != nil {
		return Account{}, fmt.Errorf("realm: activate verified account: %w", err)
	}

	if _, err := tx.Exec(
		ctx,
		`UPDATE realm_account_challenges SET consumed_at = $1 WHERE id = $2`,
		now,
		challengeID,
	); err != nil {
		return Account{}, fmt.Errorf("realm: consume email verification challenge: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Account{}, fmt.Errorf("realm: commit email verification: %w", err)
	}

	account.EmailVerified = true

	return account, nil
}

// BeginPasswordRecovery emits the canonical durable account lifecycle representation so persisted and transported
// values retain one stable shape.
func (accounts *PostgresAccounts) BeginPasswordRecovery(ctx context.Context, emailValue string) error {
	_, normalizedEmail, err := normalizeEmail(emailValue)
	if err != nil {
		return err
	}

	var accountID, accountName, recipient string

	err = accounts.pool.QueryRow(ctx, `SELECT id, name, email FROM realm_accounts
		WHERE normalized_email = $1 AND email_verified_at IS NOT NULL`, normalizedEmail).
		Scan(&accountID, &accountName, &recipient)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("realm: find recovery account: %w", err)
	}

	token, digest, err := newRealmSessionToken()
	if err != nil {
		return err
	}

	recoveryURL, err := accountActionURL(accounts.accountBaseURL, "/recover", token)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(map[string]any{"account_name": accountName, "recovery_url": recoveryURL})
	if err != nil {
		return err
	}

	now := time.Now().UTC()

	tx, err := accounts.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return fmt.Errorf("realm: begin password recovery: %w", err)
	}

	defer func() { _ = tx.Rollback(context.Background()) }()

	if _, err := tx.Exec(ctx, `UPDATE realm_account_challenges SET consumed_at = $1
		WHERE account_id = $2 AND kind = 'reset_password' AND consumed_at IS NULL`, now, accountID); err != nil {
		return fmt.Errorf("realm: retire password recovery challenges: %w", err)
	}

	if _, err := tx.Exec(ctx, `INSERT INTO realm_account_challenges
		(id, account_id, kind, token_digest, expires_at) VALUES ($1, $2, 'reset_password', $3, $4)`,
		uuid.New().String(), accountID, digest[:], now.Add(accounts.recoveryLifetime)); err != nil {
		return fmt.Errorf("realm: insert password recovery challenge: %w", err)
	}

	if _, err := tx.Exec(ctx, `INSERT INTO realm_mail_outbox
		(id, kind, recipient, payload, state, available_at) VALUES ($1, 'reset_password', $2, $3, 'pending', $4)`,
		uuid.New().String(), recipient, payload, now); err != nil {
		return fmt.Errorf("realm: insert recovery mail: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("realm: commit password recovery: %w", err)
	}

	return nil
}

// CompletePasswordRecovery executes complete password recovery through the PostgreSQL account lifecycle store so SQL
// encoding and database error translation remain centralized.
func (accounts *PostgresAccounts) CompletePasswordRecovery(ctx context.Context, token, password string) error {
	token = strings.TrimSpace(token)
	if token == "" || len(password) < minimumPasswordBytes || len(password) > maximumPasswordBytes {
		return ErrAccountChallenge
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	digest := sha256.Sum256([]byte(token))

	tx, err := accounts.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return fmt.Errorf("realm: begin password reset: %w", err)
	}

	defer func() { _ = tx.Rollback(context.Background()) }()

	var (
		challengeID, accountID string
		expiresAt              time.Time
		consumedAt             *time.Time
	)

	err = tx.QueryRow(ctx, `SELECT id, account_id, expires_at, consumed_at FROM realm_account_challenges
		WHERE kind = 'reset_password' AND token_digest = $1 FOR UPDATE`, digest[:]).
		Scan(&challengeID, &accountID, &expiresAt, &consumedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrAccountChallenge
	}

	if err != nil {
		return fmt.Errorf("realm: read password recovery challenge: %w", err)
	}

	if consumedAt != nil || !expiresAt.After(time.Now()) {
		return ErrAccountChallenge
	}

	now := time.Now().UTC()

	if _, err := tx.Exec(
		ctx,
		`UPDATE realm_accounts SET password_hash = $1 WHERE id = $2`,
		passwordHash,
		accountID,
	); err != nil {
		return fmt.Errorf("realm: update recovered password: %w", err)
	}

	if _, err := tx.Exec(
		ctx,
		`UPDATE realm_account_challenges SET consumed_at = $1 WHERE id = $2`,
		now,
		challengeID,
	); err != nil {
		return fmt.Errorf("realm: consume password recovery challenge: %w", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM realm_sessions WHERE account_id = $1`, accountID); err != nil {
		return fmt.Errorf("realm: revoke recovered account sessions: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("realm: commit password reset: %w", err)
	}

	return nil
}

type PostgresMailOutbox struct{ pool *pgxpool.Pool }

// ClaimMail executes claim mail through the PostgreSQL account lifecycle store so SQL encoding and database error
// translation remain centralized.
func (outbox *PostgresMailOutbox) ClaimMail(
	ctx context.Context,
	workerID string,
	lease time.Duration,
) (MailJob, error) {
	workerID = strings.TrimSpace(workerID)
	if outbox == nil || outbox.pool == nil || workerID == "" || lease <= 0 {
		return MailJob{}, ErrMailUnavailable
	}

	var (
		job     MailJob
		payload []byte
	)

	err := outbox.pool.QueryRow(ctx, `WITH candidate AS (
		SELECT id FROM realm_mail_outbox
		WHERE state = 'pending' AND available_at <= CURRENT_TIMESTAMP
			AND (locked_until IS NULL OR locked_until <= CURRENT_TIMESTAMP)
		ORDER BY available_at, created_at FOR UPDATE SKIP LOCKED LIMIT 1
	) UPDATE realm_mail_outbox m SET state = 'sending', attempts = attempts + 1,
		locked_by = $1, locked_until = CURRENT_TIMESTAMP + $2::interval, last_error = NULL
	FROM candidate WHERE m.id = candidate.id
	RETURNING m.id, m.kind, m.recipient, m.payload, m.attempts`, workerID, lease.String()).
		Scan(&job.ID, &job.Kind, &job.Recipient, &payload, &job.Attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return MailJob{}, ErrMailUnavailable
	}

	if err != nil {
		return MailJob{}, fmt.Errorf("realm: claim mail job: %w", err)
	}

	if err := json.Unmarshal(payload, &job.Payload); err != nil {
		return MailJob{}, fmt.Errorf("realm: decode mail job: %w", err)
	}

	return job, nil
}

// CompleteMail executes complete mail through the PostgreSQL account lifecycle store so SQL encoding and database
// error translation remain centralized.
func (outbox *PostgresMailOutbox) CompleteMail(ctx context.Context, workerID, jobID string) error {
	result, err := outbox.pool.Exec(ctx, `UPDATE realm_mail_outbox SET state = 'sent', sent_at = CURRENT_TIMESTAMP,
		locked_by = NULL, locked_until = NULL WHERE id = $1 AND state = 'sending' AND locked_by = $2`,
		strings.TrimSpace(jobID), strings.TrimSpace(workerID))
	if err != nil {
		return fmt.Errorf("realm: complete mail job: %w", err)
	}

	if result.RowsAffected() != 1 {
		return ErrMailUnavailable
	}

	return nil
}

// RetryMail executes retry mail through the PostgreSQL account lifecycle store so SQL encoding and database error
// translation remain centralized.
func (outbox *PostgresMailOutbox) RetryMail(
	ctx context.Context,
	workerID, jobID, message string,
	availableAt time.Time,
) error {
	message = strings.TrimSpace(message)
	if len(message) > 1024 {
		message = message[:1024]
	}

	result, err := outbox.pool.Exec(ctx, `UPDATE realm_mail_outbox SET state = 'pending', available_at = $1,
		locked_by = NULL, locked_until = NULL, last_error = $2
		WHERE id = $3 AND state = 'sending' AND locked_by = $4`, availableAt.UTC(), message,
		strings.TrimSpace(jobID), strings.TrimSpace(workerID))
	if err != nil {
		return fmt.Errorf("realm: retry mail job: %w", err)
	}

	if result.RowsAffected() != 1 {
		return ErrMailUnavailable
	}

	return nil
}
