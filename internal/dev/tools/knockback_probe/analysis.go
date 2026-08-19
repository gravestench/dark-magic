package main

import (
	"io"
	"math"
	"sort"
)

// report is the stable normalized output envelope consumed by later evidence
// review, so its schema and case ordering mirror the validated input.
type report struct {
	Schema string       `json:"schema"`
	Target string       `json:"target"`
	Cases  []caseReport `json:"cases"`
}

// caseReport separates observable counts from labeled hypotheses so raw
// observations never become policy merely by appearing in the same artifact.
type caseReport struct {
	ID                   string         `json:"id"`
	ControlID            string         `json:"control_id,omitempty"`
	Mechanism            string         `json:"mechanism"`
	EligibleTrials       int            `json:"eligible_trials"`
	KnockbackReactions   int            `json:"knockback_reactions"`
	MovedTrials          int            `json:"moved_trials"`
	BlockedMotionTrials  int            `json:"blocked_motion_trials"`
	ObservedRate         float64        `json:"observed_rate"`
	Rate95Low            float64        `json:"rate_95_low"`
	Rate95High           float64        `json:"rate_95_high"`
	ReactionCounts       map[string]int `json:"reaction_counts"`
	DisplacementSubtiles []float64      `json:"displacement_subtiles,omitempty"`
	ChanceObservable     bool           `json:"chance_observable"`
	Candidates           []candidate    `json:"candidates,omitempty"`
}

// candidate records a named interpretation alongside its fit to the observed
// sample, retaining the evidence label that limits how the result may be used.
type candidate struct {
	Name                 string  `json:"name"`
	Evidence             string  `json:"evidence"`
	ExpectedRate         float64 `json:"expected_rate"`
	ExpectedCount        float64 `json:"expected_count"`
	AbsoluteRateError    float64 `json:"absolute_rate_error"`
	InsideObserved95Band bool    `json:"inside_observed_95_band"`
}

// analyze enforces strict decoding and complete validation before constructing
// any report, preventing partial results from lending authority to invalid evidence.
func analyze(input io.Reader) (report, error) {
	captured, err := decodeCapture(input)
	if err != nil {
		return report{}, err
	}

	if err := validate(captured); err != nil {
		return report{}, err
	}

	return normalizeCapture(captured), nil
}

// normalizeCapture preserves capture order in the report so controls and their
// experimental cases remain easy to compare and output stays deterministic.
func normalizeCapture(captured capture) report {
	result := report{Schema: probeSchema + ".report", Target: probeTarget}
	for _, observed := range captured.Cases {
		result.Cases = append(result.Cases, normalize(observed))
	}

	return result
}

// normalize reduces one validated case to observable counts and then scores
// explicitly labeled hypotheses without changing the underlying observations.
func normalize(observed probeCase) caseReport {
	result := caseReport{
		ID:               observed.ID,
		ControlID:        observed.ControlID,
		Mechanism:        observed.Mechanism,
		ReactionCounts:   make(map[string]int),
		ChanceObservable: observed.Mechanism == "control" || observed.Target.ModeSupported,
	}
	for _, current := range observed.Trials {
		recordTrial(&result, current)
	}

	// Stable displacement ordering makes reports reproducible regardless of the
	// sequence in which equivalent trial distances were recorded.
	sort.Float64s(result.DisplacementSubtiles)

	if result.EligibleTrials > 0 {
		result.ObservedRate = float64(result.KnockbackReactions) / float64(result.EligibleTrials)
		result.Rate95Low, result.Rate95High = wilson95(result.KnockbackReactions, result.EligibleTrials)
	}

	for _, hypothesis := range hypotheses(observed, result.ChanceObservable) {
		hypothesis.ExpectedCount = hypothesis.ExpectedRate * float64(result.EligibleTrials)
		hypothesis.AbsoluteRateError = math.Abs(hypothesis.ExpectedRate - result.ObservedRate)
		hypothesis.InsideObserved95Band = result.EligibleTrials > 0 &&
			hypothesis.ExpectedRate >= result.Rate95Low &&
			hypothesis.ExpectedRate <= result.Rate95High
		result.Candidates = append(result.Candidates, hypothesis)
	}

	return result
}

// recordTrial counts every reaction but admits only interpretable combat events
// to the chance denominator, keeping blocked motion distinct from absent reactions.
func recordTrial(result *caseReport, observed trial) {
	result.ReactionCounts[observed.Reaction]++

	if !eligibleForChance(observed) {
		return
	}

	result.EligibleTrials++
	if observed.Reaction == "knockback" || observed.DisplacementSubtiles > 0 {
		result.KnockbackReactions++
	}

	if observed.DisplacementSubtiles > 0 {
		result.MovedTrials++
		result.DisplacementSubtiles = append(result.DisplacementSubtiles, observed.DisplacementSubtiles)

		return
	}

	if observed.Reaction == "knockback" {
		result.BlockedMotionTrials++
	}
}

// eligibleForChance excludes misses and combat states that mask whether the
// knockback mechanism fired, keeping the statistical denominator interpretable.
func eligibleForChance(observed trial) bool {
	return observed.Hit &&
		!observed.CombatBlocked &&
		!observed.Lethal &&
		!observed.Uninterruptible
}

// hypotheses returns candidates in evidence-review order; callers preserve
// this ordering in serialized reports for stable comparisons across captures.
func hypotheses(observed probeCase, observable bool) []candidate {
	if !observable {
		return nil
	}

	switch observed.Mechanism {
	case "control":
		return []candidate{
			{Name: "no_knockback_control", Evidence: "neutral control", ExpectedRate: 0},
		}
	case "missile_knockback":
		return []candidate{
			{
				Name:         "raw_byte_percent",
				Evidence:     "older recovered source only",
				ExpectedRate: math.Min(float64(observed.MissileKnockbackValue), 100) / 100,
			},
			{Name: "nonzero_boolean", Evidence: "alternative hypothesis", ExpectedRate: 1},
		}
	case "item_knockback":
		// The recovered hypothesis distinguishes only small and large targets;
		// every other validated size retains the normal-size rate.
		rate := 0.5

		switch observed.Target.SizeClass {
		case "small":
			rate = 1
		case "large":
			rate = 0.25
		}

		return []candidate{
			{Name: "size_weighted_128_roll", Evidence: "older recovered source only", ExpectedRate: rate},
			{Name: "always_when_present", Evidence: "alternative hypothesis", ExpectedRate: 1},
		}
	default:
		return nil
	}
}

// wilson95 computes a bounded 95-percent Wilson score interval; clamping absorbs
// floating-point edge noise so serialized rates remain valid probabilities.
func wilson95(successes, total int) (float64, float64) {
	if total == 0 {
		return 0, 0
	}

	const z = 1.959963984540054

	n := float64(total)
	p := float64(successes) / n
	denominator := 1 + z*z/n
	center := (p + z*z/(2*n)) / denominator
	margin := z * math.Sqrt(p*(1-p)/n+z*z/(4*n*n)) / denominator

	return math.Max(0, center-margin), math.Min(1, center+margin)
}
