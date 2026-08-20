package main

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
