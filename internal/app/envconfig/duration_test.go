package envconfig

import (
	"testing"
	"time"
)

// TestDurationUsesFallbackAndRejectsInvalidValues protects loops and deadlines
// from blank defaults, malformed input, and non-positive scheduling intervals.
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

// TestDurationRejectsBlankNames prevents programming mistakes from silently
// consulting an empty environment key and returning a misleading fallback.
func TestDurationRejectsBlankNames(t *testing.T) {
	if _, err := Duration("  ", time.Second); err == nil {
		t.Fatal("blank environment variable name was accepted")
	}
}

// assertDuration keeps the success cases compact while preserving the value and error contract.
func assertDuration(t *testing.T, name string, fallback, expected time.Duration) {
	t.Helper()

	value, err := Duration(name, fallback)
	if err != nil || value != expected {
		t.Fatalf("duration = %s, want %s, error = %v", value, expected, err)
	}
}
