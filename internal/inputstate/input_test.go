package inputstate

import "testing"

// TestStoreClonesPublishedAndReturnedFrames verifies that callers cannot mutate the shared input snapshot through
// either the frame they publish or a frame they receive.
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

// TestActionAndCursorHotPathDoesNotAllocate protects the per-frame polling path from introducing garbage collection
// pressure as more systems query the same input state.
func TestActionAndCursorHotPathDoesNotAllocate(t *testing.T) {
	var store Store
	store.Publish(Frame{
		Actions: map[string]ActionState{"confirm": {Pressed: true}},
		CursorX: 3,
		CursorY: 4,
	})

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

// TestRouteAssignsOneOwnerAndSuppressesCapturedSceneInput verifies that a focused overlay receives sole ownership
// while gameplay input is hidden, without discarding cursor coordinates needed by the overlay.
func TestRouteAssignsOneOwnerAndSuppressesCapturedSceneInput(t *testing.T) {
	frame := Frame{
		Actions: map[string]ActionState{"confirm": {Pressed: true}},
		Text:    "x",
		CursorX: 3,
		CursorY: 4,
	}

	routed := Route(frame, FocusOwner{Domain: FocusDebug, ID: "console"}, true, true, "center")
	if routed.Owner.Domain != FocusDebug || routed.Owner.ID != "console" || routed.Gameplay ||
		len(routed.Actions) != 0 || routed.Text != "" || routed.CursorX != 3 || routed.CursorY != 4 {
		t.Fatalf("routed frame = %#v", routed)
	}
}

// TestStoreSuppressesNonfocusedCallbackAndRestoresOwnerAccess verifies that temporary suppression is scoped to the
// callback and cannot leak into later reads by the rightful focus owner.
func TestStoreSuppressesNonfocusedCallbackAndRestoresOwnerAccess(t *testing.T) {
	var store Store
	store.Publish(Frame{
		Actions: map[string]ActionState{"confirm": {Pressed: true}},
		Text:    "x",
		CursorX: 3,
		CursorY: 4,
		Owner:   FocusOwner{Domain: FocusScene, ID: "overlay"},
	})

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

// TestStoreGameplayOnlyHidesOverlayActionsAndPreservesWorldActions verifies that gameplay callbacks see movement and
// cursor state while overlay commands and text remain confined to the UI that owns them.
func TestStoreGameplayOnlyHidesOverlayActionsAndPreservesWorldActions(t *testing.T) {
	var store Store
	store.Publish(Frame{
		Actions: map[string]ActionState{
			"right":     {Down: true},
			"inventory": {Pressed: true},
		},
		Text:    "x",
		CursorX: 3,
		CursorY: 4,
	})

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

// TestStoreGameplayOnlyGatesPointerActionToVisibleWorldHalf verifies that panel-covered pixels cannot trigger world
// interaction even though the physical pointer remains visible there.
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

// TestStoreGameplayOnlyUsesProfileWorldSplit verifies that pointer gating honors the active profile's viewport split
// instead of assuming a resolution-independent midpoint.
func TestStoreGameplayOnlyUsesProfileWorldSplit(t *testing.T) {
	var store Store
	store.Publish(Frame{
		Actions:    map[string]ActionState{"pointer_primary": {Pressed: true}},
		CursorX:    350,
		WorldView:  "left",
		WorldSplit: 320,
	})

	if err := store.GameplayOnly(func() error {
		if store.Action("pointer_primary").Pressed {
			t.Fatal("640x480 covered half accepted world pointer")
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

// TestStorePointerOnlyExposesNoKeyboardOrText verifies that pointer-only consumers cannot accidentally respond to
// keyboard actions or text intended for another input channel.
func TestStorePointerOnlyExposesNoKeyboardOrText(t *testing.T) {
	var store Store
	store.Publish(Frame{
		Actions: map[string]ActionState{
			"pointer_primary": {Pressed: true},
			"confirm":         {Pressed: true},
			"left":            {Down: true},
		},
		Text:    "x",
		CursorX: 600,
		CursorY: 300,
	})

	if err := store.PointerOnly(func() error {
		if !store.Action("pointer_primary").Pressed || store.Action("confirm").Pressed ||
			store.Action("left").Down || store.Text() != "" {
			t.Fatalf("pointer-only routing exposed another channel: %#v text=%q", store.Snapshot(), store.Text())
		}

		if x, y := store.Cursor(); x != 600 || y != 300 {
			t.Fatalf("pointer-only cursor = %v,%v", x, y)
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

// TestStoreGameplayAndPointerKeepsHUDPointerOutsideWorldView verifies that the combined channel keeps raw pointer
// actions for HUD widgets while continuing to hide unrelated overlay commands.
func TestStoreGameplayAndPointerKeepsHUDPointerOutsideWorldView(t *testing.T) {
	var store Store
	store.Publish(Frame{
		Actions:   map[string]ActionState{"pointer_primary": {Pressed: true}, "inventory": {Pressed: true}},
		CursorX:   600,
		WorldView: "left",
	})

	if err := store.GameplayAndPointer(func() error {
		if !store.Action("pointer_primary").Pressed || store.Action("inventory").Pressed {
			t.Fatalf("gameplay-plus-pointer routing = %#v", store.Snapshot().Actions)
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
