package realm

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"
)

func TestAccountsCreateAuthenticateAuthorizeAndLogout(t *testing.T) {
	accounts, err := NewAccounts(time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(100, 0)
	accounts.now = func() time.Time { return now }
	account, err := accounts.Create(t.Context(), "Alice_1", "correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if account.ID == "" || account.Name != "Alice_1" || !account.CreatedAt.Equal(now) {
		t.Fatalf("account = %#v", account)
	}
	if _, err := accounts.Create(t.Context(), " alice_1 ", "another valid password"); !errors.Is(err, ErrAccountExists) {
		t.Fatalf("case-insensitive duplicate error = %v", err)
	}
	for _, attempt := range []struct{ name, password string }{{"Alice_1", "wrong password"}, {"missing", "correct horse battery staple"}} {
		if _, err := accounts.Authenticate(t.Context(), attempt.name, attempt.password); !errors.Is(err, ErrAccountCredentials) {
			t.Fatalf("authentication error for %q = %v", attempt.name, err)
		}
	}
	session, err := accounts.Authenticate(t.Context(), "ALICE_1", "correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if session.ID == "" || session.Token == "" || session.Token == session.ID || session.Account.ID != account.ID || !session.ExpiresAt.Equal(now.Add(time.Hour)) {
		t.Fatalf("session = %#v", session)
	}
	digest := sha256.Sum256([]byte(session.Token))
	if _, stored := accounts.sessions[digest]; !stored {
		t.Fatal("session store does not retain the token digest")
	}
	principal, err := accounts.Authorize(t.Context(), session.Token)
	if err != nil {
		t.Fatal(err)
	}
	if principal.AccountID() != account.ID || principal.Name() != account.Name || principal.SessionID() != session.ID {
		t.Fatalf("principal = %#v", principal)
	}
	if err := accounts.Logout(t.Context(), session.Token); err != nil {
		t.Fatal(err)
	}
	if _, err := accounts.Authorize(t.Context(), session.Token); !errors.Is(err, ErrRealmSession) {
		t.Fatalf("logged-out authorization error = %v", err)
	}
}

func TestAccountsExpireSessionsAndValidateInputs(t *testing.T) {
	accounts, err := NewAccounts(time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(200, 0)
	accounts.now = func() time.Time { return now }
	if _, err := accounts.Create(t.Context(), "a", "password"); !errors.Is(err, ErrAccountInput) {
		t.Fatalf("short name error = %v", err)
	}
	if _, err := accounts.Create(t.Context(), "bad name", "password"); !errors.Is(err, ErrAccountInput) {
		t.Fatalf("invalid name error = %v", err)
	}
	if _, err := accounts.Create(t.Context(), "Alice", "short"); !errors.Is(err, ErrAccountInput) {
		t.Fatalf("short password error = %v", err)
	}
	if _, err := accounts.Create(t.Context(), "Alice", "long enough password"); err != nil {
		t.Fatal(err)
	}
	session, err := accounts.Authenticate(t.Context(), "Alice", "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Minute)
	if _, err := accounts.Authorize(t.Context(), session.Token); !errors.Is(err, ErrRealmSession) {
		t.Fatalf("expired authorization error = %v", err)
	}
	if err := accounts.Logout(t.Context(), session.Token); !errors.Is(err, ErrRealmSession) {
		t.Fatalf("expired logout error = %v", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := accounts.Create(cancelled, "Other", "long enough password"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled create error = %v", err)
	}
}

func TestAccountsAllowMulingButClaimEachCharacterOnce(t *testing.T) {
	accounts, err := NewAccounts(time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := accounts.Create(t.Context(), "MuleAccount", "long enough password"); err != nil {
		t.Fatal(err)
	}
	first, err := accounts.Authenticate(t.Context(), "MuleAccount", "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	second, err := accounts.Authenticate(t.Context(), "MuleAccount", "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	if err := accounts.SelectCharacter(t.Context(), first.Token, "character-main"); err != nil {
		t.Fatal(err)
	}
	if err := accounts.SelectCharacter(t.Context(), second.Token, "character-main"); !errors.Is(err, ErrCharacterOnline) {
		t.Fatalf("duplicate character claim error = %v", err)
	}
	if err := accounts.SelectCharacter(t.Context(), second.Token, "character-mule"); err != nil {
		t.Fatalf("different character on same account was rejected: %v", err)
	}
	if err := accounts.Logout(t.Context(), first.Token); err != nil {
		t.Fatal(err)
	}
	if err := accounts.SelectCharacter(t.Context(), second.Token, "character-main"); err != nil {
		t.Fatalf("released character could not be selected: %v", err)
	}
}

func TestAccountsPruneExpiredReturnsFormerPrincipal(t *testing.T) {
	accounts, err := NewAccounts(time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(300, 0)
	accounts.now = func() time.Time { return now }
	account, err := accounts.Create(t.Context(), "Alice", "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	session, err := accounts.Authenticate(t.Context(), "Alice", "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Minute)
	expired, err := accounts.PruneExpired(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(expired) != 1 || expired[0].AccountID() != account.ID || expired[0].SessionID() != session.ID {
		t.Fatalf("expired principals = %#v", expired)
	}
	if _, err := accounts.Authorize(t.Context(), session.Token); !errors.Is(err, ErrRealmSession) {
		t.Fatalf("pruned authorization error = %v", err)
	}
}

func authenticatedFixture(t *testing.T, accounts *Accounts, name string) AuthenticatedPrincipal {
	t.Helper()
	password := "fixture password"
	if _, err := accounts.Create(t.Context(), name, password); err != nil {
		t.Fatal(err)
	}
	session, err := accounts.Authenticate(t.Context(), name, password)
	if err != nil {
		t.Fatal(err)
	}
	principal, err := accounts.Authorize(t.Context(), session.Token)
	if err != nil {
		t.Fatal(err)
	}
	return principal
}
