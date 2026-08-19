package main

import "math/big"

type poolCandidate struct {
	Name        string `json:"name"`
	Factor      int    `json:"factor_percent"`
	Floor       uint64 `json:"floor"`
	Nearest     uint64 `json:"nearest"`
	Ceiling     uint64 `json:"ceiling"`
	ObservedFit string `json:"observed_fit"`
}

type shareCandidate struct {
	Name        string                     `json:"name"`
	Members     map[string]sharePrediction `json:"members"`
	ObservedFit string                     `json:"observed_fit"`
}

type sharePrediction struct {
	Observed uint64 `json:"observed"`
	Floor    uint64 `json:"floor"`
	Nearest  uint64 `json:"nearest"`
	Ceiling  uint64 `json:"ceiling"`
}

type poolHypothesis struct {
	name   string
	factor int
}

type shareHypothesis struct {
	name   string
	weight func(member) *big.Rat
}

type weightedMember struct {
	player member
	weight *big.Rat
}

// poolCandidates evaluates both documented bonus interpretations without choosing either as retail policy.
func poolCandidates(baseline, observed uint64, sameAreaMembers int) []poolCandidate {
	hypotheses := poolHypotheses(sameAreaMembers)
	result := make([]poolCandidate, 0, len(hypotheses))

	for _, hypothesis := range hypotheses {
		floor, nearest, ceiling := roundedPercentage(baseline, hypothesis.factor)
		result = append(result, poolCandidate{
			Name:        hypothesis.name,
			Factor:      hypothesis.factor,
			Floor:       floor,
			Nearest:     nearest,
			Ceiling:     ceiling,
			ObservedFit: poolRoundingFit(observed, floor, nearest, ceiling),
		})
	}

	return result
}

// poolHypotheses keeps candidate ordering stable because downstream evidence reviews compare reports structurally.
func poolHypotheses(sameAreaMembers int) []poolHypothesis {
	singleBonusFactor := 100
	if sameAreaMembers > 1 {
		singleBonusFactor = 135
	}

	return []poolHypothesis{
		{name: "single_35_percent_bonus", factor: singleBonusFactor},
		{name: "35_percent_per_additional_member", factor: 100 + 35*(sameAreaMembers-1)},
	}
}

// roundedPercentage exposes each integer-rounding possibility while retaining the original uint64 arithmetic behavior.
func roundedPercentage(value uint64, factor int) (uint64, uint64, uint64) {
	numerator := value * uint64(factor)

	return numerator / 100, (numerator + 50) / 100, (numerator + 99) / 100
}

// poolRoundingFit uses floor-first precedence when multiple rounding modes produce the same integer.
func poolRoundingFit(observed, floor, nearest, ceiling uint64) string {
	switch observed {
	case floor:
		return "floor"
	case nearest:
		return "nearest"
	case ceiling:
		return "ceiling"
	default:
		return "none"
	}
}

// shareCandidates evaluates level-weighting alternatives only for members who received XP in the observation.
func shareCandidates(observed probeCase, pool uint64) []shareCandidate {
	eligible := earningMembers(observed.Members)
	hypotheses := []shareHypothesis{
		{name: "direct_character_level", weight: directCharacterLevelWeight},
		{name: "inverse_character_level", weight: inverseCharacterLevelWeight},
		{name: "equal_shares", weight: equalShareWeight},
	}
	result := make([]shareCandidate, 0, len(hypotheses))

	for _, hypothesis := range hypotheses {
		result = append(result, evaluateShareHypothesis(hypothesis, eligible, pool))
	}

	return result
}

// earningMembers mirrors observed allocation rather than guessing eligibility from distance or party membership.
func earningMembers(members []member) []member {
	eligible := make([]member, 0, len(members))
	for _, player := range members {
		if player.ExperienceAfter > player.ExperienceBefore {
			eligible = append(eligible, player)
		}
	}

	return eligible
}

// directCharacterLevelWeight represents the hypothesis that higher-level members receive proportionally larger shares.
func directCharacterLevelWeight(player member) *big.Rat {
	return new(big.Rat).SetInt64(int64(player.Level))
}

