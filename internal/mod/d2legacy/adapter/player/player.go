// Package player defines authoritative player entity materialization.
package player

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

// EnterCommand materializes a selected durable character into session-owned ECS
// state. It is system/admin authority only; clients cannot self-admit players.
const EnterCommand = "system.player.enter"

const localEntryActor = "system:local-player-entry"

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
	Defense     int64   `json:"defense"`
	Health      int64   `json:"health"`
	MaxHealth   int64   `json:"max_health"`
	Mana        int64   `json:"mana"`
	MaxMana     int64   `json:"max_mana"`
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

// NewDestination validates a finite authoritative world and its entry anchor.
func NewDestination(x, y, width, height float64, act, levelID int64) (Destination, error) {
	if width <= 0 || height <= 0 || x < 0 || y < 0 || x >= width || y >= height || act < 1 || act > 5 || levelID <= 0 {
		return Destination{}, fmt.Errorf("player: destination requires an in-bounds spawn, act, and level")
	}
	return Destination{X: x, Y: y, Width: width, Height: height, Act: act, LevelID: levelID}, nil
}

// AdmissionCommand converts a durable character selected by a trusted host
// into one replayable entry command. Network clients may request a join, but
// they cannot choose their spawn or mint this system/admin-authority command.
func AdmissionCommand(character d2save.Character, player string, destination Destination, actor string, sequence, tick uint64, authority simulation.Authority) (simulation.Command, error) {
	if authority != simulation.AuthoritySystem && authority != simulation.AuthorityAdmin {
		return simulation.Command{}, fmt.Errorf("player: admission requires system or admin authority")
	}
	validated, err := NewDestination(destination.X, destination.Y, destination.Width, destination.Height, destination.Act, destination.LevelID)
	if err != nil {
		return simulation.Command{}, err
	}
	entry := EntryFromCharacter(character, player, validated.X, validated.Y, validated.Width, validated.Height)
	entry.Act, entry.LevelID = validated.Act, validated.LevelID
	return Command(entry, actor, sequence, tick, authority)
}

// EntrySource admits the currently selected save into the authoritative world.
// Selection remains shell state; after admission, ECS components are the live
// gameplay state and Lua only observes them.
type EntrySource struct {
	engine      *gameecs.Engine
	saves       *d2save.Store
	player      string
	destination Destination
	sequence    uint64
}

// NewEntrySource creates the offline adapter that emits one entry command for
// the selected character. Remote sessions receive equivalent commands elsewhere.
func NewEntrySource(engine *gameecs.Engine, saves *d2save.Store, player string, width, height float64) (*EntrySource, error) {
	return NewEntrySourceAt(engine, saves, player, width/2, height/2, width, height)
}

// NewEntrySourceAt admits the player at a server-selected map coordinate.
func NewEntrySourceAt(engine *gameecs.Engine, saves *d2save.Store, player string, x, y, width, height float64) (*EntrySource, error) {
	return NewEntrySourceAtLocation(engine, saves, player, x, y, width, height, 1, 1)
}

// NewEntrySourceAtLocation records the server-selected act and town level in
// the same authoritative command as the server-selected spawn coordinate.
func NewEntrySourceAtLocation(engine *gameecs.Engine, saves *d2save.Store, player string, x, y, width, height float64, act, levelID int64) (*EntrySource, error) {
	destination, err := NewDestination(x, y, width, height, act, levelID)
	if err != nil {
		return nil, err
	}
	return NewEntrySourceForDestination(engine, saves, player, destination)
}

// NewEntrySourceForDestination adapts local save selection to the same
// destination contract used by trusted remote admission.
func NewEntrySourceForDestination(engine *gameecs.Engine, saves *d2save.Store, player string, destination Destination) (*EntrySource, error) {
	player = strings.TrimSpace(player)
	validated, err := NewDestination(destination.X, destination.Y, destination.Width, destination.Height, destination.Act, destination.LevelID)
	if err != nil || engine == nil || saves == nil || player == "" {
		return nil, fmt.Errorf("player: entry source requires engine, saves, player, and positive world bounds")
	}
	return &EntrySource{engine: engine, saves: saves, player: player, destination: validated}, nil
}

// Commands emits entry intent once; it never materializes ECS state directly.
func (source *EntrySource) Commands(tick uint64) []simulation.Command {
	character, selected := source.saves.Selected()
	if !selected || source.entered(character.ID) {
		return nil
	}
	source.sequence++
	command, err := AdmissionCommand(character, source.player, source.destination, localEntryActor, source.sequence, tick, simulation.AuthoritySystem)
	if err != nil {
		return nil
	}
	return []simulation.Command{command}
}

func (source *EntrySource) entered(characterID string) bool {
	identities, found := akara.GetDynamicStore(source.engine.World(), "d2legacy.player.identity")
	if !found {
		return false
	}
	for _, entity := range identities.Entities() {
		identity, _ := identities.Get(entity)
		id, _ := identity.Get("character_id")
		if id == characterID {
			return true
		}
	}
	return false
}

// EntryFromCharacter copies the admitted durable subset into a command value.
func EntryFromCharacter(character d2save.Character, player string, x, y, width, height float64) Entry {
	entry := Entry{CharacterID: character.ID, Player: player, Name: character.Name, Class: character.Class, Level: int64(character.Level), Expansion: character.Expansion, Hardcore: character.Hardcore, X: x, Y: y, WorldWidth: width, WorldHeight: height}
	if character.Stats != nil {
		entry.Experience = int64(character.Stats.Experience)
		entry.Dexterity, entry.Defense = int64(character.Stats.Dexterity), int64(character.Stats.Defense)
		entry.Health, entry.MaxHealth = int64(character.Stats.Health), int64(character.Stats.MaxHealth)
		entry.Mana, entry.MaxMana = int64(character.Stats.Mana), int64(character.Stats.MaxMana)
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
func Command(entry Entry, actor string, sequence, tick uint64, authority simulation.Authority) (simulation.Command, error) {
	payload, err := json.Marshal(entry)
	if err != nil {
		return simulation.Command{}, err
	}
	return simulation.Command{Tick: tick, Player: actor, Authority: authority, Sequence: sequence, Kind: EnterCommand, Payload: payload}, nil
}
