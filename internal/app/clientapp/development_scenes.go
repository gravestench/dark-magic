package clientapp

import (
	"fmt"
	"math"

	gamemonster "github.com/gravestench/dark-magic/internal/game/monster"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

// developmentSceneDefaults describes only the disposable state a laboratory
// needs before its real scene code can run. Keeping this policy beside the
// composition root prevents labs from creating fake saves or swapping maps in
// Lua, and prevents cmd/darkmagic from growing one special case per lab.
type developmentSceneDefaults struct {
	characters     int
	worldLevel     int
	nearbyHostiles int
}

var developmentScenes = map[string]developmentSceneDefaults{
	// Combat Lab is the production world plus diagnostics. It therefore needs
	// an admitted hero and starts in Blood Moor, where the production monster
	// population and combat systems are active.
	"combat_lab": {characters: 1, worldLevel: 2, nearbyHostiles: 3},
}

type developmentEncounterPlacement interface {
	OpenPointNearSubtile(x, y float64) (float64, float64, bool)
	FindPath(gameworld.PathRequest) ([]gameworld.Point, error)
}

// placeDevelopmentEncounter moves only a small prefix of an already-authored
// population plan. Definitions, stats, AI, spawn authority, and combat remain
// production-owned; the lab changes only where its first test subjects stand.
func placeDevelopmentEncounter(plan gamemonster.PopulationPlan, world developmentEncounterPlacement, player [2]float64, count int) (gamemonster.PopulationPlan, error) {
	if count <= 0 || len(plan.Spawns) == 0 {
		return plan, nil
	}
	offsets := [][2]float64{{10, 0}, {7, 7}, {0, 10}, {-7, 7}, {-10, 0}, {-7, -7}, {0, -10}, {7, -7}}
	placed := 0
	for _, offset := range offsets {
		if placed >= count || placed >= len(plan.Spawns) {
			break
		}
		x, y, found := world.OpenPointNearSubtile(player[0]+offset[0], player[1]+offset[1])
		if !found || math.Hypot(x-player[0], y-player[1]) > 14 {
			continue
		}
		path, err := world.FindPath(gameworld.PathRequest{
			Start: gameworld.Point{X: player[0], Y: player[1]},
			Goal:  gameworld.Point{X: x, Y: y}, Radius: 1, StopRadius: 2.5,
		})
		if err != nil || len(path) < 2 {
			continue
		}
		plan.Spawns[placed].X, plan.Spawns[placed].Y = x, y
		placed++
	}
	if placed == 0 {
		return plan, fmt.Errorf("combat lab: no nearby reachable hostile placement")
	}
	plan.Trace = append(plan.Trace, fmt.Sprintf("combat-lab nearby-hostiles=%d", placed))
	return plan, nil
}

func applyDevelopmentSceneDefaults(options Options) Options {
	defaults, ok := developmentScenes[options.StartScene]
	if ok {
		if options.FixtureCharacters == 0 {
			options.FixtureCharacters = defaults.characters
		}
		if options.FixtureWorldLevel == 0 {
			options.FixtureWorldLevel = defaults.worldLevel
		}
	}
	return options
}
