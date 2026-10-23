package d2legacy

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

// TestMonsterScalingFollowsPopulationUntilExplicitlyOverridden verifies live
// population drives scaling only until an authority override becomes explicit.
func TestMonsterScalingFollowsPopulationUntilExplicitlyOverridden(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()

	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}

	authority, err := StartWithConfig(ctx, content.D2Legacy(), runtimeFixtureRecords{}, engine, session, Config{
		Seed: 29,
		InitialData: map[string]any{
			"d2legacy.game_rules": map[string]any{
				"target": "lod-1.14d", "expansion": true, "difficulty": 0,
				"hardcore": false, "ladder": false, "maximum_players": 2,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = authority.Stop(ctx)
		_ = session.Close()
		_ = engine.Close()
	})

	sequence := uint64(0)
	submit := func(tick uint64, kind string, payload json.RawMessage) {
		t.Helper()

		sequence++
		if err := session.Submit(simulation.Command{Tick: tick, Player: "system", Authority: simulation.AuthoritySystem,
			Sequence: sequence, Kind: kind, Payload: payload}); err != nil {
			t.Fatal(err)
		}

		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
	}

	submit(1, "system.player.enter", generatedPlayerPayload(t, "hero-alice", "alice", 0, 0))
	submit(2, "system.monster.spawn", playerCountMonsterPayload(t, "one-player", 10))
	assertMonsterPlayerCount(t, engine, "one-player", 1)

	submit(3, "system.player.enter", generatedPlayerPayload(t, "hero-bob", "bob", 20, 20))
	submit(4, "system.monster.spawn", playerCountMonsterPayload(t, "two-players", 11))
	assertMonsterPlayerCount(t, engine, "two-players", 2)

	lowerOverride, _ := json.Marshal(map[string]any{"count": 1})
	submit(5, "game.player_count.override", lowerOverride)
	submit(6, "system.monster.spawn", playerCountMonsterPayload(t, "lower-override", 12))
	assertMonsterPlayerCount(t, engine, "lower-override", 2)
	assertMonsterHealthAndExperience(t, engine, "lower-override", 6*256, 150)
	stageLethalPlayerEvent(t, engine, "bob", 6)
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	assertPlayerDead(t, engine, "bob")
	submit(8, "system.monster.spawn", playerCountMonsterPayload(t, "after-death", 13))
	assertMonsterPlayerCount(t, engine, "after-death", 1)
	assertMonsterHealthAndExperience(t, engine, "after-death", 4*256, 100)

	follow, _ := json.Marshal(map[string]any{})
	submit(9, "game.player_count.follow_population", follow)
	leave, _ := json.Marshal(map[string]any{"player": "bob"})
	submit(10, "system.player.leave", leave)
	submit(11, "system.monster.spawn", playerCountMonsterPayload(t, "after-leave", 14))
	assertMonsterPlayerCount(t, engine, "after-leave", 1)

	override, _ := json.Marshal(map[string]any{"count": 8})
	submit(12, "game.player_count.override", override)
	submit(13, "system.monster.spawn", playerCountMonsterPayload(t, "overridden", 15))
	assertMonsterPlayerCount(t, engine, "overridden", 8)

	submit(14, "game.player_count.follow_population", follow)
	submit(15, "system.monster.spawn", playerCountMonsterPayload(t, "following-again", 16))
	assertMonsterPlayerCount(t, engine, "following-again", 1)
}

// TestMaximumPlayersRejectsAdmissionWithoutBecomingGameplayCount separates the
// admission ceiling from the population value used by gameplay formulas.
func TestMaximumPlayersRejectsAdmissionWithoutBecomingGameplayCount(t *testing.T) {
	ctx := context.Background()

	engine := gameecs.New()
	defer func() { _ = engine.Close() }()

	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = session.Close() }()

	authority, err := StartWithConfig(ctx, content.D2Legacy(), runtimeFixtureRecords{}, engine, session, Config{
		Seed: 31,
		InitialData: map[string]any{
			"d2legacy.game_rules": map[string]any{
				"target": "lod-1.14d", "expansion": true, "difficulty": 0,
				"hardcore": false, "ladder": false, "maximum_players": 1,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = authority.Stop(ctx) }()

	if err := session.Submit(simulation.Command{
		Tick: 1, Player: "system", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.player.enter",
		Payload: generatedPlayerPayload(t, "hero-alice", "alice", 0, 0),
	}); err != nil {
		t.Fatal(err)
	}

	if err := session.Step(); err != nil {
		t.Fatal(err)
	}

	if err := session.Submit(simulation.Command{
		Tick: 2, Player: "system", Authority: simulation.AuthoritySystem,
		Sequence: 2, Kind: "system.player.enter",
		Payload: generatedPlayerPayload(t, "hero-bob", "bob", 20, 20),
	}); err != nil {
		t.Fatal(err)
	}

	if err := session.Step(); !errors.Is(err, gamesession.ErrCommandApply) {
		t.Fatalf("second admission error = %v, want %v", err, gamesession.ErrCommandApply)
	}
}

// TestGameRulesRejectLegacyImmutablePlayerCountConfiguration prevents an old
// bootstrap field from bypassing the live population and override contracts.
func TestGameRulesRejectLegacyImmutablePlayerCountConfiguration(t *testing.T) {
	ctx := context.Background()

	engine := gameecs.New()
	defer func() { _ = engine.Close() }()

	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = session.Close() }()

	authority, err := StartWithConfig(ctx, content.D2Legacy(), runtimeFixtureRecords{}, engine, session, Config{
		Seed: 37,
		InitialData: map[string]any{
			"d2legacy.game_rules": map[string]any{
				"target": "lod-1.14d", "expansion": true, "difficulty": 0,
				"hardcore": false, "ladder": false, "maximum_players": 8, "player_count": 1,
			},
		},
	})
	if err == nil {
		_ = authority.Stop(ctx)

		t.Fatal("legacy immutable player-count configuration was accepted")
	}
}

// playerCountMonsterPayload builds a stable monster command whose only varying
// inputs are exposed by arguments, keeping scaling assertions deterministic.
func playerCountMonsterPayload(t *testing.T, spawnID string, seed int) json.RawMessage {
	t.Helper()

	payload, err := json.Marshal(map[string]any{
		"spawn_id": spawnID, "seed": seed, "x": 90, "y": 90, "act": 1, "level_id": 2,
		"definition": map[string]any{
			"id": "fallen", "base_id": "fallen", "graphics_id": "fallen", "name_key": "Fallen",
			"ai": "fallen", "token": "FA", "weapon_class": "HTH", "components": map[string]string{},
			"life_min": 4 * 256, "life_max": 4 * 256, "level": 1, "defense": 0, "attack_rating": 0,
			"physical_min": 0, "physical_max": 0, "experience": 100, "treasure_class": "",
			"collider_radius": 1, "select_radius": 1, "velocity": 0, "think_interval": 100,
			"aggro_radius": 0, "attack_range": 1, "evil": true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	return payload
}

func assertMonsterHealthAndExperience(t *testing.T, engine *gameecs.Engine, spawnID string, wantHealth, wantExperience int64) {
	t.Helper()

	identities, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.identity")
	stats, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.stats")
	for _, entity := range identities.Entities() {
		identity, _ := identities.Get(entity)
		current, _ := identity.Get("spawn_id")
		if current != spawnID {
			continue
		}
		values, found := stats.Get(entity)
		if !found {
			t.Fatalf("monster %s has no stats", spawnID)
		}
		for field, want := range map[string]int64{
			"health": wantHealth, "max_health": wantHealth, "experience": wantExperience,
		} {
			got, _ := values.Get(field)
			if got != want {
				t.Fatalf("monster %s %s = %v, want %d", spawnID, field, got, want)
			}
		}
		return
	}
	t.Fatalf("monster %s was not created", spawnID)
}

func stageLethalPlayerEvent(t *testing.T, engine *gameecs.Engine, playerID string, tick int64) {
	t.Helper()
	identities, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.identity")
	vitals, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.vitals")
	for _, entity := range identities.Entities() {
		identity, _ := identities.Get(entity)
		current, _ := identity.Get("player")
		if current != playerID {
			continue
		}
		values, found := vitals.Get(entity)
		if !found {
			t.Fatalf("player %s has no vitals", playerID)
		}
		if err := values.Set("health", int64(0)); err != nil {
			t.Fatal(err)
		}
		combatEvents, _ := akara.GetDynamicStore(engine.World(), "d2legacy.combat.event")
		if _, err := combatEvents.Set(engine.World().MustCreateEntity(), map[string]any{
			"kind": "unit_died", "tick": tick,
			"attacker_id": "monster:fixture", "target_id": "player:" + playerID,
			"source_kind": "melee", "damage_channel": "physical",
			"rolled_damage_raw": int64(256), "damage_raw": int64(256),
			"remaining_health_raw": int64(0),
		}); err != nil {
			t.Fatal(err)
		}
		return
	}
	t.Fatalf("player %s was not created", playerID)
}

func assertPlayerDead(t *testing.T, engine *gameecs.Engine, playerID string) {
	t.Helper()
	identities, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.identity")
	deaths, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.death")
	for _, entity := range identities.Entities() {
		identity, _ := identities.Get(entity)
		current, _ := identity.Get("player")
		if current != playerID {
			continue
		}
		death, found := deaths.Get(entity)
		if !found {
			t.Fatalf("player %s has no committed death marker", playerID)
		}
		stage, _ := death.Get("stage")
		if stage != "dead" {
			t.Fatalf("player %s death stage = %v, want dead", playerID, stage)
		}
		return
	}
	t.Fatalf("player %s was not created", playerID)
}
