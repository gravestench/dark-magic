package d2legacy

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

// TestMonsterDeathUsesCreditedPlayersLivingSameLevelPartyForNoDrop verifies
// loot scaling counts only eligible nearby party members at the death tick.
func TestMonsterDeathUsesCreditedPlayersLivingSameLevelPartyForNoDrop(t *testing.T) {
	ctx := context.Background()
	authority, engine, session := startPartyFixture(t, nil)
	t.Cleanup(func() {
		_ = authority.Stop(ctx)
		_ = session.Close()
		_ = engine.Close()
	})

	population, _ := json.Marshal(map[string]any{
		"act": 1, "level_id": 2, "difficulty": 0,
		"links": []map[string]any{},
		"rooms": []map[string]any{{
			"id": "party-nodrop", "populate": true,
			"x": 0, "y": 0, "width": 10, "height": 10,
			"points": []map[string]any{{"x": 4, "y": 0}},
		}},
	})
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
	assertMonsterPlayerCount(t, engine, "level:2:room:party-nodrop:monster:1", 2)

	submitPartyCommand(t, session, 2, "alice", 1, "party.invite", map[string]any{"target": "bob"})
	stepPartySession(t, session)
	submitPartyCommand(t, session, 3, "bob", 1, "party.accept", map[string]any{"inviter": "alice"})
	stepPartySession(t, session)

	cast, _ := json.Marshal(map[string]any{
		"side": "left", "target_x": 4, "target_y": 0,
		"target_id": "monster:level:2:room:party-nodrop:monster:1",
	})

	for attempt := uint64(0); attempt < 4; attempt++ {
		if deaths, found := akara.GetDynamicStore(engine.World(), "d2legacy.monster.death"); found && deaths.Len() > 0 {
			break
		}

		if err := session.Submit(simulation.Command{
			Tick: 4 + 8*attempt, Player: "alice", Authority: simulation.AuthorityPlayer,
			Sequence: 2 + attempt, Kind: "player.use_skill", Payload: cast}); err != nil {
			t.Fatal(err)
		}

		stepSession(t, session, 8)
	}

	deaths, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.death")
	if deaths.Len() != 1 {
		stats, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.stats")

		var health any

		if stats.Len() > 0 {
			value, _ := stats.Get(stats.Entities()[0])
			health, _ = value.Get("health")
		}

		t.Fatalf("monster deaths = %d, want 1; health=%v", deaths.Len(), health)
	}

	death, _ := deaths.Get(deaths.Entities()[0])
	for field, want := range map[string]int64{
		"game_player_count":         2,
		"effective_player_count":    2,
		"nearby_party_member_count": 1,
		"monster_player_count":      2,
		"no_drop_player_count":      2,
	} {
		got, _ := death.Get(field)
		if got != want {
			t.Fatalf("death %s = %v, want %d", field, got, want)
		}
	}

	credited, _ := death.Get("credited_id")
	if credited != "player:alice" {
		t.Fatalf("death credited ID = %v, want player:alice", credited)
	}
}
