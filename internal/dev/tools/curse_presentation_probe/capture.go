package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const (
	probeSchema = "d2legacy.curse_presentation_probe/v1"
	probeTarget = "diablo-ii-lod-1.14d-expansion"
	maxPixel    = 16384
)

// capture is the complete owned-runtime observation envelope. Its field order and tags define the accepted JSON
// format, so validation operates on this representation without reshaping input evidence.
type capture struct {
	Schema  string      `json:"schema"`
	Target  string      `json:"target"`
	Source  string      `json:"source"`
	Runtime runtime     `json:"runtime"`
	Cases   []probeCase `json:"cases"`
}

// runtime records the provenance constraints that separate owned 1.14d observations from unsupported evidence.
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

// point stores capture-space coordinates. Validation bounds both axes before normalization performs arithmetic.
type point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// probeCase groups one skill observation with the anchors and visual layers needed to interpret it.
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

// layer identifies whether a referenced missile appeared and retains every observed instance in capture order.
type layer struct {
	MissileRecord string     `json:"missile_record"`
	Present       bool       `json:"present"`
	Instances     []instance `json:"instances"`
}

// instance describes an observed missile lifetime and motion relative to its declared anchor.
type instance struct {
	FirstFrame   int    `json:"first_frame"`
	LastFrame    int    `json:"last_frame"`
	Anchor       string `json:"anchor"`
	TargetIndex  *int   `json:"target_index,omitempty"`
	Start        point  `json:"start"`
	End          point  `json:"end"`
	TracksAnchor bool   `json:"tracks_anchor"`
}

// analyze reads one capture, validates its provenance and structure, then derives a normalized report. Hashing the
// original bytes ensures the report fingerprint identifies the exact evidence supplied by the caller.
func analyze(input io.Reader) (report, error) {
	rawCapture, err := io.ReadAll(input)
	if err != nil {
		return report{}, fmt.Errorf("curse presentation probe: read capture: %w", err)
	}

	captured, err := decodeCapture(rawCapture)
	if err != nil {
		return report{}, err
	}

	if err := validate(captured); err != nil {
		return report{}, err
	}

	return buildReport(captured, rawCapture), nil
}

// decodeCapture accepts exactly one JSON value and rejects unknown fields so misspelled evidence cannot be silently
// omitted from later validation.
func decodeCapture(rawCapture []byte) (capture, error) {
	decoder := json.NewDecoder(bytes.NewReader(rawCapture))
	decoder.DisallowUnknownFields()

	var captured capture
	if err := decoder.Decode(&captured); err != nil {
		return capture{}, fmt.Errorf("curse presentation probe: decode capture: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return capture{}, fmt.Errorf("curse presentation probe: capture must contain one JSON value")
	}

	return captured, nil
}
