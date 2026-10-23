package d2legacy

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const partyXPMonster = "level:2:room:party-xp:monster:1"

// TestPartyExperienceSharesByVictimDistance proves the production death path
// shares XP with a nearby party member while a far same-level member still
// contributes to the independent NoDrop party count.
func TestPartyExperienceSharesByVictimDistance(t *testing.T) {
	engine, session := startPartyXPScenario(t)
	joinPartyXP(t, session)
	stagePartyXPMonsterDeath(t, engine, "player:alice", 4)
	stepPartySession(t, session)

	for player, want := range map[string]int64{"alice": 6, "bob": 6, "charlie": 0} {
		if got := partyPlayerXP(t, engine, player); got != want {
			t.Errorf("%s experience = %d, want %d", player, got, want)
		}
	}
	deaths, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.death")
	if deaths.Len() != 1 {
		t.Fatalf("monster deaths = %d, want 1", deaths.Len())
	}
	death, _ := deaths.Get(deaths.Entities()[0])
	for field, want := range map[string]int64{
		"xp": 10, "nearby_party_member_count": 2,
	} {
		got, _ := death.Get(field)
		if got != want {
			t.Errorf("death %s = %v, want %d", field, got, want)
		}
	}

	stepPartySession(t, session)
	for player, want := range map[string]int64{"alice": 6, "bob": 6, "charlie": 0} {
		if got := partyPlayerXP(t, engine, player); got != want {
			t.Errorf("repeat tick %s experience = %d, want %d", player, got, want)
		}
	}
}

func TestPartyExperienceExcludesDeadAndOtherLevelMembers(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		change func(*testing.T, *gameecs.Engine, *gamesession.Session)
		death  int64
	}{
		{
			name: "dead member",
			change: func(t *testing.T, engine *gameecs.Engine, session *gamesession.Session) {
				stageLethalPlayerEvent(t, engine, "bob", 4)
				stepPartySession(t, session)
				assertPlayerDead(t, engine, "bob")
			},
			death: 5,
		},
		{
			name: "other level",
			change: func(t *testing.T, engine *gameecs.Engine, _ *gamesession.Session) {
				entity := partyPlayerEntity(t, engine, "bob")
				locations, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.location")
				location, _ := locations.Get(entity)
				if err := location.Set("level_id", int64(3)); err != nil {
					t.Fatal(err)
				}
			},
			death: 4,
		},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			engine, session := startPartyXPScenario(t)
			joinPartyXP(t, session)
			scenario.change(t, engine, session)
			stagePartyXPMonsterDeath(t, engine, "player:alice", scenario.death)
			stepPartySession(t, session)
			for player, want := range map[string]int64{"alice": 10, "bob": 0, "charlie": 0} {
				if got := partyPlayerXP(t, engine, player); got != want {
					t.Errorf("%s experience = %d, want %d", player, got, want)
				}
			}
		})
	}
}

func TestPartyExperienceUsesCurrentMembershipAfterLeaveAndJoin(t *testing.T) {
	engine, session := startPartyXPScenario(t)
	joinPartyXP(t, session)
	submitPartyCommand(t, session, 4, "bob", 2, "party.leave", map[string]any{})
	stepPartySession(t, session)
	submitPartyCommand(t, session, 5, "alice", 3, "party.invite", map[string]any{"target": "bob"})
	stepPartySession(t, session)
	submitPartyCommand(t, session, 6, "bob", 3, "party.accept", map[string]any{"inviter": "alice"})
	stepPartySession(t, session)
	stagePartyXPMonsterDeath(t, engine, "player:alice", 7)
	stepPartySession(t, session)
	for player, want := range map[string]int64{"alice": 6, "bob": 6, "charlie": 0} {
		if got := partyPlayerXP(t, engine, player); got != want {
			t.Errorf("%s experience after rejoin = %d, want %d", player, got, want)
		}
	}
}

