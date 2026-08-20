package player

import (
	"fmt"
	"strings"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

const localEntryActor = "system:local-player-entry"

// EntrySource admits the currently selected save into the authoritative world.
// Selection remains shell state; after admission, ECS components become live
// gameplay state and Lua only observes them.
type EntrySource struct {
	engine      *gameecs.Engine
	saves       *d2save.Store
	player      string
	destination Destination
	sequence    uint64
}

// NewEntrySource creates an offline source whose default spawn is the center
// of the provided world. Remote sessions receive equivalent commands elsewhere.
func NewEntrySource(
	engine *gameecs.Engine,
	saves *d2save.Store,
	player string,
	width, height float64,
) (*EntrySource, error) {
	return NewEntrySourceAt(engine, saves, player, width/2, height/2, width, height)
}

// NewEntrySourceAt creates an offline source at a trusted coordinate in the
// default act and level, preserving the legacy constructor's location policy.
func NewEntrySourceAt(
	engine *gameecs.Engine,
	saves *d2save.Store,
	player string,
	x, y, width, height float64,
) (*EntrySource, error) {
	return NewEntrySourceAtLocation(engine, saves, player, x, y, width, height, 1, 1)
}

// NewEntrySourceAtLocation records the server-selected act and town level in
// the same authoritative command as the server-selected spawn coordinate.
func NewEntrySourceAtLocation(
	engine *gameecs.Engine,
	saves *d2save.Store,
	player string,
	x, y, width, height float64,
	act, levelID int64,
) (*EntrySource, error) {
	destination, err := NewDestination(x, y, width, height, act, levelID)
	if err != nil {
		return nil, err
	}

	return NewEntrySourceForDestination(engine, saves, player, destination)
}

// NewEntrySourceForDestination adapts local save selection to the same trusted
// destination contract used by remote admission.
func NewEntrySourceForDestination(
	engine *gameecs.Engine,
	saves *d2save.Store,
	player string,
	destination Destination,
) (*EntrySource, error) {
	player = strings.TrimSpace(player)

	validated, err := NewDestination(
		destination.X,
		destination.Y,
		destination.Width,
		destination.Height,
		destination.Act,
		destination.LevelID,
	)
	if err != nil || engine == nil || saves == nil || player == "" {
		return nil, fmt.Errorf("player: entry source requires engine, saves, player, and positive world bounds")
	}

	return &EntrySource{engine: engine, saves: saves, player: player, destination: validated}, nil
}

// Commands emits entry intent once; it never materializes ECS state directly.
// A selected character already present in ECS produces no duplicate command.
func (source *EntrySource) Commands(tick uint64) []simulation.Command {
	character, selected := source.saves.Selected()
	if !selected || source.entered(character.ID) {
		return nil
	}

	source.sequence++

	command, err := AdmissionCommand(
		character,
		source.player,
		source.destination,
		localEntryActor,
		source.sequence,
		tick,
		simulation.AuthoritySystem,
	)
	if err != nil {
		return nil
	}

	return []simulation.Command{command}
}

// entered reports whether the durable character already has a live identity.
// Character identity, rather than connection identity, is the admission guard.
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
