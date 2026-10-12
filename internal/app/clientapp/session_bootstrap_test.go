package clientapp

import "testing"

// TestSessionInitialDataUsesSelectedDifficulty protects the offline and hosted
// game rules from silently reverting to Normal, which would also change identity.
func TestSessionInitialDataUsesSelectedDifficulty(t *testing.T) {
	t.Parallel()

	for _, difficulty := range []int{0, 1, 2} {
		app := &application{options: Options{GameDifficulty: difficulty}}

		rules, ok := app.sessionInitialData()["d2legacy.game_rules"].(map[string]any)
		if !ok {
			t.Fatal("session initial data has no d2legacy.game_rules map")
		}

		if rules["difficulty"] != difficulty {
			t.Errorf("game_rules difficulty = %v, want %d", rules["difficulty"], difficulty)
		}
	}
}