func TestPartyExperienceRestoredMembershipControlsLaterKill(t *testing.T) {
	_, session := startPartyXPScenario(t)
	joinPartyXP(t, session)
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := replay.Checkpoints[len(replay.Checkpoints)-1]

	ctx := context.Background()
	restored, engine, restoredSession := startPartyFixture(t, &checkpoint)
	t.Cleanup(func() {
		_ = restored.Stop(ctx)
		_ = restoredSession.Close()
		_ = engine.Close()
	})
	submitPartyCommand(t, restoredSession, 4, "bob", 1, "party.leave", map[string]any{})
	stepPartySession(t, restoredSession)
	stagePartyXPMonsterDeath(t, engine, "player:alice", 5)
	stepPartySession(t, restoredSession)
	stepPartySession(t, restoredSession)
	for player, want := range map[string]int64{"alice": 10, "bob": 0, "charlie": 0} {
		if got := partyPlayerXP(t, engine, player); got != want {
			t.Errorf("restored %s experience = %d, want %d", player, got, want)
		}
	}
}

func TestPartyExperienceOwnedUnitKillUsesUltimatePlayerOwner(t *testing.T) {
	engine, session := startPartyXPScenario(t)
	joinPartyXP(t, session)
	owner := partyPlayerEntity(t, engine, "alice")
	owned := engine.World().MustCreateEntity()
	selectables, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.selectable")
	if _, err := selectables.Set(owned, map[string]any{
		"id": "pet:alice", "kind": "owned", "label": "party XP fixture pet",
		"owner": "player:alice", "radius": float64(0), "priority": int64(0),
	}); err != nil {
		t.Fatal(err)
	}
	relations, _ := akara.GetDynamicStore(engine.World(), "d2legacy.owned_unit")
	if _, err := relations.Set(owned, map[string]any{
		"owner": owner, "owner_id": "player:alice", "ultimate_owner_id": "player:alice",
		"category": "fixture", "group": int64(1), "limit": int64(1),
		"replacement": "replace_oldest", "created_tick": int64(3), "expires_tick": int64(0),
		"durable_id": "", "durable": false, "unsummon": false,
		"warp_with_owner": false, "range_limited": false,
		"active": true, "survives_owner_death": false,
	}); err != nil {
		t.Fatal(err)
	}
	stagePartyXPMonsterDeath(t, engine, "pet:alice", 4)
	stepPartySession(t, session)
	for player, want := range map[string]int64{"alice": 6, "bob": 6, "charlie": 0} {
		if got := partyPlayerXP(t, engine, player); got != want {
			t.Errorf("owned-unit kill %s experience = %d, want %d", player, got, want)
		}
	}
	deaths, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.death")
	death, _ := deaths.Get(deaths.Entities()[0])
	credited, _ := death.Get("credited_id")
	if credited != "player:alice" {
		t.Errorf("owned-unit kill credit = %v, want player:alice", credited)
	}
}

func TestPartyExperienceSoloAndMissingPositionKeepExistingAward(t *testing.T) {
	for _, scenario := range []struct {
		name          string
		join          bool
		removeVictimX bool
	}{
		{name: "solo player"},
		{name: "party with missing victim position", join: true, removeVictimX: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			engine, session := startPartyXPScenario(t)
			if scenario.join {
				joinPartyXP(t, session)
			}
			if scenario.removeVictimX {
				positions, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.position")
				positions.Remove(monsterEntity(t, engine, partyXPMonster))
			}
			tick := int64(2)
			if scenario.join {
				tick = 4
			}
			stagePartyXPMonsterDeath(t, engine, "player:alice", tick)
			stepPartySession(t, session)
			for player, want := range map[string]int64{"alice": 10, "bob": 0, "charlie": 0} {
				if got := partyPlayerXP(t, engine, player); got != want {
					t.Errorf("%s experience = %d, want %d", player, got, want)
				}
			}
		})
	}
}

func TestPartyExperienceRequiresCreditedPlayerIdentityAndProgress(t *testing.T) {
	for _, component := range []string{"d2legacy.player.identity", "d2legacy.player.progress"} {
		t.Run(component, func(t *testing.T) {
			engine, session := startPartyXPScenario(t)
			joinPartyXP(t, session)
			store, _ := akara.GetDynamicStore(engine.World(), component)
			store.Remove(partyPlayerEntity(t, engine, "alice"))
			stagePartyXPMonsterDeath(t, engine, "player:alice", 4)
			stepPartySession(t, session)
			for _, player := range []string{"bob", "charlie"} {
				if got := partyPlayerXP(t, engine, player); got != 0 {
					t.Errorf("%s experience with missing credited %s = %d, want 0", player, component, got)
				}
			}
		})
	}
}

