package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestBuildSummaryPreservesMarkdownSchema protects column order and wording relied on by report readers.
func TestBuildSummaryPreservesMarkdownSchema(t *testing.T) {
	t.Parallel()

	results := []comparisonResult{
		{
			backend: "raylib",
			frame: frameTiming{
				Samples:   2,
				FrameP50:  time.Millisecond,
				FrameP95:  2 * time.Millisecond,
				UpdateP50: 300 * time.Microsecond,
				UpdateP95: 400 * time.Microsecond,
			},
			native: nativeTiming{
				LastFrameDrawCalls:     5,
				LastFrameNodesVisited:  6,
				LastFrameCompositionNS: uint64(700 * time.Microsecond),
				LastFrameRenderNS:      uint64(800 * time.Microsecond),
				TextureUploadNS:        uint64(900 * time.Microsecond),
			},
		},
	}

	summary := buildSummary("game_world", results)

	expectedFragments := []string{
		"# Native renderer comparison\n\n",
		"Matched scene: `game_world`.",
		"| Backend | Samples | Frame p50 | Frame p95 | Update p50 | Update p95 | Native render |",
		"| raylib | 2 | 1ms | 2ms | 300µs | 400µs | 800µs | 700µs | 5 | 6 | 900µs |\n",
	}
	for _, fragment := range expectedFragments {
		if !strings.Contains(summary, fragment) {
			t.Fatalf("summary missing %q:\n%s", fragment, summary)
		}
	}
}

// TestReadResultRejectsMissingScene ensures absent measurements fail instead of appearing as a zero-valued run.
func TestReadResultRejectsMissingScene(t *testing.T) {
	t.Parallel()

	profilePath := filepath.Join(t.TempDir(), "diagnostics.json")

	profile := `{"render_backend":{"name":"raylib","diagnostics":{}},"frame_timing":{}}`
	if err := os.WriteFile(profilePath, []byte(profile), 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	_, err := readResult(profilePath, "game_world")
	if err == nil || !strings.Contains(err.Error(), `has no frame timing for scene "game_world"`) {
		t.Fatalf("readResult() error = %v, want missing-scene error", err)
	}
}

// TestFormatDurationPreservesSpecialValues distinguishes absent timing from a measured sub-microsecond duration.
func TestFormatDurationPreservesSpecialValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value time.Duration
		want  string
	}{
		{name: "zero", value: 0, want: "0"},
		{name: "sub microsecond", value: time.Nanosecond, want: "<1µs"},
		{name: "rounded", value: 1_500 * time.Nanosecond, want: "2µs"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := formatDuration(test.value); got != test.want {
				t.Fatalf("formatDuration(%s) = %q, want %q", test.value, got, test.want)
			}
		})
	}
}
