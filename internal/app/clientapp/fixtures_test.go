package clientapp

import "testing"

func TestFixtureNeedsSelectionForPlayableScenes(t *testing.T) {
	t.Parallel()

	for _, scene := range []string{"game_world", "game_loading", "combat_lab", "warp_lab"} {
		if !fixtureNeedsSelection(scene) {
			t.Errorf("fixtureNeedsSelection(%q) = false; playable scenes need an admitted player", scene)
		}
	}
}

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

func TestDirectGameplayFixtureActivatesOfflineSession(t *testing.T) {
	t.Parallel()

	if !shouldActivateDevelopmentSession(Options{StartScene: "game_world", FixtureCharacters: 1}) {
		t.Fatal("direct game-world fixture did not request local-session activation")
	}
	if shouldActivateDevelopmentSession(Options{StartScene: "main_menu", FixtureCharacters: 1}) {
		t.Fatal("frontend fixture must not activate a local gameplay session")
	}
}

func TestFixtureDoesNotSelectCharacterForFrontendLab(t *testing.T) {
	t.Parallel()

	if fixtureNeedsSelection("font_lab") {
		t.Fatal("frontend-only labs must not silently select a development character")
	}
}

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

func TestOrdinarySceneKeepsNormalWorldDefault(t *testing.T) {
	t.Parallel()

	options := applyDevelopmentSceneDefaults(Options{StartScene: "game_world"})
	if options.FixtureCharacters != 0 || options.FixtureWorldLevel != 0 {
		t.Fatalf("ordinary scene acquired lab fixtures: %+v", options)
	}
}
