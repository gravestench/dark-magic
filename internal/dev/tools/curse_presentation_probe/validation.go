package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

// validate accumulates independent evidence failures so operators can repair a capture in one pass. The checks stay
// in schema-to-case order to keep diagnostics consistent with the input's hierarchy.
func validate(captured capture) error {
	var result error
	if captured.Schema != probeSchema {
		result = errors.Join(
			result,
			fmt.Errorf("curse presentation probe: schema %q, want %q", captured.Schema, probeSchema),
		)
	}

	if captured.Target != probeTarget {
		result = errors.Join(
			result,
			fmt.Errorf("curse presentation probe: target %q, want %q", captured.Target, probeTarget),
		)
	}

	if captured.Source != "owned-runtime" {
		result = errors.Join(result, fmt.Errorf("curse presentation probe: source must be owned-runtime"))
	}

	runtime := captured.Runtime
	if runtime.Patch != "1.14d" || runtime.Mode != "expansion" || runtime.Session != "single-player" {
		result = errors.Join(
			result,
			fmt.Errorf("curse presentation probe: runtime must be Expansion 1.14d single-player"),
		)
	}

	if runtime.CharacterOrigin != "probe-created" {
		result = errors.Join(result, fmt.Errorf("curse presentation probe: character must be probe-created"))
	}

	if !validSHA256(runtime.ExecutableSHA256) {
		result = errors.Join(result, fmt.Errorf("curse presentation probe: executable SHA-256 is required"))
	}

	if !oneOf(runtime.Observation, "video-frame-log", "manual-frame-log") ||
		runtime.AssetIdentification != "owned-mpq-dcc-comparison" ||
		!runtime.CameraFixed ||
		!runtime.ActorsStationary {
		result = errors.Join(
			result,
			fmt.Errorf(
				"curse presentation probe: requires fixed-camera stationary visual observation with owned-MPQ DCC comparison",
			),
		)
	}

	if len(captured.Cases) == 0 {
		result = errors.Join(result, fmt.Errorf("curse presentation probe: at least one case is required"))
	}

	seenCaseIDs := make(map[string]bool, len(captured.Cases))
	for _, observed := range captured.Cases {
		if observed.ID == "" || seenCaseIDs[observed.ID] {
			result = errors.Join(result, fmt.Errorf("curse presentation probe: case IDs must be non-empty and unique"))
		}

		seenCaseIDs[observed.ID] = true

		result = errors.Join(result, validateCase(observed))
	}

	return result
}

// validateCase checks target context first and then reconciles the observed layers with the skill's referenced
// missiles. Retaining every independent error prevents one malformed layer from hiding other evidence problems.
func validateCase(observed probeCase) error {
	var result error

	expectedMissileRecords, expectedSkillRecord, knownSkill := expectedMissiles(observed.SkillID)
	if !knownSkill ||
		observed.SkillRecord != expectedSkillRecord ||
		!oneOf(observed.Difficulty, "normal", "nightmare", "hell") ||
		observed.Area == "" {
		result = errors.Join(
			result,
			fmt.Errorf("curse presentation probe: case %q has invalid target context", observed.ID),
		)
	}

	if len(observed.TargetRecords) != len(observed.Targets) {
		result = errors.Join(
			result,
			fmt.Errorf("curse presentation probe: case %q target records/anchors differ", observed.ID),
		)
	}

	if !validPoint(observed.Caster) || !validPoint(observed.Cursor) {
		result = errors.Join(
			result,
			fmt.Errorf("curse presentation probe: case %q has invalid caster/cursor coordinates", observed.ID),
		)
	}

	for index, target := range observed.Targets {
		if observed.TargetRecords[index] == "" || !validPoint(target) {
			result = errors.Join(
				result,
				fmt.Errorf("curse presentation probe: case %q target %d is invalid", observed.ID, index),
			)
		}
	}

	layersByRecord := make(map[string]layer, len(observed.Layers))
	for _, current := range observed.Layers {
		if _, duplicate := layersByRecord[current.MissileRecord]; duplicate {
			result = errors.Join(
				result,
				fmt.Errorf(
					"curse presentation probe: case %q duplicates layer %q",
					observed.ID,
					current.MissileRecord,
				),
			)
		}

		layersByRecord[current.MissileRecord] = current
	}

	// Removing recognized records leaves only invented missiles for the final rejection pass.
	for _, missileRecord := range expectedMissileRecords {
		current, exists := layersByRecord[missileRecord]
		if !exists {
			result = errors.Join(
				result,
				fmt.Errorf(
					"curse presentation probe: case %q omits referenced missile %q",
					observed.ID,
					missileRecord,
				),
			)

			continue
		}

		result = errors.Join(result, validateLayer(observed, current))

		delete(layersByRecord, missileRecord)
	}

	for missileRecord := range layersByRecord {
		result = errors.Join(
			result,
			fmt.Errorf(
				"curse presentation probe: case %q invents missile %q",
				observed.ID,
				missileRecord,
			),
		)
	}

	return result
}

// validateLayer enforces consistency between presence, lifetime, coordinates, and anchor ownership. A contradictory
// presence marker remains the primary layer error because its instances cannot be interpreted reliably.
func validateLayer(observed probeCase, current layer) error {
	if current.Present != (len(current.Instances) > 0) {
		return fmt.Errorf(
			"curse presentation probe: case %q layer %q contradicts presence",
			observed.ID,
			current.MissileRecord,
		)
	}

	var result error

	for index, item := range current.Instances {
		if item.FirstFrame < 0 ||
			item.LastFrame < item.FirstFrame ||
			!validPoint(item.Start) ||
			!validPoint(item.End) ||
			!oneOf(item.Anchor, "caster", "cursor", "target") {
			result = errors.Join(
				result,
				fmt.Errorf(
					"curse presentation probe: case %q layer %q instance %d is invalid",
					observed.ID,
					current.MissileRecord,
					index,
				),
			)

			continue
		}

		if item.Anchor == "target" {
			if item.TargetIndex == nil || *item.TargetIndex < 0 || *item.TargetIndex >= len(observed.Targets) {
				result = errors.Join(
					result,
					fmt.Errorf(
						"curse presentation probe: case %q layer %q instance %d has invalid target anchor",
						observed.ID,
						current.MissileRecord,
						index,
					),
				)
			}
		} else if item.TargetIndex != nil {
			result = errors.Join(
				result,
				fmt.Errorf(
					"curse presentation probe: case %q layer %q instance %d has target index for non-target anchor",
					observed.ID,
					current.MissileRecord,
					index,
				),
			)
		}
	}

	return result
}

// validPoint bounds evidence coordinates before offset subtraction, limiting malformed captures to the probe's
// supported pixel range.
func validPoint(value point) bool {
	return value.X >= -maxPixel && value.X <= maxPixel && value.Y >= -maxPixel && value.Y <= maxPixel
}

// validSHA256 accepts only canonical lowercase SHA-256 text so provenance hashes have one stable representation.
func validSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}

	decoded, err := hex.DecodeString(value)

	return err == nil && hex.EncodeToString(decoded) == value
}

// oneOf centralizes exact enum matching; callers can rely on case-sensitive acceptance without aliases.
func oneOf(value string, choices ...string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}

	return false
}