// inverseCharacterLevelWeight represents the competing hypothesis that lower-level members receive larger shares.
func inverseCharacterLevelWeight(player member) *big.Rat {
	return new(big.Rat).SetFrac64(1, int64(player.Level))
}

// equalShareWeight represents a level-independent split and supplies one unit of weight for every earning member.
func equalShareWeight(member) *big.Rat {
	return new(big.Rat).SetInt64(1)
}

// evaluateShareHypothesis derives member predictions together so a reported fit always uses one common denominator.
func evaluateShareHypothesis(hypothesis shareHypothesis, eligible []member, pool uint64) shareCandidate {
	weighted, totalWeight := weightMembers(hypothesis, eligible)
	candidate := shareCandidate{
		Name:    hypothesis.name,
		Members: make(map[string]sharePrediction, len(weighted)),
	}

	for _, item := range weighted {
		exact := exactWeightedShare(pool, item.weight, totalWeight)
		floor, nearest, ceiling := roundRat(exact)
		candidate.Members[item.player.ID] = sharePrediction{
			Observed: item.player.ExperienceAfter - item.player.ExperienceBefore,
			Floor:    floor,
			Nearest:  nearest,
			Ceiling:  ceiling,
		}
	}

	candidate.ObservedFit = shareFit(candidate.Members)

	return candidate
}

// weightMembers retains exact rational weights so candidate differences cannot come from floating-point rounding.
func weightMembers(hypothesis shareHypothesis, eligible []member) ([]weightedMember, *big.Rat) {
	weighted := make([]weightedMember, 0, len(eligible))
	totalWeight := new(big.Rat)

	for _, player := range eligible {
		weight := hypothesis.weight(player)
		weighted = append(weighted, weightedMember{player: player, weight: weight})
		totalWeight.Add(totalWeight, weight)
	}

	return weighted, totalWeight
}

// exactWeightedShare allocates the pool without mutating the stored member or total weights reused by later members.
func exactWeightedShare(pool uint64, weight, totalWeight *big.Rat) *big.Rat {
	exact := new(big.Rat).SetInt(new(big.Int).SetUint64(pool))
	exact.Mul(exact, weight)
	exact.Quo(exact, totalWeight)

	return exact
}

// roundRat returns floor, nearest, and ceiling values without converting an exact candidate share to floating point.
func roundRat(value *big.Rat) (uint64, uint64, uint64) {
	numerator := value.Num()
	denominator := value.Denom()
	one := big.NewInt(1)
	two := big.NewInt(2)

	// Num and Denom expose value-owned integers, so every mutating operation starts from a defensive copy.
	floorValue := new(big.Int).Quo(new(big.Int).Set(numerator), denominator)

	ceilingNumerator := new(big.Int).Sub(new(big.Int).Set(denominator), one)
	ceilingNumerator.Add(ceilingNumerator, numerator)
	ceilingValue := new(big.Int).Quo(ceilingNumerator, denominator)

	nearestNumerator := new(big.Int).Mul(new(big.Int).Set(numerator), two)
	nearestNumerator.Add(nearestNumerator, denominator)
	nearestDenominator := new(big.Int).Mul(new(big.Int).Set(denominator), two)
	nearestValue := new(big.Int).Quo(nearestNumerator, nearestDenominator)

	return floorValue.Uint64(), nearestValue.Uint64(), ceilingValue.Uint64()
}

// shareFit reports the first rounding mode that explains every member, preserving floor-first candidate precedence.
func shareFit(predictions map[string]sharePrediction) string {
	for _, rounding := range []string{"floor", "nearest", "ceiling"} {
		if predictionsMatchRounding(predictions, rounding) {
			return rounding
		}
	}

	return "none"
}

// predictionsMatchRounding requires a single rounding rule to explain the complete observed member allocation.
func predictionsMatchRounding(predictions map[string]sharePrediction, rounding string) bool {
	for _, prediction := range predictions {
		var expected uint64

		switch rounding {
		case "floor":
			expected = prediction.Floor
		case "nearest":
			expected = prediction.Nearest
		case "ceiling":
			expected = prediction.Ceiling
		}

		if prediction.Observed != expected {
			return false
		}
	}

	return true
}
