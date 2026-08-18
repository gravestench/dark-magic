// Command fire_golem_death_probe validates and normalizes an instrumented
// observation from an owned Expansion 1.14d runtime. It records the facts
// needed to implement Fire Golem's death splash without importing an older
// patch's general monster-death routine into current gameplay policy.
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
	"sort"
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

type report struct {
	Schema             string       `json:"schema"`
	Target             string       `json:"target"`
	ExecutableSHA256   string       `json:"executable_sha256"`
	RuntimeSession     string       `json:"runtime_session"`
	Records            records      `json:"records"`
	CaptureFingerprint string       `json:"capture_fingerprint"`
	Cases              []caseReport `json:"cases"`
}

type caseReport struct {
	ID              string          `json:"id"`
	Trigger         string          `json:"trigger"`
	Difficulty      string          `json:"difficulty"`
	SkillLevel      int             `json:"skill_level"`
	PlayerLevel     int             `json:"player_level"`
	MapSeed         uint32          `json:"map_seed"`
	ExplosionCenter point           `json:"explosion_center"`
	OrderedEvents   []orderedEvent  `json:"ordered_events"`
	Targets         []targetReport  `json:"targets"`
	AffectedKinds   []string        `json:"affected_kinds"`
	UnaffectedKinds []string        `json:"unaffected_kinds"`
	RadiusBrackets  []radiusBracket `json:"radius_brackets"`
}

type radiusBracket struct {
	Kind                   string `json:"kind"`
	HostileToOwner         bool   `json:"hostile_to_owner"`
	FarthestAffectedMilli  int    `json:"farthest_affected_millisubtiles"`
	NearestUnaffectedMilli int    `json:"nearest_unaffected_millisubtiles"`
	BoundaryBracketed      bool   `json:"boundary_bracketed"`
}

type orderedEvent struct {
	Name  string `json:"name"`
	Frame int    `json:"frame"`
}

type targetReport struct {
	ID                 string   `json:"id"`
	Kind               string   `json:"kind"`
	HostileToOwner     bool     `json:"hostile_to_owner"`
	Position           point    `json:"position"`
	DistanceMilli      int      `json:"distance_millisubtiles"`
	HealthDeltaRaw     int64    `json:"health_delta_raw"`
	FireResistance     int      `json:"fire_resistance_percent"`
	PhysicalResistance int      `json:"physical_resistance_percent"`
	Channels           channels `json:"pre_mitigation_channels_raw"`
	DamageEvent        bool     `json:"damage_event"`
	HitReaction        bool     `json:"hit_reaction"`
	Died               bool     `json:"died"`
}

