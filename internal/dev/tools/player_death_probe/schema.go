package main

const (
	probeSchema = "d2legacy.player_death_probe/v1"
	probeTarget = "diablo-ii-lod-1.14d-expansion"
)

// capture mirrors the accepted evidence envelope; its field order also keeps generated fixtures and reports stable.
type capture struct {
	Schema  string      `json:"schema"`
	Target  string      `json:"target"`
	Source  string      `json:"source"`
	Runtime runtime     `json:"runtime"`
	Cases   []probeCase `json:"cases"`
}

// runtime records the provenance constraints that prevent imported or unsupported observations from being accepted.
type runtime struct {
	Patch            string `json:"patch"`
	Mode             string `json:"mode"`
	Session          string `json:"session"`
	CharacterMode    string `json:"character_mode"`
	CharacterOrigin  string `json:"character_origin"`
	ExecutableSHA256 string `json:"executable_sha256"`
	Observation      string `json:"observation"`
}

// probeCase groups one scenario's ordered observations so validation can reason about deaths independently.
type probeCase struct {
	ID           string        `json:"id"`
	Scenario     string        `json:"scenario"`
	Difficulty   string        `json:"difficulty"`
	Class        string        `json:"class"`
	Level        int           `json:"level"`
	KillerKind   string        `json:"killer_kind"`
	Observations []observation `json:"observations"`
}

// observation is a sanitized point in a death timeline; frame order establishes all temporal relationships.
type observation struct {
	Phase       string     `json:"phase"`
	DeathIndex  int        `json:"death_index"`
	Frame       int        `json:"frame"`
	Area        string     `json:"area"`
	Controlled  bool       `json:"controlled"`
	Health      int64      `json:"health"`
	MaxHealth   int64      `json:"max_health"`
	Experience  int64      `json:"experience"`
	CarriedGold int64      `json:"carried_gold"`
	StashedGold int64      `json:"stashed_gold"`
	GroundGold  int64      `json:"ground_gold"`
	CorpseCount int        `json:"corpse_count"`
	Equipment   []slotItem `json:"equipment"`
	Inventory   []slotItem `json:"inventory"`
}

// slotItem identifies sanitized item ownership without retaining any retail save representation.
type slotItem struct {
	Slot string `json:"slot"`
	ID   string `json:"id"`
}

// report preserves capture provenance while replacing raw observations with normalized consequences.
type report struct {
	Schema             string       `json:"schema"`
	Target             string       `json:"target"`
	ExecutableSHA256   string       `json:"executable_sha256"`
	CaptureFingerprint string       `json:"capture_fingerprint"`
	Cases              []caseReport `json:"cases"`
}

// caseReport summarizes consequences shared across all deaths in one scenario.
type caseReport struct {
	ID                     string        `json:"id"`
	Scenario               string        `json:"scenario"`
	Difficulty             string        `json:"difficulty"`
	Deaths                 []deathReport `json:"deaths"`
	StashedGoldUnchanged   bool          `json:"stashed_gold_unchanged"`
	SaveRejoinObserved     bool          `json:"save_rejoin_observed"`
	RejoinedCorpseCount    int           `json:"rejoined_corpse_count,omitempty"`
	RejoinedEquipmentCount int           `json:"rejoined_equipment_count,omitempty"`
	RejoinedInventoryCount int           `json:"rejoined_inventory_count,omitempty"`
}

// deathReport contains measured losses and recoveries for one validated death timeline.
type deathReport struct {
	DeathIndex                  int   `json:"death_index"`
	DeathAnimationFrames        int   `json:"death_animation_frames"`
	RespawnObserved             bool  `json:"respawn_observed"`
	RespawnInputToControlFrames int   `json:"respawn_input_to_control_frames,omitempty"`
	ExperienceLoss              int64 `json:"experience_loss,omitempty"`
	CarriedGoldLoss             int64 `json:"carried_gold_loss,omitempty"`
	GroundGoldAfterRespawn      int64 `json:"ground_gold_after_respawn,omitempty"`
	EquipmentRemovedAtRespawn   int   `json:"equipment_removed_at_respawn,omitempty"`
	InventoryRemovedAtRespawn   int   `json:"inventory_removed_at_respawn,omitempty"`
	CorpseCountAfterRespawn     int   `json:"corpse_count_after_respawn,omitempty"`
	RecoveryObserved            bool  `json:"recovery_observed"`
	ExperienceRecovered         int64 `json:"experience_recovered,omitempty"`
	EquipmentRestored           int   `json:"equipment_restored,omitempty"`
	InventoryRestored           int   `json:"inventory_restored,omitempty"`
}
