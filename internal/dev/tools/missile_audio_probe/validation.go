package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// validate checks document provenance and every captured matrix row, joining failures into one actionable report.
func validate(captured capture) error {
	var result error
	if captured.Schema != probeSchema {
		result = errors.Join(
			result,
			fmt.Errorf("missile audio probe: schema %q, want %q", captured.Schema, probeSchema),
		)
	}

	if captured.Target != probeTarget {
		result = errors.Join(
			result,
			fmt.Errorf("missile audio probe: target %q, want %q", captured.Target, probeTarget),
		)
	}

	if captured.Source != "owned-runtime" {
		result = errors.Join(result, fmt.Errorf("missile audio probe: source must be owned-runtime"))
	}

	result = errors.Join(result, validateRuntime(captured.Runtime))
	if len(captured.Cases) == 0 {
		result = errors.Join(result, fmt.Errorf("missile audio probe: at least one case is required"))
	}

	seen := make(map[string]bool, len(captured.Cases))
	for _, observed := range captured.Cases {
		if observed.ID == "" || seen[observed.ID] {
			result = errors.Join(result, fmt.Errorf("missile audio probe: case IDs must be non-empty and unique"))
		}

		seen[observed.ID] = true
		result = errors.Join(result, validateCase(observed))
	}

	return result
}

// validateRuntime enforces the controlled environment that makes separate captures meaningfully comparable.
func validateRuntime(observed runtime) error {
	var result error
	if observed.Patch != "1.14d" || observed.Mode != "expansion" || observed.Session != "single-player" ||
		observed.CharacterOrigin != "probe-created" {
		result = errors.Join(
			result,
			fmt.Errorf(
				"missile audio probe: runtime must be a probe-created Expansion 1.14d single-player character",
			),
		)
	}

	if !validSHA256(observed.ExecutableSHA256) || !validGenerationID(observed.RecordGenerationID) {
		result = errors.Join(
			result,
			fmt.Errorf(
				"missile audio probe: executable and immutable record-generation SHA-256 values are required",
			),
		)
	}

	if observed.Observation != "isolated-audio-video-frame-log" ||
		observed.SoundIdentification != "owned-mpq-waveform-comparison" ||
		observed.GameFramesPerSecond != 25 || !observed.AudioIsolated || !observed.CameraFixed ||
		!observed.ActorsStationary {
		result = errors.Join(
			result,
			fmt.Errorf(
				"missile audio probe: requires isolated 25 Hz audio/video observation with "+
					"owned-MPQ waveform comparison, a fixed camera, and stationary actors",
			),
		)
	}

	return result
}

// validateCase checks one observation against its immutable matrix row and controlled visual timeline.
func validateCase(observed probeCase) error {
	spec, found := specFor(observed.ID)
	if !found || observed.SkillID != spec.skillID || observed.SkillRecord != spec.skill ||
		observed.SkillLevel != 1 || observed.MissileRecord != spec.missile ||
		observed.Outcome != spec.outcome || observed.TargetCount != spec.targets {
		return fmt.Errorf("missile audio probe: case %q does not match its target-locked matrix row", observed.ID)
	}

	var result error
	if observed.Difficulty != "normal" || observed.Area != "blood_moor" ||
		len(observed.TargetRecords) != observed.TargetCount {
		result = errors.Join(
			result,
			fmt.Errorf(
				"missile audio probe: case %q must use the controlled Normal Blood Moor target set",
				observed.ID,
			),
		)
	}

	for _, target := range observed.TargetRecords {
		if target != "fallen1" {
			result = errors.Join(
				result,
				fmt.Errorf("missile audio probe: case %q target record must be fallen1", observed.ID),
			)
		}
	}

	result = errors.Join(result, validateTimeline(observed, spec))
	result = errors.Join(result, validateSounds(observed, spec))

	return result
}