func main() {
	input := flag.String("input", "", "sanitized owned-runtime Fire Golem death probe JSON")
	flag.Parse()
	if *input == "" {
		fmt.Fprintln(os.Stderr, "usage: fire_golem_death_probe -input <capture.json>")
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
		return report{}, fmt.Errorf("Fire Golem death probe: read capture: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var captured capture
	if err := decoder.Decode(&captured); err != nil {
		return report{}, fmt.Errorf("Fire Golem death probe: decode capture: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return report{}, fmt.Errorf("Fire Golem death probe: capture must contain one JSON value")
	}
	if err := validate(captured); err != nil {
		return report{}, err
	}
	fingerprint := sha256.Sum256(data)
	result := report{
		Schema:             probeSchema + ".report",
		Target:             probeTarget,
		ExecutableSHA256:   captured.Runtime.ExecutableSHA256,
		RuntimeSession:     captured.Runtime.Session,
		Records:            captured.Records,
		CaptureFingerprint: hex.EncodeToString(fingerprint[:]),
	}
	for _, observed := range captured.Cases {
		result.Cases = append(result.Cases, normalize(observed))
	}
	return result, nil
}

func validate(captured capture) error {
	var result error
	if captured.Schema != probeSchema {
		result = errors.Join(result, fmt.Errorf("Fire Golem death probe: schema %q, want %q", captured.Schema, probeSchema))
	}
	if captured.Target != probeTarget {
		result = errors.Join(result, fmt.Errorf("Fire Golem death probe: target %q, want %q", captured.Target, probeTarget))
	}
	if captured.Source != "owned-runtime" {
		result = errors.Join(result, fmt.Errorf("Fire Golem death probe: source must be owned-runtime"))
	}
	if captured.Runtime.Patch != "1.14d" || captured.Runtime.Mode != "expansion" ||
		!oneOf(captured.Runtime.Session, "single-player", "local-hosted-multiplayer") {
		result = errors.Join(result, fmt.Errorf("Fire Golem death probe: runtime must be Expansion 1.14d single-player or local-hosted multiplayer"))
	}
	if captured.Runtime.CharacterOrigin != "probe-created" {
		result = errors.Join(result, fmt.Errorf("Fire Golem death probe: character must be probe-created"))
	}
	if !validSHA256(captured.Runtime.ExecutableSHA256) {
		result = errors.Join(result, fmt.Errorf("Fire Golem death probe: executable SHA-256 is required"))
	}
	if captured.Runtime.Observation != "debugger-stat-log-plus-video" || !captured.Runtime.ExecutableUnmodified {
		result = errors.Join(result, fmt.Errorf("Fire Golem death probe: requires an unmodified executable and debugger stat log paired with video"))
	}
	if captured.Records.SkillID != 94 || captured.Records.SkillNameKey == "" ||
		captured.Records.LocalizedSkillName == "" || captured.Records.Locale == "" ||
		captured.Records.MonsterID != "firegolem" || !captured.Records.DeathDamageEnabled ||
		!validSHA256(captured.Records.ExtractedSHA256) {
		result = errors.Join(result, fmt.Errorf("Fire Golem death probe: owned Skills/TBL/MonStats record anchors are required"))
	}
	if len(captured.Cases) == 0 {
		result = errors.Join(result, fmt.Errorf("Fire Golem death probe: at least one case is required"))
	}
	seen := make(map[string]bool)
	for _, observed := range captured.Cases {
		if observed.ID == "" || seen[observed.ID] {
			result = errors.Join(result, fmt.Errorf("Fire Golem death probe: case IDs must be non-empty and unique"))
		}
		seen[observed.ID] = true
		result = errors.Join(result, validateCase(observed))
	}
	return result
}

func validateCase(observed probeCase) error {
	var result error
	if !oneOf(observed.Trigger, "replacement", "combat_death") ||
		!oneOf(observed.Difficulty, "normal", "nightmare", "hell") ||
		observed.SkillLevel < 1 || observed.SkillLevel > 99 ||
		observed.PlayerLevel < 30 || observed.PlayerLevel > 99 {
		result = errors.Join(result, fmt.Errorf("Fire Golem death probe: case %q has invalid target context", observed.ID))
	}
	frames := observed.EventFrames
	if frames.OldGolemRemoved == nil || frames.ExplosionStarted == nil || *frames.OldGolemRemoved < 0 || *frames.ExplosionStarted < 0 {
		result = errors.Join(result, fmt.Errorf("Fire Golem death probe: case %q lacks removal/explosion frames", observed.ID))
	}
	if observed.Trigger == "replacement" {
		if frames.NewGolemCreated == nil || *frames.NewGolemCreated < 0 {
			result = errors.Join(result, fmt.Errorf("Fire Golem death probe: replacement case %q lacks the new-golem frame", observed.ID))
		}
	} else if frames.NewGolemCreated != nil {
		result = errors.Join(result, fmt.Errorf("Fire Golem death probe: combat-death case %q creates a replacement", observed.ID))
	}
	if len(observed.Targets) == 0 {
		result = errors.Join(result, fmt.Errorf("Fire Golem death probe: case %q requires target samples", observed.ID))
	}
	seen := make(map[string]bool)
	for index, sample := range observed.Targets {
		hostilityMatchesKind := sample.HostileToOwner == oneOf(sample.Kind, "hostile_player", "hostile_monster")
		if sample.ID == "" || seen[sample.ID] ||
			!oneOf(sample.Kind, "owner", "allied_player", "hostile_player", "allied_minion", "neutral_monster", "hostile_monster") ||
			!hostilityMatchesKind ||
			sample.DistanceMilli < 0 || sample.HealthBeforeRaw < 1 ||
			sample.HealthAfterRaw < 0 || sample.HealthAfterRaw > sample.HealthBeforeRaw ||
			sample.FireResistance < -100 || sample.FireResistance > 255 ||
			sample.PhysicalResistance < -100 || sample.PhysicalResistance > 255 ||
			!sample.NoAbsorbOrFlatDR ||
			!validChannels(sample.Channels) {
			result = errors.Join(result, fmt.Errorf("Fire Golem death probe: case %q target %d is invalid", observed.ID, index))
		}
		seen[sample.ID] = true
		if distanceMilli(observed.ExplosionCenter, sample.Position) != sample.DistanceMilli {
			result = errors.Join(result, fmt.Errorf("Fire Golem death probe: case %q target %q distance does not match its coordinates", observed.ID, sample.ID))
		}
		delta := sample.HealthBeforeRaw - sample.HealthAfterRaw
		if !sample.DamageEvent && (delta != 0 || sample.HitReaction || sample.Died || channelTotal(sample.Channels) != 0) {
			result = errors.Join(result, fmt.Errorf("Fire Golem death probe: case %q unaffected target %q records damage effects", observed.ID, sample.ID))
		}
		if sample.DamageEvent && channelTotal(sample.Channels) == 0 {
			result = errors.Join(result, fmt.Errorf("Fire Golem death probe: case %q affected target %q lacks pre-mitigation channels", observed.ID, sample.ID))
		}
		if sample.Died != (sample.HealthAfterRaw == 0) {
			result = errors.Join(result, fmt.Errorf("Fire Golem death probe: case %q target %q contradicts its death state", observed.ID, sample.ID))
		}
	}
	return result
}

func normalize(observed probeCase) caseReport {
	result := caseReport{
		ID: observed.ID, Trigger: observed.Trigger, Difficulty: observed.Difficulty, SkillLevel: observed.SkillLevel,
		PlayerLevel: observed.PlayerLevel, MapSeed: observed.MapSeed, ExplosionCenter: observed.ExplosionCenter,
	}
	result.OrderedEvents = append(result.OrderedEvents,
		orderedEvent{Name: "old_golem_removed", Frame: *observed.EventFrames.OldGolemRemoved},
		orderedEvent{Name: "explosion_started", Frame: *observed.EventFrames.ExplosionStarted},
	)
	if observed.EventFrames.NewGolemCreated != nil {
		result.OrderedEvents = append(result.OrderedEvents, orderedEvent{Name: "new_golem_created", Frame: *observed.EventFrames.NewGolemCreated})
	}
	sort.SliceStable(result.OrderedEvents, func(i, j int) bool {
		return result.OrderedEvents[i].Frame < result.OrderedEvents[j].Frame
	})
	affectedKinds, unaffectedKinds := map[string]bool{}, map[string]bool{}
	type profileRange struct {
		kind              string
		hostile           bool
		farthestAffected  int
		nearestUnaffected int
	}
	profiles := map[string]*profileRange{}
	for _, sample := range observed.Targets {
		delta := sample.HealthBeforeRaw - sample.HealthAfterRaw
		result.Targets = append(result.Targets, targetReport{
			ID: sample.ID, Kind: sample.Kind, HostileToOwner: sample.HostileToOwner, Position: sample.Position,
			DistanceMilli: sample.DistanceMilli, HealthDeltaRaw: delta,
			FireResistance: sample.FireResistance, PhysicalResistance: sample.PhysicalResistance,
			Channels: sample.Channels, DamageEvent: sample.DamageEvent, HitReaction: sample.HitReaction, Died: sample.Died,
		})
		if sample.DamageEvent {
			affectedKinds[sample.Kind] = true
			key := targetProfile(sample)
			profile := profiles[key]
			if profile == nil {
				profile = &profileRange{kind: sample.Kind, hostile: sample.HostileToOwner, farthestAffected: -1, nearestUnaffected: -1}
				profiles[key] = profile
			}
			if sample.DistanceMilli > profile.farthestAffected {
				profile.farthestAffected = sample.DistanceMilli
			}
		} else {
			unaffectedKinds[sample.Kind] = true
		}
	}
	// A nearby owner or ally proves the target filter, not the radius. Only an
	// unaffected sample with the same unit kind and hostility as an affected
	// sample can bound the splash range.
	for _, sample := range observed.Targets {
		profile := profiles[targetProfile(sample)]
		if !sample.DamageEvent && profile != nil &&
			(profile.nearestUnaffected < 0 || sample.DistanceMilli < profile.nearestUnaffected) {
			profile.nearestUnaffected = sample.DistanceMilli
		}
	}
	result.AffectedKinds = sortedKeys(affectedKinds)
	result.UnaffectedKinds = sortedKeys(unaffectedKinds)
	profileKeys := make([]string, 0, len(profiles))
	for key := range profiles {
		profileKeys = append(profileKeys, key)
	}
	sort.Strings(profileKeys)
	for _, key := range profileKeys {
		profile := profiles[key]
		result.RadiusBrackets = append(result.RadiusBrackets, radiusBracket{
			Kind: profile.kind, HostileToOwner: profile.hostile,
			FarthestAffectedMilli:  profile.farthestAffected,
			NearestUnaffectedMilli: profile.nearestUnaffected,
			BoundaryBracketed:      profile.nearestUnaffected > profile.farthestAffected,
		})
	}
	return result
}

func targetProfile(sample targetSample) string {
	return fmt.Sprintf("%s:%t", sample.Kind, sample.HostileToOwner)
}

func distanceMilli(center, target point) int {
	dx := float64(center.XSubtiles - target.XSubtiles)
	dy := float64(center.YSubtiles - target.YSubtiles)
	// This is a neutral geometric observation, not a claim about the runtime's
	// internal range function. Keeping the coordinates lets a later policy
	// compare alternative 1.14d range predicates against the same capture.
	return int(math.Round(math.Hypot(dx, dy) * 1000))
}

func validChannels(value channels) bool {
	return value.Physical >= 0 && value.Fire >= 0 && value.Cold >= 0 &&
		value.Lightning >= 0 && value.Poison >= 0 && value.Magic >= 0
}

func channelTotal(value channels) int64 {
	return value.Physical + value.Fire + value.Cold + value.Lightning + value.Poison + value.Magic
}

func sortedKeys(values map[string]bool) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func validSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func oneOf(value string, candidates ...string) bool {
	for _, candidate := range candidates {
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
