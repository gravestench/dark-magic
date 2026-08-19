package envconfig

import (
	"testing"
	"time"
)

// TestDurationUsesFallbackAndRejectsInvalidValues covers fallback, configured,
// and invalid environment duration behavior.
func TestDurationUsesFallbackAndRejectsInvalidValues(t *testing.T) {
	const name = "DARK_MAGIC_TEST_DURATION"
	t.Setenv(name, "")
	assertDuration(t, name, 15*time.Second, 15*time.Second)

	t.Setenv(name, "250ms")
	assertDuration(t, name, 15*time.Second, 250*time.Millisecond)

	for _, invalid := range []string{"invalid", "0s", "-1s"} {
		t.Setenv(name, invalid)
		if _, err := Duration(name, 15*time.Second); err == nil {
			t.Fatalf("invalid duration %q was accepted", invalid)
		}
	}
}

// TestDurationRejectsBlankNames verifies that whitespace is not a variable name.
func TestDurationRejectsBlankNames(t *testing.T) {
	if _, err := Duration("  ", time.Second); err == nil {
		t.Fatal("blank environment variable name was accepted")
	}
}

// assertDuration verifies one successful Duration lookup.
func assertDuration(t *testing.T, name string, fallback, expected time.Duration) {
	t.Helper()
	value, err := Duration(name, fallback)
	if err != nil || value != expected {
		t.Fatalf("duration = %s, want %s, error = %v", value, expected, err)
	}
}
