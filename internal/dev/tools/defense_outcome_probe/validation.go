package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

// validateCapture checks provenance, every case, and control relationships before normalization can trust the data.
func validateCapture(captured capture) error {
	var result error

	// Accumulate failures in stable input order so one run describes every repair the capture requires.
	if captured.Schema != probeSchema {
		result = errors.Join(
			result,
			fmt.Errorf("defense outcome probe: schema %q, want %q", captured.Schema, probeSchema),
		)
	}

	if captured.Target != probeTarget {
		result = errors.Join(
			result,
			fmt.Errorf("defense outcome probe: target %q, want %q", captured.Target, probeTarget),
		)
	}

	if captured.Source != "owned-runtime" {
		result = errors.Join(result, fmt.Errorf("defense outcome probe: source must be owned-runtime"))
	}

	if !isTargetRuntime(captured.Runtime) {
		result = errors.Join(
			result,
			fmt.Errorf("defense outcome probe: runtime must be Expansion 1.14d single-player"),
		)
	}

	if !validSHA256(captured.Runtime.ExecutableSHA256) {
		result = errors.Join(result, fmt.Errorf("defense outcome probe: executable SHA-256 is required"))
	}

	if !isOneOf(captured.Runtime.Observation, "video-frame-log", "manual-frame-log") {
		result = errors.Join(result, fmt.Errorf("defense outcome probe: unsupported observation method"))
	}

	if len(captured.Cases) == 0 {
		result = errors.Join(result, fmt.Errorf("defense outcome probe: at least one case is required"))
	}

	casesByID := make(map[string]probeCase, len(captured.Cases))
	for _, observed := range captured.Cases {
		if observed.ID == "" {
			result = errors.Join(result, fmt.Errorf("defense outcome probe: case ID is required"))
		} else if _, exists := casesByID[observed.ID]; exists {
			result = errors.Join(result, fmt.Errorf("defense outcome probe: duplicate case %q", observed.ID))
		}

		casesByID[observed.ID] = observed
		result = errors.Join(result, validateCase(observed))
	}

	for _, observed := range captured.Cases {
		if err := validateControlReference(observed, casesByID); err != nil {
			result = errors.Join(result, err)
		}
	}

	return result
}

// isTargetRuntime keeps the accepted patch, mode, and session coupled as one provenance boundary.
func isTargetRuntime(observed runtime) bool {
	return observed.Patch == "1.14d" &&
		observed.Mode == "expansion" &&
		observed.Session == "single-player"
}

// validateControlReference rejects chained controls and comparisons whose combat contexts are not interchangeable.
func validateControlReference(observed probeCase, casesByID map[string]probeCase) error {
	if observed.Mechanism == "control" {
		if observed.ControlID != "" {
			return fmt.Errorf(
				"defense outcome probe: control %q references another control",
				observed.ID,
			)
		}

		return nil
	}

	control, exists := casesByID[observed.ControlID]
	if !exists || control.Mechanism != "control" {
		return fmt.Errorf("defense outcome probe: case %q requires a control", observed.ID)
	}

	if !sameTrialContext(control, observed) {
		return fmt.Errorf("defense outcome probe: case %q differs from control context", observed.ID)
	}

	return nil
}

// validateCase checks scenario facts before any trial is allowed to contribute to rates or damage totals.
func validateCase(observed probeCase) error {
	var result error

	if !isOneOf(observed.Mechanism, "control", "shield_block", "passive_avoidance") {
		result = errors.Join(
			result,
			fmt.Errorf("defense outcome probe: case %q has invalid mechanism", observed.ID),
		)
	}

	if !validAttackContext(observed) {
		result = errors.Join(
			result,
			fmt.Errorf("defense outcome probe: case %q has invalid attack context", observed.ID),
		)
	}

	if !validParticipantFacts(observed) {
		result = errors.Join(
			result,
			fmt.Errorf("defense outcome probe: case %q has invalid participant facts", observed.ID),
		)
	}

	if observed.Mechanism == "control" {
		if observed.EffectRecord != "" || observed.DisplayedChancePercent != 0 {
			result = errors.Join(
				result,
				fmt.Errorf("defense outcome probe: control %q has a defense effect", observed.ID),
			)
		}
	} else if !validDefenseEffect(observed) {
		result = errors.Join(
			result,
			fmt.Errorf("defense outcome probe: case %q requires an effect and displayed chance", observed.ID),
		)
	}

	if len(observed.Trials) == 0 {
		result = errors.Join(
			result,
			fmt.Errorf("defense outcome probe: case %q requires trials", observed.ID),
		)
	}

	for index, current := range observed.Trials {
		if !validTrialFacts(current) {
			result = errors.Join(
				result,
				fmt.Errorf("defense outcome probe: case %q trial %d is invalid", observed.ID, index),
			)

			continue
		}

		if !consistentOutcomeFacts(current) {
			result = errors.Join(
				result,
				fmt.Errorf(
					"defense outcome probe: case %q trial %d has inconsistent outcome facts",
					observed.ID,
					index,
				),
			)
		}
	}

	return result
}

