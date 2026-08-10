// Package player defines authoritative player entity materialization.
package player

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/persistence"
)

const EnterCommand = "system.player.enter"

const localEntryActor = "system:local-player-entry"

type Entry struct {
	CharacterID string  `json:"character_id"`
	Player      string  `json:"player"`
	Name        string  `json:"name"`
	Class       string  `json:"class"`
	Level       int64   `json:"level"`
	Experience  int64   `json:"experience"`
	Health      int64   `json:"health"`
	MaxHealth   int64   `json:"max_health"`
	Mana        int64   `json:"mana"`
	MaxMana     int64   `json:"max_mana"`
	Expansion   bool    `json:"expansion"`
	Hardcore    bool    `json:"hardcore"`
	COF         string  `json:"cof"`
	Palette     string  `json:"palette"`
	Direction   int64   `json:"direction"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	WorldWidth  float64 `json:"world_width"`
	WorldHeight float64 `json:"world_height"`
}

// EntrySource admits the currently selected save into the authoritative world.
// Selection remains shell state; after admission, ECS components are the live
// gameplay state and Lua only observes them.
type EntrySource struct {
	engine        *gameecs.Engine
	saves         *persistence.Store
	player        string
	width, height float64
	sequence      uint64
}

func NewEntrySource(engine *gameecs.Engine, saves *persistence.Store, player string, width, height float64) (*EntrySource, error) {
	player = strings.TrimSpace(player)
	if engine == nil || saves == nil || player == "" || width <= 0 || height <= 0 {
		return nil, fmt.Errorf("player: entry source requires engine, saves, player, and positive world bounds")
	}
	return &EntrySource{engine: engine, saves: saves, player: player, width: width, height: height}, nil
}

func (source *EntrySource) Commands(tick uint64) []simulation.Command {
	character, selected := source.saves.Selected()
	if !selected || source.entered(character.ID) {
		return nil
	}
	source.sequence++
	entry := EntryFromCharacter(character, source.player, source.width/2, source.height/2, source.width, source.height)
	command, err := Command(entry, localEntryActor, source.sequence, tick, simulation.AuthoritySystem)
	if err != nil {
		return nil
	}
	return []simulation.Command{command}
}

func (source *EntrySource) entered(characterID string) bool {
	identities, found := akara.GetDynamicStore(source.engine.World(), "dm.player.identity")
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

func EntryFromCharacter(character persistence.Character, player string, x, y, width, height float64) Entry {
	entry := Entry{CharacterID: character.ID, Player: player, Name: character.Name, Class: character.Class, Level: int64(character.Level), Expansion: character.Expansion, Hardcore: character.Hardcore, X: x, Y: y, WorldWidth: width, WorldHeight: height}
	if character.Stats != nil {
		entry.Experience = int64(character.Stats.Experience)
		entry.Health, entry.MaxHealth = int64(character.Stats.Health), int64(character.Stats.MaxHealth)
		entry.Mana, entry.MaxMana = int64(character.Stats.Mana), int64(character.Stats.MaxMana)
	}
	if character.Appearance != nil {
		entry.COF, entry.Palette, entry.Direction = character.Appearance.COF, character.Appearance.Palette, int64(character.Appearance.Direction)
	}
	return entry
}

func Command(entry Entry, actor string, sequence, tick uint64, authority simulation.Authority) (simulation.Command, error) {
	payload, err := json.Marshal(entry)
	if err != nil {
		return simulation.Command{}, err
	}
	return simulation.Command{Tick: tick, Player: actor, Authority: authority, Sequence: sequence, Kind: EnterCommand, Payload: payload}, nil
}

func Register(session *gamesession.Session) error {
	return session.Register(EnterCommand, gamesession.CommandHandler{
		Validate: func(command simulation.Command) error {
			_, err := decodeEntry(command.Payload)
			return err
		},
		Apply:   materialize,
		Allowed: []simulation.Authority{simulation.AuthoritySystem, simulation.AuthorityAdmin},
	})
}

func materialize(engine *gameecs.Engine, command simulation.Command) error {
	entry, err := decodeEntry(command.Payload)
	if err != nil {
		return err
	}
	stores, err := registerStores(engine.World())
	if err != nil {
		return err
	}
	for _, entity := range stores.identity.Entities() {
		component, _ := stores.identity.Get(entity)
		id, _ := component.Get("character_id")
		if id == entry.CharacterID {
			return fmt.Errorf("player: character %q already entered", entry.CharacterID)
		}
	}
	entity, err := engine.World().CreateEntity()
	if err != nil {
		return err
	}
	fail := func(err error) error {
		engine.World().DestroyEntity(entity)
		return err
	}
	components := []struct {
		store  *akara.DynamicStore
		values map[string]any
	}{
		{stores.identity, map[string]any{"character_id": entry.CharacterID, "player": entry.Player, "name": entry.Name, "class": entry.Class}},
		{stores.progress, map[string]any{"level": entry.Level, "experience": entry.Experience}},
		{stores.vitals, map[string]any{"health": entry.Health, "max_health": entry.MaxHealth, "mana": entry.Mana, "max_mana": entry.MaxMana}},
		{stores.appearance, map[string]any{"cof": entry.COF, "palette": entry.Palette, "direction": entry.Direction}},
		{stores.position, map[string]any{"x": entry.X, "y": entry.Y}},
		{stores.velocity, nil},
		{stores.movementMode, map[string]any{"running": false}},
		{stores.skillAssignment, map[string]any{"left": int64(0), "right": int64(0)}},
		{stores.belt, map[string]any{"capacity": int64(4)}},
		{stores.control, map[string]any{"player": entry.Player}},
		{stores.bounds, map[string]any{"width": entry.WorldWidth, "height": entry.WorldHeight}},
	}
	for _, component := range components {
		if _, err := component.store.Set(entity, component.values); err != nil {
			return fail(err)
		}
	}
	return nil
}

type stores struct {
	identity, progress, vitals, appearance                                   *akara.DynamicStore
	position, velocity, movementMode, skillAssignment, belt, control, bounds *akara.DynamicStore
}

func registerStores(world *akara.World) (stores, error) {
	schemas := []akara.Schema{
		{Name: "dm.player.identity", Version: 1, Fields: []akara.Field{{Name: "character_id", Kind: akara.FieldString}, {Name: "player", Kind: akara.FieldString}, {Name: "name", Kind: akara.FieldString}, {Name: "class", Kind: akara.FieldString}}},
		{Name: "dm.player.progress", Version: 1, Fields: []akara.Field{{Name: "level", Kind: akara.FieldInt64}, {Name: "experience", Kind: akara.FieldInt64}}},
		{Name: "dm.player.vitals", Version: 1, Fields: []akara.Field{{Name: "health", Kind: akara.FieldInt64}, {Name: "max_health", Kind: akara.FieldInt64}, {Name: "mana", Kind: akara.FieldInt64}, {Name: "max_mana", Kind: akara.FieldInt64}}},
		{Name: "dm.player.appearance", Version: 1, Fields: []akara.Field{{Name: "cof", Kind: akara.FieldString}, {Name: "palette", Kind: akara.FieldString}, {Name: "direction", Kind: akara.FieldInt64}}},
		{Name: "dm.world.position", Version: 1, Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}},
		{Name: "dm.world.velocity", Version: 1, Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}},
		{Name: "dm.player.movement_mode", Version: 1, Fields: []akara.Field{{Name: "running", Kind: akara.FieldBool}}},
		{Name: "dm.player.skill_assignment", Version: 1, Fields: []akara.Field{{Name: "left", Kind: akara.FieldInt64}, {Name: "right", Kind: akara.FieldInt64}}},
		{Name: "dm.player.belt", Version: 1, Fields: beltFields()},
		{Name: "dm.world.player_control", Version: 1, Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}},
		{Name: "dm.world.bounds", Version: 1, Fields: []akara.Field{{Name: "width", Kind: akara.FieldFloat64}, {Name: "height", Kind: akara.FieldFloat64}}},
	}
	registered := make([]*akara.DynamicStore, len(schemas))
	for index, schema := range schemas {
		store, err := akara.RegisterSchema(world, schema)
		if err != nil {
			return stores{}, err
		}
		registered[index] = store
	}
	return stores{identity: registered[0], progress: registered[1], vitals: registered[2], appearance: registered[3], position: registered[4], velocity: registered[5], movementMode: registered[6], skillAssignment: registered[7], belt: registered[8], control: registered[9], bounds: registered[10]}, nil
}

func beltFields() []akara.Field {
	fields := []akara.Field{{Name: "capacity", Kind: akara.FieldInt64}}
	for slot := 1; slot <= 16; slot++ {
		fields = append(fields, akara.Field{Name: fmt.Sprintf("slot_%d", slot), Kind: akara.FieldString})
	}
	return fields
}

func decodeEntry(encoded []byte) (Entry, error) {
	var entry Entry
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&entry); err != nil {
		return Entry{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Entry{}, fmt.Errorf("player: entry payload has trailing data")
	}
	entry.CharacterID, entry.Player, entry.Name, entry.Class = strings.TrimSpace(entry.CharacterID), strings.TrimSpace(entry.Player), strings.TrimSpace(entry.Name), strings.TrimSpace(entry.Class)
	if entry.CharacterID == "" || entry.Player == "" || entry.Name == "" || entry.Class == "" {
		return Entry{}, fmt.Errorf("player: character ID, player, name, and class are required")
	}
	if entry.Level < 1 || entry.Health < 0 || entry.MaxHealth < entry.Health || entry.Mana < 0 || entry.MaxMana < entry.Mana {
		return Entry{}, fmt.Errorf("player: invalid progression or vitals")
	}
	if entry.WorldWidth <= 0 || entry.WorldHeight <= 0 || entry.X < 0 || entry.X > entry.WorldWidth || entry.Y < 0 || entry.Y > entry.WorldHeight {
		return Entry{}, fmt.Errorf("player: invalid world position or bounds")
	}
	return entry, nil
}
