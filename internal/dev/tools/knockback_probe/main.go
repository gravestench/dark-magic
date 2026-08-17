// Command knockback_probe validates and normalizes sanitized observations from
// an owned Expansion 1.14d runtime. It compares observations with explicitly
// labeled hypotheses; it does not promote older recovered behavior to policy.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
)

const (
	probeSchema = "d2legacy.knockback_probe/v1"
	probeTarget = "diablo-ii-lod-1.14d-expansion"
)

type capture struct {
	Schema string      `json:"schema"`
	Target string      `json:"target"`
	Source string      `json:"source"`
	Cases  []probeCase `json:"cases"`
}

type probeCase struct {
	ID                    string  `json:"id"`
	ControlID             string  `json:"control_id,omitempty"`
	Mechanism             string  `json:"mechanism"`
	Difficulty            string  `json:"difficulty"`
	AttackerKind          string  `json:"attacker_kind"`
	Target                target  `json:"target"`
	MissileKnockbackValue int     `json:"missile_knockback_value,omitempty"`
	OpenDistanceSubtiles  float64 `json:"open_distance_subtiles"`
	Trials                []trial `json:"trials"`
}

type target struct {
	Kind          string `json:"kind"`
	Record        string `json:"record"`
	SizeClass     string `json:"size_class"`
	ModeSupported bool   `json:"mode_supported"`
}

type trial struct {
	Hit                  bool    `json:"hit"`
	CombatBlocked        bool    `json:"combat_blocked"`
	Lethal               bool    `json:"lethal"`
	Uninterruptible      bool    `json:"uninterruptible"`
	DisplacementSubtiles float64 `json:"displacement_subtiles"`
	Reaction             string  `json:"reaction"`
}

type report struct {
	Schema string       `json:"schema"`
	Target string       `json:"target"`
	Cases  []caseReport `json:"cases"`
}

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

type candidate struct {
	Name                 string  `json:"name"`
	Evidence             string  `json:"evidence"`
	ExpectedRate         float64 `json:"expected_rate"`
	ExpectedCount        float64 `json:"expected_count"`
	AbsoluteRateError    float64 `json:"absolute_rate_error"`
	InsideObserved95Band bool    `json:"inside_observed_95_band"`
}

