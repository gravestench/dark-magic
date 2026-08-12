package clientapp

import ()

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
