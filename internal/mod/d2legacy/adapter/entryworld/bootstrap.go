package entryworld

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gravestench/dark-magic/internal/game/simulation"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// InitialData builds the authoritative state fragments consumed during session creation. Deriving interactions and
// transitions from this Prepared value keeps initial Lua state aligned with the maps admitted by the authority.
func (world *Prepared) InitialData(owner string, developmentItems bool) map[string]any {
	return map[string]any{
		"d2legacy.game_rules": map[string]any{
			"target":          "lod-1.14d",
			"expansion":       true,
			"difficulty":      world.Difficulty,
			"hardcore":        false,
			"ladder":          false,
			"maximum_players": 8,
		},
		"d2legacy.development_items": map[string]any{
			"enabled":                 developmentItems,
			"create_empty_containers": !developmentItems,
		},
		"d2legacy.interactions":      InteractionData(world.Worlds, world.Zones, owner, ""),
		"d2legacy.world_transitions": TransitionData(world.Seam),
	}
}

// TransitionData serializes both directions of the prepared seam. Explicit reverse data lets Lua perform transitions
// without inferring that every topology is symmetric.
func TransitionData(seam gametransition.Seam) map[string]any {
	return map[string]any{
		"seams": []any{
			transitionEndpoint(seam.Town, seam.Wilderness),
			transitionEndpoint(seam.Wilderness, seam.Town),
		},
	}
}

// transitionEndpoint converts one directed seam into the stable Lua field contract. Destination dimensions accompany
// arrival coordinates so player placement can be validated against the world being entered.
func transitionEndpoint(source, destination gametransition.SeamEndpoint) map[string]any {
	return map[string]any{
		"source_level":      float64(source.LevelID),
		"destination_level": float64(destination.LevelID),
		"source_x":          source.X,
		"source_y":          source.Y,
		"arrival_x":         destination.ArrivalX,
		"arrival_y":         destination.ArrivalY,
		"world_width":       destination.Width,
		"world_height":      destination.Height,
	}
}

// PopulationCommand encodes the generated wilderness topology as the first system-owned population command. Tick and
// sequence remain fixed because this bootstrap must precede player-authored simulation input deterministically.
func (world *Prepared) PopulationCommand(nearby int) (simulation.Command, error) {
	payload, err := json.Marshal(world.PopulationData(nearby))
	if err != nil {
		return simulation.Command{}, err
	}

	return simulation.Command{
		Tick:      1,
		Player:    "d2legacy.population",
		Authority: simulation.AuthoritySystem,
		Sequence:  1,
		Kind:      "system.population.bootstrap",
		Payload:   payload,
	}, nil
}

// InstallCollision publishes every prepared map to the Lua runtime before gameplay systems initialize. Returning the
// first error avoids starting an authority with only a subset of level collision installed.
func (world *Prepared) InstallCollision(ctx context.Context, runtime *modruntime.Runtime) error {
	if world == nil || runtime == nil {
		return errors.New("d2legacy entry world: prepared world and runtime are required")
	}

	for levelID, collision := range world.Worlds {
		err := modruntime.SetWorldMapForLevel(
			ctx,
			runtime,
			"d2legacy.gameplay.systems.init",
			"set_collision",
			levelID,
			collision,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
