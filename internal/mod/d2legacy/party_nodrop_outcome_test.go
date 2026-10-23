package d2legacy

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

// TestPartyEligibilityChangesDurableMonsterLootOutcome catches a death path
// that stores the correct NoDrop count but rolls loot with a different count.
func TestPartyEligibilityChangesDurableMonsterLootOutcome(t *testing.T) {
	type killTiming struct {
		tick  int64
		casts int
	}
	var timings []killTiming

	for _, test := range []struct {
		name          string
		party         bool
		noDropPlayers int64
		wantCode      string
	}{
		{name: "unpartied", noDropPlayers: 1},
		{name: "partied", party: true, noDropPlayers: 2, wantCode: "rin"},
	} {
		t.Run(test.name, func(t *testing.T) {
			engine, session := startNoDropOutcomeFixture(t)

			population, err := json.Marshal(map[string]any{
				"act": 1, "level_id": 2, "difficulty": 0,
				"links": []map[string]any{},
				"rooms": []map[string]any{{
					"id": "nodrop-outcome", "populate": true,
					"x": 0, "y": 0, "width": 10, "height": 10,
					"points": []map[string]any{{"x": 4, "y": 0}},
				}},
			})
			if err != nil {
				t.Fatal(err)
			}

			for _, command := range []simulation.Command{
				{Tick: 1, Player: "population", Authority: simulation.AuthoritySystem, Sequence: 1,
					Kind: "system.population.bootstrap", Payload: population},
				{Tick: 1, Player: "system", Authority: simulation.AuthoritySystem, Sequence: 1,
					Kind: "system.player.enter", Payload: generatedPlayerPayload(t, "hero-alice", "alice", 0, 0)},
				{Tick: 1, Player: "system", Authority: simulation.AuthoritySystem, Sequence: 2,
					Kind: "system.player.enter", Payload: generatedPlayerPayload(t, "hero-bob", "bob", 8, 8)},
			} {
				if err := session.Submit(command); err != nil {
					t.Fatal(err)
				}
			}
			stepPartySession(t, session)
			assertMonsterPlayerCount(t, engine, "level:2:room:nodrop-outcome:monster:1", 2)

			if test.party {
				submitPartyCommand(t, session, 2, "alice", 1, "party.invite", map[string]any{"target": "bob"})
			}
			stepPartySession(t, session)
			if test.party {
				submitPartyCommand(t, session, 3, "bob", 1, "party.accept", map[string]any{"inviter": "alice"})
			}
			stepPartySession(t, session)

			cast, err := json.Marshal(map[string]any{
				"side": "left", "target_x": 4, "target_y": 0,
				"target_id": "monster:level:2:room:nodrop-outcome:monster:1",
			})
			if err != nil {
				t.Fatal(err)
			}
			castSequence := uint64(1)
			if test.party {
				castSequence = 2
			}
			casts := 0
			for attempt := uint64(0); attempt < 4; attempt++ {
				if deaths, found := akara.GetDynamicStore(engine.World(), "d2legacy.monster.death"); found && deaths.Len() > 0 {
					break
				}
				if err := session.Submit(simulation.Command{
					Tick: 4 + 8*attempt, Player: "alice", Authority: simulation.AuthorityPlayer,
					Sequence: castSequence + attempt, Kind: "player.use_skill", Payload: cast,
				}); err != nil {
					t.Fatal(err)
				}
				casts++
				stepSession(t, session, 8)
			}

			deaths, found := akara.GetDynamicStore(engine.World(), "d2legacy.monster.death")
			if !found {
				t.Fatal("monster death store missing")
			}
			if deaths.Len() != 1 {
				t.Fatalf("monster deaths = %d, want 1", deaths.Len())
			}
			death, _ := deaths.Get(deaths.Entities()[0])
			for field, want := range map[string]any{
				"credited_id":          "player:alice",
				"monster_player_count": int64(2),
				"game_player_count":    int64(2),
				"no_drop_player_count": test.noDropPlayers,
			} {
				got, _ := death.Get(field)
				if got != want {
					t.Fatalf("death %s = %v, want %v", field, got, want)
				}
			}

			encoded, _ := death.Get("drops")
			var drops []struct {
				Code string `json:"code"`
			}
			if err := json.Unmarshal([]byte(encoded.(string)), &drops); err != nil {
				t.Fatalf("decode durable drops: %v", err)
			}
			if drops == nil {
				t.Fatal("durable drops is null, want a JSON array")
			}
			if test.wantCode == "" && len(drops) != 0 {
				t.Fatalf("durable drops = %v, want empty", drops)
			}
			if test.wantCode != "" && (len(drops) != 1 || drops[0].Code != test.wantCode) {
				t.Fatalf("durable drops = %v, want one %s", drops, test.wantCode)
			}
			tickValue, _ := death.Get("tick")
			deathTick, ok := tickValue.(int64)
			if !ok {
				t.Fatalf("death tick = %v (%T), want int64", tickValue, tickValue)
			}
			timings = append(timings, killTiming{tick: deathTick, casts: casts})
		})
	}
	if len(timings) == 2 && timings[0] != timings[1] {
		t.Fatalf("death tick/casts differ between unpartied and partied: %v versus %v", timings[0], timings[1])
	}
}

func startNoDropOutcomeFixture(t *testing.T) (*gameecs.Engine, *gamesession.Session) {
	t.Helper()

	records := generatedHostileRecords()
	records["data/global/excel/treasureclassex.txt"] = []map[string]string{{
		"Treasure Class": "fallen-drop", "Picks": "1", "NoDrop": "3",
		"Item1": "rin", "Prob1": "1", "Unique": "1024",
	}}
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	authority, err := StartWithConfig(t.Context(), content.D2Legacy(), records, engine, session, Config{Seed: 314})
	if err != nil {
		_ = session.Close()
		_ = engine.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = authority.Stop(context.Background())
		_ = session.Close()
		_ = engine.Close()
	})
	return engine, session
}
