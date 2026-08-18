// Command cast_rate_probe validates and normalizes sanitized cast-timing
// observations from an owned Expansion 1.14d runtime. The report separates
// target evidence from candidate formulas, so older breakpoint tables cannot
// silently become Dark Magic behavior.
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
	probeSchema = "d2legacy.cast_rate_probe/v1"
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
	StatIdentification  string `json:"stat_identification"`
	Locale              string `json:"locale"`
	GameFramesPerSecond int    `json:"game_frames_per_second"`
}

type modifierSource struct {
	ItemRecord   string `json:"item_record"`
	PropertyCode string `json:"property_code"`
	Value        int    `json:"value"`
}

type probeCase struct {
	ID                 string           `json:"id"`
	SkillID            int              `json:"skill_id"`
	SkillRecord        string           `json:"skill_record"`
	CharacterClass     string           `json:"character_class"`
	AnimationMode      string           `json:"animation_mode"`
	SequenceTransition string           `json:"sequence_transition"`
	SequenceNumber     *int             `json:"sequence_number,omitempty"`
	WeaponClass        string           `json:"weapon_class"`
	RawFasterCastRate  int              `json:"raw_faster_cast_rate"`
	ModifierKey        string           `json:"modifier_key"`
	ModifierText       string           `json:"modifier_text"`
	ModifierSources    []modifierSource `json:"modifier_sources"`
	StartFrame         int              `json:"start_frame"`
	EffectFrame        int              `json:"effect_frame"`
	NeutralFrame       int              `json:"neutral_frame"`
}

type report struct {
	Schema             string       `json:"schema"`
	Target             string       `json:"target"`
	ExecutableSHA256   string       `json:"executable_sha256"`
	RecordGenerationID string       `json:"record_generation_id"`
	CaptureFingerprint string       `json:"capture_fingerprint"`
	Coverage           coverage     `json:"coverage"`
	Profiles           []caseReport `json:"profiles"`
}

type coverage struct {
	Complete bool     `json:"complete"`
	Missing  []string `json:"missing,omitempty"`
}

type caseReport struct {
	ID                string `json:"id"`
	SkillID           int    `json:"skill_id"`
	AnimationMode     string `json:"animation_mode"`
	SequenceNumber    *int   `json:"sequence_number,omitempty"`
	WeaponClass       string `json:"weapon_class"`
	RawFasterCastRate int    `json:"raw_faster_cast_rate"`
	EffectDelay       int    `json:"effect_delay_game_frames"`
	CompletionDelay   int    `json:"completion_delay_game_frames"`
}

