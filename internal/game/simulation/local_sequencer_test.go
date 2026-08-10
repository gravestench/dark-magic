package simulation

import "testing"

func TestLocalSequencerCombinesIndependentProducersPerPlayer(t *testing.T) {
	sequencer := NewLocalSequencer()
	first := sequencer.Assign([]Command{
		{Player: "alice", Sequence: 1, Kind: "entry"},
		{Player: "alice", Sequence: 1, Kind: "interaction"},
		{Player: "bob", Sequence: 99, Kind: "move"},
	})
	if first[0].Sequence != 1 || first[1].Sequence != 2 || first[2].Sequence != 1 {
		t.Fatalf("first assignment = %#v", first)
	}
	second := sequencer.Assign([]Command{{Player: "alice", Kind: "item"}, {Player: "bob", Kind: "skill"}})
	if second[0].Sequence != 3 || second[1].Sequence != 2 {
		t.Fatalf("second assignment = %#v", second)
	}
}

func TestLocalSequencerDoesNotMutateProducerSlice(t *testing.T) {
	input := []Command{{Player: "alice", Sequence: 42}}
	assigned := NewLocalSequencer().Assign(input)
	if input[0].Sequence != 42 || assigned[0].Sequence != 1 {
		t.Fatalf("input=%#v assigned=%#v", input, assigned)
	}
}
