package player

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const HUDVersion uint32 = 2

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
}

type HUDIdentity struct {
	PlayerID    string `json:"player_id"`
	CharacterID string `json:"character_id"`
	Name        string `json:"name"`
	Class       string `json:"class"`
}

type HUDVitals struct {
	Health    int64 `json:"health"`
	MaxHealth int64 `json:"max_health"`
	Mana      int64 `json:"mana"`
	MaxMana   int64 `json:"max_mana"`
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
	view.Vitals = HUDVitals{Health: intField(vitals, "health"), MaxHealth: intField(vitals, "max_health"), Mana: intField(vitals, "mana"), MaxMana: intField(vitals, "max_mana")}
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
		}
	}
	if facing, found := findComponent(snapshot, "d2legacy.world.facing"); found {
		if fields, present := findInstance(facing, entity); present {
			view.Animation.Direction = intField(fields, "direction")
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
