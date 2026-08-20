package player

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const HUDVersion uint32 = 6

var ErrHUDPlayer = errors.New("player HUD: authenticated player is absent")

// HUD is the versioned owner-private view shared by local presentation and
// remote clients. Other private domains receive separately reviewed views.
type HUD struct {
	Version   uint32       `json:"version"`
	Tick      uint64       `json:"tick"`
	Player    HUDIdentity  `json:"player"`
	Vitals    HUDVitals    `json:"vitals"`
	Progress  HUDProgress  `json:"progress"`
	Combat    HUDCombat    `json:"combat"`
	Position  HUDPosition  `json:"position"`
	Location  HUDLocation  `json:"location"`
	Animation HUDAnimation `json:"animation"`
	Movement  HUDMovement  `json:"movement"`
	Skills    HUDSkills    `json:"skills"`
	Belt      HUDBelt      `json:"belt"`
}

type HUDIdentity struct {
	PlayerID    string `json:"player_id"`
	CharacterID string `json:"character_id"`
	Name        string `json:"name"`
	Class       string `json:"class"`
}

type HUDVitals struct {
	Health        int64 `json:"health"`
	MaxHealth     int64 `json:"max_health"`
	Mana          int64 `json:"mana"`
	MaxMana       int64 `json:"max_mana"`
	Stamina       int64 `json:"stamina"`
	MaxStamina    int64 `json:"max_stamina"`
	StaminaRaw    int64 `json:"stamina_raw"`
	MaxStaminaRaw int64 `json:"max_stamina_raw"`
}

type HUDProgress struct {
	Level              int64 `json:"level"`
	Experience         int64 `json:"experience"`
	UnspentSkillPoints int64 `json:"unspent_skill_points"`
}

type HUDCombat struct {
	AttackRating int64 `json:"attack_rating"`
	Defense      int64 `json:"defense"`
}

type HUDPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type HUDLocation struct {
	Act     int64 `json:"act"`
	LevelID int64 `json:"level_id"`
}

type HUDAnimation struct {
	Mode      string `json:"mode"`
	Direction int64  `json:"direction"`
	StartTick uint64 `json:"start_tick"`
}

type HUDMovement struct {
	Running                bool        `json:"running"`
	Velocity               HUDPosition `json:"velocity"`
	Bounds                 HUDPosition `json:"bounds"`
	Radius                 float64     `json:"radius"`
	VelocityPercent        int64       `json:"velocity_percent"`
	ItemFasterMoveVelocity int64       `json:"item_faster_move_velocity"`
	RunDrain               int64       `json:"run_drain"`
	StaminaRecoveryBonus   int64       `json:"stamina_recovery_bonus"`
	StaminaDrainPercent    int64       `json:"stamina_drain_percent"`
	ArmorRunDrain          int64       `json:"armor_run_drain"`
}

type HUDSkills struct {
	Left    int64             `json:"left"`
	Right   int64             `json:"right"`
	Learned []HUDLearnedSkill `json:"learned"`
}

type HUDLearnedSkill struct {
	SkillID      int64 `json:"skill_id"`
	Level        int64 `json:"level"`
	ListRow      int64 `json:"list_row"`
	LeftAllowed  bool  `json:"left_allowed"`
	RightAllowed bool  `json:"right_allowed"`
}

type HUDBelt struct {
	Capacity int64    `json:"capacity"`
	Slots    []string `json:"slots"`
}

// ProjectHUD reads only the canonical checkpoint captured by the session. It
// cannot race the live ECS or accidentally serialize arbitrary components.
func ProjectHUD(playerID string, checkpoint simulation.Checkpoint) (json.RawMessage, error) {
	if checkpoint.Snapshot == nil || strings.TrimSpace(playerID) == "" {
		return nil, ErrHUDPlayer
	}

	snapshot := *checkpoint.Snapshot

	identities, found := findComponent(snapshot, "d2legacy.player.identity")
	if !found {
		return nil, ErrHUDPlayer
	}

	entity, identity, found := findString(identities, "player", playerID)
	if !found {
		return nil, ErrHUDPlayer
	}

	view := HUD{Version: HUDVersion, Tick: checkpoint.Tick}

	view.Player = HUDIdentity{
		PlayerID:    stringField(identity, "player"),
		CharacterID: stringField(identity, "character_id"),
		Name:        stringField(identity, "name"),
		Class:       stringField(identity, "class"),
	}
	if err := fillHUD(snapshot, entity, &view); err != nil {
		return nil, err
	}

	return json.Marshal(view)
}

// fillHUD assembles required owner state before adding optional presentation
// domains. Missing required components fail the projection; optional components
// retain zero values so older checkpoints remain displayable.
func fillHUD(snapshot gameecs.Snapshot, entity uint64, view *HUD) error {
	read := hudComponentReader{snapshot: snapshot, entity: entity}
	if err := fillRequiredHUD(read, view); err != nil {
		return err
	}

	fillOptionalHUD(read, view)
	fillHUDSkills(read, view)
	fillHUDBelt(read, view)

	return nil
}

