package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math"
)

// validate reports every independent capture defect in encounter order, giving operators one complete repair list.
func validate(captured capture) error {
	var result error

	if captured.Schema != probeSchema {
		result = errors.Join(
			result,
			probeErrorf("schema %q, want %q", captured.Schema, probeSchema),
		)
	}

	if captured.Target != probeTarget {
		result = errors.Join(
			result,
			probeErrorf("target %q, want %q", captured.Target, probeTarget),
		)
	}

	if captured.Source != "owned-runtime" {
		result = errors.Join(result, probeErrorf("source must be owned-runtime"))
	}

	if !hasSupportedRuntime(captured.Runtime) {
		result = errors.Join(
			result,
			probeErrorf(
				"runtime must be Expansion 1.14d single-player or "+
					"local-hosted multiplayer",
			),
		)
	}

	if captured.Runtime.CharacterOrigin != "probe-created" {
		result = errors.Join(result, probeErrorf("character must be probe-created"))
	}

	if !validSHA256(captured.Runtime.ExecutableSHA256) {
		result = errors.Join(result, probeErrorf("executable SHA-256 is required"))
	}

	if !hasRequiredObservationEvidence(captured.Runtime) {
		result = errors.Join(
			result,
			probeErrorf(
				"requires an unmodified executable and debugger stat log paired with video",
			),
		)
	}

	if !hasRequiredRecordAnchors(captured.Records) {
		result = errors.Join(
			result,
			probeErrorf("owned Skills/TBL/MonStats record anchors are required"),
		)
	}

	if len(captured.Cases) == 0 {
		result = errors.Join(result, probeErrorf("at least one case is required"))
	}

	seen := make(map[string]bool)
	for _, observed := range captured.Cases {
		if observed.ID == "" || seen[observed.ID] {
			result = errors.Join(
				result,
				probeErrorf("case IDs must be non-empty and unique"),
			)
		}

		seen[observed.ID] = true

		result = errors.Join(result, validateCase(observed))
	}

	return result
}

// hasSupportedRuntime confines evidence to the two owned Expansion 1.14d session modes covered by the probe.
func hasSupportedRuntime(observed runtime) bool {
	return observed.Patch == "1.14d" && observed.Mode == "expansion" &&
		oneOf(observed.Session, "single-player", "local-hosted-multiplayer")
}

// hasRequiredObservationEvidence ensures the stat log and video describe the same unmodified executable behavior.
func hasRequiredObservationEvidence(observed runtime) bool {
	return observed.Observation == "debugger-stat-log-plus-video" && observed.ExecutableUnmodified
}

// hasRequiredRecordAnchors ties the observation to the exact skill, localization, and monster records under study.
func hasRequiredRecordAnchors(observed records) bool {
	return observed.SkillID == 94 && observed.SkillNameKey != "" &&
		observed.LocalizedSkillName != "" && observed.Locale != "" &&
		observed.MonsterID == "firegolem" && observed.DeathDamageEnabled &&
		validSHA256(observed.ExtractedSHA256)
}

// validateCase accumulates context, event, and target contradictions without hiding later evidence defects.
func validateCase(observed probeCase) error {
	var result error

	if !hasValidCaseContext(observed) {
		result = errors.Join(
			result,
			probeErrorf("case %q has invalid target context", observed.ID),
		)
	}

	frames := observed.EventFrames
	if !hasRemovalAndExplosionFrames(frames) {
		result = errors.Join(
			result,
			probeErrorf("case %q lacks removal/explosion frames", observed.ID),
		)
	}

	if observed.Trigger == "replacement" {
		if frames.NewGolemCreated == nil || *frames.NewGolemCreated < 0 {
			result = errors.Join(
				result,
				probeErrorf(
					"replacement case %q lacks the new-golem frame",
					observed.ID,
				),
			)
		}
	} else if frames.NewGolemCreated != nil {
		result = errors.Join(
			result,
			probeErrorf("combat-death case %q creates a replacement", observed.ID),
		)
	}

	if len(observed.Targets) == 0 {
		result = errors.Join(
			result,
			probeErrorf("case %q requires target samples", observed.ID),
		)
	}

	seen := make(map[string]bool)
	for index, sample := range observed.Targets {
		if !isValidTargetSample(sample, seen[sample.ID]) {
			result = errors.Join(
				result,
				probeErrorf("case %q target %d is invalid", observed.ID, index),
			)
		}

		seen[sample.ID] = true

		if distanceMilli(observed.ExplosionCenter, sample.Position) != sample.DistanceMilli {
			result = errors.Join(
				result,
				probeErrorf(
					"case %q target %q distance does not match its coordinates",
					observed.ID,
					sample.ID,
				),
			)
		}

		healthDelta := sample.HealthBeforeRaw - sample.HealthAfterRaw
		if !sample.DamageEvent && recordsDamageEffects(sample, healthDelta) {
			result = errors.Join(
				result,
				probeErrorf(
					"case %q unaffected target %q records damage effects",
					observed.ID,
					sample.ID,
				),
			)
		}

		if sample.DamageEvent && channelTotal(sample.Channels) == 0 {
			result = errors.Join(
				result,
				probeErrorf(
					"case %q affected target %q lacks pre-mitigation channels",
					observed.ID,
					sample.ID,
				),
			)
		}

		if sample.Died != (sample.HealthAfterRaw == 0) {
			result = errors.Join(
				result,
				probeErrorf(
					"case %q target %q contradicts its death state",
					observed.ID,
					sample.ID,
				),
			)
		}
	}

	return result
}

