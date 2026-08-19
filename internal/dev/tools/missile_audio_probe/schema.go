package main

const (
	probeSchema = "d2legacy.missile_audio_probe/v1"
	probeTarget = "diablo-ii-lod-1.14d-expansion"
	maxFrame    = 1_000_000
)

// capture is the target-locked input document. Its field order also fixes the conventional encoded form used by
// fixtures and downstream tooling.
type capture struct {
	Schema  string      `json:"schema"`
	Target  string      `json:"target"`
	Source  string      `json:"source"`
	Runtime runtime     `json:"runtime"`
	Cases   []probeCase `json:"cases"`
}

// runtime records provenance and observation controls that distinguish owned-runtime evidence from inferred data.
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

// probeCase captures one matrix row and the visual timeline against which its sounds are interpreted.
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

// soundObservation associates a record and semantic role with every interval where the sound was identified.
type soundObservation struct {
	Record    string          `json:"record"`
	Role      string          `json:"role"`
	Present   bool            `json:"present"`
	Intervals []frameInterval `json:"intervals"`
}

// frameInterval uses inclusive capture frames so overlap and lifetime validation have unambiguous boundaries.
type frameInterval struct {
	FirstFrame int `json:"first_frame"`
	LastFrame  int `json:"last_frame"`
}

// report is the deterministic normalized document emitted after the complete capture validates.
type report struct {
	Schema             string       `json:"schema"`
	Target             string       `json:"target"`
	ExecutableSHA256   string       `json:"executable_sha256"`
	RecordGenerationID string       `json:"record_generation_id"`
	CaptureFingerprint string       `json:"capture_fingerprint"`
	Coverage           coverage     `json:"coverage"`
	Cases              []caseReport `json:"cases"`
}

// coverage identifies absent matrix rows without treating partial evidence as complete.
type coverage struct {
	Complete bool     `json:"complete"`
	Missing  []string `json:"missing,omitempty"`
}

// caseReport expresses case timing relative to stable lifecycle events rather than capture-global frame numbers.
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

// soundReport preserves the observed record while adding matrix metadata and normalized interval timing.
type soundReport struct {
	Record     string           `json:"record"`
	Role       string           `json:"role"`
	RecordLoop bool             `json:"record_loop"`
	Present    bool             `json:"present"`
	Instances  int              `json:"instances"`
	Intervals  []intervalReport `json:"intervals"`
}

// intervalReport relates a sound instance to cast, contact, and removal events for direct comparison across cases.
type intervalReport struct {
	FirstFromEffect  int  `json:"first_from_effect_frames"`
	LastFromEffect   int  `json:"last_from_effect_frames"`
	FirstFromContact *int `json:"first_from_contact_frames,omitempty"`
	LastFromRemoval  int  `json:"last_from_removal_frames"`
}
