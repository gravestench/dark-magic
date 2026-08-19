package main

import (
	"errors"
	"fmt"
	"math"
)

// validate accumulates independent capture failures in input order so one run
// reports the complete repair surface without changing which later case owns a duplicate ID.
func validate(captured capture) error {
	var result error

	if captured.Schema != probeSchema {
		result = errors.Join(
			result,
			fmt.Errorf("knockback probe: schema %q, want %q", captured.Schema, probeSchema),
		)
	}

	if captured.Target != probeTarget {
		result = errors.Join(
			result,
			fmt.Errorf("knockback probe: target %q, want %q", captured.Target, probeTarget),
		)
	}

	if captured.Source != "owned-runtime" {
		result = errors.Join(result, fmt.Errorf("knockback probe: source must be owned-runtime"))
	}

	if len(captured.Cases) == 0 {
		result = errors.Join(result, fmt.Errorf("knockback probe: at least one case is required"))
	}

	byID := make(map[string]probeCase, len(captured.Cases))
	for _, observed := range captured.Cases {
		if observed.ID == "" {
			result = errors.Join(result, fmt.Errorf("knockback probe: case ID is required"))
		} else if _, exists := byID[observed.ID]; exists {
			result = errors.Join(result, fmt.Errorf("knockback probe: duplicate case %q", observed.ID))
		}

		// Keep indexing invalid and duplicate cases so all case-local failures are
		// reported and the last duplicate remains the relationship-check context.
		byID[observed.ID] = observed
		result = errors.Join(result, validateCase(observed))
	}

	// Resolve controls only after indexing every case, which intentionally permits
	// a case to reference a control that appears later in the capture.
	for _, observed := range captured.Cases {
		if observed.Mechanism == "control" {
			if observed.ControlID != "" {
				result = errors.Join(
					result,
					fmt.Errorf("knockback probe: control case %q references another control", observed.ID),
				)
			}

			continue
		}

		control, exists := byID[observed.ControlID]
		if !exists || control.Mechanism != "control" {
			result = errors.Join(
				result,
				fmt.Errorf("knockback probe: case %q requires a control case", observed.ID),
			)

			continue
		}

		if !sameContext(control, observed) {
			result = errors.Join(
				result,
				fmt.Errorf("knockback probe: case %q differs from control context", observed.ID),
			)
		}
	}

	return result
}

// validateCase checks the local case and trial contracts while retaining every
// independent error; relationship checks remain in validate because they need the full capture.
func validateCase(observed probeCase) error {
	var result error

	if !oneOf(observed.Mechanism, "control", "item_knockback", "missile_knockback") {
		result = errors.Join(result, fmt.Errorf("knockback probe: case %q has invalid mechanism", observed.ID))
	}

	if !oneOf(observed.Difficulty, "normal", "nightmare", "hell") {
		result = errors.Join(result, fmt.Errorf("knockback probe: case %q has invalid difficulty", observed.ID))
	}

	if !oneOf(observed.AttackerKind, "player", "monster", "hireling", "summon") {
		result = errors.Join(result, fmt.Errorf("knockback probe: case %q has invalid attacker kind", observed.ID))
	}

	if !validTargetIdentity(observed.Target) {
		result = errors.Join(result, fmt.Errorf("knockback probe: case %q has invalid target identity", observed.ID))
	}

	if !oneOf(observed.Target.SizeClass, "none", "small", "normal", "large") {
		result = errors.Join(result, fmt.Errorf("knockback probe: case %q has invalid target size", observed.ID))
	}

	if !isNonnegativeFinite(observed.OpenDistanceSubtiles) {
		result = errors.Join(result, fmt.Errorf("knockback probe: case %q has invalid open distance", observed.ID))
	}

	if observed.Mechanism == "missile_knockback" {
		if observed.MissileKnockbackValue < 1 || observed.MissileKnockbackValue > 255 {
			result = errors.Join(
				result,
				fmt.Errorf("knockback probe: case %q has invalid missile KnockBack byte", observed.ID),
			)
		}
	} else if observed.MissileKnockbackValue != 0 {
		result = errors.Join(
			result,
			fmt.Errorf("knockback probe: case %q has a missile byte outside a missile case", observed.ID),
		)
	}

	if len(observed.Trials) == 0 {
		result = errors.Join(result, fmt.Errorf("knockback probe: case %q requires trials", observed.ID))
	}

	for index, current := range observed.Trials {
		if !validTrialObservation(current) {
			result = errors.Join(
				result,
				fmt.Errorf("knockback probe: case %q trial %d is invalid", observed.ID, index),
			)
		}

		if reactsWithoutHit(current) {
			result = errors.Join(
				result,
				fmt.Errorf("knockback probe: case %q trial %d reacts without a hit", observed.ID, index),
			)
		}
	}

	return result
}

// validTargetIdentity accepts only runtime actor categories with an explicit
// data-record identity, preventing unlike or anonymous targets from being compared.
func validTargetIdentity(observed target) bool {
	return oneOf(observed.Kind, "player", "monster", "hireling", "summon", "npc", "corpse") &&
		observed.Record != ""
}

// isNonnegativeFinite rejects values that cannot represent a physical distance
// or be safely sorted and compared during normalization.
func isNonnegativeFinite(value float64) bool {
	return value >= 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

// validTrialObservation keeps reaction names and displacement values within the
// finite vocabulary used by report aggregation.
func validTrialObservation(observed trial) bool {
	return isNonnegativeFinite(observed.DisplacementSubtiles) &&
		oneOf(observed.Reaction, "none", "gethit", "knockback", "death", "dead")
}

// reactsWithoutHit identifies internally inconsistent trials whose combat or
// presentation outcomes could not have followed from the recorded miss.
func reactsWithoutHit(observed trial) bool {
	return !observed.Hit &&
		(observed.CombatBlocked ||
			observed.Lethal ||
			observed.Uninterruptible ||
			observed.DisplacementSubtiles != 0 ||
			observed.Reaction != "none")
}

// sameContext requires every experimental condition and trial count to match;
// otherwise a control cannot isolate the knockback mechanism under test.
func sameContext(control, observed probeCase) bool {
	return control.Difficulty == observed.Difficulty &&
		control.AttackerKind == observed.AttackerKind &&
		control.Target == observed.Target &&
		control.OpenDistanceSubtiles == observed.OpenDistanceSubtiles &&
		len(control.Trials) == len(observed.Trials)
}

// oneOf centralizes closed-vocabulary checks so every validation site uses exact
// string equality and preserves the capture format's case sensitivity.
func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}

	return false
}
