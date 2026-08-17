// Command defense_outcome_probe validates and normalizes sanitized visual
// observations from an owned Expansion 1.14d runtime. It records evidence for
// attack/block/avoid ordering; it does not implement or infer that policy.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
)

const (
	probeSchema = "d2legacy.defense_outcome_probe/v1"
	probeTarget = "diablo-ii-lod-1.14d-expansion"
)

type capture struct {
	Schema  string      `json:"schema"`
	Target  string      `json:"target"`
	Source  string      `json:"source"`
	Runtime runtime     `json:"runtime"`
	Cases   []probeCase `json:"cases"`
}

type runtime struct {
	Patch            string `json:"patch"`
	Mode             string `json:"mode"`
	Session          string `json:"session"`
	ExecutableSHA256 string `json:"executable_sha256"`
	Observation      string `json:"observation"`
}

type probeCase struct {
	ID                     string   `json:"id"`
	ControlID              string   `json:"control_id,omitempty"`
	Mechanism              string   `json:"mechanism"`
	EffectRecord           string   `json:"effect_record,omitempty"`
	DisplayedChancePercent int      `json:"displayed_chance_percent,omitempty"`
	Difficulty             string   `json:"difficulty"`
	AttackKind             string   `json:"attack_kind"`
	AttackerKind           string   `json:"attacker_kind"`
	AttackerLevel          int      `json:"attacker_level"`
	AttackRating           int      `json:"attack_rating"`
	Defender               defender `json:"defender"`
	Trials                 []trial  `json:"trials"`
}

type defender struct {
	Kind    string `json:"kind"`
	Record  string `json:"record"`
	Level   int    `json:"level"`
	Defense int    `json:"defense"`
	State   string `json:"state"`
}

type trial struct {
	Outcome         string `json:"outcome"`
	Reaction        string `json:"reaction"`
	HealthBeforeRaw int64  `json:"health_before_raw"`
	HealthAfterRaw  int64  `json:"health_after_raw"`
}

type report struct {
	Schema             string       `json:"schema"`
	Target             string       `json:"target"`
	ExecutableSHA256   string       `json:"executable_sha256"`
	CaptureFingerprint string       `json:"capture_fingerprint"`
	Cases              []caseReport `json:"cases"`
}

type caseReport struct {
	ID           string         `json:"id"`
	ControlID    string         `json:"control_id,omitempty"`
	Mechanism    string         `json:"mechanism"`
	Trials       int            `json:"trials"`
	Counts       map[string]int `json:"counts"`
	DamageRate   interval       `json:"damage_rate"`
	BlockRate    interval       `json:"block_rate"`
	AvoidRate    interval       `json:"avoid_rate"`
	MissRate     interval       `json:"miss_rate"`
	LethalRate   interval       `json:"lethal_rate"`
	TotalDamage  int64          `json:"total_damage_raw"`
	MeanDamage   float64        `json:"mean_damage_raw"`
	ContextMatch bool           `json:"context_matches_control"`
}

type interval struct {
	Observed float64 `json:"observed"`
	Low95    float64 `json:"low_95"`
	High95   float64 `json:"high_95"`
}

