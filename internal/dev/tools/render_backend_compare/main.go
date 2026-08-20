// Command render_backend_compare summarizes matched native-backend profiles.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// comparedBackends defines report order as well as the required profile set; changing it changes the output contract.
var comparedBackends = [...]string{"raylib", "ebitengine"}

// config identifies the profile set and scene that must be compared together.
// Keeping both values together prevents later phases from accidentally reading mismatched runs.
type config struct {
	profileDirectory string
	scene            string
}

// diagnostics models only the profile fields used by this comparison tool. Unknown fields remain forward-compatible.
type diagnostics struct {
	RenderBackend struct {
		Name        string          `json:"name"`
		Diagnostics json.RawMessage `json:"diagnostics"`
	} `json:"render_backend"`
	FrameTiming map[string]frameTiming `json:"frame_timing"`
}

// frameTiming contains the steady-window application measurements displayed in the summary.
type frameTiming struct {
	Samples   int           `json:"samples"`
	FrameP50  time.Duration `json:"frame_p50"`
	FrameP95  time.Duration `json:"frame_p95"`
	UpdateP50 time.Duration `json:"update_p50"`
	UpdateP95 time.Duration `json:"update_p95"`
}

// nativeTiming contains the final-frame backend counters displayed beside the application measurements.
type nativeTiming struct {
	LastFrameDrawCalls     uint64 `json:"LastFrameDrawCalls"`
	LastFrameNodesVisited  uint64 `json:"LastFrameNodesVisited"`
	LastFrameCompositionNS uint64 `json:"LastFrameCompositionNS"`
	LastFrameRenderNS      uint64 `json:"LastFrameRenderNS"`
	TextureUploadNS        uint64 `json:"TextureUploadNS"`
}

// comparisonResult keeps one backend's application and native measurements aligned for report rendering.
type comparisonResult struct {
	backend string
	frame   frameTiming
	native  nativeTiming
}

// main reads both known backend profiles before writing and printing one deterministic Markdown comparison.
// Failing before either output prevents a partial report from looking like a valid comparison.
func main() {
	configuration := parseConfig()

	results, err := readComparisonResults(configuration)
	if err != nil {
		fatal(err)
	}

	summary := buildSummary(configuration.scene, results)

	summaryPath := filepath.Join(configuration.profileDirectory, "summary.md")
	if err := os.WriteFile(summaryPath, []byte(summary), 0o644); err != nil {
		fatal(err)
	}

	// Print the same content written to disk so automation and humans inspect an identical report.
	fmt.Print(summary)
}

// parseConfig defines the command-line contract in one place so the comparison workflow reads as data flow.
func parseConfig() config {
	profileDirectory := flag.String(
		"profile-dir",
		"profiles/render-backends",
		"directory containing raylib and ebitengine profiles",
	)
	scene := flag.String("scene", "game_world", "matched scene name")

	flag.Parse()

	return config{
		profileDirectory: *profileDirectory,
		scene:            *scene,
	}
}

// readComparisonResults loads profiles in presentation order. Stable ordering keeps generated reports diff-friendly.
func readComparisonResults(configuration config) ([]comparisonResult, error) {
	results := make([]comparisonResult, 0, len(comparedBackends))
	for _, backend := range comparedBackends {
		profilePath := filepath.Join(configuration.profileDirectory, backend, "diagnostics.json")

		entry, err := readResult(profilePath, configuration.scene)
		if err != nil {
			return nil, err
		}

		results = append(results, entry)
	}

	return results, nil
}

// readResult decodes one profile and rejects absent scene data so zeros cannot masquerade as measurements.
func readResult(path, scene string) (comparisonResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return comparisonResult{}, fmt.Errorf("read %s: %w", path, err)
	}

	var document diagnostics
	if err := json.Unmarshal(data, &document); err != nil {
		return comparisonResult{}, fmt.Errorf("decode %s: %w", path, err)
	}

	frame, ok := document.FrameTiming[scene]
	if !ok {
		return comparisonResult{}, fmt.Errorf("%s has no frame timing for scene %q", path, scene)
	}

	var native nativeTiming
	if err := json.Unmarshal(document.RenderBackend.Diagnostics, &native); err != nil {
		return comparisonResult{}, fmt.Errorf("decode native diagnostics in %s: %w", path, err)
	}

	return comparisonResult{backend: document.RenderBackend.Name, frame: frame, native: native}, nil
}

// buildSummary renders the established Markdown schema without reordering backends or changing duration precision.
func buildSummary(scene string, results []comparisonResult) string {
	var summary strings.Builder
	summary.WriteString("# Native renderer comparison\n\nMatched scene: `")
	summary.WriteString(scene)
	summary.WriteString(
		"`. Native audio is disabled for both runs. Durations are steady-window application metrics; native " +
			"render and composition values are the final captured frame.\n\n",
	)
	summary.WriteString(
		"| Backend | Samples | Frame p50 | Frame p95 | Update p50 | Update p95 | Native render | " +
			"Composition | Draws | Nodes | Upload CPU |\n",
	)
	summary.WriteString("| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n")

	for _, entry := range results {
		summary.WriteString(formatSummaryRow(entry))
	}

	return summary.String()
}

// formatSummaryRow keeps column ordering explicit because downstream readers may treat the Markdown as structured data.
func formatSummaryRow(entry comparisonResult) string {
	return fmt.Sprintf(
		"| %s | %d | %s | %s | %s | %s | %s | %s | %d | %d | %s |\n",
		entry.backend,
		entry.frame.Samples,
		formatDuration(entry.frame.FrameP50),
		formatDuration(entry.frame.FrameP95),
		formatDuration(entry.frame.UpdateP50),
		formatDuration(entry.frame.UpdateP95),
		formatDuration(time.Duration(entry.native.LastFrameRenderNS)),
		formatDuration(time.Duration(entry.native.LastFrameCompositionNS)),
		entry.native.LastFrameDrawCalls,
		entry.native.LastFrameNodesVisited,
		formatDuration(time.Duration(entry.native.TextureUploadNS)),
	)
}

// formatDuration preserves useful sub-millisecond precision while giving zero and sub-microsecond values distinct text.
func formatDuration(value time.Duration) string {
	if value == 0 {
		return "0"
	}

	if rounded := value.Round(time.Microsecond); rounded > 0 {
		return rounded.String()
	}

	return "<1µs"
}

// fatal reports a user-facing failure without extra decoration and terminates before producing misleading output.
func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
