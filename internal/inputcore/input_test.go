package inputcore

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
