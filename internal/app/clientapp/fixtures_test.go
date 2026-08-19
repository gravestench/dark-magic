package clientapp

import "testing"

// TestFixtureNeedsSelectionForPlayableScenes ensures direct gameplay cannot start without an admitted fixture player.
func TestFixtureNeedsSelectionForPlayableScenes(t *testing.T) {
	t.Parallel()

	for _, scene := range []string{"game_world", "game_loading", "combat_lab", "spell_lab", "warp_lab"} {
		if !fixtureNeedsSelection(scene) {
			t.Errorf("fixtureNeedsSelection(%q) = false; playable scenes need an admitted player", scene)
		}
	}
}

// TestSpellLabSuppliesProductionSpellFixture proves the lab changes disposable admission data, not gameplay systems.
func TestSpellLabSuppliesProductionSpellFixture(t *testing.T) {
	t.Parallel()

	options := applyDevelopmentSceneDefaults(Options{StartScene: "spell_lab"})
	if options.FixtureCharacters != 1 || options.FixtureWorldLevel != 2 {
		t.Fatalf("Spell Lab defaults = %+v, want one player in Blood Moor", options)
	}

	if !shouldActivateDevelopmentSession(options) || !developmentGameplayScene("spell_lab") {
		t.Fatal("Spell Lab must activate the production gameplay session")
	}
	characters := developmentCharactersForScene("spell_lab", options.FixtureCharacters)
	if len(characters) != 1 || characters[0].Class != "Sorceress" || characters[0].Level != 30 {
		t.Fatalf("Spell Lab character = %+v", characters)
	}
	if characters[0].Stats.Mana != 4096 || characters[0].Stats.MaxMana != 4096 {
		t.Fatalf("Spell Lab mana = %+v", characters[0].Stats)
	}

	app := &application{options: options}
	skills := app.developmentSkillsBootstrapData()
	if skills["all_implemented"] != true || skills["skill_ids"] != nil {
		t.Fatalf("Spell Lab skill source = %#v, want the target-locked implementation manifest", skills)
	}
}

// TestWarpLabSuppliesProductionTransitionState keeps the lab on production transition and gameplay-input paths.
func TestWarpLabSuppliesProductionTransitionState(t *testing.T) {
	t.Parallel()

	options := applyDevelopmentSceneDefaults(Options{StartScene: "warp_lab"})
	if options.FixtureCharacters != 1 || options.FixtureWorldLevel != 1 || options.FixtureWorldSpawn != "entry" {
		t.Fatalf("Warp Lab defaults = %+v, want one player at the town-side warp fixture", options)
	}
	if !shouldActivateDevelopmentSession(options) {
		t.Fatal("Warp Lab must activate its direct-start offline session")
	}
	if !developmentGameplayScene("warp_lab") {
		t.Fatal("Warp Lab must receive routed gameplay input")
	}
}

// TestDirectGameplayFixtureActivatesOfflineSession distinguishes playable fixtures from frontend-only scenes.
func TestDirectGameplayFixtureActivatesOfflineSession(t *testing.T) {
	t.Parallel()

	if !shouldActivateDevelopmentSession(Options{StartScene: "game_world", FixtureCharacters: 1}) {
		t.Fatal("direct game-world fixture did not request local-session activation")
	}
	if shouldActivateDevelopmentSession(Options{StartScene: "main_menu", FixtureCharacters: 1}) {
		t.Fatal("frontend fixture must not activate a local gameplay session")
	}
}

// TestFixtureDoesNotSelectCharacterForFrontendLab prevents visual labs from acquiring gameplay save ownership.
func TestFixtureDoesNotSelectCharacterForFrontendLab(t *testing.T) {
	t.Parallel()

	if fixtureNeedsSelection("font_lab") {
		t.Fatal("frontend-only labs must not silently select a development character")
	}
}

// TestCombatLabSuppliesItsOwnDevelopmentState ensures direct launch reaches the intended production wilderness setup.
func TestCombatLabSuppliesItsOwnDevelopmentState(t *testing.T) {
	t.Parallel()

	options := applyDevelopmentSceneDefaults(Options{StartScene: "combat_lab"})
	if options.FixtureCharacters != 1 {
		t.Fatalf("FixtureCharacters = %d, want 1", options.FixtureCharacters)
	}
	if options.FixtureWorldLevel != 2 {
		t.Fatalf("FixtureWorldLevel = %d, want Blood Moor level 2", options.FixtureWorldLevel)
	}
}

// TestCombatLabDoesNotReplaceExplicitDevelopmentState preserves caller intent over convenience defaults.
func TestCombatLabDoesNotReplaceExplicitDevelopmentState(t *testing.T) {
	t.Parallel()

	options := applyDevelopmentSceneDefaults(Options{
		StartScene:        "combat_lab",
		FixtureCharacters: 3,
		FixtureWorldLevel: 1,
	})
	if options.FixtureCharacters != 3 || options.FixtureWorldLevel != 1 {
		t.Fatalf("explicit fixture options were replaced: %+v", options)
	}
}

// TestOrdinarySceneKeepsNormalWorldDefault prevents laboratory policy from leaking into regular client startup.
func TestOrdinarySceneKeepsNormalWorldDefault(t *testing.T) {
	t.Parallel()

	options := applyDevelopmentSceneDefaults(Options{StartScene: "game_world"})
	if options.FixtureCharacters != 0 || options.FixtureWorldLevel != 0 {
		t.Fatalf("ordinary scene acquired lab fixtures: %+v", options)
	}
}
