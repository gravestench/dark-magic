// Command player_death_probe validates and normalizes sanitized visual
// observations from a probe-created character in an owned Expansion 1.14d
// runtime. It records death/corpse consequence evidence; it never reads or
// writes retail save data and does not implement inferred gameplay policy.
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
	probeSchema = "d2legacy.player_death_probe/v1"
	probeTarget = "diablo-ii-lod-1.14d-expansion"
)

type capture struct {
	Schema  string      `json:"schema"`
	Target  string      `json:"target"`
	Source  string      `json:"source"`
	Runtime runtime     `json:"runtime"`
	Cases   []probeCase `json:"cases"`
}

type runtime struct {
	Patch            string `json:"patch"`
	Mode             string `json:"mode"`
	Session          string `json:"session"`
	CharacterMode    string `json:"character_mode"`
	CharacterOrigin  string `json:"character_origin"`
	ExecutableSHA256 string `json:"executable_sha256"`
	Observation      string `json:"observation"`
}

type probeCase struct {
	ID           string        `json:"id"`
	Scenario     string        `json:"scenario"`
	Difficulty   string        `json:"difficulty"`
	Class        string        `json:"class"`
	Level        int           `json:"level"`
	KillerKind   string        `json:"killer_kind"`
	Observations []observation `json:"observations"`
}

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

type slotItem struct {
	Slot string `json:"slot"`
	ID   string `json:"id"`
}

type report struct {
	Schema             string       `json:"schema"`
	Target             string       `json:"target"`
	ExecutableSHA256   string       `json:"executable_sha256"`
	CaptureFingerprint string       `json:"capture_fingerprint"`
	Cases              []caseReport `json:"cases"`
}

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

