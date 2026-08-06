package inputstate

import "testing"

func TestStoreClonesPublishedAndReturnedFrames(t *testing.T) {
	t.Parallel()

	var store Store
	frame := Frame{Actions: map[string]ActionState{"confirm": {Pressed: true}}}
	store.Publish(frame)
	frame.Actions["confirm"] = ActionState{}
	got := store.Snapshot()
	if !got.Actions["confirm"].Pressed {
		t.Fatal("published frame was mutated by caller")
	}
	got.Actions["confirm"] = ActionState{}
	if !store.Snapshot().Actions["confirm"].Pressed {
		t.Fatal("stored frame was mutated through snapshot")
	}
}

func TestActionAndCursorHotPathDoesNotAllocate(t *testing.T) {
	var store Store
	store.Publish(Frame{Actions: map[string]ActionState{"confirm": {Pressed: true}}, CursorX: 3, CursorY: 4})
	allocations := testing.AllocsPerRun(1000, func() {
		if !store.Action("confirm").Pressed {
			panic("missing action")
		}
		x, y := store.Cursor()
		if x != 3 || y != 4 {
			panic("wrong cursor")
		}
	})
	if allocations != 0 {
		t.Fatalf("hot-path allocations = %v", allocations)
	}
}
