package clientapp

import ()

// developmentSceneDefaults describes only the disposable state a laboratory
// needs before its real scene code can run. Keeping this policy beside the
// composition root prevents labs from creating fake saves or swapping maps in
// Lua, and prevents cmd/client from growing one special case per lab.
type developmentSceneDefaults struct {
	characters     int
	worldLevel     int
	nearbyHostiles int
	gameplay       bool
}

var developmentScenes = map[string]developmentSceneDefaults{
	// Combat Lab is the production world plus diagnostics. It therefore needs
	// an admitted hero and starts in Blood Moor, where the production monster
	// population and combat systems are active.
	"combat_lab": {characters: 1, worldLevel: 2, nearbyHostiles: 3, gameplay: true},
	// Warp Lab delegates to the production game world and adds a configured pair
	// of authoritative town/wilderness warp entities plus read-only diagnostics.
	"warp_lab": {characters: 1, worldLevel: 1, gameplay: true},
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
	if options.FixtureWorldSpawn == "" {
		options.FixtureWorldSpawn = "entry"
	}
	return options
}

func developmentGameplayScene(scene string) bool {
	return developmentScenes[scene].gameplay
}

func shouldActivateDevelopmentSession(options Options) bool {
	return options.StartScene != "" && options.FixtureCharacters > 0 &&
		(fixtureNeedsSelection(options.StartScene) || developmentGameplayScene(options.StartScene))
}
