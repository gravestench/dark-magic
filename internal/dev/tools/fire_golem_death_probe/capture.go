package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const (
	probeSchema = "d2legacy.fire_golem_death_probe/v1"
	probeTarget = "diablo-ii-lod-1.14d-expansion"
)

type capture struct {
	Schema  string      `json:"schema"`
	Target  string      `json:"target"`
	Source  string      `json:"source"`
	Runtime runtime     `json:"runtime"`
	Records records     `json:"records"`
	Cases   []probeCase `json:"cases"`
}

type runtime struct {
	Patch                string `json:"patch"`
	Mode                 string `json:"mode"`
	Session              string `json:"session"`
	CharacterOrigin      string `json:"character_origin"`
	ExecutableSHA256     string `json:"executable_sha256"`
	Observation          string `json:"observation"`
	ExecutableUnmodified bool   `json:"executable_unmodified"`
}

type records struct {
	SkillID            int    `json:"skill_id"`
	SkillNameKey       string `json:"skill_name_key"`
	LocalizedSkillName string `json:"localized_skill_name"`
	Locale             string `json:"locale"`
	MonsterID          string `json:"monster_id"`
	DeathDamageEnabled bool   `json:"death_damage_enabled"`
	ExtractedSHA256    string `json:"extracted_records_sha256"`
}

type probeCase struct {
	ID              string         `json:"id"`
	Trigger         string         `json:"trigger"`
	Difficulty      string         `json:"difficulty"`
	SkillLevel      int            `json:"skill_level"`
	PlayerLevel     int            `json:"player_level"`
	MapSeed         uint32         `json:"map_seed"`
	EventFrames     eventFrames    `json:"event_frames"`
	ExplosionCenter point          `json:"explosion_center"`
	Targets         []targetSample `json:"targets"`
}

type eventFrames struct {
	OldGolemRemoved  *int `json:"old_golem_removed"`
	ExplosionStarted *int `json:"explosion_started"`
	NewGolemCreated  *int `json:"new_golem_created,omitempty"`
}

type point struct {
	XSubtiles int `json:"x_subtiles"`
	YSubtiles int `json:"y_subtiles"`
}

type targetSample struct {
	ID                 string   `json:"id"`
	Kind               string   `json:"kind"`
	HostileToOwner     bool     `json:"hostile_to_owner"`
	Position           point    `json:"position"`
	DistanceMilli      int      `json:"distance_millisubtiles"`
	HealthBeforeRaw    int64    `json:"health_before_raw"`
	HealthAfterRaw     int64    `json:"health_after_raw"`
	FireResistance     int      `json:"fire_resistance_percent"`
	PhysicalResistance int      `json:"physical_resistance_percent"`
	NoAbsorbOrFlatDR   bool     `json:"no_absorb_or_flat_damage_reduction"`
	Channels           channels `json:"pre_mitigation_channels_raw"`
	DamageEvent        bool     `json:"damage_event"`
	HitReaction        bool     `json:"hit_reaction"`
	Died               bool     `json:"died"`
}

type channels struct {
	Physical  int64 `json:"physical"`
	Fire      int64 `json:"fire"`
	Cold      int64 `json:"cold"`
	Lightning int64 `json:"lightning"`
	Poison    int64 `json:"poison"`
	Magic     int64 `json:"magic"`
}

// analyze validates one strict capture before normalizing it, so rejected evidence never produces a partial report.
func analyze(input io.Reader) (report, error) {
	data, err := io.ReadAll(input)
	if err != nil {
		return report{}, probeErrorf("read capture: %w", err)
	}

	captured, err := decodeCapture(data)
	if err != nil {
		return report{}, err
	}

	if err := validate(captured); err != nil {
		return report{}, err
	}

	// The fingerprint covers the exact source bytes, preserving provenance independently of JSON normalization.
	fingerprint := sha256.Sum256(data)
	result := report{
		Schema:             probeSchema + ".report",
		Target:             probeTarget,
		ExecutableSHA256:   captured.Runtime.ExecutableSHA256,
		RuntimeSession:     captured.Runtime.Session,
		Records:            captured.Records,
		CaptureFingerprint: hex.EncodeToString(fingerprint[:]),
	}

	// Capture case order is evidence order and remains stable in the report.
	for _, observed := range captured.Cases {
		result.Cases = append(result.Cases, normalize(observed))
	}

	return result, nil
}

// decodeCapture accepts exactly one known-schema JSON value, preventing trailing or unknown evidence from being lost.
func decodeCapture(data []byte) (capture, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var captured capture
	if err := decoder.Decode(&captured); err != nil {
		return capture{}, probeErrorf("decode capture: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return capture{}, probeErrorf("capture must contain one JSON value")
	}

	return captured, nil
}

// probeErrorf preserves the established diagnostic prefix so all command errors remain externally consistent.
func probeErrorf(format string, arguments ...any) error {
	return fmt.Errorf("Fire Golem death probe: "+format, arguments...)
}