// validAttackContext limits categorical facts to values the probe knows how to compare without inference.
func validAttackContext(observed probeCase) bool {
	return isOneOf(observed.Difficulty, "normal", "nightmare", "hell") &&
		isOneOf(observed.AttackKind, "melee", "missile") &&
		isOneOf(observed.AttackerKind, "player", "monster")
}

// validParticipantFacts requires physical combat values and a named defender record before comparison.
func validParticipantFacts(observed probeCase) bool {
	return observed.AttackerLevel >= 1 &&
		observed.AttackRating >= 0 &&
		observed.Defender.Level >= 1 &&
		observed.Defender.Defense >= 0 &&
		isOneOf(observed.Defender.Kind, "player", "monster") &&
		observed.Defender.Record != "" &&
		isOneOf(observed.Defender.State, "standing", "walking", "running", "attacking", "casting")
}

// validDefenseEffect prevents non-control mechanisms from claiming an absent record or impossible displayed chance.
func validDefenseEffect(observed probeCase) bool {
	return observed.EffectRecord != "" &&
		observed.DisplayedChancePercent >= 1 &&
		observed.DisplayedChancePercent <= 100
}

// validTrialFacts checks the independent categorical and health-range constraints before relational checks.
func validTrialFacts(observed trial) bool {
	return isOneOf(observed.Outcome, "miss", "damage", "block", "avoid", "lethal") &&
		isOneOf(observed.Reaction, "none", "gethit", "block", "avoid", "death") &&
		observed.HealthBeforeRaw >= 0 &&
		observed.HealthAfterRaw >= 0 &&
		observed.HealthAfterRaw <= observed.HealthBeforeRaw
}

// consistentOutcomeFacts binds outcome labels to health and animation evidence so rates reflect visible events.
func consistentOutcomeFacts(observed trial) bool {
	damage := observed.HealthBeforeRaw - observed.HealthAfterRaw

	return (observed.Outcome == "damage" || observed.Outcome == "lethal") == (damage > 0) &&
		(observed.Outcome == "lethal") ==
			(observed.HealthBeforeRaw > 0 && observed.HealthAfterRaw == 0) &&
		(observed.Outcome == "lethal") == (observed.Reaction == "death") &&
		(observed.Outcome == "miss") == (observed.Reaction == "none") &&
		(observed.Outcome == "block") == (observed.Reaction == "block") &&
		(observed.Outcome == "avoid") == (observed.Reaction == "avoid")
}

// sameTrialContext requires equal scenario facts and sample sizes before a mechanism can be compared with its control.
func sameTrialContext(left, right probeCase) bool {
	return left.Difficulty == right.Difficulty &&
		left.AttackKind == right.AttackKind &&
		left.AttackerKind == right.AttackerKind &&
		left.AttackerLevel == right.AttackerLevel &&
		left.AttackRating == right.AttackRating &&
		left.Defender == right.Defender &&
		len(left.Trials) == len(right.Trials)
}

// validSHA256 accepts only a full hexadecimal digest, preventing ambiguous executable provenance.
func validSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}

	_, err := hex.DecodeString(value)

	return err == nil
}

// isOneOf centralizes closed-set membership checks so every validator uses exact, case-sensitive values.
func isOneOf(value string, choices ...string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}

	return false
}
