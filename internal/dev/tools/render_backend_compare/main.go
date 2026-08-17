// Command render_backend_compare summarizes matched native-backend profiles.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type diagnostics struct {
	RenderBackend struct {
		Name        string          `json:"name"`
		Diagnostics json.RawMessage `json:"diagnostics"`
	} `json:"render_backend"`
	FrameTiming map[string]frameTiming `json:"frame_timing"`
}

type frameTiming struct {
	Samples   int           `json:"samples"`
	FrameP50  time.Duration `json:"frame_p50"`
	FrameP95  time.Duration `json:"frame_p95"`
	UpdateP50 time.Duration `json:"update_p50"`
	UpdateP95 time.Duration `json:"update_p95"`
}

type nativeTiming struct {
	LastFrameDrawCalls     uint64 `json:"LastFrameDrawCalls"`
	LastFrameNodesVisited  uint64 `json:"LastFrameNodesVisited"`
	LastFrameCompositionNS uint64 `json:"LastFrameCompositionNS"`
	LastFrameRenderNS      uint64 `json:"LastFrameRenderNS"`
	TextureUploadNS        uint64 `json:"TextureUploadNS"`
}

type result struct {
	backend string
	frame   frameTiming
	native  nativeTiming
}

func main() {
	directory := flag.String("profile-dir", "profiles/render-backends", "directory containing raylib and ebitengine profiles")
	scene := flag.String("scene", "game_world", "matched scene name")
	flag.Parse()

	results := make([]result, 0, 2)
	for _, backend := range []string{"raylib", "ebitengine"} {
		entry, err := readResult(filepath.Join(*directory, backend, "diagnostics.json"), *scene)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		results = append(results, entry)
	}

	summary := fmt.Sprintf("# Native renderer comparison\n\nMatched scene: `%s`. Native audio is disabled for both runs. Durations are steady-window application metrics; native render and composition values are the final captured frame.\n\n", *scene)
	summary += "| Backend | Samples | Frame p50 | Frame p95 | Update p50 | Update p95 | Native render | Composition | Draws | Nodes | Upload CPU |\n"
	summary += "| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n"
	for _, entry := range results {
		summary += fmt.Sprintf("| %s | %d | %s | %s | %s | %s | %s | %s | %d | %d | %s |\n",
			entry.backend, entry.frame.Samples, duration(entry.frame.FrameP50), duration(entry.frame.FrameP95), duration(entry.frame.UpdateP50), duration(entry.frame.UpdateP95),
			duration(time.Duration(entry.native.LastFrameRenderNS)), duration(time.Duration(entry.native.LastFrameCompositionNS)), entry.native.LastFrameDrawCalls, entry.native.LastFrameNodesVisited, duration(time.Duration(entry.native.TextureUploadNS)))
	}
	path := filepath.Join(*directory, "summary.md")
	if err := os.WriteFile(path, []byte(summary), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(summary)
}

func readResult(path, scene string) (result, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return result{}, fmt.Errorf("read %s: %w", path, err)
	}
	var document diagnostics
	if err := json.Unmarshal(data, &document); err != nil {
		return result{}, fmt.Errorf("decode %s: %w", path, err)
	}
	frame, ok := document.FrameTiming[scene]
	if !ok {
		return result{}, fmt.Errorf("%s has no frame timing for scene %q", path, scene)
	}
	var native nativeTiming
	if err := json.Unmarshal(document.RenderBackend.Diagnostics, &native); err != nil {
		return result{}, fmt.Errorf("decode native diagnostics in %s: %w", path, err)
	}
	return result{backend: document.RenderBackend.Name, frame: frame, native: native}, nil
}

func duration(value time.Duration) string {
	if value == 0 {
		return "0"
	}
	if rounded := value.Round(time.Microsecond); rounded > 0 {
		return rounded.String()
	}
	return "<1µs"
}