type hudComponentReader struct {
	snapshot gameecs.Snapshot
	entity   uint64
}

// required returns an authenticated player's component or a diagnostic error.
// Keeping this policy in one place makes required HUD schema additions explicit.
func (reader hudComponentReader) required(name string) (map[string]gameecs.ValueSnapshot, error) {
	component, found := findComponent(reader.snapshot, name)
	if !found {
		return nil, fmt.Errorf("player HUD: component %q is absent", name)
	}

	fields, found := findInstance(component, reader.entity)
	if !found {
		return nil, fmt.Errorf("player HUD: component %q is absent for player", name)
	}

	return fields, nil
}

// optional returns an authenticated player's component when the checkpoint
// contains it. Absence deliberately maps to zero-valued presentation fields.
func (reader hudComponentReader) optional(name string) (map[string]gameecs.ValueSnapshot, bool) {
	component, found := findComponent(reader.snapshot, name)
	if !found {
		return nil, false
	}

	return findInstance(component, reader.entity)
}

// fillRequiredHUD copies the canonical durable and spatial fields that every
// valid player checkpoint must contain, failing before a partial view escapes.
func fillRequiredHUD(reader hudComponentReader, view *HUD) error {
	vitals, err := reader.required("d2legacy.player.vitals")
	if err != nil {
		return err
	}

	view.Vitals = HUDVitals{
		Health:        intField(vitals, "health"),
		MaxHealth:     intField(vitals, "max_health"),
		Mana:          intField(vitals, "mana"),
		MaxMana:       intField(vitals, "max_mana"),
		Stamina:       intField(vitals, "stamina"),
		MaxStamina:    intField(vitals, "max_stamina"),
		StaminaRaw:    intField(vitals, "stamina_raw"),
		MaxStaminaRaw: intField(vitals, "max_stamina_raw"),
	}

	progress, err := reader.required("d2legacy.player.progress")
	if err != nil {
		return err
	}

	view.Progress = HUDProgress{
		Level:              intField(progress, "level"),
		Experience:         intField(progress, "experience"),
		UnspentSkillPoints: intField(progress, "unspent_skill_points"),
	}

	combat, err := reader.required("d2legacy.player.combat_stats")
	if err != nil {
		return err
	}

	view.Combat = HUDCombat{AttackRating: intField(combat, "attack_rating"), Defense: intField(combat, "defense")}

	position, err := reader.required("d2legacy.world.position")
	if err != nil {
		return err
	}

	view.Position = HUDPosition{X: floatField(position, "x"), Y: floatField(position, "y")}

	location, err := reader.required("d2legacy.world.location")
	if err != nil {
		return err
	}

	view.Location = HUDLocation{Act: intField(location, "act"), LevelID: intField(location, "level_id")}

	return nil
}

// fillOptionalHUD adds live presentation and movement details when available.
// Each component stays optional to preserve compatibility with older snapshots.
func fillOptionalHUD(reader hudComponentReader, view *HUD) {
	if fields, found := reader.optional("d2legacy.player.animation"); found {
		view.Animation.Mode = stringField(fields, "mode")
		view.Animation.StartTick = uint64(max(0, intField(fields, "start_tick")))
	}

	if fields, found := reader.optional("d2legacy.world.facing"); found {
		view.Animation.Direction = intField(fields, "direction")
	}

	if fields, found := reader.optional("d2legacy.player.movement_mode"); found {
		view.Movement.Running = boolField(fields, "running")
	}

	if fields, found := reader.optional("d2legacy.player.movement_stats"); found {
		view.Movement.VelocityPercent = intField(fields, "velocitypercent")
		view.Movement.ItemFasterMoveVelocity = intField(fields, "item_fastermovevelocity")
		view.Movement.RunDrain = intField(fields, "run_drain")
		view.Movement.StaminaRecoveryBonus = intField(fields, "staminarecoverybonus")
		view.Movement.StaminaDrainPercent = intField(fields, "item_staminadrainpct")
		view.Movement.ArmorRunDrain = intField(fields, "armor_run_drain")
	}

	if fields, found := reader.optional("d2legacy.world.velocity"); found {
		view.Movement.Velocity = HUDPosition{X: floatField(fields, "x"), Y: floatField(fields, "y")}
	}

	if fields, found := reader.optional("d2legacy.world.bounds"); found {
		view.Movement.Bounds = HUDPosition{X: floatField(fields, "width"), Y: floatField(fields, "height")}
	}

	if fields, found := reader.optional("d2legacy.world.collider"); found {
		view.Movement.Radius = floatField(fields, "radius")
	}

	if fields, found := reader.optional("d2legacy.player.skill_assignment"); found {
		view.Skills.Left = intField(fields, "left")
		view.Skills.Right = intField(fields, "right")
	}
}