// hasValidCaseContext bounds observations to supported triggers, difficulties, and attainable probe levels.
func hasValidCaseContext(observed probeCase) bool {
	return oneOf(observed.Trigger, "replacement", "combat_death") &&
		oneOf(observed.Difficulty, "normal", "nightmare", "hell") &&
		observed.SkillLevel >= 1 && observed.SkillLevel <= 99 &&
		observed.PlayerLevel >= 30 && observed.PlayerLevel <= 99
}

// hasRemovalAndExplosionFrames requires both observable death phases and rejects impossible negative timestamps.
func hasRemovalAndExplosionFrames(frames eventFrames) bool {
	return frames.OldGolemRemoved != nil && frames.ExplosionStarted != nil &&
		*frames.OldGolemRemoved >= 0 && *frames.ExplosionStarted >= 0
}

// isValidTargetSample checks sample identity and measurement bounds before cross-field contradictions are reported.
func isValidTargetSample(sample targetSample, duplicateID bool) bool {
	// Hostility is implied by the target class; accepting a conflicting flag would corrupt radius grouping.
	hostilityMatchesKind := sample.HostileToOwner ==
		oneOf(sample.Kind, "hostile_player", "hostile_monster")

	return sample.ID != "" && !duplicateID &&
		oneOf(
			sample.Kind,
			"owner",
			"allied_player",
			"hostile_player",
			"allied_minion",
			"neutral_monster",
			"hostile_monster",
		) &&
		hostilityMatchesKind &&
		sample.DistanceMilli >= 0 && sample.HealthBeforeRaw >= 1 &&
		sample.HealthAfterRaw >= 0 && sample.HealthAfterRaw <= sample.HealthBeforeRaw &&
		sample.FireResistance >= -100 && sample.FireResistance <= 255 &&
		sample.PhysicalResistance >= -100 && sample.PhysicalResistance <= 255 &&
		sample.NoAbsorbOrFlatDR && validChannels(sample.Channels)
}

// recordsDamageEffects detects evidence that contradicts a target marked as unaffected by the explosion.
func recordsDamageEffects(sample targetSample, healthDelta int64) bool {
	return healthDelta != 0 || sample.HitReaction || sample.Died || channelTotal(sample.Channels) != 0
}

// distanceMilli derives a neutral geometric measurement while leaving runtime range policy out of the capture format.
func distanceMilli(center, target point) int {
	dx := float64(center.XSubtiles - target.XSubtiles)
	dy := float64(center.YSubtiles - target.YSubtiles)

	// Keeping coordinates permits later policies to compare alternative 1.14d predicates against the same evidence.
	return int(math.Round(math.Hypot(dx, dy) * 1000))
}

// validChannels rejects impossible negative pre-mitigation damage without imposing a gameplay-specific damage mix.
func validChannels(value channels) bool {
	return value.Physical >= 0 && value.Fire >= 0 && value.Cold >= 0 &&
		value.Lightning >= 0 && value.Poison >= 0 && value.Magic >= 0
}

// channelTotal preserves raw fixed-point channel arithmetic when checking whether any damage was recorded.
func channelTotal(value channels) int64 {
	return value.Physical + value.Fire + value.Cold + value.Lightning + value.Poison + value.Magic
}

// validSHA256 accepts only full hexadecimal digests so provenance identifiers cannot be mistaken for partial hashes.
func validSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}

	_, err := hex.DecodeString(value)

	return err == nil
}

// oneOf centralizes closed string sets so validation call sites show the complete accepted vocabulary.
func oneOf(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if value == candidate {
			return true
		}
	}

	return false
}
