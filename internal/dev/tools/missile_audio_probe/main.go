// Command missile_audio_probe validates and normalizes sanitized audio/video
// observations from an owned Expansion 1.14d runtime. It resolves when and how
// often record-referenced missile sounds occur without importing older engine,
// server, save, memory-tool, or community behavior.
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
	"strings"
)

const (
	probeSchema = "d2legacy.missile_audio_probe/v1"
	probeTarget = "diablo-ii-lod-1.14d-expansion"
	maxFrame    = 1_000_000
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
	RecordGenerationID  string `json:"record_generation_id"`
	Observation         string `json:"observation"`
	SoundIdentification string `json:"sound_identification"`
	GameFramesPerSecond int    `json:"game_frames_per_second"`
	AudioIsolated       bool   `json:"audio_isolated"`
	CameraFixed         bool   `json:"camera_fixed"`
	ActorsStationary    bool   `json:"actors_stationary"`
}

type probeCase struct {
	ID                    string             `json:"id"`
	SkillID               int                `json:"skill_id"`
	SkillRecord           string             `json:"skill_record"`
	SkillLevel            int                `json:"skill_level"`
	MissileRecord         string             `json:"missile_record"`
	Difficulty            string             `json:"difficulty"`
	Area                  string             `json:"area"`
	Outcome               string             `json:"outcome"`
	TargetCount           int                `json:"target_count"`
	TargetRecords         []string           `json:"target_records"`
	ProjectileVisualCount int                `json:"projectile_visual_count"`
	CastEffectFrame       int                `json:"cast_effect_frame"`
	ContactFrame          *int               `json:"contact_frame,omitempty"`
	MissileRemovedFrame   int                `json:"missile_removed_frame"`
	Sounds                []soundObservation `json:"sounds"`
}

type soundObservation struct {
	Record    string          `json:"record"`
	Role      string          `json:"role"`
	Present   bool            `json:"present"`
	Intervals []frameInterval `json:"intervals"`
}

type frameInterval struct {
	FirstFrame int `json:"first_frame"`
	LastFrame  int `json:"last_frame"`
}

type report struct {
	Schema             string       `json:"schema"`
	Target             string       `json:"target"`
	ExecutableSHA256   string       `json:"executable_sha256"`
	RecordGenerationID string       `json:"record_generation_id"`
	CaptureFingerprint string       `json:"capture_fingerprint"`
	Coverage           coverage     `json:"coverage"`
	Cases              []caseReport `json:"cases"`
}

type coverage struct {
	Complete bool     `json:"complete"`
	Missing  []string `json:"missing,omitempty"`
}

type caseReport struct {
	ID                    string        `json:"id"`
	SkillID               int           `json:"skill_id"`
	SkillLevel            int           `json:"skill_level"`
	MissileRecord         string        `json:"missile_record"`
	Outcome               string        `json:"outcome"`
	TargetCount           int           `json:"target_count"`
	ProjectileVisualCount int           `json:"projectile_visual_count"`
	ContactFromEffect     *int          `json:"contact_from_effect_frames,omitempty"`
	LifetimeFrames        int           `json:"lifetime_frames"`
	Sounds                []soundReport `json:"sounds"`
}

type soundReport struct {
	Record     string           `json:"record"`
	Role       string           `json:"role"`
	RecordLoop bool             `json:"record_loop"`
	Present    bool             `json:"present"`
	Instances  int              `json:"instances"`
	Intervals  []intervalReport `json:"intervals"`
}

type intervalReport struct {
	FirstFromEffect  int  `json:"first_from_effect_frames"`
	LastFromEffect   int  `json:"last_from_effect_frames"`
	FirstFromContact *int `json:"first_from_contact_frames,omitempty"`
	LastFromRemoval  int  `json:"last_from_removal_frames"`
}

type soundSpec struct {
	record string
	role   string
	loop   bool
}

