package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// budget captures the persisted limits for one scene. Zero timing limits disable their optional checks, matching
// historical budget files that predate frame-timing enforcement.
type budget struct {
	MaxRetainedTextureBytes uint64 `json:"max_retained_texture_bytes"`
	MaxActiveResources      int    `json:"max_active_resources"`
	MaxDecodedWeight        int    `json:"max_decoded_weight"`
	MaxDecodeTimeMS         int64  `json:"max_decode_time_ms"`
	MinFrameSamples         int    `json:"min_frame_samples"`
	MaxFrameP95MS           int64  `json:"max_frame_p95_ms"`
	MaxUpdateP95MS          int64  `json:"max_update_p95_ms"`
}

// check validates every configured scene and accumulates failures so one run identifies the full repair surface.
// Sorted scene names and filepath.Glob's sorted matches keep the combined error stable across runs.
func check(profileDirectory, budgetPath string) error {
	budgets, err := readBudgets(budgetPath)
	if err != nil {
		return err
	}

	var result error

	for _, scene := range sortedSceneNames(budgets) {
		paths, err := sceneDiagnosticPaths(profileDirectory, scene)
		if err != nil || len(paths) == 0 {
			result = errors.Join(result, fmt.Errorf("profile check: scene %q has no diagnostics", scene))
			continue
		}

		for _, path := range paths {
			if err := checkSnapshot(scene, path, budgets[scene]); err != nil {
				result = errors.Join(result, err)
			}
		}
	}

	return result
}

// readBudgets owns budget-file decoding so callers receive errors that distinguish unavailable input from invalid
// JSON. The prefixes are part of the command's diagnostic contract and identify which setup phase failed.
func readBudgets(path string) (map[string]budget, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("profile check: read budgets: %w", err)
	}

	budgets := make(map[string]budget)
	if err := json.Unmarshal(data, &budgets); err != nil {
		return nil, fmt.Errorf("profile check: parse budgets: %w", err)
	}

	return budgets, nil
}

// sortedSceneNames converts map-backed configuration into deterministic validation and error-reporting order.
func sortedSceneNames(budgets map[string]budget) []string {
	names := make([]string, 0, len(budgets))
	for name := range budgets {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

// sceneDiagnosticPaths centralizes the artifact layout contract used by profile capture and verification. Glob
// returns lexical order, which check preserves when it aggregates per-snapshot failures.
func sceneDiagnosticPaths(profileDirectory, scene string) ([]string, error) {
	pattern := filepath.Join(profileDirectory, "scenes", scene, "diagnostics-*.json")

	return filepath.Glob(pattern)
}
