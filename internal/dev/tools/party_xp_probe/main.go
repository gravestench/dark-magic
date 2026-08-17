// Command party_xp_probe validates and normalizes sanitized observations from
// an owned expansion 1.14d runtime. It does not emulate party XP: its output is
// evidence used to choose exact distance and integer-ordering vectors later.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
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

type report struct {
	Schema string       `json:"schema"`
	Target string       `json:"target"`
	Cases  []caseReport `json:"cases"`
}

type caseReport struct {
	ID                   string            `json:"id"`
	BaselineID           string            `json:"baseline_id,omitempty"`
	TotalDelta           uint64            `json:"total_delta"`
	MemberDeltas         map[string]uint64 `json:"member_deltas"`
	EarningMembers       []string          `json:"earning_members"`
	SameAreaMembers      int               `json:"same_area_members"`
	BaselineKillerDelta  uint64            `json:"baseline_killer_delta,omitempty"`
	PoolRatioNumerator   uint64            `json:"pool_ratio_numerator,omitempty"`
	PoolRatioDenominator uint64            `json:"pool_ratio_denominator,omitempty"`
	PenaltyFreeMatrix    bool              `json:"penalty_free_matrix"`
	PoolCandidates       []poolCandidate   `json:"pool_candidates,omitempty"`
}

type poolCandidate struct {
	Name        string `json:"name"`
	Factor      int    `json:"factor_percent"`
	Floor       uint64 `json:"floor"`
	Nearest     uint64 `json:"nearest"`
	Ceiling     uint64 `json:"ceiling"`
	ObservedFit string `json:"observed_fit"`
}

func main() {
	input := flag.String("input", "", "sanitized owned-runtime party-XP probe JSON")
	flag.Parse()
	if *input == "" {
		fmt.Fprintln(os.Stderr, "usage: party_xp_probe -input <capture.json>")
		os.Exit(2)
	}
	file, err := os.Open(*input)
	if err != nil {
		fatal(err)
	}
	defer file.Close()
	result, err := analyze(file)
	if err != nil {
		fatal(err)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fatal(err)
	}
}

