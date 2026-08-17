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
	view.Player = HUDIdentity{PlayerID: stringField(identity, "player"), CharacterID: stringField(identity, "character_id"), Name: stringField(identity, "name"), Class: stringField(identity, "class")}
	if err := fillHUD(snapshot, entity, &view); err != nil {
		return nil, err
	}
	return json.Marshal(view)
}

func fillHUD(snapshot gameecs.Snapshot, entity uint64, view *HUD) error {
	read := func(name string) (map[string]gameecs.ValueSnapshot, error) {
		component, found := findComponent(snapshot, name)
		if !found {
			return nil, fmt.Errorf("player HUD: component %q is absent", name)
		}
		fields, found := findInstance(component, entity)
		if !found {
			return nil, fmt.Errorf("player HUD: component %q is absent for player", name)
		}
		return fields, nil
	}
	vitals, err := read("d2legacy.player.vitals")
	if err != nil {
		return err
	}
	view.Vitals = HUDVitals{
		Health: intField(vitals, "health"), MaxHealth: intField(vitals, "max_health"),
		Mana: intField(vitals, "mana"), MaxMana: intField(vitals, "max_mana"),
		Stamina: intField(vitals, "stamina"), MaxStamina: intField(vitals, "max_stamina"),
		StaminaRaw: intField(vitals, "stamina_raw"), MaxStaminaRaw: intField(vitals, "max_stamina_raw"),
	}
	progress, err := read("d2legacy.player.progress")
	if err != nil {
		return err
	}
	view.Progress = HUDProgress{Level: intField(progress, "level"), Experience: intField(progress, "experience"), UnspentSkillPoints: intField(progress, "unspent_skill_points")}
	combat, err := read("d2legacy.player.combat_stats")
	if err != nil {
		return err
	}
	view.Combat = HUDCombat{AttackRating: intField(combat, "attack_rating"), Defense: intField(combat, "defense")}
	position, err := read("d2legacy.world.position")
	if err != nil {
		return err
	}
	view.Position = HUDPosition{X: floatField(position, "x"), Y: floatField(position, "y")}
	location, err := read("d2legacy.world.location")
	if err != nil {
		return err
	}
	view.Location = HUDLocation{Act: intField(location, "act"), LevelID: intField(location, "level_id")}
	if animation, found := findComponent(snapshot, "d2legacy.player.animation"); found {
		if fields, present := findInstance(animation, entity); present {
			view.Animation.Mode = stringField(fields, "mode")
			view.Animation.StartTick = uint64(max(0, intField(fields, "start_tick")))
		}
	}
	if facing, found := findComponent(snapshot, "d2legacy.world.facing"); found {
		if fields, present := findInstance(facing, entity); present {
			view.Animation.Direction = intField(fields, "direction")
		}
	}
	if movement, found := findComponent(snapshot, "d2legacy.player.movement_mode"); found {
		if fields, present := findInstance(movement, entity); present {
			view.Movement.Running = boolField(fields, "running")
		}
	}
	if movement, found := findComponent(snapshot, "d2legacy.player.movement_stats"); found {
		if fields, present := findInstance(movement, entity); present {
			view.Movement.VelocityPercent = intField(fields, "velocitypercent")
			view.Movement.ItemFasterMoveVelocity = intField(fields, "item_fastermovevelocity")
			view.Movement.RunDrain = intField(fields, "run_drain")
			view.Movement.StaminaRecoveryBonus = intField(fields, "staminarecoverybonus")
			view.Movement.StaminaDrainPercent = intField(fields, "item_staminadrainpct")
			view.Movement.ArmorRunDrain = intField(fields, "armor_run_drain")
		}
	}
	if velocity, found := findComponent(snapshot, "d2legacy.world.velocity"); found {
		if fields, present := findInstance(velocity, entity); present {
			view.Movement.Velocity = HUDPosition{X: floatField(fields, "x"), Y: floatField(fields, "y")}
		}
	}
	if bounds, found := findComponent(snapshot, "d2legacy.world.bounds"); found {
		if fields, present := findInstance(bounds, entity); present {
			view.Movement.Bounds = HUDPosition{X: floatField(fields, "width"), Y: floatField(fields, "height")}
		}
	}
	if colliders, found := findComponent(snapshot, "d2legacy.world.collider"); found {
		if fields, present := findInstance(colliders, entity); present {
			view.Movement.Radius = floatField(fields, "radius")
		}
	}
	if assignment, found := findComponent(snapshot, "d2legacy.player.skill_assignment"); found {
		if fields, present := findInstance(assignment, entity); present {
			view.Skills.Left, view.Skills.Right = intField(fields, "left"), intField(fields, "right")
		}
	}
	view.Skills.Learned = []HUDLearnedSkill{}
	if learned, found := findComponent(snapshot, "d2legacy.player.learned_skill"); found {
		for _, instance := range learned.Instances {
			fields, present := findInstance(learned, instance.Entity)
			if !present || entityField(fields, "owner") != entity {
				continue
			}
			view.Skills.Learned = append(view.Skills.Learned, HUDLearnedSkill{
				SkillID: intField(fields, "skill_id"), Level: intField(fields, "level"), ListRow: intField(fields, "list_row"),
				LeftAllowed: boolField(fields, "left_allowed"), RightAllowed: boolField(fields, "right_allowed"),
			})
		}
	}
	sort.Slice(view.Skills.Learned, func(i, j int) bool {
		if view.Skills.Learned[i].ListRow == view.Skills.Learned[j].ListRow {
			return view.Skills.Learned[i].SkillID < view.Skills.Learned[j].SkillID
		}
		return view.Skills.Learned[i].ListRow < view.Skills.Learned[j].ListRow
	})
	view.Belt.Slots = make([]string, 16)
	if belt, found := findComponent(snapshot, "d2legacy.player.belt"); found {
		if fields, present := findInstance(belt, entity); present {
			view.Belt.Capacity = intField(fields, "capacity")
			for slot := 1; slot <= len(view.Belt.Slots); slot++ {
				view.Belt.Slots[slot-1] = stringField(fields, fmt.Sprintf("slot_%d", slot))
			}
		}
	}
	return nil
}

func findComponent(snapshot gameecs.Snapshot, name string) (gameecs.ComponentSnapshot, bool) {
	for _, component := range snapshot.Components {
		if component.Name == name {
			return component, true
		}
	}
	return gameecs.ComponentSnapshot{}, false
}

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

func findString(component gameecs.ComponentSnapshot, field, value string) (uint64, map[string]gameecs.ValueSnapshot, bool) {
	for _, instance := range component.Instances {
		fields, found := findInstance(component, instance.Entity)
		if found && stringField(fields, field) == value {
			return instance.Entity, fields, true
		}
	}
	return 0, nil, false
}

func stringField(fields map[string]gameecs.ValueSnapshot, name string) string {
	if value := fields[name].String; value != nil {
		return *value
	}
	return ""
}

func intField(fields map[string]gameecs.ValueSnapshot, name string) int64 {
	if value := fields[name].Int; value != nil {
		return *value
	}
	return 0
}

func floatField(fields map[string]gameecs.ValueSnapshot, name string) float64 {
	if value := fields[name].Float; value != nil {
		return math.Float64frombits(*value)
	}
	return 0
}

func boolField(fields map[string]gameecs.ValueSnapshot, name string) bool {
	if value := fields[name].Bool; value != nil {
		return *value
	}
	return false
}

func entityField(fields map[string]gameecs.ValueSnapshot, name string) uint64 {
	if value := fields[name].Entity; value != nil {
		return *value
	}
	return 0
}
