package d2legacy

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

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

	leave, _ := json.Marshal(map[string]any{"player": "bob"})
	submit(5, "system.player.leave", leave)
	submit(6, "system.monster.spawn", playerCountMonsterPayload(t, "after-leave", 12))
	assertMonsterPlayerCount(t, engine, "after-leave", 1)

	override, _ := json.Marshal(map[string]any{"count": 8})
	submit(7, "game.player_count.override", override)
	submit(8, "system.monster.spawn", playerCountMonsterPayload(t, "overridden", 13))
	assertMonsterPlayerCount(t, engine, "overridden", 8)

	follow, _ := json.Marshal(map[string]any{})
	submit(9, "game.player_count.follow_population", follow)
	submit(10, "system.monster.spawn", playerCountMonsterPayload(t, "following-again", 14))
	assertMonsterPlayerCount(t, engine, "following-again", 1)
}

func TestMaximumPlayersRejectsAdmissionWithoutBecomingGameplayCount(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	defer engine.Close()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
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
	defer authority.Stop(ctx)

	if err := session.Submit(simulation.Command{Tick: 1, Player: "system", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.player.enter", Payload: generatedPlayerPayload(t, "hero-alice", "alice", 0, 0)}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	if err := session.Submit(simulation.Command{Tick: 2, Player: "system", Authority: simulation.AuthoritySystem,
		Sequence: 2, Kind: "system.player.enter", Payload: generatedPlayerPayload(t, "hero-bob", "bob", 20, 20)}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); !errors.Is(err, gamesession.ErrCommandApply) {
		t.Fatalf("second admission error = %v, want %v", err, gamesession.ErrCommandApply)
	}
}

func TestGameRulesRejectLegacyImmutablePlayerCountConfiguration(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	defer engine.Close()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
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

func playerCountMonsterPayload(t *testing.T, spawnID string, seed int) json.RawMessage {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"spawn_id": spawnID, "seed": seed, "x": 90, "y": 90, "act": 1, "level_id": 2,
		"definition": map[string]any{
			"id": "fallen", "base_id": "fallen", "graphics_id": "fallen", "name_key": "Fallen",
			"ai": "fallen", "token": "FA", "weapon_class": "HTH", "components": map[string]string{},
			"life_min": 256, "life_max": 256, "level": 1, "defense": 0, "attack_rating": 0,
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
