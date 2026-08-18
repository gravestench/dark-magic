// Command curse_presentation_probe validates and normalizes sanitized visual
// observations from an owned Expansion 1.14d runtime. It identifies the
// client-function-30 presentation roles without reading memory/save data and
// without treating an older or community implementation as behavior evidence.
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
	"os"
	"sort"
)

const (
	probeSchema = "d2legacy.curse_presentation_probe/v1"
	probeTarget = "diablo-ii-lod-1.14d-expansion"
	maxPixel    = 16384
)

type capture struct {
	Schema  string      `json:"schema"`
	Target  string      `json:"target"`
	Source  string      `json:"source"`
	Runtime runtime     `json:"runtime"`
	Cases   []probeCase `json:"cases"`
}

type runtime struct {
	Patch               string `json:"patch"`
	Mode                string `json:"mode"`
	Session             string `json:"session"`
	CharacterOrigin     string `json:"character_origin"`
	ExecutableSHA256    string `json:"executable_sha256"`
	Observation         string `json:"observation"`
	AssetIdentification string `json:"asset_identification"`
	CameraFixed         bool   `json:"camera_fixed"`
	ActorsStationary    bool   `json:"actors_stationary"`
}

type point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type probeCase struct {
	ID            string   `json:"id"`
	SkillID       int      `json:"skill_id"`
	SkillRecord   string   `json:"skill_record"`
	Difficulty    string   `json:"difficulty"`
	Area          string   `json:"area"`
	Caster        point    `json:"caster"`
	Cursor        point    `json:"cursor"`
	TargetRecords []string `json:"target_records"`
	Targets       []point  `json:"targets"`
	Layers        []layer  `json:"layers"`
}

type layer struct {
	MissileRecord string     `json:"missile_record"`
	Present       bool       `json:"present"`
	Instances     []instance `json:"instances"`
}

type instance struct {
	FirstFrame   int    `json:"first_frame"`
	LastFrame    int    `json:"last_frame"`
	Anchor       string `json:"anchor"`
	TargetIndex  *int   `json:"target_index,omitempty"`
	Start        point  `json:"start"`
	End          point  `json:"end"`
	TracksAnchor bool   `json:"tracks_anchor"`
}

type report struct {
	Schema             string       `json:"schema"`
	Target             string       `json:"target"`
	ExecutableSHA256   string       `json:"executable_sha256"`
	CaptureFingerprint string       `json:"capture_fingerprint"`
	Coverage           coverage     `json:"coverage"`
	Cases              []caseReport `json:"cases"`
}

type coverage struct {
	Complete bool     `json:"complete"`
	Missing  []string `json:"missing,omitempty"`
}

type caseReport struct {
	ID         string        `json:"id"`
	SkillID    int           `json:"skill_id"`
	TargetBand string        `json:"target_band"`
	Layers     []layerReport `json:"layers"`
}

type layerReport struct {
	MissileRecord string           `json:"missile_record"`
	Present       bool             `json:"present"`
	Instances     []instanceReport `json:"instances"`
}

type instanceReport struct {
	FirstFrame   int    `json:"first_frame"`
	LastFrame    int    `json:"last_frame"`
	Frames       int    `json:"frames"`
	Anchor       string `json:"anchor"`
	TargetIndex  *int   `json:"target_index,omitempty"`
	StartOffset  point  `json:"start_offset"`
	EndOffset    point  `json:"end_offset"`
	Translated   bool   `json:"translated"`
	TracksAnchor bool   `json:"tracks_anchor"`
}