func main() {
	input := flag.String("input", "", "sanitized owned-runtime player-death probe JSON")
	flag.Parse()
	if *input == "" {
		fmt.Fprintln(os.Stderr, "usage: player_death_probe -input <capture.json>")
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
		return report{}, fmt.Errorf("player death probe: read capture: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var captured capture
	if err := decoder.Decode(&captured); err != nil {
		return report{}, fmt.Errorf("player death probe: decode capture: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return report{}, fmt.Errorf("player death probe: capture must contain one JSON value")
	}
	if err := validate(captured); err != nil {
		return report{}, err
	}
	fingerprint := sha256.Sum256(data)
	result := report{
		Schema:             probeSchema + ".report",
		Target:             probeTarget,
		ExecutableSHA256:   captured.Runtime.ExecutableSHA256,
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
		result = errors.Join(result, fmt.Errorf("player death probe: schema %q, want %q", captured.Schema, probeSchema))
	}
	if captured.Target != probeTarget {
		result = errors.Join(result, fmt.Errorf("player death probe: target %q, want %q", captured.Target, probeTarget))
	}
	if captured.Source != "owned-runtime" {
		result = errors.Join(result, fmt.Errorf("player death probe: source must be owned-runtime"))
	}
	if captured.Runtime.Patch != "1.14d" || captured.Runtime.Mode != "expansion" ||
		captured.Runtime.Session != "single-player" || captured.Runtime.CharacterMode != "softcore" {
		result = errors.Join(result, fmt.Errorf("player death probe: runtime must be softcore Expansion 1.14d single-player"))
	}
	if captured.Runtime.CharacterOrigin != "probe-created" {
		result = errors.Join(result, fmt.Errorf("player death probe: character must be created for the probe, not imported save data"))
	}
	if !validSHA256(captured.Runtime.ExecutableSHA256) {
		result = errors.Join(result, fmt.Errorf("player death probe: executable SHA-256 is required"))
	}
	if !oneOf(captured.Runtime.Observation, "video-frame-log", "manual-frame-log") {
		result = errors.Join(result, fmt.Errorf("player death probe: unsupported observation method"))
	}
	if len(captured.Cases) == 0 {
		result = errors.Join(result, fmt.Errorf("player death probe: at least one case is required"))
	}
	seen := make(map[string]bool)
	for _, observed := range captured.Cases {
		if observed.ID == "" || seen[observed.ID] {
			result = errors.Join(result, fmt.Errorf("player death probe: case IDs must be non-empty and unique"))
		}
		seen[observed.ID] = true
		result = errors.Join(result, validateCase(observed))
	}
	return result
}

func validateCase(observed probeCase) error {
	var result error
	if !oneOf(observed.Scenario, "single_recovery", "single_no_recovery", "multiple_corpses", "save_exit_dead", "save_exit_respawned") ||
		!oneOf(observed.Difficulty, "normal", "nightmare", "hell") ||
		!oneOf(observed.Class, "amazon", "sorceress", "necromancer", "paladin", "barbarian", "druid", "assassin") ||
		observed.Level < 1 || observed.KillerKind != "monster" {
		return fmt.Errorf("player death probe: case %q has invalid target context", observed.ID)
	}
	if len(observed.Observations) == 0 {
		return fmt.Errorf("player death probe: case %q requires observations", observed.ID)
	}
	byDeath := make(map[int]map[string]observation)
	lastFrame := -1
	stash := observed.Observations[0].StashedGold
	for index, current := range observed.Observations {
		if current.Frame <= lastFrame || current.Area == "" || current.DeathIndex < 1 ||
			current.Health < 0 || current.MaxHealth < 1 || current.Health > current.MaxHealth ||
			current.Experience < 0 || current.CarriedGold < 0 || current.StashedGold < 0 ||
			current.GroundGold < 0 || current.CorpseCount < 0 ||
			!oneOf(current.Phase, "before_death", "death_started", "death_animation_complete", "respawn_input", "town_control", "corpse_recovered", "save_exit", "rejoined") {
			result = errors.Join(result, fmt.Errorf("player death probe: case %q observation %d is invalid", observed.ID, index))
		}
		lastFrame = current.Frame
		if current.StashedGold != stash {
			result = errors.Join(result, fmt.Errorf("player death probe: case %q changes stash gold between observations", observed.ID))
		}
		if err := validateItems(current.Equipment); err != nil {
			result = errors.Join(result, fmt.Errorf("player death probe: case %q observation %d equipment: %w", observed.ID, index, err))
		}
		if err := validateItems(current.Inventory); err != nil {
			result = errors.Join(result, fmt.Errorf("player death probe: case %q observation %d inventory: %w", observed.ID, index, err))
		}
		knownControlPhase := oneOf(current.Phase, "before_death", "death_started", "death_animation_complete", "respawn_input", "town_control", "corpse_recovered")
		shouldBeControlled := oneOf(current.Phase, "before_death", "town_control", "corpse_recovered")
		if knownControlPhase && current.Controlled != shouldBeControlled ||
			oneOf(current.Phase, "death_started", "death_animation_complete", "respawn_input") && current.Health != 0 ||
			current.Phase == "town_control" && current.Health == 0 {
			result = errors.Join(result, fmt.Errorf("player death probe: case %q observation %d contradicts control/health state", observed.ID, index))
		}
		phases := byDeath[current.DeathIndex]
		if phases == nil {
			phases = make(map[string]observation)
			byDeath[current.DeathIndex] = phases
		}
		if _, exists := phases[current.Phase]; exists {
			result = errors.Join(result, fmt.Errorf("player death probe: case %q repeats phase %q for death %d", observed.ID, current.Phase, current.DeathIndex))
		}
		phases[current.Phase] = current
	}
	deathCount := len(byDeath)
	for deathIndex := 1; deathIndex <= deathCount; deathIndex++ {
		phases, exists := byDeath[deathIndex]
		if !exists {
			result = errors.Join(result, fmt.Errorf("player death probe: case %q skips death index %d", observed.ID, deathIndex))
			continue
		}
		for _, required := range []string{"before_death", "death_started", "death_animation_complete"} {
			if _, exists := phases[required]; !exists {
				result = errors.Join(result, fmt.Errorf("player death probe: case %q death %d lacks %s", observed.ID, deathIndex, required))
			}
		}
		if before, ok := phases["before_death"]; ok {
			if started, ok := phases["death_started"]; ok && started.Frame <= before.Frame {
				result = errors.Join(result, fmt.Errorf("player death probe: case %q death %d starts before its baseline", observed.ID, deathIndex))
			}
		}
		if started, ok := phases["death_started"]; ok {
			if complete, ok := phases["death_animation_complete"]; ok && complete.Frame <= started.Frame {
				result = errors.Join(result, fmt.Errorf("player death probe: case %q death %d completes before it starts", observed.ID, deathIndex))
			}
		}
		if observed.Scenario != "save_exit_dead" {
			for _, required := range []string{"respawn_input", "town_control"} {
				if _, exists := phases[required]; !exists {
					result = errors.Join(result, fmt.Errorf("player death probe: case %q death %d lacks %s", observed.ID, deathIndex, required))
				}
			}
			if input, ok := phases["respawn_input"]; ok {
				if complete, ok := phases["death_animation_complete"]; ok && input.Frame <= complete.Frame {
					result = errors.Join(result, fmt.Errorf("player death probe: case %q death %d respawn input precedes death completion", observed.ID, deathIndex))
				}
				if town, ok := phases["town_control"]; ok && town.Frame <= input.Frame {
					result = errors.Join(result, fmt.Errorf("player death probe: case %q death %d regains control before respawn input", observed.ID, deathIndex))
				}
			}
			if town, ok := phases["town_control"]; ok {
				if recovered, ok := phases["corpse_recovered"]; ok && recovered.Frame <= town.Frame {
					result = errors.Join(result, fmt.Errorf("player death probe: case %q death %d recovers a corpse before town control", observed.ID, deathIndex))
				}
			}
		}
		if saveExit, ok := phases["save_exit"]; ok {
			if rejoined, ok := phases["rejoined"]; ok && rejoined.Frame <= saveExit.Frame {
				result = errors.Join(result, fmt.Errorf("player death probe: case %q death %d rejoins before save/exit", observed.ID, deathIndex))
			}
		}
	}
	if observed.Scenario == "single_recovery" && (deathCount != 1 || !hasPhase(byDeath, "corpse_recovered")) {
		result = errors.Join(result, fmt.Errorf("player death probe: single_recovery requires one recovered death"))
	}
	if observed.Scenario == "single_no_recovery" && deathCount != 1 {
		result = errors.Join(result, fmt.Errorf("player death probe: single_no_recovery requires one death"))
	}
	if observed.Scenario == "single_no_recovery" && hasPhase(byDeath, "corpse_recovered") {
		result = errors.Join(result, fmt.Errorf("player death probe: single_no_recovery cannot include corpse recovery"))
	}
	if observed.Scenario == "multiple_corpses" && deathCount < 2 {
		result = errors.Join(result, fmt.Errorf("player death probe: multiple_corpses requires at least two deaths"))
	}
	if oneOf(observed.Scenario, "save_exit_dead", "save_exit_respawned") &&
		(!hasPhase(byDeath, "save_exit") || !hasPhase(byDeath, "rejoined")) {
		result = errors.Join(result, fmt.Errorf("player death probe: save/exit scenario requires save_exit and rejoined observations"))
	}
	return result
}

func validateItems(items []slotItem) error {
	seenSlots := make(map[string]bool)
	seenIDs := make(map[string]bool)
	for _, item := range items {
		if item.Slot == "" || item.ID == "" || seenSlots[item.Slot] || seenIDs[item.ID] {
			return fmt.Errorf("slot and sanitized item ID must be non-empty and unique")
		}
		seenSlots[item.Slot] = true
		seenIDs[item.ID] = true
	}
	return nil
}

func normalize(observed probeCase) caseReport {
	result := caseReport{ID: observed.ID, Scenario: observed.Scenario, Difficulty: observed.Difficulty, StashedGoldUnchanged: true}
	byDeath := make(map[int]map[string]observation)
	for _, current := range observed.Observations {
		if byDeath[current.DeathIndex] == nil {
			byDeath[current.DeathIndex] = make(map[string]observation)
		}
		byDeath[current.DeathIndex][current.Phase] = current
		if current.Phase == "rejoined" {
			result.SaveRejoinObserved = true
			result.RejoinedCorpseCount = current.CorpseCount
			result.RejoinedEquipmentCount = len(current.Equipment)
			result.RejoinedInventoryCount = len(current.Inventory)
		}
	}
	indexes := make([]int, 0, len(byDeath))
	for index := range byDeath {
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)
	for _, index := range indexes {
		phases := byDeath[index]
		before, started := phases["before_death"], phases["death_started"]
		complete := phases["death_animation_complete"]
		report := deathReport{
			DeathIndex: index, DeathAnimationFrames: complete.Frame - started.Frame,
		}
		input, hasInput := phases["respawn_input"]
		town, hasTown := phases["town_control"]
		if hasInput && hasTown {
			report.RespawnObserved = true
			report.RespawnInputToControlFrames = town.Frame - input.Frame
			report.ExperienceLoss = max64(0, before.Experience-town.Experience)
			report.CarriedGoldLoss = max64(0, before.CarriedGold-town.CarriedGold)
			report.GroundGoldAfterRespawn = town.GroundGold
			report.CorpseCountAfterRespawn = town.CorpseCount
			report.EquipmentRemovedAtRespawn = missing(before.Equipment, town.Equipment)
			report.InventoryRemovedAtRespawn = missing(before.Inventory, town.Inventory)
		}
		if recovered, exists := phases["corpse_recovered"]; exists && hasTown {
			report.RecoveryObserved = true
			report.ExperienceRecovered = max64(0, recovered.Experience-town.Experience)
			report.EquipmentRestored = restored(before.Equipment, town.Equipment, recovered.Equipment)
			report.InventoryRestored = restored(before.Inventory, town.Inventory, recovered.Inventory)
		}
		result.Deaths = append(result.Deaths, report)
	}
	return result
}

func missing(before, after []slotItem) int {
	afterIDs := itemIDs(after)
	count := 0
	for id := range itemIDs(before) {
		if !afterIDs[id] {
			count++
		}
	}
	return count
}

func restored(before, after, recovered []slotItem) int {
	afterIDs, recoveredIDs := itemIDs(after), itemIDs(recovered)
	count := 0
	for id := range itemIDs(before) {
		if !afterIDs[id] && recoveredIDs[id] {
			count++
		}
	}
	return count
}

func itemIDs(items []slotItem) map[string]bool {
	result := make(map[string]bool, len(items))
	for _, item := range items {
		result[item.ID] = true
	}
	return result
}

func hasPhase(byDeath map[int]map[string]observation, phase string) bool {
	for _, phases := range byDeath {
		if _, exists := phases[phase]; exists {
			return true
		}
	}
	return false
}

func max64(left, right int64) int64 {
	if left > right {
		return left
	}
	return right
}

func validSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func oneOf(value string, choices ...string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}

func fatal(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
