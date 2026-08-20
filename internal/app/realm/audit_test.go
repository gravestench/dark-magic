package realm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type capturedAudit struct {
	mu     sync.Mutex
	events []AuditEvent
}

// Record supplies the test double's record behavior, keeping the scenario deterministic and independent of external
// services.
func (capture *capturedAudit) Record(_ context.Context, event AuditEvent) {
	capture.mu.Lock()
	defer capture.mu.Unlock()

	capture.events = append(capture.events, event)
}

// snapshot supplies the test double's snapshot behavior, keeping the scenario deterministic and independent of
// external services.
func (capture *capturedAudit) snapshot() []AuditEvent {
	capture.mu.Lock()
	defer capture.mu.Unlock()

	return append([]AuditEvent(nil), capture.events...)
}

// TestControlPlaneAuditsRealmIdentityLifecycleWithoutSecrets verifies control plane audits realm identity lifecycle
// without secrets. The scenario keeps the audit contract visible to maintainers.
func TestControlPlaneAuditsRealmIdentityLifecycleWithoutSecrets(t *testing.T) {
	capture := &capturedAudit{}

	control, err := NewControlPlane(orchestratedControlConfig(capture))
	if err != nil {
		t.Fatal(err)
	}

	const password = "audit secret password"

	account, err := control.CreateAccount(t.Context(), "Alice", password)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := control.Authenticate(t.Context(), "Alice", "wrong password"); !errors.Is(err, ErrAccountCredentials) {
		t.Fatalf("bad-password error = %v", err)
	}

	session, err := control.Authenticate(t.Context(), "Alice", password)
	if err != nil {
		t.Fatal(err)
	}

	character, err := control.CreateCharacter(t.Context(), session.Token, CreateCharacterRequest{
		Name: "AuditHero", Class: "Amazon", Expansion: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := control.JoinChannel(t.Context(), session.Token, "Dark Magic"); err != nil {
		t.Fatal(err)
	}

	if _, err := control.SendChannelMessage(t.Context(), session.Token, "this body must not be audited"); err != nil {
		t.Fatal(err)
	}

	game, err := control.CreateGame(t.Context(), session.Token, CreateGameRequest{
		Name: "Audit Game", Password: "game secret", Difficulty: DifficultyNormal, Maximum: 8, Expansion: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := control.ResolveGameJoin(t.Context(), session.Token, game.Game.Entry.Name, "game secret"); err != nil {
		t.Fatal(err)
	}

	if _, err := control.LeaveGame(t.Context(), session.Token, game.Game.Entry.GameID); err != nil {
		t.Fatal(err)
	}

	if err := control.Logout(t.Context(), session.Token); err != nil {
		t.Fatal(err)
	}

	events := capture.snapshot()

	wantOperations := []string{AuditAccountCreate, AuditAccountLogin, AuditAccountLogin, AuditCharacterCreate,
		AuditChannelJoin, AuditChannelMessage, AuditGameCreate, AuditGameResolve, AuditGameLeave, AuditAccountLogout}
	if len(events) != len(wantOperations) {
		t.Fatalf("events = %#v", events)
	}

	for index, operation := range wantOperations {
		if events[index].Version != RealmAuditVersion || events[index].Operation != operation {
			t.Fatalf("event[%d] = %#v", index, events[index])
		}
	}

	if events[1].Outcome != "denied" || events[1].ErrorCode != "unauthorized" {
		t.Fatalf("failed login event = %#v", events[1])
	}

	if events[2].Outcome != "success" || events[2].AccountID != account.ID || events[2].SessionID != session.ID {
		t.Fatalf("successful login event = %#v", events[2])
	}

	if events[3].CharacterID != character.Character.ID || events[3].CharacterName != character.Character.Name {
		t.Fatalf("character event = %#v", events[3])
	}

	if events[5].MessageBytes != len("this body must not be audited") {
		t.Fatalf("message event = %#v", events[5])
	}

	serialized := fmt.Sprintf("%#v", events)
	for _, secret := range []string{password, session.Token, "this body must not be audited", "game secret"} {
		if strings.Contains(serialized, secret) {
			t.Fatalf("audit events contain secret %q: %s", secret, serialized)
		}
	}
}

// TestRealmHTTPAuditUsesSocketPeerAddress verifies realm httpaudit uses socket peer address. The scenario keeps the
// audit contract visible to maintainers.
func TestRealmHTTPAuditUsesSocketPeerAddress(t *testing.T) {
	capture := &capturedAudit{}

	control, err := NewControlPlane(ControlPlaneConfig{Audit: capture})
	if err != nil {
		t.Fatal(err)
	}

	handler, err := NewHTTPHandler(control)
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/v1/accounts", strings.NewReader(
		`{"Name":"Alice","Password":"long enough password"}`))
	request.RemoteAddr = "192.0.2.10:43210"
	request.Header.Set("X-Forwarded-For", "203.0.113.9")

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", response.Code, response.Body.String())
	}

	events := capture.snapshot()
	if len(events) != 1 || events[0].ClientAddress != "192.0.2.10" {
		t.Fatalf("events = %#v", events)
	}
}

// TestControlPlaneAuditsExpiredSessionCleanup verifies control plane audits expired session cleanup. The scenario
// keeps the audit contract visible to maintainers.
func TestControlPlaneAuditsExpiredSessionCleanup(t *testing.T) {
	capture := &capturedAudit{}

	control, err := NewControlPlane(ControlPlaneConfig{SessionLifetime: time.Minute, Audit: capture})
	if err != nil {
		t.Fatal(err)
	}

	now := time.Unix(100, 0)

	control.accounts.(*Accounts).now = func() time.Time { return now }
	if _, err := control.CreateAccount(t.Context(), "Alice", "long enough password"); err != nil {
		t.Fatal(err)
	}

	session, err := control.Authenticate(t.Context(), "Alice", "long enough password")
	if err != nil {
		t.Fatal(err)
	}

	now = now.Add(time.Minute)

	count, err := control.PruneExpiredSessions(t.Context())
	if err != nil || count != 1 {
		t.Fatalf("count=%d error=%v", count, err)
	}

	events := capture.snapshot()
	if len(events) != 3 || events[2].Operation != AuditAccountExpire || events[2].Outcome != "success" ||
		events[2].AccountID != session.Account.ID || events[2].SessionID != session.ID {
		t.Fatalf("events = %#v", events)
	}
}

// TestAuditClassifiesWorkerUnavailabilityWithoutLeakingInternals verifies audit classifies worker unavailability
// without leaking internals. The scenario keeps the audit contract visible to maintainers.
func TestAuditClassifiesWorkerUnavailabilityWithoutLeakingInternals(t *testing.T) {
	for _, err := range []error{ErrGameUnavailable, ErrWorker, ErrAdmission, fmt.Errorf("wrapped: %w", ErrWorker)} {
		outcome, code := auditResult(err)
		if outcome != "failed" || code != "unavailable" {
			t.Fatalf("auditResult(%v) = %q, %q", err, outcome, code)
		}
	}
}

// TestAuditClassifiesCharacterDifferenceAsDenied verifies audit classifies character difference as denied. The
// scenario keeps the audit contract visible to maintainers.
func TestAuditClassifiesCharacterDifferenceAsDenied(t *testing.T) {
	outcome, code := auditResult(ErrGameLevelRange)
	if outcome != "denied" || code != "level_restricted" {
		t.Fatalf("auditResult = %q, %q", outcome, code)
	}
}
