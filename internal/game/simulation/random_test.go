package simulation

import (
	"bytes"
	"testing"
)

// TestNamedStreamsAreIndependentAndRestorable protects domain isolation and exact continuation after restore.
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

// TestUint64nIsBoundedAndZeroSafe covers the total primitive used beneath stricter registry validation.
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

// TestRandomStreamsRestoreExactNamedSequences ensures checkpoints resume each purpose stream independently.
func TestRandomStreamsRestoreExactNamedSequences(t *testing.T) {
	streams := NewRandomStreams(42)
	if err := streams.Register("d2legacy.combat.hit"); err != nil {
		t.Fatal(err)
	}

	if err := streams.Register("d2legacy.loot.quality"); err != nil {
		t.Fatal(err)
	}

	_, _ = streams.Uint64n("d2legacy.combat.hit", 100)

	snapshot, err := streams.SnapshotState()
	if err != nil {
		t.Fatal(err)
	}

	want, err := streams.Uint64n("d2legacy.combat.hit", 100)
	if err != nil {
		t.Fatal(err)
	}

	if err := streams.RestoreState(snapshot); err != nil {
		t.Fatal(err)
	}

	got, err := streams.Uint64n("d2legacy.combat.hit", 100)
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Fatalf("restored roll = %d, want %d", got, want)
	}

	restored, err := streams.SnapshotState()
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(snapshot, restored) {
		t.Fatal("drawing after restore did not advance the checkpointed stream")
	}
}

// TestRandomStreamsRejectUnknownAndChangedRegistrations prevents topology drift during authoritative replay.
func TestRandomStreamsRejectUnknownAndChangedRegistrations(t *testing.T) {
	streams := NewRandomStreams(7)
	if err := streams.Register("combat"); err != nil {
		t.Fatal(err)
	}

	if _, err := streams.Uint64n("loot", 10); err == nil {
		t.Fatal("unknown stream was accepted")
	}

	snapshot, err := streams.SnapshotState()
	if err != nil {
		t.Fatal(err)
	}

	different := NewRandomStreams(7)
	if err := different.Register("loot"); err != nil {
		t.Fatal(err)
	}

	if err := different.RestoreState(snapshot); err == nil {
		t.Fatal("changed registration restored successfully")
	}
}