func main() {
	input := flag.String("input", "", "sanitized owned-runtime knockback probe JSON")
	flag.Parse()
	if *input == "" {
		fmt.Fprintln(os.Stderr, "usage: knockback_probe -input <capture.json>")
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
		return report{}, fmt.Errorf("knockback probe: decode capture: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return report{}, fmt.Errorf("knockback probe: capture must contain one JSON value")
	}
	if err := validate(captured); err != nil {
		return report{}, err
	}
	result := report{Schema: probeSchema + ".report", Target: probeTarget}
	for _, observed := range captured.Cases {
		result.Cases = append(result.Cases, normalize(observed))
	}
	return result, nil
}

func validate(captured capture) error {
	var result error
	if captured.Schema != probeSchema {
		result = errors.Join(result, fmt.Errorf("knockback probe: schema %q, want %q", captured.Schema, probeSchema))
	}
	if captured.Target != probeTarget {
		result = errors.Join(result, fmt.Errorf("knockback probe: target %q, want %q", captured.Target, probeTarget))
	}
	if captured.Source != "owned-runtime" {
		result = errors.Join(result, fmt.Errorf("knockback probe: source must be owned-runtime"))
	}
	if len(captured.Cases) == 0 {
		result = errors.Join(result, fmt.Errorf("knockback probe: at least one case is required"))
	}
	byID := make(map[string]probeCase, len(captured.Cases))
	for _, observed := range captured.Cases {
		if observed.ID == "" {
			result = errors.Join(result, fmt.Errorf("knockback probe: case ID is required"))
		} else if _, exists := byID[observed.ID]; exists {
			result = errors.Join(result, fmt.Errorf("knockback probe: duplicate case %q", observed.ID))
		}
		byID[observed.ID] = observed
		result = errors.Join(result, validateCase(observed))
	}
	for _, observed := range captured.Cases {
		if observed.Mechanism == "control" {
			if observed.ControlID != "" {
				result = errors.Join(result, fmt.Errorf("knockback probe: control case %q references another control", observed.ID))
			}
			continue
		}
		control, exists := byID[observed.ControlID]
		if !exists || control.Mechanism != "control" {
			result = errors.Join(result, fmt.Errorf("knockback probe: case %q requires a control case", observed.ID))
			continue
		}
		if !sameContext(control, observed) {
			result = errors.Join(result, fmt.Errorf("knockback probe: case %q differs from control context", observed.ID))
		}
	}
	return result
}

func validateCase(observed probeCase) error {
	var result error
	if !oneOf(observed.Mechanism, "control", "item_knockback", "missile_knockback") {
		result = errors.Join(result, fmt.Errorf("knockback probe: case %q has invalid mechanism", observed.ID))
	}
	if !oneOf(observed.Difficulty, "normal", "nightmare", "hell") {
		result = errors.Join(result, fmt.Errorf("knockback probe: case %q has invalid difficulty", observed.ID))
	}
	if !oneOf(observed.AttackerKind, "player", "monster", "hireling", "summon") {
		result = errors.Join(result, fmt.Errorf("knockback probe: case %q has invalid attacker kind", observed.ID))
	}
	if !oneOf(observed.Target.Kind, "player", "monster", "hireling", "summon", "npc", "corpse") || observed.Target.Record == "" {
		result = errors.Join(result, fmt.Errorf("knockback probe: case %q has invalid target identity", observed.ID))
	}
	if !oneOf(observed.Target.SizeClass, "none", "small", "normal", "large") {
		result = errors.Join(result, fmt.Errorf("knockback probe: case %q has invalid target size", observed.ID))
	}
	if observed.OpenDistanceSubtiles < 0 || math.IsNaN(observed.OpenDistanceSubtiles) || math.IsInf(observed.OpenDistanceSubtiles, 0) {
		result = errors.Join(result, fmt.Errorf("knockback probe: case %q has invalid open distance", observed.ID))
	}
	if observed.Mechanism == "missile_knockback" {
		if observed.MissileKnockbackValue < 1 || observed.MissileKnockbackValue > 255 {
			result = errors.Join(result, fmt.Errorf("knockback probe: case %q has invalid missile KnockBack byte", observed.ID))
		}
	} else if observed.MissileKnockbackValue != 0 {
		result = errors.Join(result, fmt.Errorf("knockback probe: case %q has a missile byte outside a missile case", observed.ID))
	}
	if len(observed.Trials) == 0 {
		result = errors.Join(result, fmt.Errorf("knockback probe: case %q requires trials", observed.ID))
	}
	for index, current := range observed.Trials {
		if current.DisplacementSubtiles < 0 || math.IsNaN(current.DisplacementSubtiles) || math.IsInf(current.DisplacementSubtiles, 0) ||
			!oneOf(current.Reaction, "none", "gethit", "knockback", "death", "dead") {
			result = errors.Join(result, fmt.Errorf("knockback probe: case %q trial %d is invalid", observed.ID, index))
		}
		if !current.Hit && (current.CombatBlocked || current.Lethal || current.Uninterruptible || current.DisplacementSubtiles != 0 || current.Reaction != "none") {
			result = errors.Join(result, fmt.Errorf("knockback probe: case %q trial %d reacts without a hit", observed.ID, index))
		}
	}
	return result
}

func sameContext(control, observed probeCase) bool {
	return control.Difficulty == observed.Difficulty &&
		control.AttackerKind == observed.AttackerKind &&
		control.Target == observed.Target &&
		control.OpenDistanceSubtiles == observed.OpenDistanceSubtiles &&
		len(control.Trials) == len(observed.Trials)
}

func normalize(observed probeCase) caseReport {
	result := caseReport{
		ID: observed.ID, ControlID: observed.ControlID, Mechanism: observed.Mechanism,
		ReactionCounts:   make(map[string]int),
		ChanceObservable: observed.Mechanism == "control" || observed.Target.ModeSupported,
	}
	for _, current := range observed.Trials {
		result.ReactionCounts[current.Reaction]++
		if !current.Hit || current.CombatBlocked || current.Lethal || current.Uninterruptible {
			continue
		}
		result.EligibleTrials++
		knockback := current.Reaction == "knockback" || current.DisplacementSubtiles > 0
		if knockback {
			result.KnockbackReactions++
		}
		if current.DisplacementSubtiles > 0 {
			result.MovedTrials++
			result.DisplacementSubtiles = append(result.DisplacementSubtiles, current.DisplacementSubtiles)
		} else if current.Reaction == "knockback" {
			result.BlockedMotionTrials++
		}
	}
	sort.Float64s(result.DisplacementSubtiles)
	if result.EligibleTrials > 0 {
		result.ObservedRate = float64(result.KnockbackReactions) / float64(result.EligibleTrials)
		result.Rate95Low, result.Rate95High = wilson95(result.KnockbackReactions, result.EligibleTrials)
	}
	for _, hypothesis := range hypotheses(observed, result.ChanceObservable) {
		hypothesis.ExpectedCount = hypothesis.ExpectedRate * float64(result.EligibleTrials)
		hypothesis.AbsoluteRateError = math.Abs(hypothesis.ExpectedRate - result.ObservedRate)
		hypothesis.InsideObserved95Band = result.EligibleTrials > 0 &&
			hypothesis.ExpectedRate >= result.Rate95Low && hypothesis.ExpectedRate <= result.Rate95High
		result.Candidates = append(result.Candidates, hypothesis)
	}
	return result
}

func hypotheses(observed probeCase, observable bool) []candidate {
	if !observable {
		return nil
	}
	switch observed.Mechanism {
	case "control":
		return []candidate{{Name: "no_knockback_control", Evidence: "neutral control", ExpectedRate: 0}}
	case "missile_knockback":
		return []candidate{
			{Name: "raw_byte_percent", Evidence: "older recovered source only", ExpectedRate: math.Min(float64(observed.MissileKnockbackValue), 100) / 100},
			{Name: "nonzero_boolean", Evidence: "alternative hypothesis", ExpectedRate: 1},
		}
	case "item_knockback":
		rate := 0.5
		if observed.Target.SizeClass == "small" {
			rate = 1
		} else if observed.Target.SizeClass == "large" {
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

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
