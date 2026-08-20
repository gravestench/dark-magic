// Package player defines authoritative player entity materialization.
package player

import (
	"encoding/json"
	"fmt"

	"github.com/gravestench/dark-magic/internal/game/simulation"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

// EnterCommand materializes a selected durable character into session-owned ECS
// state. It is system/admin authority only; clients cannot self-admit players.
const EnterCommand = "system.player.enter"

// Entry is the complete validated intent needed to create the initial player
// archetype. After application, ECS components—not this payload—are live state.
type Entry struct {
	CharacterID string  `json:"character_id"`
	Player      string  `json:"player"`
	Name        string  `json:"name"`
	Class       string  `json:"class"`
	Level       int64   `json:"level"`
	Experience  int64   `json:"experience"`
	Dexterity   int64   `json:"dexterity"`
	Vitality    int64   `json:"vitality"`
	Defense     int64   `json:"defense"`
	Health      int64   `json:"health"`
	MaxHealth   int64   `json:"max_health"`
	Mana        int64   `json:"mana"`
	MaxMana     int64   `json:"max_mana"`
	Stamina     int64   `json:"stamina"`
	MaxStamina  int64   `json:"max_stamina"`
	Expansion   bool    `json:"expansion"`
	Hardcore    bool    `json:"hardcore"`
	COF         string  `json:"cof"`
	Palette     string  `json:"palette,omitempty"`
	Direction   *int64  `json:"direction,omitempty"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	WorldWidth  float64 `json:"world_width"`
	WorldHeight float64 `json:"world_height"`
	Act         int64   `json:"act"`
	LevelID     int64   `json:"level_id"`
}

// Destination is the server-selected place where an admitted character enters
// a game. Keeping this separate from a local save means offline selection and a
// remote realm join can share exactly the same trusted town-spawn policy.
type Destination struct {
	X, Y, Width, Height float64
	Act, LevelID        int64
}

// NewDestination requires positive world bounds and an in-bounds entry anchor.
// The checks preserve the command's established spawn and location contract.
func NewDestination(x, y, width, height float64, act, levelID int64) (Destination, error) {
	if width <= 0 || height <= 0 || x < 0 || y < 0 || x >= width || y >= height || act <= 0 || levelID <= 0 {
		return Destination{}, fmt.Errorf("player: destination requires an in-bounds spawn, act, and level")
	}

	return Destination{X: x, Y: y, Width: width, Height: height, Act: act, LevelID: levelID}, nil
}

// AdmissionCommand converts a durable character selected by a trusted host
// into one replayable entry command. Network clients may request a join, but
// they cannot choose their spawn or mint this system/admin-authority command.
func AdmissionCommand(
	character d2save.Character,
	player string,
	destination Destination,
	actor string,
	sequence, tick uint64,
	authority simulation.Authority,
) (simulation.Command, error) {
	if authority != simulation.AuthoritySystem && authority != simulation.AuthorityAdmin {
		return simulation.Command{}, fmt.Errorf("player: admission requires system or admin authority")
	}

	validated, err := NewDestination(
		destination.X,
		destination.Y,
		destination.Width,
		destination.Height,
		destination.Act,
		destination.LevelID,
	)
	if err != nil {
		return simulation.Command{}, err
	}

	entry := EntryFromCharacter(character, player, validated.X, validated.Y, validated.Width, validated.Height)
	entry.Act, entry.LevelID = validated.Act, validated.LevelID

	return Command(entry, actor, sequence, tick, authority)
}

// EntryFromCharacter copies the admitted durable subset into a command value.
func EntryFromCharacter(character d2save.Character, player string, x, y, width, height float64) Entry {
	entry := Entry{
		CharacterID: character.ID,
		Player:      player,
		Name:        character.Name,
		Class:       character.Class,
		Level:       int64(character.Level),
		Expansion:   character.Expansion,
		Hardcore:    character.Hardcore,
		X:           x,
		Y:           y,
		WorldWidth:  width,
		WorldHeight: height,
	}

	// Stats and appearance are optional in older saves. Their zero values are
	// intentionally retained so admission remains compatible with those saves.
	if character.Stats != nil {
		entry.Experience = int64(character.Stats.Experience)
		entry.Dexterity = int64(character.Stats.Dexterity)
		entry.Vitality = int64(character.Stats.Vitality)
		entry.Defense = int64(character.Stats.Defense)
		entry.Health, entry.MaxHealth = int64(character.Stats.Health), int64(character.Stats.MaxHealth)
		entry.Mana, entry.MaxMana = int64(character.Stats.Mana), int64(character.Stats.MaxMana)
		entry.Stamina, entry.MaxStamina = int64(character.Stats.Stamina), int64(character.Stats.MaxStamina)
	}

	if character.Appearance != nil {
		direction := int64(character.Appearance.Direction)

		entry.COF, entry.Direction = character.Appearance.COF, &direction
		if character.Appearance.Palette != "" {
			entry.Palette = character.Appearance.Palette
		}
	}

	return entry
}

// Command encodes player-entry intent for deterministic admission and replay.
func Command(
	entry Entry,
	actor string,
	sequence, tick uint64,
	authority simulation.Authority,
) (simulation.Command, error) {
	payload, err := json.Marshal(entry)
	if err != nil {
		return simulation.Command{}, err
	}

	return simulation.Command{
		Tick: tick, Player: actor, Authority: authority, Sequence: sequence, Kind: EnterCommand, Payload: payload,
	}, nil
}