func startPartyXPScenario(t *testing.T) (*gameecs.Engine, *gamesession.Session) {
	t.Helper()
	ctx := context.Background()
	authority, engine, session := startPartyFixture(t, nil)
	t.Cleanup(func() {
		_ = authority.Stop(ctx)
		_ = session.Close()
		_ = engine.Close()
	})

	population, err := json.Marshal(map[string]any{
		"act": 1, "level_id": 2, "difficulty": 0,
		"links": []map[string]any{},
		"rooms": []map[string]any{{
			"id": "party-xp", "populate": true,
			"x": 0, "y": 0, "width": 100, "height": 100,
			"points": []map[string]any{{"x": 4, "y": 4}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, command := range []simulation.Command{
		{Tick: 1, Player: "population", Authority: simulation.AuthoritySystem, Sequence: 1,
			Kind: "system.population.bootstrap", Payload: population},
		{Tick: 1, Player: "system", Authority: simulation.AuthoritySystem, Sequence: 1,
			Kind: "system.player.enter", Payload: generatedPlayerPayload(t, "hero-alice", "alice", 1, 1)},
		{Tick: 1, Player: "system", Authority: simulation.AuthoritySystem, Sequence: 2,
			Kind: "system.player.enter", Payload: generatedPlayerPayload(t, "hero-bob", "bob", 8, 8)},
		{Tick: 1, Player: "system", Authority: simulation.AuthoritySystem, Sequence: 3,
			Kind: "system.player.enter", Payload: generatedPlayerPayload(t, "hero-charlie", "charlie", 90, 90)},
	} {
		if err := session.Submit(command); err != nil {
			t.Fatal(err)
		}
	}
	stepPartySession(t, session)
	assertMonsterPlayerCount(t, engine, partyXPMonster, 3)
	return engine, session
}

func joinPartyXP(t *testing.T, session *gamesession.Session) {
	t.Helper()
	submitPartyCommand(t, session, 2, "alice", 1, "party.invite", map[string]any{"target": "bob"})
	submitPartyCommand(t, session, 2, "alice", 2, "party.invite", map[string]any{"target": "charlie"})
	stepPartySession(t, session)
	submitPartyCommand(t, session, 3, "bob", 1, "party.accept", map[string]any{"inviter": "alice"})
	submitPartyCommand(t, session, 3, "charlie", 1, "party.accept", map[string]any{"inviter": "alice"})
	stepPartySession(t, session)
}

func stagePartyXPMonsterDeath(t *testing.T, engine *gameecs.Engine, killer string, tick int64) {
	t.Helper()
	monster := monsterEntity(t, engine, partyXPMonster)
	stats, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.stats")
	value, found := stats.Get(monster)
	if !found {
		t.Fatal("party XP monster has no stats")
	}
	if err := value.Set("health", int64(0)); err != nil {
		t.Fatal(err)
	}
	events, _ := akara.GetDynamicStore(engine.World(), "d2legacy.combat.event")
	if _, err := events.Set(engine.World().MustCreateEntity(), map[string]any{
		"kind": "unit_died", "tick": tick,
		"attacker_id": killer, "target_id": "monster:" + partyXPMonster,
		"source_kind": "melee", "damage_channel": "physical",
		"rolled_damage_raw": int64(256), "damage_raw": int64(256),
		"remaining_health_raw": int64(0),
	}); err != nil {
		t.Fatal(err)
	}
}

func partyPlayerXP(t *testing.T, engine *gameecs.Engine, player string) int64 {
	t.Helper()
	entity := partyPlayerEntity(t, engine, player)
	progress, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.progress")
	value, found := progress.Get(entity)
	if !found {
		t.Fatalf("player %s has no progress", player)
	}
	xp, _ := value.Get("experience")
	return xp.(int64)
}

func partyPlayerEntity(t *testing.T, engine *gameecs.Engine, player string) akara.Entity {
	t.Helper()
	identities, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.identity")
	for _, entity := range identities.Entities() {
		identity, _ := identities.Get(entity)
		current, _ := identity.Get("player")
		if current == player {
			return entity
		}
	}
	t.Fatalf("player %s not found", player)
	return 0
}
