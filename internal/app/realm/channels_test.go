package realm

import (
	"errors"
	"testing"
	"time"

	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

func TestChannelsProjectCharacterPresenceAndPublicMessages(t *testing.T) {
	accounts, err := NewAccounts(time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	alice := authenticatedFixture(t, accounts, "Alice")
	bob := authenticatedFixture(t, accounts, "Bob")
	channels := NewChannels(8)
	now := time.Unix(100, 0)
	channels.now = func() time.Time { now = now.Add(time.Second); return now }
	alicePresence := CharacterPresence{CharacterID: "character:alice", Name: "Alyssa", Class: "Assassin", Level: 12,
		Expansion: true, Appearance: &d2save.Appearance{COF: "AITNHTH.cof", Components: map[string]string{"HD": "head.dcc"}}}
	view, err := channels.Join(t.Context(), alice, " Diablo II En-1 ", alicePresence)
	if err != nil {
		t.Fatal(err)
	}
	if view.Version != ChannelViewVersion || view.ID != "diablo ii en-1" || view.Name != "Diablo II En-1" || len(view.Members) != 1 || view.LastEvent != 1 {
		t.Fatalf("first view = %#v", view)
	}
	view.Members[0].Character.Appearance.Components["HD"] = "mutated"
	bobPresence := CharacterPresence{CharacterID: "character:bob", Name: "Borin", Class: "Barbarian", Level: 9}
	view, err = channels.Join(t.Context(), bob, "diablo ii en-1", bobPresence)
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Members) != 2 || view.Members[0].Account != "Alice" || view.Members[1].Account != "Bob" ||
		view.Members[0].Character.Appearance.Components["HD"] != "head.dcc" {
		t.Fatalf("joined view = %#v", view)
	}
	event, err := channels.Send(t.Context(), bob, "  hello realm  ")
	if err != nil {
		t.Fatal(err)
	}
	if event.Kind != ChatEventMessage || event.Text != "hello realm" || event.Sender == nil || event.Sender.Character.Name != "Borin" || event.Sequence != 3 {
		t.Fatalf("message = %#v", event)
	}
	events, err := channels.EventsAfter(t.Context(), alice, 1, 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[0].Kind != ChatEventJoined || events[1].Kind != ChatEventMessage {
		t.Fatalf("events = %#v", events)
	}
	events[0].Sender.Character.Name = "mutated"
	current, err := channels.View(t.Context(), alice)
	if err != nil || current.Members[1].Character.Name != "Borin" {
		t.Fatalf("defensive view = %#v error=%v", current, err)
	}
}

func TestChannelsMoveSessionsAndBoundHistory(t *testing.T) {
	accounts, err := NewAccounts(time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	alice := authenticatedFixture(t, accounts, "Alice")
	channels := NewChannels(2)
	presence := CharacterPresence{CharacterID: "character", Name: "Hero", Class: "Amazon", Level: 1}
	if _, err := channels.Join(t.Context(), alice, "First", presence); err != nil {
		t.Fatal(err)
	}
	if _, err := channels.Send(t.Context(), alice, "one"); err != nil {
		t.Fatal(err)
	}
	if _, err := channels.Send(t.Context(), alice, "two"); err != nil {
		t.Fatal(err)
	}
	if _, err := channels.Join(t.Context(), alice, "Second", presence); err != nil {
		t.Fatal(err)
	}
	view, err := channels.View(t.Context(), alice)
	if err != nil || view.ID != "second" || len(view.Members) != 1 {
		t.Fatalf("moved view = %#v error=%v", view, err)
	}
	first := channels.channels["first"]
	if len(first.members) != 0 || len(first.events) != 2 || first.events[0].Text != "two" || first.events[1].Kind != ChatEventLeft {
		t.Fatalf("bounded first channel = %#v", first)
	}
	if err := channels.Leave(t.Context(), alice); err != nil {
		t.Fatal(err)
	}
	if _, err := channels.View(t.Context(), alice); !errors.Is(err, ErrChannelMember) {
		t.Fatalf("view after leave error = %v", err)
	}
	if _, err := channels.Send(t.Context(), AuthenticatedPrincipal{}, "spoof"); !errors.Is(err, ErrChannelInput) {
		t.Fatalf("spoofed principal error = %v", err)
	}
}

func TestChannelsDefendAgainstDuplicateCharacterPresence(t *testing.T) {
	accounts, err := NewAccounts(time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := accounts.Create(t.Context(), "MuleAccount", "fixture password"); err != nil {
		t.Fatal(err)
	}
	principal := func() AuthenticatedPrincipal {
		session, err := accounts.Authenticate(t.Context(), "MuleAccount", "fixture password")
		if err != nil {
			t.Fatal(err)
		}
		principal, err := accounts.Authorize(t.Context(), session.Token)
		if err != nil {
			t.Fatal(err)
		}
		return principal
	}
	first, second := principal(), principal()
	channels := NewChannels(8)
	presence := CharacterPresence{CharacterID: "shared-character", Name: "Shared", Class: "Amazon", Level: 1}
	if _, err := channels.Join(t.Context(), first, "Diablo II", presence); err != nil {
		t.Fatal(err)
	}
	if _, err := channels.Join(t.Context(), second, "Diablo II", presence); !errors.Is(err, ErrCharacterOnline) {
		t.Fatalf("duplicate channel presence error = %v", err)
	}
	mule := CharacterPresence{CharacterID: "mule-character", Name: "Mule", Class: "Barbarian", Level: 1}
	view, err := channels.Join(t.Context(), second, "Diablo II", mule)
	if err != nil || len(view.Members) != 2 {
		t.Fatalf("same-account mule presence = %#v, %v", view, err)
	}
	if err := channels.Leave(t.Context(), first); err != nil {
		t.Fatal(err)
	}
	if _, err := channels.Join(t.Context(), second, "Diablo II", presence); err != nil {
		t.Fatalf("released character remained claimed: %v", err)
	}
}

func TestChannelsPruneOnlyPresenceThatStoppedRenewing(t *testing.T) {
	accounts, err := NewAccounts(time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	alice := authenticatedFixture(t, accounts, "Alice")
	bob := authenticatedFixture(t, accounts, "Bob")
	channels := NewChannels(8)
	now := time.Unix(100, 0)
	channels.now = func() time.Time { return now }
	presence := func(id, name string) CharacterPresence {
		return CharacterPresence{CharacterID: id, Name: name, Class: "Amazon", Level: 1}
	}
	if _, err := channels.Join(t.Context(), alice, "Diablo II", presence("alice", "AliceHero")); err != nil {
		t.Fatal(err)
	}
	if _, err := channels.Join(t.Context(), bob, "Diablo II", presence("bob", "BobHero")); err != nil {
		t.Fatal(err)
	}
	now = now.Add(20 * time.Second)
	if _, err := channels.View(t.Context(), alice); err != nil {
		t.Fatal(err)
	}
	pruned, err := channels.PruneInactive(t.Context(), now.Add(-10*time.Second))
	if err != nil || pruned != 1 {
		t.Fatalf("pruned=%d error=%v", pruned, err)
	}
	view, err := channels.View(t.Context(), alice)
	if err != nil || len(view.Members) != 1 || view.Members[0].Character.CharacterID != "alice" {
		t.Fatalf("active view=%#v error=%v", view, err)
	}
	if _, err := channels.View(t.Context(), bob); !errors.Is(err, ErrChannelMember) {
		t.Fatalf("inactive Bob still present: %v", err)
	}
}