type caseSpec struct {
	id, skill, missile, outcome string
	skillID, targets            int
	sounds                      []soundSpec
}

var requiredCases = []caseSpec{
	{id: "fire-bolt-expire", skillID: 36, skill: "Fire Bolt", missile: "firebolt", outcome: "expired", targets: 0,
		sounds: []soundSpec{{"sorceress_firebolt_1", "travel", true}, {"sorceress_firebolt_impact_1", "hit", false}}},
	{id: "fire-bolt-hit", skillID: 36, skill: "Fire Bolt", missile: "firebolt", outcome: "hit", targets: 1,
		sounds: []soundSpec{{"sorceress_firebolt_1", "travel", true}, {"sorceress_firebolt_impact_1", "hit", false}}},
	{id: "fire-ball-hit", skillID: 47, skill: "Fire Ball", missile: "fireball", outcome: "hit", targets: 1,
		sounds: []soundSpec{{"sorceress_fireball_1", "travel", true}, {"sorceress_fireball_impact_1", "hit", false}}},
	{id: "nova-empty", skillID: 48, skill: "Nova", missile: "nova", outcome: "expired", targets: 0,
		sounds: []soundSpec{{"sorceress_nova", "travel", false}}},
	{id: "nova-three-targets", skillID: 48, skill: "Nova", missile: "nova", outcome: "multi-contact", targets: 3,
		sounds: []soundSpec{{"sorceress_nova", "travel", false}}},
	{id: "ice-blast-hit", skillID: 45, skill: "Ice Blast", missile: "iceblast", outcome: "hit", targets: 1,
		sounds: []soundSpec{{"sorceress_icebolt_1", "travel", true}, {"sorceress_iceblast_impact_1", "hit", false}}},
	{id: "glacial-spike-hit", skillID: 55, skill: "Glacial Spike", missile: "glacialspike", outcome: "hit", targets: 1,
		sounds: []soundSpec{{"sorceress_glacialspike_1", "travel", true}, {"sorceress_iceblast_impact_1", "hit", false}}},
}

