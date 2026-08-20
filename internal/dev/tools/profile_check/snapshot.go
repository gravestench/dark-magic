package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

// snapshot models only the profiler fields governed by budgets. Unrecognized diagnostic fields remain forward
// compatible because encoding/json ignores them during this focused decode.
type snapshot struct {
	Composition compositionSnapshot            `json:"composition"`
	FrameTiming map[string]frameTimingSnapshot `json:"frame_timing"`
}

// compositionSnapshot groups resource retention and decoding measurements that describe scene construction cost.
type compositionSnapshot struct {
	Decoded  struct{ Weight int }
	Retained struct {
		ActiveResources      int
		RetainedTextureBytes uint64
	}
	DecodeTime time.Duration
}

// frameTimingSnapshot contains the sampled steady-state timings for one scene. Duration values retain the profiler's
// native nanosecond JSON representation, avoiding any conversion ambiguity before comparison.
type frameTimingSnapshot struct {
	Samples   int
	FrameP95  time.Duration `json:"frame_p95"`
	UpdateP95 time.Duration `json:"update_p95"`
}

// budgetViolations preserves discovery order while using errors.Join so callers can inspect every underlying failure.
type budgetViolations struct {
	combined error
}

// add appends one violation without wrapping successful checks. Skipping nil avoids changing the join tree between
// failures, preserving both historical error ordering and errors.Is/errors.As traversal.
func (violations *budgetViolations) add(err error) {
	if err == nil {
		return
	}

	violations.combined = errors.Join(violations.combined, err)
}

// checkSnapshot separates artifact I/O from budget evaluation so malformed snapshots cannot be mistaken for budget
// regressions. Read and parse failures stop this snapshot only; check continues validating later artifacts.
func checkSnapshot(scene, path string, limits budget) error {
	got, err := readSnapshot(path)
	if err != nil {
		return err
	}

	return validateSnapshot(scene, got, limits)
}

// readSnapshot decodes the profiler's persisted representation and preserves distinct read and parse diagnostics.
func readSnapshot(path string) (snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return snapshot{}, fmt.Errorf("profile check: read %s: %w", path, err)
	}

	var got snapshot
	if err := json.Unmarshal(data, &got); err != nil {
		return snapshot{}, fmt.Errorf("profile check: parse %s: %w", path, err)
	}

	return got, nil
}

// validateSnapshot evaluates composition before frame timing to retain the command's established error order. All
// checks run even after a violation, giving maintainers one complete report for the snapshot.
func validateSnapshot(scene string, got snapshot, limits budget) error {
	var violations budgetViolations

	recordCompositionViolations(&violations, scene, got.Composition, limits)
	recordFrameTimingViolations(&violations, scene, got.FrameTiming[scene], limits)

	return violations.combined
}

// recordCompositionViolations compares the hard construction budgets in their stable reporting order. Unlike the
// optional timing limits, zero composition maxima deliberately permit no retained or decoded cost.
func recordCompositionViolations(
	violations *budgetViolations,
	scene string,
	got compositionSnapshot,
	limits budget,
) {
	if got.Retained.RetainedTextureBytes > limits.MaxRetainedTextureBytes {
		violations.add(fmt.Errorf(
			"profile check: %s retained texture bytes %d exceed %d",
			scene,
			got.Retained.RetainedTextureBytes,
			limits.MaxRetainedTextureBytes,
		))
	}

	if got.Retained.ActiveResources > limits.MaxActiveResources {
		violations.add(fmt.Errorf(
			"profile check: %s active resources %d exceed %d",
			scene,
			got.Retained.ActiveResources,
			limits.MaxActiveResources,
		))
	}

	if got.Decoded.Weight > limits.MaxDecodedWeight {
		violations.add(fmt.Errorf(
			"profile check: %s decoded weight %d exceeds %d",
			scene,
			got.Decoded.Weight,
			limits.MaxDecodedWeight,
		))
	}

	if got.DecodeTime > time.Duration(limits.MaxDecodeTimeMS)*time.Millisecond {
		violations.add(fmt.Errorf(
			"profile check: %s cumulative decode time %s exceeds %dms",
			scene,
			got.DecodeTime,
			limits.MaxDecodeTimeMS,
		))
	}
}

// recordFrameTimingViolations applies only configured timing limits. Looking up a missing scene produces zero values,
// which intentionally fails a positive sample minimum while leaving disabled percentile checks untouched.
func recordFrameTimingViolations(
	violations *budgetViolations,
	scene string,
	got frameTimingSnapshot,
	limits budget,
) {
	if limits.MinFrameSamples > 0 && got.Samples < limits.MinFrameSamples {
		violations.add(fmt.Errorf(
			"profile check: %s frame samples %d below %d",
			scene,
			got.Samples,
			limits.MinFrameSamples,
		))
	}

	if limits.MaxFrameP95MS > 0 && got.FrameP95 > time.Duration(limits.MaxFrameP95MS)*time.Millisecond {
		violations.add(fmt.Errorf(
			"profile check: %s p95 frame interval %s exceeds %dms",
			scene,
			got.FrameP95,
			limits.MaxFrameP95MS,
		))
	}

	if limits.MaxUpdateP95MS > 0 && got.UpdateP95 > time.Duration(limits.MaxUpdateP95MS)*time.Millisecond {
		violations.add(fmt.Errorf(
			"profile check: %s p95 update time %s exceeds %dms",
			scene,
			got.UpdateP95,
			limits.MaxUpdateP95MS,
		))
	}
}
