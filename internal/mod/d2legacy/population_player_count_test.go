package d2legacy

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

// TestRoomPopulationUsesLivePlayerCount exercises the room activation path
// after eligibility changes, then compares it with direct monster spawning.
func TestRoomPopulationUsesLivePlayerCount(t *testing.T) {
	for _, scenario := range []struct {
		name       string
		dead       bool
		inactive   bool
		wantCount  int64
		wantHealth int64
		wantXP     int64
		directHP   int64
		directXP   int64
	}{
		{name: "two living", wantCount: 2, wantHealth: 4 * 256, wantXP: 7, directHP: 6 * 256, directXP: 150},
		{name: "one committed dead", dead: true, wantCount: 1, wantHealth: 3 * 256, wantXP: 5, directHP: 4 * 256, directXP: 100},
		{name: "one inactive", inactive: true, wantCount: 1, wantHealth: 3 * 256, wantXP: 5, directHP: 4 * 256, directXP: 100},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			authority, engine, session := startPartyFixture(t, nil)
			t.Cleanup(func() {
				_ = authority.Stop(context.Background())
				_ = session.Close()
				_ = engine.Close()
			})

			population, err := json.Marshal(map[string]any{
				"act": 1, "level_id": 2, "difficulty": 0,
				"links": []map[string]any{},
				"rooms": []map[string]any{{
					"id": "count-room", "populate": true,
					"x": 10, "y": 0, "width": 10, "height": 10,
					"points": []map[string]any{{"x": 14, "y": 1}},
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
					Kind: "system.player.enter", Payload: generatedPlayerPayload(t, "hero-bob", "bob", 1, 1)},
			} {
				if err := session.Submit(command); err != nil {
					t.Fatal(err)
				}
			}
			stepPartySession(t, session)
			if got := monsterCount(engine); got != 0 {
				t.Fatalf("monster count before target room activation = %d, want 0", got)
			}

			if scenario.dead {
				stageLethalPlayerEvent(t, engine, "bob", 1)
				stepPartySession(t, session)
				assertPlayerDead(t, engine, "bob")
			}
			if scenario.inactive {
				inactive, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.inactive")
				if _, err := inactive.Set(playerEntity(t, engine, "bob"), map[string]any{}); err != nil {
					t.Fatal(err)
				}
			}

			position, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.position")
			setEntityPosition(t, position, playerEntity(t, engine, "alice"), 11, 1)
			stepPartySession(t, session)
			const roomMonster = "level:2:room:count-room:monster:1"
			assertMonsterPlayerCount(t, engine, roomMonster, scenario.wantCount)
			assertMonsterHealthAndExperience(t, engine, roomMonster, scenario.wantHealth, scenario.wantXP)

			directTick := uint64(3)
			if scenario.dead {
				directTick = 4
			}
			if err := session.Submit(simulation.Command{
				Tick: directTick, Player: "system", Authority: simulation.AuthoritySystem,
				Sequence: 3, Kind: "system.monster.spawn", Payload: playerCountMonsterPayload(t, "direct-count", 21),
			}); err != nil {
				t.Fatal(err)
			}
			stepPartySession(t, session)
			assertMonsterPlayerCount(t, engine, "direct-count", scenario.wantCount)
			assertMonsterHealthAndExperience(t, engine, "direct-count", scenario.directHP, scenario.directXP)
		})
	}
}
