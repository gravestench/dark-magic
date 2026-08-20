package realm

import (
	"context"
	"errors"
	"net/mail"
	"net/url"
	"strings"
	"time"
)

const (
	defaultVerificationLifetime = 30 * time.Minute
	defaultRecoveryLifetime     = 15 * time.Minute
	maximumEmailBytes           = 320
)

var (
	ErrAccountUnverified = errors.New("realm: account email is not verified")
	ErrAccountChallenge  = errors.New("realm: invalid or expired account challenge")
	ErrMailUnavailable   = errors.New("realm: mail delivery unavailable")
)

// AccountLifecycle is the production identity boundary. It deliberately sits
// beside AccountRepository so deterministic fixtures can keep creating active
// accounts without pretending to implement email delivery.
type AccountLifecycle interface {
	Signup(context.Context, SignupRequest) (Account, error)
	VerifyEmail(context.Context, string) (Account, error)
	BeginPasswordRecovery(context.Context, string) error
	CompletePasswordRecovery(context.Context, string, string) error
}

type SignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type MailJob struct {
	ID        string         `json:"id"`
	Kind      string         `json:"kind"`
	Recipient string         `json:"recipient"`
	Payload   map[string]any `json:"payload"`
	Attempts  int            `json:"attempts"`
}

// MailOutbox provides transactionally-created jobs to a provider-neutral mail
// worker. Claiming is leased so a crashed worker does not lose delivery.
type MailOutbox interface {
	ClaimMail(context.Context, string, time.Duration) (MailJob, error)
	CompleteMail(context.Context, string, string) error
	RetryMail(context.Context, string, string, string, time.Time) error
}

// normalizeEmail checks the account lifecycle invariant before state changes, keeping invalid values off shared paths.
func normalizeEmail(value string) (string, string, error) {
	display := strings.TrimSpace(value)
	if display == "" || len(display) > maximumEmailBytes {
		return "", "", ErrAccountInput
	}

	parsed, err := mail.ParseAddress(display)
	if err != nil || parsed.Address != display || strings.Count(display, "@") != 1 {
		return "", "", ErrAccountInput
	}

	local, domain, found := strings.Cut(display, "@")
	if !found || local == "" || domain == "" {
		return "", "", ErrAccountInput
	}
	// Domain names are case-insensitive. Lower-casing the complete address is
	// an intentional Realm policy that gives users predictable uniqueness.
	return display, strings.ToLower(local + "@" + domain), nil
}

// accountActionURL contains account action url within the account lifecycle boundary so callers do not duplicate its
// domain-specific policy.
func accountActionURL(baseURL, path, token string) (string, error) {
	base, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || base.Scheme != "https" || base.Host == "" {
		return "", ErrAccountInput
	}

	reference := &url.URL{Path: path, RawQuery: url.Values{"token": {token}}.Encode()}

	return base.ResolveReference(reference).String(), nil
}
