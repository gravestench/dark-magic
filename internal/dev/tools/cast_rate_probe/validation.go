package main

import (
	"errors"
	"fmt"
	"strings"
)

// validate accumulates capture and case failures in encounter order so callers receive all actionable corrections.
func validate(captured capture) error {
	var result error

	result = appendCaptureIdentityErrors(result, captured)
	result = appendRuntimeErrors(result, captured.Runtime)

	if len(captured.Cases) == 0 {
		result = errors.Join(result, fmt.Errorf("cast-rate probe: at least one case is required"))
	}

	seenIDs, seenProfiles := map[string]bool{}, map[string]bool{}

	for _, observed := range captured.Cases {
		result = appendCaseUniquenessErrors(result, observed, seenIDs, seenProfiles)
		result = errors.Join(result, validateCase(observed))
	}

	return result
}

// appendCaptureIdentityErrors preserves the schema, target, and provenance boundary for owned-runtime evidence.
func appendCaptureIdentityErrors(result error, captured capture) error {
	if captured.Schema != probeSchema {
		result = errors.Join(
			result,
			fmt.Errorf("cast-rate probe: schema %q, want %q", captured.Schema, probeSchema),
		)
	}

	if captured.Target != probeTarget {
		result = errors.Join(
			result,
			fmt.Errorf("cast-rate probe: target %q, want %q", captured.Target, probeTarget),
		)
	}

	if captured.Source != "owned-runtime" {
		result = errors.Join(result, fmt.Errorf("cast-rate probe: source must be owned-runtime"))
	}

	return result
}

// appendRuntimeErrors enforces the runtime provenance required to interpret frames as target observations.
func appendRuntimeErrors(result error, observedRuntime runtime) error {
	if observedRuntime.Patch != "1.14d" ||
		observedRuntime.Mode != "expansion" ||
		observedRuntime.Session != "single-player" {
		result = errors.Join(
			result,
			fmt.Errorf("cast-rate probe: runtime must be Expansion 1.14d single-player"),
		)
	}

	if observedRuntime.CharacterOrigin != "probe-created" {
		result = errors.Join(result, fmt.Errorf("cast-rate probe: character must be probe-created"))
	}

	if !validSHA256(observedRuntime.ExecutableSHA256) ||
		!validGenerationID(observedRuntime.RecordGenerationID) {
		result = errors.Join(
			result,
			fmt.Errorf("cast-rate probe: executable and owned-record SHA-256 identities are required"),
		)
	}

	if observedRuntime.Observation != "video-frame-log" ||
		observedRuntime.StatIdentification != "owned-itemstatcost-properties-tbl" ||
		observedRuntime.Locale != "eng" ||
		observedRuntime.GameFramesPerSecond != 25 {
		result = errors.Join(
			result,
			fmt.Errorf(
				"cast-rate probe: requires a 25 Hz visual log and owned ItemStatCost/Properties/TBL identification",
			),
		)
	}

	return result
}

// appendCaseUniquenessErrors records identifiers before validation so every later duplicate is reported consistently.
func appendCaseUniquenessErrors(
	result error,
	observed probeCase,
	seenIDs map[string]bool,
	seenProfiles map[string]bool,
) error {
	if observed.ID == "" || seenIDs[observed.ID] {
		result = errors.Join(result, fmt.Errorf("cast-rate probe: case IDs must be non-empty and unique"))
	}

	seenIDs[observed.ID] = true

	profile := fmt.Sprintf("%d/%s/%d", observed.SkillID, observed.WeaponClass, observed.RawFasterCastRate)
	if seenProfiles[profile] {
		result = errors.Join(result, fmt.Errorf("cast-rate probe: duplicate profile %s", profile))
	}

	seenProfiles[profile] = true

	return result
}

// validateCase checks one observation's skill, equipment, provenance, and frame boundaries without short-circuiting.
func validateCase(observed probeCase) error {
	var result error

	wantRecord, wantMode, wantSequence, knownSkill := expectedSkill(observed.SkillID)
	result = appendSkillContextErrors(result, observed, wantRecord, wantMode, knownSkill)
	result = appendSequenceErrors(result, observed, wantSequence)
	result = appendCastProfileErrors(result, observed)
	result = appendModifierIdentityErrors(result, observed)
	result = appendModifierSourceErrors(result, observed)
	result = appendFrameBoundaryErrors(result, observed)

	return result
}

