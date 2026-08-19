package realm

import (
	"context"
	"errors"
	"strings"
)

// Signup delegates account enrollment to the configured lifecycle service and
// records the attempted identity without exposing the submitted password.
func (control *ControlPlane) Signup(
	ctx context.Context,
	request SignupRequest,
) (account Account, err error) {
	defer func() {
		control.recordAudit(ctx, AuditEvent{
			Operation:   AuditAccountSignup,
			AccountID:   account.ID,
			AccountName: firstNonEmpty(account.Name, strings.TrimSpace(request.Name)),
		}, err)
	}()
	if control == nil || control.accountLifecycle == nil {
		return Account{}, ErrAccountInput
	}
	return control.accountLifecycle.Signup(ctx, request)
}

// VerifyEmail consumes a lifecycle challenge and audits only the resulting
// account identity, keeping the bearer challenge out of diagnostics.
func (control *ControlPlane) VerifyEmail(
	ctx context.Context,
	challenge string,
) (account Account, err error) {
	defer func() {
		control.recordAudit(ctx, AuditEvent{
			Operation:   AuditAccountVerify,
			AccountID:   account.ID,
			AccountName: account.Name,
		}, err)
	}()
	if control == nil || control.accountLifecycle == nil {
		return Account{}, ErrAccountChallenge
	}
	return control.accountLifecycle.VerifyEmail(ctx, challenge)
}

// BeginPasswordRecovery starts the configured recovery workflow while keeping
// the supplied email address out of the audit event.
func (control *ControlPlane) BeginPasswordRecovery(ctx context.Context, email string) (err error) {
	defer func() {
		control.recordAudit(ctx, AuditEvent{Operation: AuditAccountRecoveryBegin}, err)
	}()
	if control == nil || control.accountLifecycle == nil {
		return ErrAccountInput
	}
	return control.accountLifecycle.BeginPasswordRecovery(ctx, email)
}

// CompletePasswordRecovery consumes a recovery challenge without logging the
// challenge or replacement password.
func (control *ControlPlane) CompletePasswordRecovery(
	ctx context.Context,
	challenge string,
	password string,
) (err error) {
	defer func() {
		control.recordAudit(ctx, AuditEvent{Operation: AuditAccountRecoveryComplete}, err)
	}()
	if control == nil || control.accountLifecycle == nil {
		return ErrAccountChallenge
	}
	return control.accountLifecycle.CompletePasswordRecovery(ctx, challenge, password)
}

// CreateAccount creates a directly authenticated account for deployments that
// do not require the enrollment lifecycle. Passwords never enter audit data.
func (control *ControlPlane) CreateAccount(
	ctx context.Context,
	name string,
	password string,
) (account Account, err error) {
	defer func() {
		control.recordAudit(ctx, AuditEvent{
			Operation:   AuditAccountCreate,
			AccountID:   account.ID,
			AccountName: firstNonEmpty(account.Name, strings.TrimSpace(name)),
		}, err)
	}()
	if control == nil || control.accounts == nil {
		return Account{}, ErrAccountInput
	}
	return control.accounts.Create(ctx, name, password)
}

// Authenticate exchanges account credentials for a realm session and records
// only the stable account and session identity.
func (control *ControlPlane) Authenticate(
	ctx context.Context,
	name string,
	password string,
) (session RealmSession, err error) {
	defer func() {
		control.recordAudit(ctx, AuditEvent{
			Operation:   AuditAccountLogin,
			AccountID:   session.Account.ID,
			AccountName: firstNonEmpty(session.Account.Name, strings.TrimSpace(name)),
			SessionID:   session.ID,
		}, err)
	}()
	if control == nil || control.accounts == nil {
		return RealmSession{}, ErrAccountCredentials
	}
	return control.accounts.Authenticate(ctx, name, password)
}

// Logout removes channel presence before invalidating the session. Failure to
// find channel membership is harmless; an invalid or expired session still
// fails so callers cannot mistake an unauthenticated request for a logout.
func (control *ControlPlane) Logout(ctx context.Context, token string) (err error) {
	event := AuditEvent{Operation: AuditAccountLogout}
	defer func() { control.recordAudit(ctx, event, err) }()

	principal, err := control.authorize(ctx, token)
	if err != nil {
		return err
	}
	event.AccountID = principal.accountID
	event.AccountName = principal.name
	event.SessionID = principal.sessionID

	leaveErr := control.channels.Leave(ctx, principal)
	if errors.Is(leaveErr, ErrChannelMember) {
		leaveErr = nil
	}
	return errors.Join(leaveErr, control.accounts.Logout(ctx, token))
}

// authorize prunes expired sessions before every semantic operation so normal
// realm traffic also removes stale lobby presence.
func (control *ControlPlane) authorize(
	ctx context.Context,
	token string,
) (AuthenticatedPrincipal, error) {
	if control == nil ||
		control.accounts == nil ||
		control.channels == nil ||
		control.games == nil ||
		control.characters == nil {
		return AuthenticatedPrincipal{}, ErrRealmSession
	}
	if _, err := control.PruneExpiredSessions(ctx); err != nil {
		return AuthenticatedPrincipal{}, err
	}
	return control.accounts.Authorize(ctx, token)
}

// PruneExpiredSessions is safe to call from a maintenance ticker and is also
// invoked before authorization so ordinary realm traffic clears ghost presence.
func (control *ControlPlane) PruneExpiredSessions(ctx context.Context) (int, error) {
	if control == nil || control.accounts == nil || control.channels == nil {
		return 0, ErrRealmSession
	}

	expired, err := control.accounts.PruneExpired(ctx)
	if err != nil {
		return 0, err
	}
	for _, principal := range expired {
		leaveErr := control.channels.Leave(ctx, principal)
		if errors.Is(leaveErr, ErrChannelMember) {
			leaveErr = nil
		}
		control.recordAudit(ctx, AuditEvent{
			Operation:   AuditAccountExpire,
			AccountID:   principal.accountID,
			AccountName: principal.name,
			SessionID:   principal.sessionID,
		}, leaveErr)
		if leaveErr != nil {
			return 0, leaveErr
		}
	}
	return len(expired), nil
}

// firstNonEmpty selects the first populated audit value so successful results
// can replace untrusted request text without losing failure context.
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