func main() {
	input := flag.String("input", "", "sanitized owned-runtime defense-outcome probe JSON")
	flag.Parse()
	if *input == "" {
		fmt.Fprintln(os.Stderr, "usage: defense_outcome_probe -input <capture.json>")
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
	data, err := io.ReadAll(input)
	if err != nil {
		return report{}, fmt.Errorf("defense outcome probe: read capture: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var captured capture
	if err := decoder.Decode(&captured); err != nil {
		return report{}, fmt.Errorf("defense outcome probe: decode capture: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return report{}, fmt.Errorf("defense outcome probe: capture must contain one JSON value")
	}
	if err := validate(captured); err != nil {
		return report{}, err
	}
	fingerprint := sha256.Sum256(data)
	result := report{
		Schema:             probeSchema + ".report",
		Target:             probeTarget,
		ExecutableSHA256:   captured.Runtime.ExecutableSHA256,
		CaptureFingerprint: hex.EncodeToString(fingerprint[:]),
	}
	byID := make(map[string]probeCase, len(captured.Cases))
	for _, observed := range captured.Cases {
		byID[observed.ID] = observed
	}
	for _, observed := range captured.Cases {
		result.Cases = append(result.Cases, normalize(observed, byID[observed.ControlID]))
	}
	return result, nil
}

func validate(captured capture) error {
	var result error
	if captured.Schema != probeSchema {
		result = errors.Join(result, fmt.Errorf("defense outcome probe: schema %q, want %q", captured.Schema, probeSchema))
	}
	if captured.Target != probeTarget {
		result = errors.Join(result, fmt.Errorf("defense outcome probe: target %q, want %q", captured.Target, probeTarget))
	}
	if captured.Source != "owned-runtime" {
		result = errors.Join(result, fmt.Errorf("defense outcome probe: source must be owned-runtime"))
	}
	if captured.Runtime.Patch != "1.14d" || captured.Runtime.Mode != "expansion" || captured.Runtime.Session != "single-player" {
		result = errors.Join(result, fmt.Errorf("defense outcome probe: runtime must be Expansion 1.14d single-player"))
	}
	if !validSHA256(captured.Runtime.ExecutableSHA256) {
		result = errors.Join(result, fmt.Errorf("defense outcome probe: executable SHA-256 is required"))
	}
	if !oneOf(captured.Runtime.Observation, "video-frame-log", "manual-frame-log") {
		result = errors.Join(result, fmt.Errorf("defense outcome probe: unsupported observation method"))
	}
	if len(captured.Cases) == 0 {
		result = errors.Join(result, fmt.Errorf("defense outcome probe: at least one case is required"))
	}
	byID := make(map[string]probeCase, len(captured.Cases))
	for _, observed := range captured.Cases {
		if observed.ID == "" {
			result = errors.Join(result, fmt.Errorf("defense outcome probe: case ID is required"))
		} else if _, exists := byID[observed.ID]; exists {
			result = errors.Join(result, fmt.Errorf("defense outcome probe: duplicate case %q", observed.ID))
		}
		byID[observed.ID] = observed
		result = errors.Join(result, validateCase(observed))
	}
	for _, observed := range captured.Cases {
		if observed.Mechanism == "control" {
			if observed.ControlID != "" {
				result = errors.Join(result, fmt.Errorf("defense outcome probe: control %q references another control", observed.ID))
			}
			continue
		}
		control, exists := byID[observed.ControlID]
		if !exists || control.Mechanism != "control" {
			result = errors.Join(result, fmt.Errorf("defense outcome probe: case %q requires a control", observed.ID))
		} else if !sameContext(control, observed) {
			result = errors.Join(result, fmt.Errorf("defense outcome probe: case %q differs from control context", observed.ID))
		}
	}
	return result
}

func validateCase(observed probeCase) error {
	var result error
	if !oneOf(observed.Mechanism, "control", "shield_block", "passive_avoidance") {
		result = errors.Join(result, fmt.Errorf("defense outcome probe: case %q has invalid mechanism", observed.ID))
	}
	if !oneOf(observed.Difficulty, "normal", "nightmare", "hell") ||
		!oneOf(observed.AttackKind, "melee", "missile") ||
		!oneOf(observed.AttackerKind, "player", "monster") {
		result = errors.Join(result, fmt.Errorf("defense outcome probe: case %q has invalid attack context", observed.ID))
	}
	if observed.AttackerLevel < 1 || observed.AttackRating < 0 || observed.Defender.Level < 1 || observed.Defender.Defense < 0 ||
		!oneOf(observed.Defender.Kind, "player", "monster") || observed.Defender.Record == "" ||
		!oneOf(observed.Defender.State, "standing", "walking", "running", "attacking", "casting") {
		result = errors.Join(result, fmt.Errorf("defense outcome probe: case %q has invalid participant facts", observed.ID))
	}
	if observed.Mechanism == "control" {
		if observed.EffectRecord != "" || observed.DisplayedChancePercent != 0 {
			result = errors.Join(result, fmt.Errorf("defense outcome probe: control %q has a defense effect", observed.ID))
		}
	} else if observed.EffectRecord == "" || observed.DisplayedChancePercent < 1 || observed.DisplayedChancePercent > 100 {
		result = errors.Join(result, fmt.Errorf("defense outcome probe: case %q requires an effect and displayed chance", observed.ID))
	}
	if len(observed.Trials) == 0 {
		result = errors.Join(result, fmt.Errorf("defense outcome probe: case %q requires trials", observed.ID))
	}
	for index, current := range observed.Trials {
		if !oneOf(current.Outcome, "miss", "damage", "block", "avoid", "lethal") ||
			!oneOf(current.Reaction, "none", "gethit", "block", "avoid", "death") ||
			current.HealthBeforeRaw < 0 || current.HealthAfterRaw < 0 || current.HealthAfterRaw > current.HealthBeforeRaw {
			result = errors.Join(result, fmt.Errorf("defense outcome probe: case %q trial %d is invalid", observed.ID, index))
			continue
		}
		damage := current.HealthBeforeRaw - current.HealthAfterRaw
		if (current.Outcome == "damage" || current.Outcome == "lethal") != (damage > 0) ||
			(current.Outcome == "lethal") != (current.HealthBeforeRaw > 0 && current.HealthAfterRaw == 0) ||
			(current.Outcome == "lethal") != (current.Reaction == "death") ||
			(current.Outcome == "miss") != (current.Reaction == "none") ||
			(current.Outcome == "block") != (current.Reaction == "block") ||
			(current.Outcome == "avoid") != (current.Reaction == "avoid") {
			result = errors.Join(result, fmt.Errorf("defense outcome probe: case %q trial %d has inconsistent outcome facts", observed.ID, index))
		}
	}
	return result
}

func sameContext(left, right probeCase) bool {
	return left.Difficulty == right.Difficulty && left.AttackKind == right.AttackKind &&
		left.AttackerKind == right.AttackerKind && left.AttackerLevel == right.AttackerLevel &&
		left.AttackRating == right.AttackRating && left.Defender == right.Defender && len(left.Trials) == len(right.Trials)
}

func normalize(observed, control probeCase) caseReport {
	result := caseReport{ID: observed.ID, ControlID: observed.ControlID, Mechanism: observed.Mechanism,
		Trials: len(observed.Trials), Counts: make(map[string]int),
		ContextMatch: observed.Mechanism == "control" || sameContext(control, observed)}
	for _, current := range observed.Trials {
		result.Counts[current.Outcome]++
		result.TotalDamage += current.HealthBeforeRaw - current.HealthAfterRaw
	}
	if result.Trials > 0 {
		result.MeanDamage = float64(result.TotalDamage) / float64(result.Trials)
	}
	result.DamageRate = rate(result.Counts["damage"], result.Trials)
	result.BlockRate = rate(result.Counts["block"], result.Trials)
	result.AvoidRate = rate(result.Counts["avoid"], result.Trials)
	result.MissRate = rate(result.Counts["miss"], result.Trials)
	result.LethalRate = rate(result.Counts["lethal"], result.Trials)
	return result
}

func rate(successes, trials int) interval {
	if trials == 0 {
		return interval{}
	}
	low, high := wilson95(successes, trials)
	return interval{Observed: float64(successes) / float64(trials), Low95: low, High95: high}
}

func wilson95(successes, trials int) (float64, float64) {
	const z = 1.959963984540054
	n := float64(trials)
	p := float64(successes) / n
	z2 := z * z
	center := (p + z2/(2*n)) / (1 + z2/n)
	margin := z * math.Sqrt((p*(1-p)+z2/(4*n))/n) / (1 + z2/n)
	return center - margin, center + margin
}

func validSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func oneOf(value string, choices ...string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}

func fatal(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