func main() {
	input := flag.String("input", "", "sanitized owned-runtime cast-rate probe JSON")
	flag.Parse()
	if *input == "" {
		fmt.Fprintln(os.Stderr, "usage: cast_rate_probe -input <capture.json>")
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
		return report{}, fmt.Errorf("cast-rate probe: read capture: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var captured capture
	if err := decoder.Decode(&captured); err != nil {
		return report{}, fmt.Errorf("cast-rate probe: decode capture: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return report{}, fmt.Errorf("cast-rate probe: capture must contain one JSON value")
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
		result.Profiles = append(result.Profiles, caseReport{
			ID: observed.ID, SkillID: observed.SkillID, AnimationMode: observed.AnimationMode,
			SequenceNumber: observed.SequenceNumber, WeaponClass: observed.WeaponClass,
			RawFasterCastRate: observed.RawFasterCastRate,
			EffectDelay:       observed.EffectFrame - observed.StartFrame,
			CompletionDelay:   observed.NeutralFrame - observed.StartFrame,
		})
	}
	sort.Slice(result.Profiles, func(i, j int) bool {
		left, right := result.Profiles[i], result.Profiles[j]
		if left.SkillID != right.SkillID {
			return left.SkillID < right.SkillID
		}
		if left.WeaponClass != right.WeaponClass {
			return left.WeaponClass < right.WeaponClass
		}
		if left.RawFasterCastRate != right.RawFasterCastRate {
			return left.RawFasterCastRate < right.RawFasterCastRate
		}
		return left.ID < right.ID
	})
	return result, nil
}

func validate(captured capture) error {
	var result error
	if captured.Schema != probeSchema {
		result = errors.Join(result, fmt.Errorf("cast-rate probe: schema %q, want %q", captured.Schema, probeSchema))
	}
	if captured.Target != probeTarget {
		result = errors.Join(result, fmt.Errorf("cast-rate probe: target %q, want %q", captured.Target, probeTarget))
	}
	if captured.Source != "owned-runtime" {
		result = errors.Join(result, fmt.Errorf("cast-rate probe: source must be owned-runtime"))
	}
	runtime := captured.Runtime
	if runtime.Patch != "1.14d" || runtime.Mode != "expansion" || runtime.Session != "single-player" {
		result = errors.Join(result, fmt.Errorf("cast-rate probe: runtime must be Expansion 1.14d single-player"))
	}
	if runtime.CharacterOrigin != "probe-created" {
		result = errors.Join(result, fmt.Errorf("cast-rate probe: character must be probe-created"))
	}
	if !validSHA256(runtime.ExecutableSHA256) || !validGenerationID(runtime.RecordGenerationID) {
		result = errors.Join(result, fmt.Errorf("cast-rate probe: executable and owned-record SHA-256 identities are required"))
	}
	if runtime.Observation != "video-frame-log" || runtime.StatIdentification != "owned-itemstatcost-properties-tbl" ||
		runtime.Locale != "eng" || runtime.GameFramesPerSecond != 25 {
		result = errors.Join(result, fmt.Errorf("cast-rate probe: requires a 25 Hz visual log and owned ItemStatCost/Properties/TBL identification"))
	}
	if len(captured.Cases) == 0 {
		result = errors.Join(result, fmt.Errorf("cast-rate probe: at least one case is required"))
	}
	seenIDs, seenProfiles := map[string]bool{}, map[string]bool{}
	for _, observed := range captured.Cases {
		if observed.ID == "" || seenIDs[observed.ID] {
			result = errors.Join(result, fmt.Errorf("cast-rate probe: case IDs must be non-empty and unique"))
		}
		seenIDs[observed.ID] = true
		profile := fmt.Sprintf("%d/%s/%d", observed.SkillID, observed.WeaponClass, observed.RawFasterCastRate)
		if seenProfiles[profile] {
			result = errors.Join(result, fmt.Errorf("cast-rate probe: duplicate profile %s", profile))
		}
		seenProfiles[profile] = true
		result = errors.Join(result, validateCase(observed))
	}
	return result
}

func validateCase(observed probeCase) error {
	var result error
	wantRecord, wantMode, wantSequence, known := expectedSkill(observed.SkillID)
	if !known || observed.SkillRecord != wantRecord || observed.CharacterClass != "sor" ||
		observed.AnimationMode != wantMode || observed.SequenceTransition != "SC" {
		result = errors.Join(result, fmt.Errorf("cast-rate probe: case %q has invalid owned skill context", observed.ID))
	}
	if wantSequence == nil {
		if observed.SequenceNumber != nil {
			result = errors.Join(result, fmt.Errorf("cast-rate probe: case %q assigns a sequence to an SC cast", observed.ID))
		}
	} else if observed.SequenceNumber == nil || *observed.SequenceNumber != *wantSequence {
		result = errors.Join(result, fmt.Errorf("cast-rate probe: case %q must use sequence %d", observed.ID, *wantSequence))
	}
	if !oneOf(observed.WeaponClass, "HTH", "1HS", "STF") || observed.RawFasterCastRate < 0 || observed.RawFasterCastRate > 200 {
		result = errors.Join(result, fmt.Errorf("cast-rate probe: case %q has an invalid weapon/FCR profile", observed.ID))
	}
	if observed.ModifierKey != "ModStr4v" || observed.ModifierText != "Faster Cast Rate" {
		result = errors.Join(result, fmt.Errorf("cast-rate probe: case %q does not preserve the owned ModStr4v text", observed.ID))
	}
	sourceTotal := 0
	for index, source := range observed.ModifierSources {
		if source.ItemRecord == "" || !oneOf(source.PropertyCode, "cast1", "cast2", "cast3") || source.Value <= 0 {
			result = errors.Join(result, fmt.Errorf("cast-rate probe: case %q modifier source %d is invalid", observed.ID, index))
		}
		sourceTotal += source.Value
	}
	if sourceTotal != observed.RawFasterCastRate {
		result = errors.Join(result, fmt.Errorf("cast-rate probe: case %q modifier sources total %d, want %d", observed.ID, sourceTotal, observed.RawFasterCastRate))
	}
	if observed.StartFrame < 0 || observed.EffectFrame <= observed.StartFrame || observed.NeutralFrame <= observed.EffectFrame || observed.NeutralFrame > maxFrame {
		result = errors.Join(result, fmt.Errorf("cast-rate probe: case %q has invalid visual action boundaries", observed.ID))
	}
	return result
}

func coverageFor(cases []probeCase) coverage {
	required := map[string]bool{}
	// Paired values on both sides of each candidate transition make the
	// observation useful even when the target disproves that transition. This
	// list declares required measurements only; it encodes no expected delay.
	for _, rate := range []int{0, 8, 9, 19, 20, 36, 37, 62, 63, 104, 105, 199, 200} {
		required[fmt.Sprintf("sc-hth-fcr-%d", rate)] = false
	}
	for _, weapon := range []string{"1HS", "STF"} {
		for _, rate := range []int{0, 105} {
			required[fmt.Sprintf("sc-%s-fcr-%d", strings.ToLower(weapon), rate)] = false
		}
	}
	for _, rate := range []int{0, 105} {
		required[fmt.Sprintf("sq-hth-fcr-%d", rate)] = false
	}
	for _, observed := range cases {
		prefix := "sc"
		if observed.SkillID == 49 {
			prefix = "sq"
		}
		key := fmt.Sprintf("%s-%s-fcr-%d", prefix, strings.ToLower(observed.WeaponClass), observed.RawFasterCastRate)
		if _, exists := required[key]; exists {
			required[key] = true
		}
	}
	result := coverage{Complete: true}
	for key, present := range required {
		if !present {
			result.Complete = false
			result.Missing = append(result.Missing, key)
		}
	}
	sort.Strings(result.Missing)
	return result
}

func expectedSkill(id int) (string, string, *int, bool) {
	switch id {
	case 36:
		return "Fire Bolt", "SC", nil, true
	case 49:
		sequence := 12
		return "Lightning", "SQ", &sequence, true
	default:
		return "", "", nil, false
	}
}

func validSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return false
		}
	}
	return true
}

func validGenerationID(value string) bool {
	return strings.HasPrefix(value, "sha256:") && validSHA256(strings.TrimPrefix(value, "sha256:"))
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
