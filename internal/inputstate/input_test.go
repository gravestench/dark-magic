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

func TestRouteAssignsOneOwnerAndSuppressesCapturedSceneInput(t *testing.T) {
	frame := Frame{Actions: map[string]ActionState{"confirm": {Pressed: true}}, Text: "x", CursorX: 3, CursorY: 4}
	routed := Route(frame, FocusOwner{Domain: FocusDebug, ID: "console"}, true, true, "center")
	if routed.Owner.Domain != FocusDebug || routed.Owner.ID != "console" || routed.Gameplay || len(routed.Actions) != 0 || routed.Text != "" || routed.CursorX != 3 || routed.CursorY != 4 {
		t.Fatalf("routed frame = %#v", routed)
	}
}

func TestStoreSuppressesNonfocusedCallbackAndRestoresOwnerAccess(t *testing.T) {
	var store Store
	store.Publish(Frame{Actions: map[string]ActionState{"confirm": {Pressed: true}}, Text: "x", CursorX: 3, CursorY: 4, Owner: FocusOwner{Domain: FocusScene, ID: "overlay"}})
	if err := store.Suppress(func() error {
		if store.Action("confirm").Pressed || store.Text() != "" {
			t.Fatal("suppressed callback observed actions or text")
		}
		if x, y := store.Cursor(); x != 0 || y != 0 {
			t.Fatalf("suppressed cursor = %v,%v", x, y)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !store.Action("confirm").Pressed || store.Text() != "x" || store.Owner().ID != "overlay" {
		t.Fatal("owner access was not restored")
	}
}

func TestStoreGameplayOnlyHidesOverlayActionsAndPreservesWorldActions(t *testing.T) {
	var store Store
	store.Publish(Frame{Actions: map[string]ActionState{"right": {Down: true}, "inventory": {Pressed: true}}, Text: "x", CursorX: 3, CursorY: 4})
	if err := store.GameplayOnly(func() error {
		if !store.Action("right").Down || store.Action("inventory").Pressed || store.Text() != "" {
			t.Fatal("gameplay channel routing is incorrect")
		}
		if x, y := store.Cursor(); x != 3 || y != 4 {
			t.Fatalf("gameplay cursor = %v,%v", x, y)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestStoreGameplayOnlyGatesPointerActionToVisibleWorldHalf(t *testing.T) {
	var store Store
	store.Publish(Frame{
		Actions:   map[string]ActionState{"pointer_primary": {Pressed: true}},
		CursorX:   600,
		WorldView: "left",
	})
	if err := store.GameplayOnly(func() error {
		if store.Action("pointer_primary").Pressed {
			t.Fatal("pointer action passed through the covered panel half")
		}
		if store.Snapshot().Actions["pointer_primary"].Pressed {
			t.Fatal("snapshot exposed pointer action over the covered panel half")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
