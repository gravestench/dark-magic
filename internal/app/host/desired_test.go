package host

import "testing"

func TestParseDesired(t *testing.T) {
	desired, err := ParseDesired("", "boot")
	if err != nil || !desired["boot"] {
		t.Fatalf("desired=%v err=%v", desired, err)
	}
	desired, err = ParseDesired("none", "boot")
	if err != nil || len(desired) != 0 {
		t.Fatalf("desired=%v err=%v", desired, err)
	}
}