// validateTimeline rejects impossible frame relationships before normalization could make them look plausible.
func validateTimeline(observed probeCase, spec caseSpec) error {
	var result error
	if !validFrame(observed.CastEffectFrame) || !validFrame(observed.MissileRemovedFrame) ||
		observed.MissileRemovedFrame < observed.CastEffectFrame || observed.ProjectileVisualCount <= 0 {
		result = errors.Join(
			result,
			fmt.Errorf("missile audio probe: case %q has an invalid visual timeline", observed.ID),
		)
	}

	if spec.outcome == "expired" {
		if observed.ContactFrame != nil {
			result = errors.Join(
				result,
				fmt.Errorf("missile audio probe: expired case %q cannot have contact", observed.ID),
			)
		}

		return result
	}

	if observed.ContactFrame == nil || !validFrame(*observed.ContactFrame) ||
		*observed.ContactFrame < observed.CastEffectFrame ||
		*observed.ContactFrame > observed.MissileRemovedFrame {
		result = errors.Join(
			result,
			fmt.Errorf(
				"missile audio probe: contact case %q requires an in-lifetime contact frame",
				observed.ID,
			),
		)
	}

	return result
}

// validateSounds requires exactly the records named by the matrix and validates each observed interval.
func validateSounds(observed probeCase, spec caseSpec) error {
	var result error

	byRecord := make(map[string]soundObservation, len(observed.Sounds))
	for _, sound := range observed.Sounds {
		if _, duplicate := byRecord[sound.Record]; duplicate {
			result = errors.Join(
				result,
				fmt.Errorf("missile audio probe: case %q duplicates sound %q", observed.ID, sound.Record),
			)
		}

		byRecord[sound.Record] = sound
	}

	for _, expected := range spec.sounds {
		sound, exists := byRecord[expected.record]
		if !exists || sound.Role != expected.role {
			result = errors.Join(
				result,
				fmt.Errorf(
					"missile audio probe: case %q omits or mislabels %s sound %q",
					observed.ID,
					expected.role,
					expected.record,
				),
			)

			continue
		}

		result = errors.Join(result, validateSound(observed.ID, observed.CastEffectFrame, sound))

		delete(byRecord, expected.record)
	}

	for record := range byRecord {
		result = errors.Join(
			result,
			fmt.Errorf("missile audio probe: case %q invents sound record %q", observed.ID, record),
		)
	}

	return result
}

// validateSound keeps presence flags and frame intervals internally consistent and non-overlapping.
func validateSound(caseID string, castEffectFrame int, sound soundObservation) error {
	if sound.Present != (len(sound.Intervals) > 0) {
		return fmt.Errorf(
			"missile audio probe: case %q sound %q contradicts presence",
			caseID,
			sound.Record,
		)
	}

	previous := -1
	for index, interval := range sound.Intervals {
		if !validFrame(interval.FirstFrame) || !validFrame(interval.LastFrame) ||
			interval.FirstFrame < castEffectFrame || interval.LastFrame < interval.FirstFrame ||
			interval.FirstFrame <= previous {
			return fmt.Errorf(
				"missile audio probe: case %q sound %q interval %d is invalid or overlaps",
				caseID,
				sound.Record,
				index,
			)
		}

		previous = interval.LastFrame
	}

	return nil
}

// validFrame bounds observations to a generous but finite capture window, catching sentinel and corrupt values.
func validFrame(value int) bool {
	return value >= 0 && value <= maxFrame
}

// validSHA256 accepts only canonical lowercase digests so equivalent identities have one representation.
func validSHA256(value string) bool {
	if len(value) != 64 || value != strings.ToLower(value) {
		return false
	}

	_, err := hex.DecodeString(value)

	return err == nil
}

// validGenerationID requires the explicit algorithm prefix used by immutable generated-record snapshots.
func validGenerationID(value string) bool {
	return strings.HasPrefix(value, "sha256:") && validSHA256(strings.TrimPrefix(value, "sha256:"))
}
