package gameserver

import (
	"context"
	"testing"
	"time"
)

// newTestTicketAuthority gives ticket tests a valid signing key while keeping session binding explicit at each call.
func newTestTicketAuthority(t *testing.T, sessionID string) *TicketAuthority {
	t.Helper()

	authority, err := NewTicketAuthority([]byte("0123456789abcdef0123456789abcdef"), sessionID)
	if err != nil {
		t.Fatal(err)
	}

	return authority
}

// testTicketPrincipal returns all identity fields that must survive a signed ticket round trip.
func testTicketPrincipal() Principal {
	return Principal{ID: "account", CharacterID: "character", PlayerID: "player"}
}

// issueTestTicket fails at the point of fixture construction so assertions can focus on authentication outcomes.
func issueTestTicket(t *testing.T, authority *TicketAuthority, lifetime time.Duration) string {
	t.Helper()

	ticket, err := authority.Issue(testTicketPrincipal(), lifetime)
	if err != nil {
		t.Fatal(err)
	}

	return ticket
}

// TestTicketAuthorityIssuesSessionBoundOneTimeTickets verifies both replay prevention and session isolation.
func TestTicketAuthorityIssuesSessionBoundOneTimeTickets(t *testing.T) {
	authority := newTestTicketAuthority(t, "game-1")
	now := time.Unix(1000, 0)
	authority.now = func() time.Time { return now }
	ticket := issueTestTicket(t, authority, time.Minute)

	principal, err := authority.Authenticate(context.Background(), ticket)
	if err != nil || principal.PlayerID != "player" {
		t.Fatalf("principal=%#v error=%v", principal, err)
	}

	if _, err := authority.Authenticate(context.Background(), ticket); err != ErrAuthentication {
		t.Fatalf("replay error = %v", err)
	}

	other := newTestTicketAuthority(t, "game-2")
	other.now = authority.now

	second := issueTestTicket(t, authority, time.Minute)
	if _, err := other.Authenticate(context.Background(), second); err != ErrAuthentication {
		t.Fatalf("wrong session error = %v", err)
	}
}

// TestTicketAuthorityRejectsTamperingAndExpiry covers both cryptographic and temporal invalidation.
func TestTicketAuthorityRejectsTamperingAndExpiry(t *testing.T) {
	authority := newTestTicketAuthority(t, "game")
	now := time.Unix(1000, 0)
	authority.now = func() time.Time { return now }
	ticket := issueTestTicket(t, authority, time.Second)

	if _, err := authority.Authenticate(context.Background(), ticket+"x"); err != ErrAuthentication {
		t.Fatalf("tamper error = %v", err)
	}

	now = now.Add(2 * time.Second)

	if _, err := authority.Authenticate(context.Background(), ticket); err != ErrAuthentication {
		t.Fatalf("expiry error = %v", err)
	}
}

// TestTicketAuthorityRevokesUnconsumedTicket proves explicit cancellation closes the same
// admission path as consumption.
func TestTicketAuthorityRevokesUnconsumedTicket(t *testing.T) {
	authority := newTestTicketAuthority(t, "game")
	ticket := issueTestTicket(t, authority, time.Minute)

	if err := authority.Revoke(ticket); err != nil {
		t.Fatal(err)
	}

	if _, err := authority.Authenticate(context.Background(), ticket); err != ErrAuthentication {
		t.Fatalf("revoked ticket error = %v", err)
	}
}
