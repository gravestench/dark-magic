package simulation

import "testing"

func TestNamedStreamsAreIndependentAndRestorable(t *testing.T) {
	loot := NewStream(42, "loot")
	combat := NewStream(42, "combat")
	if loot.Uint64() == combat.Uint64() {
		t.Fatal("distinct stream names produced the same first value")
	}
	state := loot.State()
	want := loot.Uint64()
	if got := RestoreStream(state).Uint64(); got != want {
		t.Fatalf("restored value = %d, want %d", got, want)
	}
	untouched := NewStream(42, "combat")
	_ = NewStream(42, "loot").Uint64()
	if got, want := untouched.Uint64(), NewStream(42, "combat").Uint64(); got != want {
		t.Fatalf("another stream perturbed combat: %d != %d", got, want)
	}
}

func TestUint64nIsBoundedAndZeroSafe(t *testing.T) {
	stream := NewStream(7, "test")
	if stream.Uint64n(0) != 0 {
		t.Fatal("zero limit did not return zero")
	}
	for range 1000 {
		if value := stream.Uint64n(7); value >= 7 {
			t.Fatalf("value %d outside limit", value)
		}
	}
}
