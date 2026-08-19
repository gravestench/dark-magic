package main

import (
	"io"
	"sort"
)

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
	ShareCandidates      []shareCandidate  `json:"share_candidates,omitempty"`
}

// analyze validates the complete evidence set before deriving reports, preventing partial output from invalid captures.
func analyze(input io.Reader) (report, error) {
	captured, err := decodeCapture(input)
	if err != nil {
		return report{}, err
	}

	if err := validate(captured); err != nil {
		return report{}, err
	}

	casesByID := indexCases(captured.Cases)

	result := report{
		Schema: probeSchema + ".report",
		Target: probeTarget,
	}
	for _, observed := range captured.Cases {
		result.Cases = append(result.Cases, analyzeCase(observed, casesByID))
	}

	return result, nil
}

// indexCases makes baseline lookup explicit while report iteration continues to preserve capture order.
func indexCases(cases []probeCase) map[string]probeCase {
	casesByID := make(map[string]probeCase, len(cases))
	for _, observed := range cases {
		casesByID[observed.ID] = observed
	}

	return casesByID
}

// analyzeCase adds control-relative hypotheses only when levels make a penalty-free comparison possible.
func analyzeCase(observed probeCase, casesByID map[string]probeCase) caseReport {
	normalized := normalize(observed)
	if observed.BaselineID == "" {
		return normalized
	}

	baselineDelta := killerDelta(casesByID[observed.BaselineID])
	normalized.BaselineKillerDelta = baselineDelta
	normalized.PoolRatioNumerator = normalized.TotalDelta
	normalized.PoolRatioDenominator = baselineDelta

	normalized.PenaltyFreeMatrix = penaltyFree(observed)
	if !normalized.PenaltyFreeMatrix {
		return normalized
	}

	normalized.PoolCandidates = poolCandidates(
		baselineDelta,
		normalized.TotalDelta,
		normalized.SameAreaMembers,
	)
	normalized.ShareCandidates = shareCandidates(observed, normalized.TotalDelta)

	return normalized
}

// normalize converts raw XP counters into deterministic per-case deltas without interpreting the retail formula.
func normalize(observed probeCase) caseReport {
	result := caseReport{
		ID:           observed.ID,
		BaselineID:   observed.BaselineID,
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

	// Member maps encode values, while this slice is sorted to make the explicit earning order stable in report JSON.
	sort.Strings(result.EarningMembers)

	return result
}

// penaltyFree identifies cases where level penalties cannot confound pool and share rounding hypotheses.
func penaltyFree(observed probeCase) bool {
	for _, player := range observed.Members {
		if player.SameArea && abs(player.Level-observed.MonsterLevel) > 5 {
			return false
		}
	}

	return true
}

// abs supports bounded validated levels, so negation cannot overflow while measuring a level gap.
func abs(value int) int {
	if value < 0 {
		return -value
	}

	return value
}
