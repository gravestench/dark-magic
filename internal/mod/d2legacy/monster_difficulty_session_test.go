package d2legacy

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func difficultyMonsterRecords() fixtureRecords {
	records := generatedHostileRecords()
	area := records["data/global/excel/levels.txt"][0]
	area["NumMon"] = "1"
	area["nmon1"] = "fallen"
	area["MonDen(N)"] = "100000"
	area["MonDen(H)"] = "100000"
	area["MonLvl2Ex"] = "4"
	area["MonLvl3Ex"] = "5"

	stats := records["data/global/excel/monstats.txt"][0]
	stats["noRatio"] = "0"
	stats["Level"] = "1"
	stats["Level(N)"] = "2"
	stats["Level(H)"] = "3"
	for column, value := range map[string]string{
		"minHP": "100", "maxHP": "100", "AC": "100", "Exp": "100", "TreasureClass1": "normal-drop",
		"minHP(N)": "120", "maxHP(N)": "120", "AC(N)": "120", "Exp(N)": "150", "TreasureClass1(N)": "nightmare-drop",
		"minHP(H)": "140", "maxHP(H)": "140", "AC(H)": "150", "Exp(H)": "200", "TreasureClass1(H)": "hell-drop",
	} {
		stats[column] = value
	}
	records["data/global/excel/monlvl.txt"] = []map[string]string{
		{"Level": "1", "HP": "10", "AC": "20", "XP": "50"},
		{"Level": "4", "HP(N)": "20", "AC(N)": "30", "XP(N)": "80"},
		{"Level": "5", "HP(H)": "30", "AC(H)": "40", "XP(H)": "120"},
	}
	return records
}

func startDifficultyMonsterSession(t *testing.T, difficulty int) (*Authority, *gameecs.Engine, *gamesession.Session) {
	t.Helper()
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	authority, err := StartWithConfig(t.Context(), content.D2Legacy(), difficultyMonsterRecords(), engine, session,
		Config{Seed: 314, InitialData: map[string]any{
			"d2legacy.game_rules": map[string]any{
				"target": "lod-1.14d", "expansion": true, "difficulty": difficulty,
				"hardcore": false, "ladder": false, "maximum_players": 2,
			},
		}})
	if err != nil {
		_ = session.Close()
		_ = engine.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = authority.Stop(t.Context())
		_ = session.Close()
		_ = engine.Close()
	})
	return authority, engine, session
}

func difficultyPopulationPayload(t *testing.T, difficulty int) json.RawMessage {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"act": 1, "level_id": 2, "difficulty": difficulty,
		"links": []map[string]any{},
		"rooms": []map[string]any{{
			"id": "difficulty-room", "populate": true,
			"x": 0, "y": 0, "width": 10, "height": 10,
			"points": []map[string]any{{"x": 5, "y": 5}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

// TestGeneratedMonsterUsesSessionDifficulty verifies the immutable game's
// difficulty reaches area level, MonLvl scaling, count bonus, and loot class.
func TestGeneratedMonsterUsesSessionDifficulty(t *testing.T) {
	for _, difficulty := range []struct {
		name  string
		index int
		level int64
		hp    int64
		armor int64
		xp    int64
		class string
	}{
		{name: "normal", index: 0, level: 1, hp: 10, armor: 20, xp: 50, class: "normal-drop"},
		{name: "nightmare", index: 1, level: 4, hp: 24, armor: 36, xp: 120, class: "nightmare-drop"},
		{name: "hell", index: 2, level: 5, hp: 42, armor: 60, xp: 240, class: "hell-drop"},
	} {
		for _, players := range []int{1, 2} {
			t.Run(fmt.Sprintf("%s/%d_players", difficulty.name, players), func(t *testing.T) {
				_, engine, session := startDifficultyMonsterSession(t, difficulty.index)
				commands := []simulation.Command{{
					Tick: 1, Player: "population", Authority: simulation.AuthoritySystem, Sequence: 1,
					Kind: "system.population.bootstrap", Payload: difficultyPopulationPayload(t, difficulty.index),
				}}
				for index := 0; index < players; index++ {
					player := fmt.Sprintf("player-%d", index)
					commands = append(commands, simulation.Command{
						Tick: 1, Player: "system", Authority: simulation.AuthoritySystem, Sequence: uint64(index + 1),
						Kind: "system.player.enter", Payload: generatedPlayerPayload(t, "hero-"+player, player, float64(index+1), 1),
					})
				}
				for _, command := range commands {
					if err := session.Submit(command); err != nil {
						t.Fatal(err)
					}
				}
				if err := session.Step(); err != nil {
					t.Fatal(err)
				}

				const id = "level:2:room:difficulty-room:monster:1"
				entity := monsterEntity(t, engine, id)
				stats, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.stats")
				values, _ := stats.Get(entity)
				life, experience := difficulty.hp, difficulty.xp
				if players == 2 {
					life += life / 2
					experience += experience / 2
				}
				for field, want := range map[string]int64{
					"level": difficulty.level, "health": life * 256, "max_health": life * 256,
					"defense": difficulty.armor, "experience": experience, "spawn_player_count": int64(players),
				} {
					got, _ := values.Get(field)
					if got != want {
						t.Errorf("%s = %v, want %d", field, got, want)
					}
				}
				identities, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.identity")
				identity, _ := identities.Get(entity)
				if got, _ := identity.Get("treasure_class"); got != difficulty.class {
					t.Errorf("treasure class = %v, want %q", got, difficulty.class)
				}
			})
		}
	}
}

func TestPopulationRejectsDifferentSessionDifficulty(t *testing.T) {
	authority, _, session := startDifficultyMonsterSession(t, 2)
	beforeRules, _ := authority.State.Read("d2legacy.game_rules")
	beforePlan, _ := authority.State.Read("d2legacy.population.plan")
	if err := session.Submit(simulation.Command{
		Tick: 1, Player: "population", Authority: simulation.AuthoritySystem, Sequence: 1,
		Kind: "system.population.bootstrap", Payload: difficultyPopulationPayload(t, 1),
	}); err == nil || !strings.Contains(err.Error(), "population difficulty differs from immutable game rules") {
		t.Fatalf("mismatched bootstrap error = %v", err)
	}
	afterRules, _ := authority.State.Read("d2legacy.game_rules")
	afterPlan, _ := authority.State.Read("d2legacy.population.plan")
	if !reflect.DeepEqual(afterRules, beforeRules) || !reflect.DeepEqual(afterPlan, beforePlan) {
		t.Fatalf("mismatched bootstrap changed rules or plan: rules=%#v plan=%#v", afterRules, afterPlan)
	}
	var rules struct {
		Difficulty int `json:"difficulty"`
	}
	if err := json.Unmarshal(afterRules.Data, &rules); err != nil {
		t.Fatal(err)
	}
	var plan struct {
		Installed bool `json:"installed"`
	}
	if err := json.Unmarshal(afterPlan.Data, &plan); err != nil {
		t.Fatal(err)
	}
	if rules.Difficulty != 2 || plan.Installed {
		t.Fatalf("rules difficulty=%d, plan installed=%v; want 2 and false", rules.Difficulty, plan.Installed)
	}
}