// fillHUDSkills projects only skills owned by the authenticated player and
// sorts them so snapshots with different ECS iteration order encode identically.
func fillHUDSkills(reader hudComponentReader, view *HUD) {
	view.Skills.Learned = []HUDLearnedSkill{}

	if learned, found := findComponent(reader.snapshot, "d2legacy.player.learned_skill"); found {
		for _, instance := range learned.Instances {
			fields, present := findInstance(learned, instance.Entity)
			if !present || entityField(fields, "owner") != reader.entity {
				continue
			}

			view.Skills.Learned = append(view.Skills.Learned, HUDLearnedSkill{
				SkillID:      intField(fields, "skill_id"),
				Level:        intField(fields, "level"),
				ListRow:      intField(fields, "list_row"),
				LeftAllowed:  boolField(fields, "left_allowed"),
				RightAllowed: boolField(fields, "right_allowed"),
			})
		}
	}

	sort.Slice(view.Skills.Learned, func(i, j int) bool {
		if view.Skills.Learned[i].ListRow == view.Skills.Learned[j].ListRow {
			return view.Skills.Learned[i].SkillID < view.Skills.Learned[j].SkillID
		}

		return view.Skills.Learned[i].ListRow < view.Skills.Learned[j].ListRow
	})
}

// fillHUDBelt reserves the stable sixteen-slot wire shape even when no belt
// component exists, preventing optional ECS state from changing JSON shape.
func fillHUDBelt(reader hudComponentReader, view *HUD) {
	view.Belt.Slots = make([]string, 16)
	if fields, found := reader.optional("d2legacy.player.belt"); found {
		view.Belt.Capacity = intField(fields, "capacity")
		for slot := 1; slot <= len(view.Belt.Slots); slot++ {
			view.Belt.Slots[slot-1] = stringField(fields, fmt.Sprintf("slot_%d", slot))
		}
	}
}

// findComponent locates a schema by its stable name without exposing live ECS
// stores; every projection therefore reads only the immutable checkpoint copy.
func findComponent(snapshot gameecs.Snapshot, name string) (gameecs.ComponentSnapshot, bool) {
	for _, component := range snapshot.Components {
		if component.Name == name {
			return component, true
		}
	}

	return gameecs.ComponentSnapshot{}, false
}

// findInstance maps one positional snapshot row back to named fields. Rows with
// mismatched schemas are rejected instead of being partially interpreted.
func findInstance(component gameecs.ComponentSnapshot, entity uint64) (map[string]gameecs.ValueSnapshot, bool) {
	for _, instance := range component.Instances {
		if instance.Entity != entity || len(instance.Values) != len(component.Fields) {
			continue
		}

		fields := make(map[string]gameecs.ValueSnapshot, len(component.Fields))
		for index, field := range component.Fields {
			fields[field.Name] = instance.Values[index]
		}

		return fields, true
	}

	return nil, false
}

// findString identifies an entity by one string field. It preserves component
// iteration order, which matches the existing first-match authentication rule.
func findString(
	component gameecs.ComponentSnapshot,
	field, value string,
) (uint64, map[string]gameecs.ValueSnapshot, bool) {
	for _, instance := range component.Instances {
		fields, found := findInstance(component, instance.Entity)
		if found && stringField(fields, field) == value {
			return instance.Entity, fields, true
		}
	}

	return 0, nil, false
}

// stringField reads a string snapshot value, treating absent optional fields as
// their wire-format zero value rather than panicking on schema evolution.
func stringField(fields map[string]gameecs.ValueSnapshot, name string) string {
	if value := fields[name].String; value != nil {
		return *value
	}

	return ""
}

// intField reads an integer snapshot value, preserving zero-value behavior for
// optional or older component fields.
func intField(fields map[string]gameecs.ValueSnapshot, name string) int64 {
	if value := fields[name].Int; value != nil {
		return *value
	}

	return 0
}

// floatField reconstructs a float from its deterministic snapshot bit pattern;
// absent optional fields remain zero.
func floatField(fields map[string]gameecs.ValueSnapshot, name string) float64 {
	if value := fields[name].Float; value != nil {
		return math.Float64frombits(*value)
	}

	return 0
}

// boolField reads a boolean snapshot value while keeping absent optional fields
// backward-compatible with false.
func boolField(fields map[string]gameecs.ValueSnapshot, name string) bool {
	if value := fields[name].Bool; value != nil {
		return *value
	}

	return false
}

// entityField reads an ECS relationship without leaking a live entity handle;
// zero continues to mean an absent relationship.
func entityField(fields map[string]gameecs.ValueSnapshot, name string) uint64 {
	if value := fields[name].Entity; value != nil {
		return *value
	}

	return 0
}
