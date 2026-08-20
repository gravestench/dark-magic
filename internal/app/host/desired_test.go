package host

import "testing"

// TestParseDesired verifies the compact CLI syntax preserves its default and explicit-disable meanings.
func TestParseDesired(t *testing.T) {
	desired, err := ParseDesired("", "boot")
	if err != nil || !desired["boot"] {
		t.Fatalf("desired=%v err=%v", desired, err)
	}

	// "none" must override the default because callers use it to request an intentionally empty runtime.
	desired, err = ParseDesired("none", "boot")
	if err != nil || len(desired) != 0 {
		t.Fatalf("desired=%v err=%v", desired, err)
	}
}
