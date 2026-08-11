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
	"github.com/gravestench/dark-magic/internal/game/targeting"
	"github.com/gravestench/dark-magic/internal/persistence"
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
	Health      int64   `json:"health"`
	MaxHealth   int64   `json:"max_health"`
	Mana        int64   `json:"mana"`
	MaxMana     int64   `json:"max_mana"`
	Expansion   bool    `json:"expansion"`
	Hardcore    bool    `json:"hardcore"`
	COF         string  `json:"cof"`
	Token       string  `json:"token"`
	Palette     string  `json:"palette"`
	Direction   int64   `json:"direction"`
	Mode        string  `json:"mode"`
	WeaponClass string  `json:"weapon_class"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	WorldWidth  float64 `json:"world_width"`
	WorldHeight float64 `json:"world_height"`
	Act         int64   `json:"act"`
	LevelID     int64   `json:"level_id"`
	Skills      []Skill `json:"skills,omitempty"`
}

// Skill is one learned action admitted with the character. Presentation may
// describe it, but only this session-owned record decides whether it is usable.
type Skill struct {
	ID           int64 `json:"id"`
	Level        int64 `json:"level"`
	ListRow      int64 `json:"list_row"`
	LeftAllowed  bool  `json:"left_allowed"`
	RightAllowed bool  `json:"right_allowed"`
}

// Destination is the server-selected place where an admitted character enters
// a game. Keeping this separate from a local save means offline selection and a
// remote realm join can share exactly the same trusted town-spawn policy.
type Destination struct {
	X, Y, Width, Height float64
	Act, LevelID        int64
}

// PlayerColliderRadius is one gameplay subtile. Riiablo models players as
// Size.MEDIUM (diameter 2) and uses size/2 for the physics circle. A map tile is
// five subtiles, so this is deliberately much smaller than the rendered hero.
const PlayerColliderRadius = 1.0

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
func AdmissionCommand(character persistence.Character, player string, destination Destination, skills []Skill, actor string, sequence, tick uint64, authority simulation.Authority) (simulation.Command, error) {
	if authority != simulation.AuthoritySystem && authority != simulation.AuthorityAdmin {
		return simulation.Command{}, fmt.Errorf("player: admission requires system or admin authority")
	}
	validated, err := NewDestination(destination.X, destination.Y, destination.Width, destination.Height, destination.Act, destination.LevelID)
	if err != nil {
		return simulation.Command{}, err
	}
	entry := EntryFromCharacter(character, player, validated.X, validated.Y, validated.Width, validated.Height)
	entry.Act, entry.LevelID, entry.Skills = validated.Act, validated.LevelID, append([]Skill(nil), skills...)
	return Command(entry, actor, sequence, tick, authority)
}

// SkillProvider translates durable character knowledge into authoritative entry
// facts without coupling player materialization to the typed data catalog.
type SkillProvider func(persistence.Character) []Skill

// EntrySource admits the currently selected save into the authoritative world.
// Selection remains shell state; after admission, ECS components are the live
// gameplay state and Lua only observes them.
type EntrySource struct {
	engine      *gameecs.Engine
	saves       *persistence.Store
	player      string
	destination Destination
	sequence    uint64
	skills      SkillProvider
}

// NewEntrySource creates the offline adapter that emits one entry command for
// the selected character. Remote sessions receive equivalent commands elsewhere.
func NewEntrySource(engine *gameecs.Engine, saves *persistence.Store, player string, width, height float64, skills SkillProvider) (*EntrySource, error) {
	return NewEntrySourceAt(engine, saves, player, width/2, height/2, width, height, skills)
}

// NewEntrySourceAt admits the player at a server-selected map coordinate.
func NewEntrySourceAt(engine *gameecs.Engine, saves *persistence.Store, player string, x, y, width, height float64, skills SkillProvider) (*EntrySource, error) {
	return NewEntrySourceAtLocation(engine, saves, player, x, y, width, height, 1, 1, skills)
}

// NewEntrySourceAtLocation records the server-selected act and town level in
// the same authoritative command as the server-selected spawn coordinate.
func NewEntrySourceAtLocation(engine *gameecs.Engine, saves *persistence.Store, player string, x, y, width, height float64, act, levelID int64, skills SkillProvider) (*EntrySource, error) {
	destination, err := NewDestination(x, y, width, height, act, levelID)
	if err != nil {
		return nil, err
	}
	return NewEntrySourceForDestination(engine, saves, player, destination, skills)
}

// NewEntrySourceForDestination adapts local save selection to the same
// destination contract used by trusted remote admission.
func NewEntrySourceForDestination(engine *gameecs.Engine, saves *persistence.Store, player string, destination Destination, skills SkillProvider) (*EntrySource, error) {
	player = strings.TrimSpace(player)
	validated, err := NewDestination(destination.X, destination.Y, destination.Width, destination.Height, destination.Act, destination.LevelID)
	if err != nil || engine == nil || saves == nil || player == "" {
		return nil, fmt.Errorf("player: entry source requires engine, saves, player, and positive world bounds")
	}
	return &EntrySource{engine: engine, saves: saves, player: player, destination: validated, skills: skills}, nil
}

// Commands emits entry intent once; it never materializes ECS state directly.
func (source *EntrySource) Commands(tick uint64) []simulation.Command {
	character, selected := source.saves.Selected()
	if !selected || source.entered(character.ID) {
		return nil
	}
	source.sequence++
	var skills []Skill
	if source.skills != nil {
		skills = source.skills(character)
	}
	command, err := AdmissionCommand(character, source.player, source.destination, skills, localEntryActor, source.sequence, tick, simulation.AuthoritySystem)
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

// EntryFromCharacter copies the admitted durable subset into a command value.
func EntryFromCharacter(character persistence.Character, player string, x, y, width, height float64) Entry {
	entry := Entry{CharacterID: character.ID, Player: player, Name: character.Name, Class: character.Class, Level: int64(character.Level), Expansion: character.Expansion, Hardcore: character.Hardcore, Token: classToken(character.Class), Palette: "data/global/Palette/units/pal.dat", Direction: 0, Mode: "NU", WeaponClass: "HTH", X: x, Y: y, WorldWidth: width, WorldHeight: height, Act: 1, LevelID: 1}
	if character.Stats != nil {
		entry.Experience = int64(character.Stats.Experience)
		entry.Health, entry.MaxHealth = int64(character.Stats.Health), int64(character.Stats.MaxHealth)
		entry.Mana, entry.MaxMana = int64(character.Stats.Mana), int64(character.Stats.MaxMana)
	}
	if character.Appearance != nil {
		entry.COF, entry.Direction = character.Appearance.COF, int64(character.Appearance.Direction)
		if character.Appearance.Palette != "" {
			entry.Palette = character.Appearance.Palette
		}
	}
	return entry
}

// classToken translates the durable player-facing class name into Diablo II's
// two-letter composite namespace. This is simulation seed data, not a filename
// guessed by the renderer.
func classToken(class string) string {
	switch strings.ToLower(strings.TrimSpace(class)) {
	case "amazon":
		return "AM"
	case "sorceress":
		return "SO"
	case "necromancer":
		return "NE"
	case "paladin":
		return "PA"
	case "barbarian":
		return "BA"
	case "assassin":
		return "AI"
	case "druid":
		return "DZ"
	default:
		return ""
	}
}

// Command encodes player-entry intent for deterministic admission and replay.
func Command(entry Entry, actor string, sequence, tick uint64, authority simulation.Authority) (simulation.Command, error) {
	payload, err := json.Marshal(entry)
	if err != nil {
		return simulation.Command{}, err
	}
	return simulation.Command{Tick: tick, Player: actor, Authority: authority, Sequence: sequence, Kind: EnterCommand, Payload: payload}, nil
}

// Register installs the trusted player-materialization handler.
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
	leftSkill, rightSkill := int64(0), int64(0)
	for _, skill := range entry.Skills {
		if leftSkill == 0 && skill.LeftAllowed {
			leftSkill = skill.ID
		}
		if rightSkill == 0 && skill.RightAllowed {
			rightSkill = skill.ID
		}
	}
	components := []struct {
		store  *akara.DynamicStore
		values map[string]any
	}{
		{stores.identity, map[string]any{"character_id": entry.CharacterID, "player": entry.Player, "name": entry.Name, "class": entry.Class}},
		{stores.progress, map[string]any{"level": entry.Level, "experience": entry.Experience}},
		{stores.vitals, map[string]any{"health": entry.Health, "max_health": entry.MaxHealth, "mana": entry.Mana, "max_mana": entry.MaxMana}},
		{stores.appearance, map[string]any{"cof": entry.COF, "token": entry.Token, "palette": entry.Palette, "weapon_class": entry.WeaponClass}},
		{stores.animation, map[string]any{"direction": entry.Direction, "mode": entry.Mode}},
		{stores.position, map[string]any{"x": entry.X, "y": entry.Y}},
		{stores.velocity, nil},
		{stores.movementMode, map[string]any{"running": false}},
		{stores.skillAssignment, map[string]any{"left": leftSkill, "right": rightSkill}},
		{stores.skillIntent, map[string]any{"side": "", "skill_id": int64(0), "target_x": entry.X, "target_y": entry.Y, "target_id": ""}},
		{stores.belt, map[string]any{"capacity": int64(4)}},
		{stores.control, map[string]any{"player": entry.Player}},
		{stores.bounds, map[string]any{"width": entry.WorldWidth, "height": entry.WorldHeight}},
		{stores.location, map[string]any{"act": entry.Act, "level_id": entry.LevelID}},
		{stores.collider, map[string]any{"radius": PlayerColliderRadius}},
		{stores.selectable, map[string]any{"id": "player:" + entry.Player, "kind": targeting.KindPlayer, "label": entry.Name, "owner": entry.Player, "radius": 0.75, "priority": int64(10)}},
	}
	for _, component := range components {
		if _, err := component.store.Set(entity, component.values); err != nil {
			return fail(err)
		}
	}
	if err := materializeSkills(engine.World(), entity, entry.Skills); err != nil {
		return fail(err)
	}
	return nil
}

func materializeSkills(world *akara.World, owner akara.Entity, skills []Skill) error {
	store, err := akara.RegisterSchema(world, akara.Schema{Name: "dm.player.learned_skill", Version: 1, Fields: []akara.Field{
		{Name: "owner", Kind: akara.FieldEntity}, {Name: "skill_id", Kind: akara.FieldInt64},
		{Name: "level", Kind: akara.FieldInt64}, {Name: "list_row", Kind: akara.FieldInt64},
		{Name: "left_allowed", Kind: akara.FieldBool}, {Name: "right_allowed", Kind: akara.FieldBool},
	}})
	if err != nil {
		return err
	}
	created := make([]akara.Entity, 0, len(skills))
	rollback := func() {
		for _, entity := range created {
			world.DestroyEntity(entity)
		}
	}
	for _, skill := range skills {
		entity, err := world.CreateEntity()
		if err != nil {
			rollback()
			return err
		}
		created = append(created, entity)
		if _, err := store.Set(entity, map[string]any{"owner": owner, "skill_id": skill.ID, "level": skill.Level, "list_row": skill.ListRow, "left_allowed": skill.LeftAllowed, "right_allowed": skill.RightAllowed}); err != nil {
			rollback()
			return err
		}
	}
	return nil
}

type stores struct {
	identity, progress, vitals, appearance, animation                                                                     *akara.DynamicStore
	position, velocity, movementMode, skillAssignment, skillIntent, belt, control, bounds, location, collider, selectable *akara.DynamicStore
}

func registerStores(world *akara.World) (stores, error) {
	schemas := []akara.Schema{
		{Name: "dm.player.identity", Version: 1, Fields: []akara.Field{{Name: "character_id", Kind: akara.FieldString}, {Name: "player", Kind: akara.FieldString}, {Name: "name", Kind: akara.FieldString}, {Name: "class", Kind: akara.FieldString}}},
		{Name: "dm.player.progress", Version: 1, Fields: []akara.Field{{Name: "level", Kind: akara.FieldInt64}, {Name: "experience", Kind: akara.FieldInt64}}},
		{Name: "dm.player.vitals", Version: 1, Fields: []akara.Field{{Name: "health", Kind: akara.FieldInt64}, {Name: "max_health", Kind: akara.FieldInt64}, {Name: "mana", Kind: akara.FieldInt64}, {Name: "max_mana", Kind: akara.FieldInt64}}},
		{Name: "dm.player.appearance", Version: 1, Fields: []akara.Field{{Name: "cof", Kind: akara.FieldString}, {Name: "token", Kind: akara.FieldString}, {Name: "palette", Kind: akara.FieldString}, {Name: "weapon_class", Kind: akara.FieldString}}},
		{Name: "dm.player.animation", Version: 1, Fields: []akara.Field{{Name: "direction", Kind: akara.FieldInt64}, {Name: "mode", Kind: akara.FieldString}}},
		{Name: "dm.world.position", Version: 1, Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}},
		{Name: "dm.world.velocity", Version: 1, Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}},
		{Name: "dm.player.movement_mode", Version: 1, Fields: []akara.Field{{Name: "running", Kind: akara.FieldBool}}},
		{Name: "dm.player.skill_assignment", Version: 1, Fields: []akara.Field{{Name: "left", Kind: akara.FieldInt64}, {Name: "right", Kind: akara.FieldInt64}}},
		{Name: "dm.player.skill_intent", Version: 1, Fields: []akara.Field{{Name: "side", Kind: akara.FieldString}, {Name: "skill_id", Kind: akara.FieldInt64}, {Name: "target_x", Kind: akara.FieldFloat64}, {Name: "target_y", Kind: akara.FieldFloat64}, {Name: "target_id", Kind: akara.FieldString}}},
		{Name: "dm.player.belt", Version: 1, Fields: beltFields()},
		{Name: "dm.world.player_control", Version: 1, Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}},
		{Name: "dm.world.bounds", Version: 1, Fields: []akara.Field{{Name: "width", Kind: akara.FieldFloat64}, {Name: "height", Kind: akara.FieldFloat64}}},
		{Name: "dm.world.location", Version: 1, Fields: []akara.Field{{Name: "act", Kind: akara.FieldInt64}, {Name: "level_id", Kind: akara.FieldInt64}}},
		{Name: "dm.world.collider", Version: 1, Fields: []akara.Field{{Name: "radius", Kind: akara.FieldFloat64}}},
		targeting.Schema(),
	}
	registered := make([]*akara.DynamicStore, len(schemas))
	for index, schema := range schemas {
		store, err := akara.RegisterSchema(world, schema)
		if err != nil {
			return stores{}, err
		}
		registered[index] = store
	}
	return stores{identity: registered[0], progress: registered[1], vitals: registered[2], appearance: registered[3], animation: registered[4], position: registered[5], velocity: registered[6], movementMode: registered[7], skillAssignment: registered[8], skillIntent: registered[9], belt: registered[10], control: registered[11], bounds: registered[12], location: registered[13], collider: registered[14], selectable: registered[15]}, nil
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
	entry.Token, entry.Palette, entry.Mode, entry.WeaponClass = strings.ToUpper(strings.TrimSpace(entry.Token)), strings.TrimSpace(entry.Palette), strings.ToUpper(strings.TrimSpace(entry.Mode)), strings.ToUpper(strings.TrimSpace(entry.WeaponClass))
	if entry.CharacterID == "" || entry.Player == "" || entry.Name == "" || entry.Class == "" {
		return Entry{}, fmt.Errorf("player: character ID, player, name, and class are required")
	}
	if entry.Level < 1 || entry.Health < 0 || entry.MaxHealth < entry.Health || entry.Mana < 0 || entry.MaxMana < entry.Mana {
		return Entry{}, fmt.Errorf("player: invalid progression or vitals")
	}
	if len(entry.Token) != 2 || entry.Palette == "" || entry.Mode != "NU" || entry.WeaponClass != "HTH" || entry.Direction < 0 || entry.Direction > 7 {
		return Entry{}, fmt.Errorf("player: invalid initial composite appearance")
	}
	for _, skill := range entry.Skills {
		if skill.ID < 0 || skill.Level < 1 || skill.ListRow < 0 || !skill.LeftAllowed && !skill.RightAllowed {
			return Entry{}, fmt.Errorf("player: invalid learned skill")
		}
	}
	if entry.WorldWidth <= 0 || entry.WorldHeight <= 0 || entry.X < 0 || entry.X >= entry.WorldWidth || entry.Y < 0 || entry.Y >= entry.WorldHeight {
		return Entry{}, fmt.Errorf("player: invalid world position or bounds")
	}
	if entry.Act < 1 || entry.Act > 5 || entry.LevelID <= 0 {
		return Entry{}, fmt.Errorf("player: entry act and level are invalid")
	}
	return entry, nil
}