func main() {
	input := flag.String("input", "", "sanitized owned-runtime curse-presentation probe JSON")
	flag.Parse()
	if *input == "" {
		fmt.Fprintln(os.Stderr, "usage: curse_presentation_probe -input <capture.json>")
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
		return report{}, fmt.Errorf("curse presentation probe: read capture: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var captured capture
	if err := decoder.Decode(&captured); err != nil {
		return report{}, fmt.Errorf("curse presentation probe: decode capture: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return report{}, fmt.Errorf("curse presentation probe: capture must contain one JSON value")
	}
	if err := validate(captured); err != nil {
		return report{}, err
	}
	fingerprint := sha256.Sum256(data)
	result := report{
		Schema: probeSchema + ".report", Target: probeTarget,
		ExecutableSHA256:   captured.Runtime.ExecutableSHA256,
		CaptureFingerprint: hex.EncodeToString(fingerprint[:]),
		Coverage:           coverageFor(captured.Cases),
	}
	for _, observed := range captured.Cases {
		result.Cases = append(result.Cases, normalize(observed))
	}
	return result, nil
}

func validate(captured capture) error {
	var result error
	if captured.Schema != probeSchema {
		result = errors.Join(result, fmt.Errorf("curse presentation probe: schema %q, want %q", captured.Schema, probeSchema))
	}
	if captured.Target != probeTarget {
		result = errors.Join(result, fmt.Errorf("curse presentation probe: target %q, want %q", captured.Target, probeTarget))
	}
	if captured.Source != "owned-runtime" {
		result = errors.Join(result, fmt.Errorf("curse presentation probe: source must be owned-runtime"))
	}
	runtime := captured.Runtime
	if runtime.Patch != "1.14d" || runtime.Mode != "expansion" || runtime.Session != "single-player" {
		result = errors.Join(result, fmt.Errorf("curse presentation probe: runtime must be Expansion 1.14d single-player"))
	}
	if runtime.CharacterOrigin != "probe-created" {
		result = errors.Join(result, fmt.Errorf("curse presentation probe: character must be probe-created"))
	}
	if !validSHA256(runtime.ExecutableSHA256) {
		result = errors.Join(result, fmt.Errorf("curse presentation probe: executable SHA-256 is required"))
	}
	if !oneOf(runtime.Observation, "video-frame-log", "manual-frame-log") ||
		runtime.AssetIdentification != "owned-mpq-dcc-comparison" || !runtime.CameraFixed || !runtime.ActorsStationary {
		result = errors.Join(result, fmt.Errorf("curse presentation probe: requires fixed-camera stationary visual observation with owned-MPQ DCC comparison"))
	}
	if len(captured.Cases) == 0 {
		result = errors.Join(result, fmt.Errorf("curse presentation probe: at least one case is required"))
	}
	seen := make(map[string]bool, len(captured.Cases))
	for _, observed := range captured.Cases {
		if observed.ID == "" || seen[observed.ID] {
			result = errors.Join(result, fmt.Errorf("curse presentation probe: case IDs must be non-empty and unique"))
		}
		seen[observed.ID] = true
		result = errors.Join(result, validateCase(observed))
	}
	return result
}

func validateCase(observed probeCase) error {
	var result error
	expected, record, known := expectedMissiles(observed.SkillID)
	if !known || observed.SkillRecord != record || !oneOf(observed.Difficulty, "normal", "nightmare", "hell") || observed.Area == "" {
		result = errors.Join(result, fmt.Errorf("curse presentation probe: case %q has invalid target context", observed.ID))
	}
	if len(observed.TargetRecords) != len(observed.Targets) {
		result = errors.Join(result, fmt.Errorf("curse presentation probe: case %q target records/anchors differ", observed.ID))
	}
	if !validPoint(observed.Caster) || !validPoint(observed.Cursor) {
		result = errors.Join(result, fmt.Errorf("curse presentation probe: case %q has invalid caster/cursor coordinates", observed.ID))
	}
	for index, target := range observed.Targets {
		if observed.TargetRecords[index] == "" || !validPoint(target) {
			result = errors.Join(result, fmt.Errorf("curse presentation probe: case %q target %d is invalid", observed.ID, index))
		}
	}
	byRecord := make(map[string]layer, len(observed.Layers))
	for _, current := range observed.Layers {
		if _, duplicate := byRecord[current.MissileRecord]; duplicate {
			result = errors.Join(result, fmt.Errorf("curse presentation probe: case %q duplicates layer %q", observed.ID, current.MissileRecord))
		}
		byRecord[current.MissileRecord] = current
	}
	for _, missile := range expected {
		current, exists := byRecord[missile]
		if !exists {
			result = errors.Join(result, fmt.Errorf("curse presentation probe: case %q omits referenced missile %q", observed.ID, missile))
			continue
		}
		result = errors.Join(result, validateLayer(observed, current))
		delete(byRecord, missile)
	}
	for missile := range byRecord {
		result = errors.Join(result, fmt.Errorf("curse presentation probe: case %q invents missile %q", observed.ID, missile))
	}
	return result
}

func validateLayer(observed probeCase, current layer) error {
	if current.Present != (len(current.Instances) > 0) {
		return fmt.Errorf("curse presentation probe: case %q layer %q contradicts presence", observed.ID, current.MissileRecord)
	}
	var result error
	for index, item := range current.Instances {
		if item.FirstFrame < 0 || item.LastFrame < item.FirstFrame || !validPoint(item.Start) || !validPoint(item.End) ||
			!oneOf(item.Anchor, "caster", "cursor", "target") {
			result = errors.Join(result, fmt.Errorf("curse presentation probe: case %q layer %q instance %d is invalid", observed.ID, current.MissileRecord, index))
			continue
		}
		if item.Anchor == "target" {
			if item.TargetIndex == nil || *item.TargetIndex < 0 || *item.TargetIndex >= len(observed.Targets) {
				result = errors.Join(result, fmt.Errorf("curse presentation probe: case %q layer %q instance %d has invalid target anchor", observed.ID, current.MissileRecord, index))
			}
		} else if item.TargetIndex != nil {
			result = errors.Join(result, fmt.Errorf("curse presentation probe: case %q layer %q instance %d has target index for non-target anchor", observed.ID, current.MissileRecord, index))
		}
	}
	return result
}

func normalize(observed probeCase) caseReport {
	result := caseReport{ID: observed.ID, SkillID: observed.SkillID, TargetBand: targetBand(len(observed.Targets))}
	for _, current := range observed.Layers {
		layerResult := layerReport{MissileRecord: current.MissileRecord, Present: current.Present}
		for _, item := range current.Instances {
			anchor := anchorPoint(observed, item)
			layerResult.Instances = append(layerResult.Instances, instanceReport{
				FirstFrame: item.FirstFrame, LastFrame: item.LastFrame, Frames: item.LastFrame - item.FirstFrame + 1,
				Anchor: item.Anchor, TargetIndex: item.TargetIndex,
				StartOffset: subtract(item.Start, anchor), EndOffset: subtract(item.End, anchor),
				Translated: item.Start != item.End, TracksAnchor: item.TracksAnchor,
			})
		}
		result.Layers = append(result.Layers, layerResult)
	}
	return result
}

func coverageFor(cases []probeCase) coverage {
	seen := make(map[string]bool)
	for _, observed := range cases {
		seen[fmt.Sprintf("skill-%d:%s", observed.SkillID, targetBand(len(observed.Targets)))] = true
	}
	result := coverage{}
	for _, skill := range []int{66, 72} {
		for _, band := range []string{"empty", "single", "multiple"} {
			key := fmt.Sprintf("skill-%d:%s", skill, band)
			if !seen[key] {
				result.Missing = append(result.Missing, key)
			}
		}
	}
	sort.Strings(result.Missing)
	result.Complete = len(result.Missing) == 0
	return result
}

func expectedMissiles(skillID int) ([]string, string, bool) {
	switch skillID {
	case 66:
		return []string{"curseamplifydamage", "cursecast"}, "Amplify Damage", true
	case 72:
		return []string{"curseweaken", "cursecast"}, "Weaken", true
	default:
		return nil, "", false
	}
}

func targetBand(count int) string {
	if count == 0 {
		return "empty"
	}
	if count == 1 {
		return "single"
	}
	return "multiple"
}

func anchorPoint(observed probeCase, item instance) point {
	switch item.Anchor {
	case "caster":
		return observed.Caster
	case "cursor":
		return observed.Cursor
	case "target":
		return observed.Targets[*item.TargetIndex]
	default:
		panic("validated anchor")
	}
}

func subtract(value, origin point) point { return point{X: value.X - origin.X, Y: value.Y - origin.Y} }

func validPoint(value point) bool {
	return value.X >= -maxPixel && value.X <= maxPixel && value.Y >= -maxPixel && value.Y <= maxPixel
}

func validSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && hex.EncodeToString(decoded) == value
}

func oneOf(value string, choices ...string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
