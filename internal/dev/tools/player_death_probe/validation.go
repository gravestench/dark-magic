package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

// validate accumulates independent capture failures so operators can repair all provenance and timeline defects
// at once.
func validate(captured capture) error {
	var result error
	if captured.Schema != probeSchema {
		result = errors.Join(
			result,
			fmt.Errorf("player death probe: schema %q, want %q", captured.Schema, probeSchema),
		)
	}

	if captured.Target != probeTarget {
		result = errors.Join(
			result,
			fmt.Errorf("player death probe: target %q, want %q", captured.Target, probeTarget),
		)
	}

	if captured.Source != "owned-runtime" {
		result = errors.Join(result, fmt.Errorf("player death probe: source must be owned-runtime"))
	}

	if !isSupportedRuntime(captured.Runtime) {
		result = errors.Join(
			result,
			fmt.Errorf("player death probe: runtime must be softcore Expansion 1.14d single-player"),
		)
	}

	if captured.Runtime.CharacterOrigin != "probe-created" {
		result = errors.Join(
			result,
			fmt.Errorf("player death probe: character must be created for the probe, not imported save data"),
		)
	}

	if !validSHA256(captured.Runtime.ExecutableSHA256) {
		result = errors.Join(result, fmt.Errorf("player death probe: executable SHA-256 is required"))
	}

	if !oneOf(captured.Runtime.Observation, "video-frame-log", "manual-frame-log") {
		result = errors.Join(result, fmt.Errorf("player death probe: unsupported observation method"))
	}

	if len(captured.Cases) == 0 {
		result = errors.Join(result, fmt.Errorf("player death probe: at least one case is required"))
	}

	seenCaseIDs := make(map[string]bool)
	for _, observed := range captured.Cases {
		if observed.ID == "" || seenCaseIDs[observed.ID] {
			result = errors.Join(result, fmt.Errorf("player death probe: case IDs must be non-empty and unique"))
		}

		seenCaseIDs[observed.ID] = true

		result = errors.Join(result, validateCase(observed))
	}

	return result
}

// isSupportedRuntime keeps the target tuple atomic because partial matches are not meaningful evidence.
func isSupportedRuntime(observed runtime) bool {
	return observed.Patch == "1.14d" &&
		observed.Mode == "expansion" &&
		observed.Session == "single-player" &&
		observed.CharacterMode == "softcore"
}

// validateCase checks context before timelines because observations are only meaningful for the intended target case.
func validateCase(observed probeCase) error {
	if !hasSupportedCaseContext(observed) {
		return fmt.Errorf("player death probe: case %q has invalid target context", observed.ID)
	}

	if len(observed.Observations) == 0 {
		return fmt.Errorf("player death probe: case %q requires observations", observed.ID)
	}

	byDeath, result := validateObservations(observed)
	result = errors.Join(result, validateDeathTimelines(observed.ID, observed.Scenario, byDeath))
	result = errors.Join(result, validateScenario(observed.Scenario, byDeath))

	return result
}

// hasSupportedCaseContext rejects cases whose gameplay context would make target-specific conclusions ambiguous.
func hasSupportedCaseContext(observed probeCase) bool {
	return oneOf(
		observed.Scenario,
		"single_recovery",
		"single_no_recovery",
		"multiple_corpses",
		"save_exit_dead",
		"save_exit_respawned",
	) && oneOf(observed.Difficulty, "normal", "nightmare", "hell") &&
		oneOf(
			observed.Class,
			"amazon",
			"sorceress",
			"necromancer",
			"paladin",
			"barbarian",
			"druid",
			"assassin",
		) && observed.Level >= 1 && observed.KillerKind == "monster"
}

// validateObservations preserves input order while building per-death phase indexes for later temporal checks.
func validateObservations(observed probeCase) (map[int]map[string]observation, error) {
	byDeath := make(map[int]map[string]observation)
	lastFrame := -1
	stashedGold := observed.Observations[0].StashedGold

	var result error

	for index, current := range observed.Observations {
		result = errors.Join(
			result,
			validateObservation(observed.ID, index, current, lastFrame, stashedGold),
		)
		// Even an invalid frame remains the predecessor so every adjacent ordering defect is reported.
		lastFrame = current.Frame

		result = errors.Join(result, recordObservationPhase(observed.ID, current, byDeath))
	}

	return byDeath, result
}