func main() {
	input := flag.String("input", "", "sanitized owned-runtime missile-audio probe JSON")
	flag.Parse()
	if *input == "" {
		fmt.Fprintln(os.Stderr, "usage: missile_audio_probe -input <capture.json>")
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
		return report{}, fmt.Errorf("missile audio probe: read capture: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var captured capture
	if err := decoder.Decode(&captured); err != nil {
		return report{}, fmt.Errorf("missile audio probe: decode capture: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return report{}, fmt.Errorf("missile audio probe: capture must contain one JSON value")
	}
	if err := validate(captured); err != nil {
		return report{}, err
	}
	fingerprint := sha256.Sum256(data)
	result := report{
		Schema: probeSchema + ".report", Target: probeTarget,
		ExecutableSHA256: captured.Runtime.ExecutableSHA256, RecordGenerationID: captured.Runtime.RecordGenerationID,
		CaptureFingerprint: hex.EncodeToString(fingerprint[:]), Coverage: coverageFor(captured.Cases),
	}
	for _, observed := range captured.Cases {
		result.Cases = append(result.Cases, normalize(observed))
	}
	sort.Slice(result.Cases, func(i, j int) bool { return result.Cases[i].ID < result.Cases[j].ID })
	return result, nil
}

func validate(captured capture) error {
	var result error
	if captured.Schema != probeSchema {
		result = errors.Join(result, fmt.Errorf("missile audio probe: schema %q, want %q", captured.Schema, probeSchema))
	}
	if captured.Target != probeTarget {
		result = errors.Join(result, fmt.Errorf("missile audio probe: target %q, want %q", captured.Target, probeTarget))
	}
	if captured.Source != "owned-runtime" {
		result = errors.Join(result, fmt.Errorf("missile audio probe: source must be owned-runtime"))
	}
	runtime := captured.Runtime
	if runtime.Patch != "1.14d" || runtime.Mode != "expansion" || runtime.Session != "single-player" ||
		runtime.CharacterOrigin != "probe-created" {
		result = errors.Join(result, fmt.Errorf("missile audio probe: runtime must be a probe-created Expansion 1.14d single-player character"))
	}
	if !validSHA256(runtime.ExecutableSHA256) || !validGenerationID(runtime.RecordGenerationID) {
		result = errors.Join(result, fmt.Errorf("missile audio probe: executable and immutable record-generation SHA-256 values are required"))
	}
	if runtime.Observation != "isolated-audio-video-frame-log" ||
		runtime.SoundIdentification != "owned-mpq-waveform-comparison" || runtime.GameFramesPerSecond != 25 ||
		!runtime.AudioIsolated || !runtime.CameraFixed || !runtime.ActorsStationary {
		result = errors.Join(result, fmt.Errorf("missile audio probe: requires isolated 25 Hz audio/video observation with owned-MPQ waveform comparison, a fixed camera, and stationary actors"))
	}
	if len(captured.Cases) == 0 {
		result = errors.Join(result, fmt.Errorf("missile audio probe: at least one case is required"))
	}
	seen := make(map[string]bool, len(captured.Cases))
	for _, observed := range captured.Cases {
		if observed.ID == "" || seen[observed.ID] {
			result = errors.Join(result, fmt.Errorf("missile audio probe: case IDs must be non-empty and unique"))
		}
		seen[observed.ID] = true
		result = errors.Join(result, validateCase(observed))
	}
	return result
}

func validateCase(observed probeCase) error {
	spec, found := specFor(observed.ID)
	if !found || observed.SkillID != spec.skillID || observed.SkillRecord != spec.skill || observed.SkillLevel != 1 ||
		observed.MissileRecord != spec.missile || observed.Outcome != spec.outcome || observed.TargetCount != spec.targets {
		return fmt.Errorf("missile audio probe: case %q does not match its target-locked matrix row", observed.ID)
	}
	var result error
	if observed.Difficulty != "normal" || observed.Area != "blood_moor" || len(observed.TargetRecords) != observed.TargetCount {
		result = errors.Join(result, fmt.Errorf("missile audio probe: case %q must use the controlled Normal Blood Moor target set", observed.ID))
	}
	for _, target := range observed.TargetRecords {
		if target != "fallen1" {
			result = errors.Join(result, fmt.Errorf("missile audio probe: case %q target record must be fallen1", observed.ID))
		}
	}
	if !validFrame(observed.CastEffectFrame) || !validFrame(observed.MissileRemovedFrame) ||
		observed.MissileRemovedFrame < observed.CastEffectFrame || observed.ProjectileVisualCount <= 0 {
		result = errors.Join(result, fmt.Errorf("missile audio probe: case %q has an invalid visual timeline", observed.ID))
	}
	if spec.outcome == "expired" {
		if observed.ContactFrame != nil {
			result = errors.Join(result, fmt.Errorf("missile audio probe: expired case %q cannot have contact", observed.ID))
		}
	} else if observed.ContactFrame == nil || !validFrame(*observed.ContactFrame) ||
		*observed.ContactFrame < observed.CastEffectFrame || *observed.ContactFrame > observed.MissileRemovedFrame {
		result = errors.Join(result, fmt.Errorf("missile audio probe: contact case %q requires an in-lifetime contact frame", observed.ID))
	}
	byRecord := make(map[string]soundObservation, len(observed.Sounds))
	for _, sound := range observed.Sounds {
		if _, duplicate := byRecord[sound.Record]; duplicate {
			result = errors.Join(result, fmt.Errorf("missile audio probe: case %q duplicates sound %q", observed.ID, sound.Record))
		}
		byRecord[sound.Record] = sound
	}
	for _, expected := range spec.sounds {
		sound, exists := byRecord[expected.record]
		if !exists || sound.Role != expected.role {
			result = errors.Join(result, fmt.Errorf("missile audio probe: case %q omits or mislabels %s sound %q", observed.ID, expected.role, expected.record))
			continue
		}
		result = errors.Join(result, validateSound(observed.ID, observed.CastEffectFrame, sound))
		delete(byRecord, expected.record)
	}
	for record := range byRecord {
		result = errors.Join(result, fmt.Errorf("missile audio probe: case %q invents sound record %q", observed.ID, record))
	}
	return result
}

func validateSound(caseID string, castEffectFrame int, sound soundObservation) error {
	if sound.Present != (len(sound.Intervals) > 0) {
		return fmt.Errorf("missile audio probe: case %q sound %q contradicts presence", caseID, sound.Record)
	}
	previous := -1
	for index, interval := range sound.Intervals {
		if !validFrame(interval.FirstFrame) || !validFrame(interval.LastFrame) || interval.FirstFrame < castEffectFrame || interval.LastFrame < interval.FirstFrame ||
			interval.FirstFrame <= previous {
			return fmt.Errorf("missile audio probe: case %q sound %q interval %d is invalid or overlaps", caseID, sound.Record, index)
		}
		previous = interval.LastFrame
	}
	return nil
}

func normalize(observed probeCase) caseReport {
	spec, _ := specFor(observed.ID)
	result := caseReport{
		ID: observed.ID, SkillID: observed.SkillID, SkillLevel: observed.SkillLevel,
		MissileRecord: observed.MissileRecord, Outcome: observed.Outcome,
		TargetCount: observed.TargetCount, ProjectileVisualCount: observed.ProjectileVisualCount,
		LifetimeFrames: observed.MissileRemovedFrame - observed.CastEffectFrame,
	}
	if observed.ContactFrame != nil {
		value := *observed.ContactFrame - observed.CastEffectFrame
		result.ContactFromEffect = &value
	}
	for _, sound := range observed.Sounds {
		expected, _ := soundSpecFor(spec, sound.Record)
		normalized := soundReport{Record: sound.Record, Role: sound.Role, RecordLoop: expected.loop, Present: sound.Present, Instances: len(sound.Intervals)}
		for _, interval := range sound.Intervals {
			item := intervalReport{
				FirstFromEffect: interval.FirstFrame - observed.CastEffectFrame,
				LastFromEffect:  interval.LastFrame - observed.CastEffectFrame,
				LastFromRemoval: interval.LastFrame - observed.MissileRemovedFrame,
			}
			if observed.ContactFrame != nil {
				value := interval.FirstFrame - *observed.ContactFrame
				item.FirstFromContact = &value
			}
			normalized.Intervals = append(normalized.Intervals, item)
		}
		result.Sounds = append(result.Sounds, normalized)
	}
	sort.Slice(result.Sounds, func(i, j int) bool { return result.Sounds[i].Record < result.Sounds[j].Record })
	return result
}

func coverageFor(cases []probeCase) coverage {
	seen := make(map[string]bool, len(cases))
	for _, observed := range cases {
		seen[observed.ID] = true
	}
	result := coverage{Complete: true}
	for _, required := range requiredCases {
		if !seen[required.id] {
			result.Complete = false
			result.Missing = append(result.Missing, required.id)
		}
	}
	return result
}

func specFor(id string) (caseSpec, bool) {
	for _, spec := range requiredCases {
		if spec.id == id {
			return spec, true
		}
	}
	return caseSpec{}, false
}

func soundSpecFor(spec caseSpec, record string) (soundSpec, bool) {
	for _, sound := range spec.sounds {
		if sound.record == record {
			return sound, true
		}
	}
	return soundSpec{}, false
}

func validFrame(value int) bool { return value >= 0 && value <= maxFrame }

func validSHA256(value string) bool {
	if len(value) != 64 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func validGenerationID(value string) bool {
	return strings.HasPrefix(value, "sha256:") && validSHA256(strings.TrimPrefix(value, "sha256:"))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
