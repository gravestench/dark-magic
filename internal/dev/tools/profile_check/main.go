// Command profile_check verifies scene diagnostic snapshots against tracked
// performance budgets. It intentionally consumes profiler artifacts rather
// than starting the graphical client itself.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type budget struct {
	MaxRetainedTextureBytes uint64 `json:"max_retained_texture_bytes"`
	MaxActiveResources      int    `json:"max_active_resources"`
	MaxDecodedWeight        int    `json:"max_decoded_weight"`
	MaxDecodeTimeMS         int64  `json:"max_decode_time_ms"`
	MinFrameSamples         int    `json:"min_frame_samples"`
	MaxFrameP95MS           int64  `json:"max_frame_p95_ms"`
	MaxUpdateP95MS          int64  `json:"max_update_p95_ms"`
}

type snapshot struct {
	Composition struct {
		Decoded  struct{ Weight int }
		Retained struct {
			ActiveResources      int
			RetainedTextureBytes uint64
		}
		DecodeTime time.Duration
	} `json:"composition"`
	FrameTiming map[string]struct {
		Samples   int
		FrameP95  time.Duration `json:"frame_p95"`
		UpdateP95 time.Duration `json:"update_p95"`
	} `json:"frame_timing"`
}

func main() {
	profileDirectory := flag.String("profile-dir", "./profiles/acceptance", "profiling artifact directory")
	budgetPath := flag.String("budgets", "./docs/profile-budgets.json", "scene budget JSON")
	flag.Parse()
	if err := check(*profileDirectory, *budgetPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func check(profileDirectory, budgetPath string) error {
	data, err := os.ReadFile(budgetPath)
	if err != nil {
		return fmt.Errorf("profile check: read budgets: %w", err)
	}
	budgets := make(map[string]budget)
	if err := json.Unmarshal(data, &budgets); err != nil {
		return fmt.Errorf("profile check: parse budgets: %w", err)
	}
	names := make([]string, 0, len(budgets))
	for name := range budgets {
		names = append(names, name)
	}
	sort.Strings(names)
	var result error
	for _, name := range names {
		paths, err := filepath.Glob(filepath.Join(profileDirectory, "scenes", name, "diagnostics-*.json"))
		if err != nil || len(paths) == 0 {
			result = errors.Join(result, fmt.Errorf("profile check: scene %q has no diagnostics", name))
			continue
		}
		for _, path := range paths {
			if err := checkSnapshot(name, path, budgets[name]); err != nil {
				result = errors.Join(result, err)
			}
		}
	}
	return result
}

func checkSnapshot(scene, path string, limits budget) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("profile check: read %s: %w", path, err)
	}
	var got snapshot
	if err := json.Unmarshal(data, &got); err != nil {
		return fmt.Errorf("profile check: parse %s: %w", path, err)
	}
	var result error
	if got.Composition.Retained.RetainedTextureBytes > limits.MaxRetainedTextureBytes {
		result = errors.Join(result, fmt.Errorf("profile check: %s retained texture bytes %d exceed %d", scene, got.Composition.Retained.RetainedTextureBytes, limits.MaxRetainedTextureBytes))
	}
	if got.Composition.Retained.ActiveResources > limits.MaxActiveResources {
		result = errors.Join(result, fmt.Errorf("profile check: %s active resources %d exceed %d", scene, got.Composition.Retained.ActiveResources, limits.MaxActiveResources))
	}
	if got.Composition.Decoded.Weight > limits.MaxDecodedWeight {
		result = errors.Join(result, fmt.Errorf("profile check: %s decoded weight %d exceeds %d", scene, got.Composition.Decoded.Weight, limits.MaxDecodedWeight))
	}
	if got.Composition.DecodeTime > time.Duration(limits.MaxDecodeTimeMS)*time.Millisecond {
		result = errors.Join(result, fmt.Errorf("profile check: %s cumulative decode time %s exceeds %dms", scene, got.Composition.DecodeTime, limits.MaxDecodeTimeMS))
	}
	timing := got.FrameTiming[scene]
	if limits.MinFrameSamples > 0 && timing.Samples < limits.MinFrameSamples {
		result = errors.Join(result, fmt.Errorf("profile check: %s frame samples %d below %d", scene, timing.Samples, limits.MinFrameSamples))
	}
	if limits.MaxFrameP95MS > 0 && timing.FrameP95 > time.Duration(limits.MaxFrameP95MS)*time.Millisecond {
		result = errors.Join(result, fmt.Errorf("profile check: %s p95 frame interval %s exceeds %dms", scene, timing.FrameP95, limits.MaxFrameP95MS))
	}
	if limits.MaxUpdateP95MS > 0 && timing.UpdateP95 > time.Duration(limits.MaxUpdateP95MS)*time.Millisecond {
		result = errors.Join(result, fmt.Errorf("profile check: %s p95 update time %s exceeds %dms", scene, timing.UpdateP95, limits.MaxUpdateP95MS))
	}
	return result
}
