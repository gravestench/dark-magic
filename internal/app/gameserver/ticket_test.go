package gameserver

import (
	"context"
	"testing"
	"time"
)

func TestTicketAuthorityIssuesSessionBoundOneTimeTickets(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	authority, err := NewTicketAuthority(key, "game-1")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1000, 0)
	authority.now = func() time.Time { return now }
	ticket, err := authority.Issue(Principal{ID: "account", CharacterID: "character", PlayerID: "player"}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	principal, err := authority.Authenticate(context.Background(), ticket)
	if err != nil || principal.PlayerID != "player" {
		t.Fatalf("principal=%#v error=%v", principal, err)
	}
	if _, err := authority.Authenticate(context.Background(), ticket); err != ErrAuthentication {
		t.Fatalf("replay error = %v", err)
	}

	other, err := NewTicketAuthority(key, "game-2")
	if err != nil {
		t.Fatal(err)
	}
	other.now = authority.now
	second, err := authority.Issue(Principal{ID: "account", CharacterID: "character", PlayerID: "player"}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := other.Authenticate(context.Background(), second); err != ErrAuthentication {
		t.Fatalf("wrong session error = %v", err)
	}
}

func TestTicketAuthorityRejectsTamperingAndExpiry(t *testing.T) {
	authority, err := NewTicketAuthority([]byte("0123456789abcdef0123456789abcdef"), "game")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1000, 0)
	authority.now = func() time.Time { return now }
	ticket, err := authority.Issue(Principal{ID: "account", CharacterID: "character", PlayerID: "player"}, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := authority.Authenticate(context.Background(), ticket+"x"); err != ErrAuthentication {
		t.Fatalf("tamper error = %v", err)
	}
	now = now.Add(2 * time.Second)
	if _, err := authority.Authenticate(context.Background(), ticket); err != ErrAuthentication {
		t.Fatalf("expiry error = %v", err)
	}
}

func TestTicketAuthorityRevokesUnconsumedTicket(t *testing.T) {
	authority, err := NewTicketAuthority([]byte("0123456789abcdef0123456789abcdef"), "game")
	if err != nil {
		t.Fatal(err)
	}
	ticket, err := authority.Issue(Principal{ID: "account", CharacterID: "character", PlayerID: "player"}, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := authority.Revoke(ticket); err != nil {
		t.Fatal(err)
	}
	if _, err := authority.Authenticate(context.Background(), ticket); err != ErrAuthentication {
		t.Fatalf("revoked ticket error = %v", err)
	}
}