// validateObservation reports independent field, ownership, and control-state contradictions in stable order.
func validateObservation(caseID string, index int, current observation, lastFrame int, stashedGold int64) error {
	var result error
	if hasInvalidObservationValues(current, lastFrame) {
		result = errors.Join(
			result,
			fmt.Errorf("player death probe: case %q observation %d is invalid", caseID, index),
		)
	}

	if current.StashedGold != stashedGold {
		result = errors.Join(
			result,
			fmt.Errorf("player death probe: case %q changes stash gold between observations", caseID),
		)
	}

	if err := validateItems(current.Equipment); err != nil {
		result = errors.Join(
			result,
			fmt.Errorf("player death probe: case %q observation %d equipment: %w", caseID, index, err),
		)
	}

	if err := validateItems(current.Inventory); err != nil {
		result = errors.Join(
			result,
			fmt.Errorf("player death probe: case %q observation %d inventory: %w", caseID, index, err),
		)
	}

	if contradictsControlState(current) {
		result = errors.Join(
			result,
			fmt.Errorf(
				"player death probe: case %q observation %d contradicts control/health state",
				caseID,
				index,
			),
		)
	}

	return result
}

// hasInvalidObservationValues centralizes scalar bounds and phase vocabulary without hiding reporting order.
func hasInvalidObservationValues(current observation, lastFrame int) bool {
	return current.Frame <= lastFrame ||
		current.Area == "" ||
		current.DeathIndex < 1 ||
		current.Health < 0 ||
		current.MaxHealth < 1 ||
		current.Health > current.MaxHealth ||
		current.Experience < 0 ||
		current.CarriedGold < 0 ||
		current.StashedGold < 0 ||
		current.GroundGold < 0 ||
		current.CorpseCount < 0 ||
		!oneOf(
			current.Phase,
			"before_death",
			"death_started",
			"death_animation_complete",
			"respawn_input",
			"town_control",
			"corpse_recovered",
			"save_exit",
			"rejoined",
		)
}

// contradictsControlState enforces the observed input boundary and the health states implied by each visual phase.
func contradictsControlState(current observation) bool {
	knownControlPhase := oneOf(
		current.Phase,
		"before_death",
		"death_started",
		"death_animation_complete",
		"respawn_input",
		"town_control",
		"corpse_recovered",
	)
	shouldBeControlled := oneOf(current.Phase, "before_death", "town_control", "corpse_recovered")
	deadPhase := oneOf(current.Phase, "death_started", "death_animation_complete", "respawn_input")

	return knownControlPhase && current.Controlled != shouldBeControlled ||
		deadPhase && current.Health != 0 ||
		current.Phase == "town_control" && current.Health == 0
}

// recordObservationPhase retains the latest repeated phase while reporting the duplicate, matching timeline semantics.
func recordObservationPhase(
	caseID string,
	current observation,
	byDeath map[int]map[string]observation,
) error {
	phases := byDeath[current.DeathIndex]
	if phases == nil {
		phases = make(map[string]observation)
		byDeath[current.DeathIndex] = phases
	}

	var result error
	if _, exists := phases[current.Phase]; exists {
		result = errors.Join(
			result,
			fmt.Errorf(
				"player death probe: case %q repeats phase %q for death %d",
				caseID,
				current.Phase,
				current.DeathIndex,
			),
		)
	}

	phases[current.Phase] = current

	return result
}

// validateDeathTimelines checks contiguous death indexes before validating each indexed phase sequence.
func validateDeathTimelines(caseID string, scenario string, byDeath map[int]map[string]observation) error {
	deathCount := len(byDeath)

	var result error

	for deathIndex := 1; deathIndex <= deathCount; deathIndex++ {
		phases, exists := byDeath[deathIndex]
		if !exists {
			result = errors.Join(
				result,
				fmt.Errorf("player death probe: case %q skips death index %d", caseID, deathIndex),
			)

			continue
		}

		result = errors.Join(result, validateDeathTimeline(caseID, deathIndex, scenario, phases))
	}

	return result
}

// validateDeathTimeline preserves the causal phase order required before normalization subtracts frame values.
func validateDeathTimeline(
	caseID string,
	deathIndex int,
	scenario string,
	phases map[string]observation,
) error {
	var result error

	result = errors.Join(
		result,
		validateRequiredPhases(caseID, deathIndex, phases, "before_death", "death_started", "death_animation_complete"),
	)

	before, hasBefore := phases["before_death"]

	started, hasStarted := phases["death_started"]
	if hasBefore && hasStarted && started.Frame <= before.Frame {
		result = errors.Join(
			result,
			fmt.Errorf("player death probe: case %q death %d starts before its baseline", caseID, deathIndex),
		)
	}

	complete, hasComplete := phases["death_animation_complete"]
	if hasStarted && hasComplete && complete.Frame <= started.Frame {
		result = errors.Join(
			result,
			fmt.Errorf("player death probe: case %q death %d completes before it starts", caseID, deathIndex),
		)
	}

	if scenario != "save_exit_dead" {
		result = errors.Join(result, validateRespawnTimeline(caseID, deathIndex, phases))
	}

	result = errors.Join(result, validateSaveExitOrder(caseID, deathIndex, phases))

	return result
}

