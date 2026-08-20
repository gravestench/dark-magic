package realm

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	minimumAccountNameBytes = 3
	maximumAccountNameBytes = 32
	minimumPasswordBytes    = 8
	maximumPasswordBytes    = 72 // bcrypt's defined input limit.
	defaultSessionLifetime  = 24 * time.Hour
)

var (
	ErrAccountExists      = errors.New("realm: account already exists")
	ErrAccountCredentials = errors.New("realm: invalid account credentials")
	ErrAccountInput       = errors.New("realm: invalid account input")
	ErrRealmSession       = errors.New("realm: invalid realm session")
)

// Account is the durable public identity of one realm account. Password hashes
// and authentication-session details are deliberately absent.
type Account struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

// RealmSession is returned only at successful authentication. Token is an
// opaque bearer credential; stores retain only its SHA-256 digest.
type RealmSession struct {
	ID        string    `json:"id"`
	Account   Account   `json:"account"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AuthenticatedPrincipal cannot be fabricated by another package because all
// fields are private. Realm services accept it instead of caller-supplied IDs.
type AuthenticatedPrincipal struct {
	accountID string
	name      string
	sessionID string
}

// AccountRepository is the persistence boundary for Realm identities and
// digest-only sessions. Accounts is its deterministic in-memory reference;
// production single-machine operation uses PostgreSQL.
type AccountRepository interface {
	Create(context.Context, string, string) (Account, error)
	Authenticate(context.Context, string, string) (RealmSession, error)
	Authorize(context.Context, string) (AuthenticatedPrincipal, error)
	SelectCharacter(context.Context, string, string) error
	SelectedCharacter(context.Context, string) (string, error)
	Logout(context.Context, string) error
	PruneExpired(context.Context) ([]AuthenticatedPrincipal, error)
}

// AccountID exposes the authenticated principal's account id without giving callers mutable access to account or
// session storage.
func (principal AuthenticatedPrincipal) AccountID() string { return principal.accountID }

// Name exposes the authenticated principal's name without giving callers mutable access to account or session storage.
func (principal AuthenticatedPrincipal) Name() string { return principal.name }

// SessionID exposes the authenticated principal's session id without giving callers mutable access to account or
// session storage.
func (principal AuthenticatedPrincipal) SessionID() string { return principal.sessionID }

// valid checks the accounts invariant before state changes, keeping invalid values off shared paths.
func (principal AuthenticatedPrincipal) valid() bool {
	return principal.accountID != "" && principal.name != "" && principal.sessionID != ""
}

type memoryAccount struct {
	account      Account
	passwordHash []byte
}

type memoryRealmSession struct {
	id                  string
	accountID           string
	selectedCharacterID string
	expiresAt           time.Time
}

// SelectCharacter coordinates select character through the owning accounts synchronization boundary so shared state is
// published only after a complete transition.
func (accounts *Accounts) SelectCharacter(ctx context.Context, token, characterID string) error {
	if err := contextErr(ctx); err != nil {
		return err
	}

	characterID = strings.TrimSpace(characterID)
	if accounts == nil || strings.TrimSpace(token) == "" || characterID == "" {
		return ErrRealmSession
	}

	digest := sha256.Sum256([]byte(token))

	accounts.mu.Lock()
	defer accounts.mu.Unlock()

	session, found := accounts.sessions[digest]
	if !found || !session.expiresAt.After(accounts.now()) {
		delete(accounts.sessions, digest)
		return ErrRealmSession
	}

	for otherDigest, other := range accounts.sessions {
		if otherDigest == digest || !other.expiresAt.After(accounts.now()) {
			continue
		}

		if other.selectedCharacterID == characterID {
			return ErrCharacterOnline
		}
	}

	session.selectedCharacterID = characterID
	accounts.sessions[digest] = session

	return nil
}

// SelectedCharacter coordinates selected character through the owning accounts synchronization boundary so shared
// state is published only after a complete transition.
func (accounts *Accounts) SelectedCharacter(ctx context.Context, token string) (string, error) {
	if err := contextErr(ctx); err != nil {
		return "", err
	}

	if accounts == nil || strings.TrimSpace(token) == "" {
		return "", ErrRealmSession
	}

	digest := sha256.Sum256([]byte(token))

	accounts.mu.Lock()
	defer accounts.mu.Unlock()

	session, found := accounts.sessions[digest]
	if !found || !session.expiresAt.After(accounts.now()) {
		delete(accounts.sessions, digest)
		return "", ErrRealmSession
	}

	if session.selectedCharacterID == "" {
		return "", ErrCharacterNotFound
	}

	return session.selectedCharacterID, nil
}

// Accounts is the in-memory reference implementation of realm account creation
// and authentication. Durable adapters must retain the same normalized-name
// uniqueness and opaque, expiring session semantics.
type Accounts struct {
	mu              sync.Mutex
	now             func() time.Time
	sessionLifetime time.Duration
	dummyHash       []byte
	byName          map[string]*memoryAccount
	byID            map[string]*memoryAccount
	sessions        map[[sha256.Size]byte]memoryRealmSession
}

// NewAccounts constructs the accounts boundary and validates dependencies before callers can publish or mutate shared
// state.
func NewAccounts(sessionLifetime time.Duration) (*Accounts, error) {
	if sessionLifetime == 0 {
		sessionLifetime = defaultSessionLifetime
	}

	if sessionLifetime < time.Minute {
		return nil, ErrAccountInput
	}

	dummyHash, err := bcrypt.GenerateFromPassword([]byte("dark-magic-invalid-password"), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &Accounts{
		now:             time.Now,
		sessionLifetime: sessionLifetime,
		dummyHash:       dummyHash,
		byName:          make(map[string]*memoryAccount),
		byID:            make(map[string]*memoryAccount),
		sessions:        make(map[[sha256.Size]byte]memoryRealmSession),
	}, nil
}

// Create coordinates create through the owning accounts synchronization boundary so shared state is published only
// after a complete transition.
func (accounts *Accounts) Create(ctx context.Context, name, password string) (Account, error) {
	if err := contextErr(ctx); err != nil {
		return Account{}, err
	}

	if accounts == nil {
		return Account{}, ErrAccountInput
	}

	displayName, normalizedName, err := validateAccountName(name)
	if err != nil || len(password) < minimumPasswordBytes || len(password) > maximumPasswordBytes {
		return Account{}, ErrAccountInput
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return Account{}, err
	}

	if err := contextErr(ctx); err != nil {
		return Account{}, err
	}

	accounts.mu.Lock()
	defer accounts.mu.Unlock()

	if _, exists := accounts.byName[normalizedName]; exists {
		return Account{}, ErrAccountExists
	}
	// Direct Create is the trusted fixture/administration path. Public signup
	// uses AccountLifecycle and remains inactive until email verification.
	account := Account{ID: uuid.New().String(), Name: displayName, EmailVerified: true, CreatedAt: accounts.now().UTC()}
	stored := &memoryAccount{account: account, passwordHash: passwordHash}
	accounts.byName[normalizedName], accounts.byID[account.ID] = stored, stored

	return account, nil
}

// Authenticate coordinates authenticate through the owning accounts synchronization boundary so shared state is
// published only after a complete transition.
func (accounts *Accounts) Authenticate(ctx context.Context, name, password string) (RealmSession, error) {
	if err := contextErr(ctx); err != nil {
		return RealmSession{}, err
	}

	if accounts == nil || len(password) > maximumPasswordBytes {
		return RealmSession{}, ErrAccountCredentials
	}

	_, normalizedName, nameErr := validateAccountName(name)

	accounts.mu.Lock()
	stored := accounts.byName[normalizedName]
	dummyHash := append([]byte(nil), accounts.dummyHash...)
	accounts.mu.Unlock()

	hash := dummyHash
	if stored != nil {
		hash = append([]byte(nil), stored.passwordHash...)
	}

	if nameErr != nil || bcrypt.CompareHashAndPassword(hash, []byte(password)) != nil || stored == nil {
		return RealmSession{}, ErrAccountCredentials
	}

	if err := contextErr(ctx); err != nil {
		return RealmSession{}, err
	}

	token, digest, err := newRealmSessionToken()
	if err != nil {
		return RealmSession{}, err
	}

	accounts.mu.Lock()
	defer accounts.mu.Unlock()
	// Re-read by ID in case a future durable adapter supports deletion while a
	// password comparison is in flight.
	current := accounts.byID[stored.account.ID]
	if current == nil || current != stored {
		return RealmSession{}, ErrAccountCredentials
	}

	session := memoryRealmSession{
		id:        uuid.New().String(),
		accountID: stored.account.ID,
		expiresAt: accounts.now().Add(accounts.sessionLifetime).UTC(),
	}
	accounts.sessions[digest] = session

	return RealmSession{ID: session.id, Account: stored.account, Token: token, ExpiresAt: session.expiresAt}, nil
}

// Authorize coordinates authorize through the owning accounts synchronization boundary so shared state is published
// only after a complete transition.
func (accounts *Accounts) Authorize(ctx context.Context, token string) (AuthenticatedPrincipal, error) {
	if err := contextErr(ctx); err != nil {
		return AuthenticatedPrincipal{}, err
	}

	if accounts == nil || strings.TrimSpace(token) == "" {
		return AuthenticatedPrincipal{}, ErrRealmSession
	}

	digest := sha256.Sum256([]byte(token))

	accounts.mu.Lock()
	defer accounts.mu.Unlock()

	session, found := accounts.sessions[digest]
	if !found || !session.expiresAt.After(accounts.now()) {
		delete(accounts.sessions, digest)
		return AuthenticatedPrincipal{}, ErrRealmSession
	}

	account := accounts.byID[session.accountID]
	if account == nil {
		delete(accounts.sessions, digest)
		return AuthenticatedPrincipal{}, ErrRealmSession
	}

	return AuthenticatedPrincipal{accountID: account.account.ID, name: account.account.Name, sessionID: session.id}, nil
}

// Logout coordinates logout through the owning accounts synchronization boundary so shared state is published only
// after a complete transition.
func (accounts *Accounts) Logout(ctx context.Context, token string) error {
	if err := contextErr(ctx); err != nil {
		return err
	}

	if accounts == nil || strings.TrimSpace(token) == "" {
		return ErrRealmSession
	}

	digest := sha256.Sum256([]byte(token))

	accounts.mu.Lock()
	defer accounts.mu.Unlock()

	session, found := accounts.sessions[digest]
	if !found || !session.expiresAt.After(accounts.now()) {
		delete(accounts.sessions, digest)
		return ErrRealmSession
	}

	delete(accounts.sessions, digest)

	return nil
}

// PruneExpired removes expired bearer credentials and returns their former
// principals so the control plane can remove associated ephemeral presence.
func (accounts *Accounts) PruneExpired(ctx context.Context) ([]AuthenticatedPrincipal, error) {
	if err := contextErr(ctx); err != nil {
		return nil, err
	}

	if accounts == nil {
		return nil, ErrRealmSession
	}

	accounts.mu.Lock()
	defer accounts.mu.Unlock()

	now := accounts.now()
	result := make([]AuthenticatedPrincipal, 0)

	for digest, session := range accounts.sessions {
		if session.expiresAt.After(now) {
			continue
		}

		if account := accounts.byID[session.accountID]; account != nil {
			result = append(
				result,
				AuthenticatedPrincipal{accountID: account.account.ID, name: account.account.Name, sessionID: session.id},
			)
		}

		delete(accounts.sessions, digest)
	}

	return result, nil
}

// validateAccountName checks the accounts invariant before state changes, keeping invalid values off shared paths.
func validateAccountName(name string) (string, string, error) {
	display := strings.TrimSpace(name)
	if len(display) < minimumAccountNameBytes || len(display) > maximumAccountNameBytes {
		return "", "", ErrAccountInput
	}

	for _, value := range display {
		if (value < 'a' || value > 'z') &&
			(value < 'A' || value > 'Z') &&
			(value < '0' || value > '9') &&
			value != '_' && value != '-' && value != '.' {
			return "", "", ErrAccountInput
		}
	}

	return display, strings.ToLower(display), nil
}

// newRealmSessionToken constructs the accounts boundary and validates dependencies before callers can publish or
// mutate shared state.
func newRealmSessionToken() (string, [sha256.Size]byte, error) {
	var bytes [32]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", [sha256.Size]byte{}, err
	}

	token := base64.RawURLEncoding.EncodeToString(bytes[:])

	return token, sha256.Sum256([]byte(token)), nil
}

// contextErr contains context err within the accounts boundary so callers do not duplicate its domain-specific policy.
func contextErr(ctx context.Context) error {
	if ctx == nil {
		return context.Canceled
	}

	return ctx.Err()
}
