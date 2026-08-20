package main

import "math"

// normalizeCase reduces one validated scenario while retaining its control link and every outcome category.
func normalizeCase(observed, control probeCase) caseReport {
	result := caseReport{
		ID:           observed.ID,
		ControlID:    observed.ControlID,
		Mechanism:    observed.Mechanism,
		Trials:       len(observed.Trials),
		Counts:       make(map[string]int),
		ContextMatch: observed.Mechanism == "control" || sameTrialContext(control, observed),
	}

	for _, current := range observed.Trials {
		result.Counts[current.Outcome]++
		result.TotalDamage += current.HealthBeforeRaw - current.HealthAfterRaw
	}

	if result.Trials > 0 {
		result.MeanDamage = float64(result.TotalDamage) / float64(result.Trials)
	}

	result.DamageRate = observedRate(result.Counts["damage"], result.Trials)
	result.BlockRate = observedRate(result.Counts["block"], result.Trials)
	result.AvoidRate = observedRate(result.Counts["avoid"], result.Trials)
	result.MissRate = observedRate(result.Counts["miss"], result.Trials)
	result.LethalRate = observedRate(result.Counts["lethal"], result.Trials)

	return result
}

// observedRate returns zero values for an empty sample and otherwise pairs the proportion with stable 95-percent
// bounds.
func observedRate(successes, trials int) interval {
	if trials == 0 {
		return interval{}
	}

	low, high := wilson95(successes, trials)

	return interval{
		Observed: float64(successes) / float64(trials),
		Low95:    low,
		High95:   high,
	}
}

// wilson95 uses the fixed 95-percent z-score so reports remain reproducible across captures and tool invocations.
func wilson95(successes, trials int) (float64, float64) {
	const z = 1.959963984540054

	n := float64(trials)
	p := float64(successes) / n
	z2 := z * z
	center := (p + z2/(2*n)) / (1 + z2/n)
	margin := z * math.Sqrt((p*(1-p)+z2/(4*n))/n) / (1 + z2/n)

	return center - margin, center + margin
}
