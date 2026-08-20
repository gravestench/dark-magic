package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const (
	probeSchema = "d2legacy.party_xp_probe/v1"
	probeTarget = "lod-1.14d"
)

type capture struct {
	Schema string      `json:"schema"`
	Target string      `json:"target"`
	Source string      `json:"source"`
	Cases  []probeCase `json:"cases"`
}

type probeCase struct {
	ID           string   `json:"id"`
	BaselineID   string   `json:"baseline_id,omitempty"`
	Difficulty   string   `json:"difficulty"`
	Area         string   `json:"area"`
	Monster      string   `json:"monster"`
	MonsterLevel int      `json:"monster_level"`
	GamePlayers  int      `json:"game_players"`
	Party        bool     `json:"party"`
	Members      []member `json:"members"`
}

type member struct {
	ID               string  `json:"id"`
	Level            int     `json:"level"`
	Killer           bool    `json:"killer"`
	SameArea         bool    `json:"same_area"`
	DistanceSubtiles float64 `json:"distance_subtiles"`
	ExperienceBefore uint64  `json:"experience_before"`
	ExperienceAfter  uint64  `json:"experience_after"`
}

// decodeCapture accepts exactly one strict JSON value so typos or appended evidence cannot be silently ignored.
func decodeCapture(input io.Reader) (capture, error) {
	decoder := json.NewDecoder(input)
	decoder.DisallowUnknownFields()

	var captured capture
	if err := decoder.Decode(&captured); err != nil {
		return capture{}, fmt.Errorf("party XP probe: decode capture: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return capture{}, fmt.Errorf("party XP probe: capture must contain one JSON value")
	}

	return captured, nil
}

// validate accumulates independent capture defects in input order so one repair pass can address the entire fixture.
func validate(captured capture) error {
	var result error

	result = joinValidationErrors(result, validateCaptureMetadata(captured))

	casesByID := make(map[string]probeCase, len(captured.Cases))
	for _, observed := range captured.Cases {
		result = joinValidationErrors(result, validateProbeCase(observed, casesByID))
		// Retaining the latest duplicate matches the historical baseline lookup behavior after reporting the duplicate.
		casesByID[observed.ID] = observed
	}

	for _, observed := range captured.Cases {
		result = joinValidationErrors(result, validateBaselineReference(observed, casesByID))
	}

	return result
}

// joinValidationErrors preserves the original deterministic error ordering while helpers expose validation domains.
func joinValidationErrors(result error, validationErrors []error) error {
	for _, validationErr := range validationErrors {
		result = errors.Join(result, validationErr)
	}

	return result
}

// validateCaptureMetadata locks evidence to the one schema, runtime target, and provenance accepted by this probe.
func validateCaptureMetadata(captured capture) []error {
	var validationErrors []error

	if captured.Schema != probeSchema {
		validationErrors = append(
			validationErrors,
			fmt.Errorf("party XP probe: schema %q, want %q", captured.Schema, probeSchema),
		)
	}

	if captured.Target != probeTarget {
		validationErrors = append(
			validationErrors,
			fmt.Errorf("party XP probe: target %q, want %q", captured.Target, probeTarget),
		)
	}

	if captured.Source != "owned-runtime" {
		validationErrors = append(validationErrors, fmt.Errorf("party XP probe: source must be owned-runtime"))
	}

	if len(captured.Cases) == 0 {
		validationErrors = append(validationErrors, fmt.Errorf("party XP probe: at least one case is required"))
	}

	return validationErrors
}

// validateProbeCase checks standalone case facts before cross-case baseline relationships are considered.
func validateProbeCase(observed probeCase, casesByID map[string]probeCase) []error {
	var validationErrors []error

	if observed.ID == "" {
		validationErrors = append(validationErrors, fmt.Errorf("party XP probe: case ID is required"))
	} else if _, exists := casesByID[observed.ID]; exists {
		validationErrors = append(validationErrors, fmt.Errorf("party XP probe: duplicate case %q", observed.ID))
	}

	if observed.Difficulty != "normal" && observed.Difficulty != "nightmare" && observed.Difficulty != "hell" {
		validationErrors = append(
			validationErrors,
			fmt.Errorf("party XP probe: case %q has invalid difficulty", observed.ID),
		)
	}

	if observed.Area == "" || observed.Monster == "" || observed.MonsterLevel < 1 || observed.MonsterLevel > 110 {
		validationErrors = append(
			validationErrors,
			fmt.Errorf("party XP probe: case %q has invalid monster context", observed.ID),
		)
	}

	if observed.GamePlayers < 1 || observed.GamePlayers > 8 || len(observed.Members) != observed.GamePlayers {
		validationErrors = append(
			validationErrors,
			fmt.Errorf("party XP probe: case %q must list every connected player", observed.ID),
		)
	}

	return append(validationErrors, validateMemberRoster(observed)...)
}

// validateMemberRoster enforces unique player facts and the single in-area killer required for meaningful deltas.
func validateMemberRoster(observed probeCase) []error {
	var validationErrors []error

	killerCount := 0
	killerInArea := false
	memberIDs := make(map[string]bool, len(observed.Members))

	for _, player := range observed.Members {
		if hasInvalidMemberData(player, memberIDs[player.ID]) {
			validationErrors = append(
				validationErrors,
				fmt.Errorf("party XP probe: case %q has invalid member data", observed.ID),
			)
		}

		memberIDs[player.ID] = true

		if player.Killer {
			killerCount++
			killerInArea = player.SameArea
		} else if !observed.Party && player.ExperienceAfter != player.ExperienceBefore {
			validationErrors = append(
				validationErrors,
				fmt.Errorf("party XP probe: neutral case %q awarded a non-killer", observed.ID),
			)
		}
	}

	if killerCount != 1 {
		validationErrors = append(
			validationErrors,
			fmt.Errorf("party XP probe: case %q must identify exactly one killer", observed.ID),
		)
	} else if !killerInArea {
		validationErrors = append(
			validationErrors,
			fmt.Errorf("party XP probe: case %q killer must be in the monster area", observed.ID),
		)
	}

	return validationErrors
}

// hasInvalidMemberData centralizes bounds that make XP subtraction and roster comparisons safe and meaningful.
func hasInvalidMemberData(player member, duplicateID bool) bool {
	return player.ID == "" || duplicateID || player.Level < 1 || player.Level > 99 ||
		player.DistanceSubtiles < 0 || player.ExperienceAfter < player.ExperienceBefore
}

// validateBaselineReference ensures a party observation differs from its neutral control only in measured party facts.
func validateBaselineReference(observed probeCase, casesByID map[string]probeCase) []error {
	if observed.BaselineID == "" {
		if observed.Party {
			return []error{fmt.Errorf("party XP probe: party case %q requires a baseline", observed.ID)}
		}

		return nil
	}

	baseline, exists := casesByID[observed.BaselineID]
	if !exists || baseline.Party || !observed.Party {
		return []error{fmt.Errorf("party XP probe: case %q has invalid neutral baseline", observed.ID)}
	}

	var validationErrors []error
	if !matchingBaselineContext(baseline, observed) {
		validationErrors = append(
			validationErrors,
			fmt.Errorf("party XP probe: case %q differs from baseline context", observed.ID),
		)
	}

	if killerDelta(baseline) == 0 {
		validationErrors = append(
			validationErrors,
			fmt.Errorf("party XP probe: baseline %q has no killer XP delta", baseline.ID),
		)
	}

	return validationErrors
}

// matchingBaselineContext protects pool comparisons from mixing different encounters or connected-player rosters.
func matchingBaselineContext(baseline, observed probeCase) bool {
	return baseline.Difficulty == observed.Difficulty &&
		baseline.Area == observed.Area &&
		baseline.Monster == observed.Monster &&
		baseline.MonsterLevel == observed.MonsterLevel &&
		baseline.GamePlayers == observed.GamePlayers &&
		sameRoster(baseline.Members, observed.Members)
}

// sameRoster compares identity, level, and killer role while allowing fixture member order to differ between captures.
func sameRoster(left, right []member) bool {
	type identity struct {
		level  int
		killer bool
	}

	identities := make(map[string]identity, len(left))
	for _, player := range left {
		identities[player.ID] = identity{level: player.Level, killer: player.Killer}
	}

	if len(identities) != len(right) {
		return false
	}

	for _, player := range right {
		if identities[player.ID] != (identity{level: player.Level, killer: player.Killer}) {
			return false
		}
	}

	return true
}

// killerDelta returns the neutral control award; zero remains the sentinel for malformed or non-earning controls.
func killerDelta(observed probeCase) uint64 {
	for _, player := range observed.Members {
		if player.Killer {
			return player.ExperienceAfter - player.ExperienceBefore
		}
	}

	return 0
}
