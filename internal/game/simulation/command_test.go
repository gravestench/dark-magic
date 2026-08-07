package simulation

import (
	"errors"
	"fmt"
	"testing"
)

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
	if err := ordered.Admit(Command{Tick: 12, Player: "p", Sequence: 1, Kind: "move", Payload: []byte(`{}`)}, 10); err != nil {
		t.Fatal(err)
	}
	if err := ordered.Admit(Command{Tick: 11, Player: "p", Sequence: 2, Kind: "move", Payload: []byte(`{}`)}, 10); !errors.Is(err, ErrCommandTick) {
		t.Fatalf("backward player tick error = %v", err)
	}
}