// appendSkillContextErrors ties each supported skill to its owned record and animation evidence.
func appendSkillContextErrors(
	result error,
	observed probeCase,
	wantRecord string,
	wantMode string,
	knownSkill bool,
) error {
	if !knownSkill ||
		observed.SkillRecord != wantRecord ||
		observed.CharacterClass != "sor" ||
		observed.AnimationMode != wantMode ||
		observed.SequenceTransition != "SC" {
		result = errors.Join(
			result,
			fmt.Errorf("cast-rate probe: case %q has invalid owned skill context", observed.ID),
		)
	}

	return result
}

// appendSequenceErrors distinguishes direct SC casts from SQ animations whose sequence number is part of the evidence.
func appendSequenceErrors(result error, observed probeCase, wantSequence *int) error {
	if wantSequence == nil && observed.SequenceNumber != nil {
		return errors.Join(
			result,
			fmt.Errorf("cast-rate probe: case %q assigns a sequence to an SC cast", observed.ID),
		)
	}

	if wantSequence != nil && (observed.SequenceNumber == nil || *observed.SequenceNumber != *wantSequence) {
		return errors.Join(
			result,
			fmt.Errorf("cast-rate probe: case %q must use sequence %d", observed.ID, *wantSequence),
		)
	}

	return result
}

// appendCastProfileErrors limits weapon classes and raw rates to the deliberately measured probe space.
func appendCastProfileErrors(result error, observed probeCase) error {
	validProfile := oneOf(observed.WeaponClass, "HTH", "1HS", "STF") &&
		observed.RawFasterCastRate >= 0 && observed.RawFasterCastRate <= 200
	if !validProfile {
		result = errors.Join(
			result,
			fmt.Errorf("cast-rate probe: case %q has an invalid weapon/FCR profile", observed.ID),
		)
	}

	return result
}

// appendModifierIdentityErrors ensures localized evidence retains the owned stat key and its English rendering.
func appendModifierIdentityErrors(result error, observed probeCase) error {
	if observed.ModifierKey != "ModStr4v" || observed.ModifierText != "Faster Cast Rate" {
		result = errors.Join(
			result,
			fmt.Errorf("cast-rate probe: case %q does not preserve the owned ModStr4v text", observed.ID),
		)
	}

	return result
}

// appendModifierSourceErrors validates every owned item source and requires the sources to explain the reported rate.
func appendModifierSourceErrors(result error, observed probeCase) error {
	sourceTotal := 0

	for index, source := range observed.ModifierSources {
		if source.ItemRecord == "" ||
			!oneOf(source.PropertyCode, "cast1", "cast2", "cast3") ||
			source.Value <= 0 {
			result = errors.Join(
				result,
				fmt.Errorf("cast-rate probe: case %q modifier source %d is invalid", observed.ID, index),
			)
		}

		sourceTotal += source.Value
	}

	if sourceTotal != observed.RawFasterCastRate {
		result = errors.Join(
			result,
			fmt.Errorf(
				"cast-rate probe: case %q modifier sources total %d, want %d",
				observed.ID,
				sourceTotal,
				observed.RawFasterCastRate,
			),
		)
	}

	return result
}

// appendFrameBoundaryErrors requires a bounded start-effect-neutral progression before deriving elapsed frames.
func appendFrameBoundaryErrors(result error, observed probeCase) error {
	validBoundaries := observed.StartFrame >= 0 &&
		observed.EffectFrame > observed.StartFrame &&
		observed.NeutralFrame > observed.EffectFrame &&
		observed.NeutralFrame <= maxFrame
	if !validBoundaries {
		result = errors.Join(
			result,
			fmt.Errorf("cast-rate probe: case %q has invalid visual action boundaries", observed.ID),
		)
	}

	return result
}

// expectedSkill returns the owned record and animation identity; unknown IDs remain invalid rather than inferred.
func expectedSkill(id int) (string, string, *int, bool) {
	switch id {
	case 36:
		return "Fire Bolt", "SC", nil, true
	case 49:
		sequence := 12

		return "Lightning", "SQ", &sequence, true
	default:
		return "", "", nil, false
	}
}

// validSHA256 accepts only lowercase hexadecimal identities so equivalent captures have one canonical spelling.
func validSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}

	for _, character := range value {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return false
		}
	}

	return true
}

// validGenerationID verifies the namespaced record identity while reusing the executable hash's canonical rules.
func validGenerationID(value string) bool {
	return strings.HasPrefix(value, "sha256:") && validSHA256(strings.TrimPrefix(value, "sha256:"))
}

// oneOf keeps small closed-set checks readable without changing their case-sensitive matching semantics.
func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}

	return false
}