func analyze(input io.Reader) (report, error) {
	decoder := json.NewDecoder(input)
	decoder.DisallowUnknownFields()
	var captured capture
	if err := decoder.Decode(&captured); err != nil {
		return report{}, fmt.Errorf("party XP probe: decode capture: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return report{}, fmt.Errorf("party XP probe: capture must contain one JSON value")
	}
	if err := validate(captured); err != nil {
		return report{}, err
	}
	byID := make(map[string]probeCase, len(captured.Cases))
	for _, candidate := range captured.Cases {
		byID[candidate.ID] = candidate
	}
	result := report{Schema: probeSchema + ".report", Target: probeTarget}
	for _, observed := range captured.Cases {
		normalized := normalize(observed)
		if observed.BaselineID != "" {
			baseline := byID[observed.BaselineID]
			normalized.BaselineKillerDelta = killerDelta(baseline)
			normalized.PoolRatioNumerator = normalized.TotalDelta
			normalized.PoolRatioDenominator = normalized.BaselineKillerDelta
			normalized.PenaltyFreeMatrix = penaltyFree(observed)
			if normalized.PenaltyFreeMatrix {
				normalized.PoolCandidates = poolCandidates(
					normalized.BaselineKillerDelta,
					normalized.TotalDelta,
					normalized.SameAreaMembers,
				)
			}
		}
		result.Cases = append(result.Cases, normalized)
	}
	return result, nil
}

func validate(captured capture) error {
	var result error
	if captured.Schema != probeSchema {
		result = errors.Join(result, fmt.Errorf("party XP probe: schema %q, want %q", captured.Schema, probeSchema))
	}
	if captured.Target != probeTarget {
		result = errors.Join(result, fmt.Errorf("party XP probe: target %q, want %q", captured.Target, probeTarget))
	}
	if captured.Source != "owned-runtime" {
		result = errors.Join(result, fmt.Errorf("party XP probe: source must be owned-runtime"))
	}
	if len(captured.Cases) == 0 {
		result = errors.Join(result, fmt.Errorf("party XP probe: at least one case is required"))
	}
	byID := make(map[string]probeCase, len(captured.Cases))
	for _, observed := range captured.Cases {
		if observed.ID == "" {
			result = errors.Join(result, fmt.Errorf("party XP probe: case ID is required"))
		} else if _, exists := byID[observed.ID]; exists {
			result = errors.Join(result, fmt.Errorf("party XP probe: duplicate case %q", observed.ID))
		}
		byID[observed.ID] = observed
		if observed.Difficulty != "normal" && observed.Difficulty != "nightmare" && observed.Difficulty != "hell" {
			result = errors.Join(result, fmt.Errorf("party XP probe: case %q has invalid difficulty", observed.ID))
		}
		if observed.Area == "" || observed.Monster == "" || observed.MonsterLevel < 1 || observed.MonsterLevel > 110 {
			result = errors.Join(result, fmt.Errorf("party XP probe: case %q has invalid monster context", observed.ID))
		}
		if observed.GamePlayers < 1 || observed.GamePlayers > 8 || len(observed.Members) != observed.GamePlayers {
			result = errors.Join(result, fmt.Errorf("party XP probe: case %q must list every connected player", observed.ID))
		}
		killerCount := 0
		killerInArea := false
		memberIDs := map[string]bool{}
		for _, player := range observed.Members {
			if player.ID == "" || memberIDs[player.ID] || player.Level < 1 || player.Level > 99 ||
				player.DistanceSubtiles < 0 || player.ExperienceAfter < player.ExperienceBefore {
				result = errors.Join(result, fmt.Errorf("party XP probe: case %q has invalid member data", observed.ID))
			}
			memberIDs[player.ID] = true
			if player.Killer {
				killerCount++
				killerInArea = player.SameArea
			} else if !observed.Party && player.ExperienceAfter != player.ExperienceBefore {
				result = errors.Join(result, fmt.Errorf("party XP probe: neutral case %q awarded a non-killer", observed.ID))
			}
		}
		if killerCount != 1 {
			result = errors.Join(result, fmt.Errorf("party XP probe: case %q must identify exactly one killer", observed.ID))
		} else if !killerInArea {
			result = errors.Join(result, fmt.Errorf("party XP probe: case %q killer must be in the monster area", observed.ID))
		}
	}
	for _, observed := range captured.Cases {
		if observed.BaselineID == "" {
			if observed.Party {
				result = errors.Join(result, fmt.Errorf("party XP probe: party case %q requires a baseline", observed.ID))
			}
			continue
		}
		baseline, exists := byID[observed.BaselineID]
		if !exists || baseline.Party || !observed.Party {
			result = errors.Join(result, fmt.Errorf("party XP probe: case %q has invalid neutral baseline", observed.ID))
			continue
		}
		if baseline.Difficulty != observed.Difficulty || baseline.Area != observed.Area ||
			baseline.Monster != observed.Monster || baseline.MonsterLevel != observed.MonsterLevel ||
			baseline.GamePlayers != observed.GamePlayers || !sameRoster(baseline.Members, observed.Members) {
			result = errors.Join(result, fmt.Errorf("party XP probe: case %q differs from baseline context", observed.ID))
		}
		if killerDelta(baseline) == 0 {
			result = errors.Join(result, fmt.Errorf("party XP probe: baseline %q has no killer XP delta", baseline.ID))
		}
	}
	return result
}

func normalize(observed probeCase) caseReport {
	result := caseReport{
		ID: observed.ID, BaselineID: observed.BaselineID,
		MemberDeltas: make(map[string]uint64, len(observed.Members)),
	}
	for _, player := range observed.Members {
		delta := player.ExperienceAfter - player.ExperienceBefore
		result.MemberDeltas[player.ID] = delta
		result.TotalDelta += delta
		if delta > 0 {
			result.EarningMembers = append(result.EarningMembers, player.ID)
		}
		if player.SameArea {
			result.SameAreaMembers++
		}
	}
	sort.Strings(result.EarningMembers)
	return result
}

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

func killerDelta(observed probeCase) uint64 {
	for _, player := range observed.Members {
		if player.Killer {
			return player.ExperienceAfter - player.ExperienceBefore
		}
	}
	return 0
}

func penaltyFree(observed probeCase) bool {
	for _, player := range observed.Members {
		if player.SameArea && abs(player.Level-observed.MonsterLevel) > 5 {
			return false
		}
	}
	return true
}

func poolCandidates(baseline, observed uint64, sameAreaMembers int) []poolCandidate {
	singleBonusFactor := 100
	if sameAreaMembers > 1 {
		singleBonusFactor = 135
	}
	factors := []struct {
		name   string
		factor int
	}{
		{name: "single_35_percent_bonus", factor: singleBonusFactor},
		{name: "35_percent_per_additional_member", factor: 100 + 35*(sameAreaMembers-1)},
	}
	result := make([]poolCandidate, 0, len(factors))
	for _, candidate := range factors {
		numerator := baseline * uint64(candidate.factor)
		floor := numerator / 100
		nearest := (numerator + 50) / 100
		ceiling := (numerator + 99) / 100
		fit := "none"
		switch observed {
		case floor:
			fit = "floor"
		case nearest:
			fit = "nearest"
		case ceiling:
			fit = "ceiling"
		}
		result = append(result, poolCandidate{
			Name: candidate.name, Factor: candidate.factor,
			Floor: floor, Nearest: nearest, Ceiling: ceiling, ObservedFit: fit,
		})
	}
	return result
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
