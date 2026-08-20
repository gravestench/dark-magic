package player

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/gravestench/dark-magic/internal/game/simulation"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

// ProjectCharacter merges session-owned fields from one canonical checkpoint
// into the durable baseline acquired by the trusted host. Fields not yet owned
// by the session remain unchanged instead of being erased by a partial view.
func ProjectCharacter(
	playerID string,
	baseline d2save.Character,
	checkpoint simulation.Checkpoint,
) (d2save.Character, error) {
	payload, err := ProjectHUD(playerID, checkpoint)
	if err != nil {
		return d2save.Character{}, err
	}

	view, err := DecodeHUD(payload)
	if err != nil {
		return d2save.Character{}, err
	}

	return MergeCharacter(baseline, view)
}

// MergeCharacter copies only session-owned durable fields from an authenticated
// owner projection. Player-profile-only fields, including appearance and game
// mode flags, remain sourced from the selected local baseline. Realm commits
// apply their separate lease and revision policy before calling ProjectCharacter.
func MergeCharacter(baseline d2save.Character, view HUD) (d2save.Character, error) {
	if baseline.ID == "" || view.Version != HUDVersion || view.Player.CharacterID != baseline.ID {
		return d2save.Character{}, fmt.Errorf("%w: checkpoint character differs", ErrHUDPlayer)
	}

	values := []int64{
		view.Progress.Level, view.Progress.Experience, view.Vitals.Health,
		view.Vitals.MaxHealth, view.Vitals.Mana, view.Vitals.MaxMana,
		view.Vitals.Stamina, view.Vitals.MaxStamina, view.Combat.Defense,
	}

	// Validate every converted integer before mutating the clone. This prevents
	// partial saves and preserves compatibility on 32-bit hosts.
	for _, value := range values {
		if value < 0 || (strconv.IntSize == 32 && value > math.MaxInt32) {
			return d2save.Character{}, fmt.Errorf("%w: durable numeric value is out of range", ErrHUDPlayer)
		}
	}

	if view.Vitals.Health > view.Vitals.MaxHealth || view.Vitals.Mana > view.Vitals.MaxMana ||
		view.Vitals.Stamina > view.Vitals.MaxStamina {
		return d2save.Character{}, fmt.Errorf("%w: durable vitals are inconsistent", ErrHUDPlayer)
	}

	result := cloneDurableCharacter(baseline)

	result.Name, result.Class, result.Level = view.Player.Name, view.Player.Class, int(view.Progress.Level)
	if result.Stats == nil {
		result.Stats = &d2save.Stats{}
	}

	result.Stats.Experience = int(view.Progress.Experience)
	result.Stats.Health, result.Stats.MaxHealth = int(view.Vitals.Health), int(view.Vitals.MaxHealth)
	result.Stats.Mana, result.Stats.MaxMana = int(view.Vitals.Mana), int(view.Vitals.MaxMana)
	result.Stats.Stamina, result.Stats.MaxStamina = int(view.Vitals.Stamina), int(view.Vitals.MaxStamina)
	result.Stats.Defense = int(view.Combat.Defense)

	return result, nil
}

// DecodeHUD rejects malformed or version-skewed checkpoints before durable
// projection, so callers never merge a schema they do not understand.
func DecodeHUD(payload []byte) (HUD, error) {
	var view HUD
	if err := json.Unmarshal(payload, &view); err != nil || view.Version != HUDVersion {
		return HUD{}, fmt.Errorf("%w: invalid HUD projection", ErrHUDPlayer)
	}

	return view, nil
}

// cloneDurableCharacter delegates to the save store's established deep-copy
// behavior, preventing checkpoint projection from aliasing baseline pointers.
func cloneDurableCharacter(character d2save.Character) d2save.Character {
	store := d2save.New(character)
	return store.Characters()[0]
}
