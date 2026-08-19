package main

// capture is the complete owned-runtime evidence envelope, keeping provenance and ordered cases in one trust boundary.
type capture struct {
	Schema  string      `json:"schema"`
	Target  string      `json:"target"`
	Source  string      `json:"source"`
	Runtime runtime     `json:"runtime"`
	Cases   []probeCase `json:"cases"`
}

// runtime identifies the exact executable and observation method so evidence cannot drift across runtime variants.
type runtime struct {
	Patch            string `json:"patch"`
	Mode             string `json:"mode"`
	Session          string `json:"session"`
	ExecutableSHA256 string `json:"executable_sha256"`
	Observation      string `json:"observation"`
}

// probeCase holds one defense mechanism scenario and the trials observed under its fixed combat context.
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

// defender captures every defender fact that must match before a control comparison is meaningful.
type defender struct {
	Kind    string `json:"kind"`
	Record  string `json:"record"`
	Level   int    `json:"level"`
	Defense int    `json:"defense"`
	State   string `json:"state"`
}

// trial records a single visual outcome with enough health and reaction evidence to reject contradictions.
type trial struct {
	Outcome         string `json:"outcome"`
	Reaction        string `json:"reaction"`
	HealthBeforeRaw int64  `json:"health_before_raw"`
	HealthAfterRaw  int64  `json:"health_after_raw"`
}

// report is the normalized evidence artifact, retaining both executable and byte-exact capture provenance.
type report struct {
	Schema             string       `json:"schema"`
	Target             string       `json:"target"`
	ExecutableSHA256   string       `json:"executable_sha256"`
	CaptureFingerprint string       `json:"capture_fingerprint"`
	Cases              []caseReport `json:"cases"`
}

// caseReport summarizes one scenario without discarding its counts, damage, rates, or control relationship.
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

// interval reports the observed proportion beside its 95-percent Wilson confidence bounds.
type interval struct {
	Observed float64 `json:"observed"`
	Low95    float64 `json:"low_95"`
	High95   float64 `json:"high_95"`
}
