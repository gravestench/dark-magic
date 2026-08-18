package clientapp

import d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"

// developmentSceneDefaults describes only the disposable state a laboratory
// needs before its real scene code can run. Keeping this policy beside the
// composition root prevents labs from creating fake saves or swapping maps in
// Lua, and prevents cmd/client from growing one special case per lab.
type developmentSceneDefaults struct {
	characters     int
	worldLevel     int
	nearbyHostiles int
	gameplay       bool
	characterClass string
	characterLevel int
	mana           int
}

var developmentScenes = map[string]developmentSceneDefaults{
	// Combat Lab is the production world plus diagnostics. It therefore needs
	// an admitted hero and starts in Blood Moor, where the production monster
	// population and combat systems are active.
	"combat_lab": {characters: 1, worldLevel: 2, nearbyHostiles: 3, gameplay: true},
	// Spell Lab uses the production Blood Moor session and HUD with every exact-ID
	// behavior currently admitted as an ephemeral learned skill. A deep mana pool
	// keeps repeated visual inspection from depending on unimplemented regeneration.
	"spell_lab": {
		characters: 1, worldLevel: 2, nearbyHostiles: 3, gameplay: true,
		characterClass: "Sorceress", characterLevel: 30, mana: 4096,
	},
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

func developmentCharactersForScene(scene string, count int) []d2save.Character {
	characters := DevelopmentCharacters(count)
	defaults := developmentScenes[scene]
	if len(characters) == 0 {
		return characters
	}
	if defaults.characterClass != "" {
		characters[0].Class = defaults.characterClass
	}
	if defaults.characterLevel > 0 {
		characters[0].Level = defaults.characterLevel
	}
	if defaults.mana > 0 {
		characters[0].Stats.Mana = defaults.mana
		characters[0].Stats.MaxMana = defaults.mana
	}
	return characters
}

func shouldActivateDevelopmentSession(options Options) bool {
	return options.StartScene != "" && options.FixtureCharacters > 0 &&
		(fixtureNeedsSelection(options.StartScene) || developmentGameplayScene(options.StartScene))
}

func (app *application) developmentSkillsBootstrapData() map[string]any {
	if app.options.StartScene != "spell_lab" {
		return map[string]any{"enabled": false}
	}
	return map[string]any{
		"enabled": true,
		"replace": true,
		"skill_ids": []any{
			float64(0), float64(36), float64(40), float64(45), float64(47), float64(48),
			float64(52), float64(54), float64(55), float64(66), float64(72), float64(98),
		},
		"skill_level": float64(20),
		"left":        float64(36),
		"right":       float64(66),
	}
}
