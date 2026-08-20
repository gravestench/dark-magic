package simulation

import (
	"errors"
	"fmt"
	"testing"
)

// TestAdmitterEnforcesAuthorityBoundaryTransactionally proves failed validation cannot consume sequence state.
func TestAdmitterEnforcesAuthorityBoundaryTransactionally(t *testing.T) {
	admitter := NewAdmitter(2)
	if err := admitter.Register("move", func(command Command) error {
		if string(command.Payload) != `{"x":1}` {
			return fmt.Errorf("bad move")
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}

	valid := Command{Tick: 11, Player: "player-1", Sequence: 1, Kind: "move", Payload: []byte(`{"x":1}`)}
	if err := admitter.Admit(valid, 10); err != nil {
		t.Fatal(err)
	}

	invalid := Command{Tick: 11, Player: "player-1", Sequence: 2, Kind: "move", Payload: []byte(`{"x":2}`)}
	if err := admitter.Admit(invalid, 10); !errors.Is(err, ErrCommandPayload) {
		t.Fatalf("payload error = %v", err)
	}

	valid.Sequence = 2
	if err := admitter.Admit(valid, 10); err != nil {
		t.Fatalf("rejected sequence after failed validation: %v", err)
	}

	if err := admitter.Admit(valid, 10); !errors.Is(err, ErrCommandSequence) {
		t.Fatalf("replay error = %v", err)
	}

	valid.Sequence = 3

	valid.Tick = 13
	if err := admitter.Admit(valid, 10); !errors.Is(err, ErrCommandTick) {
		t.Fatalf("future tick error = %v", err)
	}

	ordered := NewAdmitter(5)
	if err := ordered.Register("move", func(Command) error { return nil }); err != nil {
		t.Fatal(err)
	}

	first := Command{Tick: 12, Player: "p", Sequence: 1, Kind: "move", Payload: []byte(`{}`)}
	if err := ordered.Admit(first, 10); err != nil {
		t.Fatal(err)
	}

	backward := Command{Tick: 11, Player: "p", Sequence: 2, Kind: "move", Payload: []byte(`{}`)}
	if err := ordered.Admit(backward, 10); !errors.Is(err, ErrCommandTick) {
		t.Fatalf("backward player tick error = %v", err)
	}
}

// TestAdmitterRequiresHandlerGrantedAdministrativeAuthority keeps privilege in trusted registration policy.
func TestAdmitterRequiresHandlerGrantedAdministrativeAuthority(t *testing.T) {
	admitter := NewAdmitter(1)
	if err := admitter.Register("move", func(Command) error { return nil }); err != nil {
		t.Fatal(err)
	}

	if err := admitter.RegisterAuthorities(
		"admin.spawn_item",
		func(Command) error { return nil },
		AuthorityAdmin,
		AuthoritySystem,
	); err != nil {
		t.Fatal(err)
	}

	admin := Command{Tick: 1, Player: "operator-1", Authority: AuthorityAdmin, Sequence: 1, Payload: []byte(`{}`)}

	admin.Kind = "move"
	if err := admitter.Admit(admin, 0); !errors.Is(err, ErrCommandAuthority) {
		t.Fatalf("admin player-command error = %v", err)
	}

	admin.Kind = "admin.spawn_item"
	if err := admitter.Admit(admin, 0); err != nil {
		t.Fatal(err)
	}

	player := Command{
		Tick: 1, Player: "player-1", Authority: AuthorityPlayer,
		Sequence: 1, Kind: "admin.spawn_item", Payload: []byte(`{}`),
	}
	if err := admitter.Admit(player, 0); !errors.Is(err, ErrCommandAuthority) {
		t.Fatalf("player admin-command error = %v", err)
	}
}
