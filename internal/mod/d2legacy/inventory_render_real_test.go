package d2legacy

import "testing"

// The normal inventory must show owned items after earlier modal scenes release nodes.
func TestOwnedInventoryContentsDrawAbovePanelsAfterSceneTeardown(t *testing.T) {
	testOwnedOverlayAfterSceneTeardown(t, "inventory", 4, 7)
}