// validateRequiredPhases reports missing phases in caller-provided order so accumulated errors remain deterministic.
func validateRequiredPhases(
	caseID string,
	deathIndex int,
	phases map[string]observation,
	requiredPhases ...string,
) error {
	var result error

	for _, required := range requiredPhases {
		if _, exists := phases[required]; exists {
			continue
		}

		result = errors.Join(
			result,
			fmt.Errorf("player death probe: case %q death %d lacks %s", caseID, deathIndex, required),
		)
	}

	return result
}

// validateRespawnTimeline requires both respawn boundaries and rejects consequences measured before control returns.
func validateRespawnTimeline(caseID string, deathIndex int, phases map[string]observation) error {
	var result error

	result = errors.Join(
		result,
		validateRequiredPhases(caseID, deathIndex, phases, "respawn_input", "town_control"),
	)

	input, hasInput := phases["respawn_input"]

	complete, hasComplete := phases["death_animation_complete"]
	if hasInput && hasComplete && input.Frame <= complete.Frame {
		result = errors.Join(
			result,
			fmt.Errorf(
				"player death probe: case %q death %d respawn input precedes death completion",
				caseID,
				deathIndex,
			),
		)
	}

	town, hasTown := phases["town_control"]
	if hasInput && hasTown && town.Frame <= input.Frame {
		result = errors.Join(
			result,
			fmt.Errorf(
				"player death probe: case %q death %d regains control before respawn input",
				caseID,
				deathIndex,
			),
		)
	}

	recovered, hasRecovered := phases["corpse_recovered"]
	if hasTown && hasRecovered && recovered.Frame <= town.Frame {
		result = errors.Join(
			result,
			fmt.Errorf(
				"player death probe: case %q death %d recovers a corpse before town control",
				caseID,
				deathIndex,
			),
		)
	}

	return result
}

// validateSaveExitOrder ensures rejoin snapshots cannot precede the save/exit event they claim to observe.
func validateSaveExitOrder(caseID string, deathIndex int, phases map[string]observation) error {
	saveExit, hasSaveExit := phases["save_exit"]

	rejoined, hasRejoined := phases["rejoined"]
	if !hasSaveExit || !hasRejoined || rejoined.Frame > saveExit.Frame {
		return nil
	}

	return fmt.Errorf(
		"player death probe: case %q death %d rejoins before save/exit",
		caseID,
		deathIndex,
	)
}

// validateScenario checks cross-death requirements after every individual timeline has been indexed and checked.
func validateScenario(scenario string, byDeath map[int]map[string]observation) error {
	deathCount := len(byDeath)

	var result error

	if scenario == "single_recovery" && (deathCount != 1 || !hasPhase(byDeath, "corpse_recovered")) {
		result = errors.Join(
			result,
			fmt.Errorf("player death probe: single_recovery requires one recovered death"),
		)
	}

	if scenario == "single_no_recovery" && deathCount != 1 {
		result = errors.Join(
			result,
			fmt.Errorf("player death probe: single_no_recovery requires one death"),
		)
	}

	if scenario == "single_no_recovery" && hasPhase(byDeath, "corpse_recovered") {
		result = errors.Join(
			result,
			fmt.Errorf("player death probe: single_no_recovery cannot include corpse recovery"),
		)
	}

	if scenario == "multiple_corpses" && deathCount < 2 {
		result = errors.Join(
			result,
			fmt.Errorf("player death probe: multiple_corpses requires at least two deaths"),
		)
	}

	if oneOf(scenario, "save_exit_dead", "save_exit_respawned") &&
		(!hasPhase(byDeath, "save_exit") || !hasPhase(byDeath, "rejoined")) {
		result = errors.Join(
			result,
			fmt.Errorf("player death probe: save/exit scenario requires save_exit and rejoined observations"),
		)
	}

	return result
}

// validateItems rejects ambiguous ownership evidence because normalization compares sanitized identities across phases.
func validateItems(items []slotItem) error {
	seenSlots := make(map[string]bool)

	seenIDs := make(map[string]bool)
	for _, item := range items {
		if item.Slot == "" || item.ID == "" || seenSlots[item.Slot] || seenIDs[item.ID] {
			return fmt.Errorf("slot and sanitized item ID must be non-empty and unique")
		}

		seenSlots[item.Slot] = true
		seenIDs[item.ID] = true
	}

	return nil
}

// hasPhase answers cross-death scenario questions without relying on nondeterministic map iteration order.
func hasPhase(byDeath map[int]map[string]observation, phase string) bool {
	for _, phases := range byDeath {
		if _, exists := phases[phase]; exists {
			return true
		}
	}

	return false
}

// validSHA256 accepts only a complete hexadecimal digest so executable provenance has an unambiguous representation.
func validSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}

	_, err := hex.DecodeString(value)

	return err == nil
}

// oneOf keeps closed capture vocabularies explicit at each validation boundary.
func oneOf(value string, choices ...string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}

	return false
}
