package d2legacy

import (
	"context"
	"encoding/json"
	"math"
	"reflect"
	"testing"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	lua "github.com/yuin/gopher-lua"
)

func TestConfigureRuntimePreservesClientCatalogOverrides(t *testing.T) {
	runtime := modruntime.New()
	for _, name := range []string{"d2legacy.quest_catalog/v1", "d2legacy.map_catalog/v1"} {
		if err := runtime.RegisterModule(modruntime.Module{
			Name: name,
			Help: modruntime.ModuleHelp{Summary: "client override"},
			Loader: func(state *lua.LState) int {
				state.Push(state.NewTable())
				return 1
			},
		}); err != nil {
			t.Fatal(err)
		}
	}
	engine := gameecs.New()
	defer engine.Close()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := ConfigureRuntime(runtime, content.D2Legacy(), runtimeFixtureRecords{}, engine, session,
		simulation.NewStateStore(), simulation.NewRandomStreams(1), nil); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"d2legacy.quest_catalog/v1", "d2legacy.map_catalog/v1"} {
		if got := runtime.ModuleHelp()[name].Summary; got != "client override" {
			t.Fatalf("%s summary = %q, want client override", name, got)
		}
		count := 0
		for _, registered := range runtime.ModuleNames() {
			if registered == name {
				count++
			}
		}
		if count != 1 {
			t.Fatalf("%s registered %d times", name, count)
		}
	}
}

func TestClientCatalogsReplaceDefaultsAfterConfigureRuntime(t *testing.T) {
	runtime := modruntime.New()
	engine := gameecs.New()
	defer engine.Close()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := ConfigureRuntime(runtime, content.D2Legacy(), runtimeFixtureRecords{}, engine, session,
		simulation.NewStateStore(), simulation.NewRandomStreams(1), nil); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"d2legacy.quest_catalog/v1", "d2legacy.map_catalog/v1"} {
		if err := runtime.RegisterModuleOverride(modruntime.Module{
			Name: name,
			Help: modruntime.ModuleHelp{Summary: "late client override"},
			Loader: func(state *lua.LState) int {
				state.Push(state.NewTable())
				return 1
			},
		}); err != nil {
			t.Fatal(err)
		}
		if got := runtime.ModuleHelp()[name].Summary; got != "late client override" {
			t.Fatalf("%s summary = %q, want late client override", name, got)
		}
	}
}

func TestAuthorityMaterializesPlayerEntryThroughLua(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	defer engine.Close()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	authority, err := Start(ctx, content.D2Legacy(), runtimeFixtureRecords{}, engine, session, 7)
	if err != nil {
		t.Fatal(err)
	}
	defer authority.Stop(ctx)
	payload, _ := json.Marshal(map[string]any{
		"character_id": "hero", "player": "alice", "name": "Hero", "class": "Amazon",
		"level": 1, "experience": 0, "dexterity": 20, "defense": 20,
		"health": 50, "max_health": 50, "mana": 20, "max_mana": 20,
		"expansion": true, "hardcore": false, "cof": "",
		"palette": "data/global/Palette/units/pal.dat", "direction": 0, "mode": "NU",
		"x": 10, "y": 20, "world_width": 100, "world_height": 100, "act": 1, "level_id": 1,
		"skills": []map[string]any{{"id": 0, "level": 1, "list_row": 0, "left_allowed": true, "right_allowed": true}},
	})
	if err := session.Submit(simulation.Command{Tick: 1, Player: "system", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.player.enter", Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	identities, found := akara.GetDynamicStore(engine.World(), "d2legacy.player.identity")
	if !found || len(identities.Entities()) != 1 {
		t.Fatalf("Lua entry created %d players", len(identities.Entities()))
	}
	player := identities.Entities()[0]
	assignments, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.skill_assignment")
	assignment, _ := assignments.Get(player)
	left, _ := assignment.Get("left")
	if left != int64(36) {
		t.Fatalf("initial left skill = %v, want Lua-selected Fire Bolt 36", left)
	}
	learned, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.learned_skill")
	if len(learned.Entities()) != 1 {
		t.Fatalf("Lua entry created %d learned skills", len(learned.Entities()))
	}
}

func TestGameRulesCheckpointRestoreAndIdentityDrift(t *testing.T) {
	start := func(initial map[string]any, restore []simulation.ParticipantState) (*Authority, *gameecs.Engine, *gamesession.Session, error) {
		engine := gameecs.New()
		session, err := gamesession.New(engine, gamesession.Config{})
		if err != nil {
			_ = engine.Close()
			return nil, nil, nil, err
		}
		authority, err := StartWithConfig(t.Context(), content.D2Legacy(), runtimeFixtureRecords{}, engine, session,
			Config{Seed: 7, InitialData: initial, Restore: restore})
		if err != nil {
			_ = session.Close()
			_ = engine.Close()
			return nil, nil, nil, err
		}
		return authority, engine, session, nil
	}
	initial := map[string]any{
		"engine.game_data_generation_id": "sha256:test-generation",
		"d2legacy.game_rules": map[string]any{"target": "lod-1.14d", "expansion": true,
			"difficulty": 1, "hardcore": true, "maximum_players": 2},
	}
	authority, engine, session, err := start(initial, nil)
	if err != nil {
		t.Fatal(err)
	}
	override, _ := json.Marshal(map[string]any{"count": 8})
	if err := session.Submit(simulation.Command{Tick: 1, Player: "host", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "game.player_count.override", Payload: override}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	checkpoint, err := session.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	_ = authority.Stop(t.Context())
	_ = session.Close()
	_ = engine.Close()

	restored, restoredEngine, restoredSession, err := start(initial, checkpoint.Participants)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Stop(t.Context())
	defer restoredSession.Close()
	defer restoredEngine.Close()
	value, found := restored.State.Read("d2legacy.game_rules")
	if !found {
		t.Fatal("restored game rules are missing")
	}
	var rules map[string]any
	if err := json.Unmarshal(value.Data, &rules); err != nil {
		t.Fatal(err)
	}
	if rules["schema"] != "d2legacy.game_rules/v2" || rules["difficulty"] != float64(1) ||
		rules["hardcore"] != true || rules["maximum_players"] != float64(2) || rules["player_count"] != nil {
		t.Fatalf("restored rules = %#v", rules)
	}
	countValue, found := restored.State.Read("d2legacy.player_count")
	if !found {
		t.Fatal("restored player-count authority state is missing")
	}
	var count map[string]any
	if err := json.Unmarshal(countValue.Data, &count); err != nil {
		t.Fatal(err)
	}
	if count["schema"] != "d2legacy.player_count/v1" || count["override"] != float64(8) {
		t.Fatalf("restored player-count state = %#v", count)
	}
	clear, _ := json.Marshal(map[string]any{})
	if err := restoredSession.Submit(simulation.Command{Tick: 1, Player: "host", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "game.player_count.follow_population", Payload: clear}); err != nil {
		t.Fatal(err)
	}
	if err := restoredSession.Step(); err != nil {
		t.Fatal(err)
	}
	countValue, _ = restored.State.Read("d2legacy.player_count")
	count = nil
	if err := json.Unmarshal(countValue.Data, &count); err != nil {
		t.Fatal(err)
	}
	if count["override"] != nil || count["revision"] != float64(2) {
		t.Fatalf("cleared player-count state = %#v", count)
	}

	drifted := map[string]any{
		"engine.game_data_generation_id": "sha256:test-generation",
		"d2legacy.game_rules": map[string]any{"target": "lod-1.14d", "expansion": true,
			"difficulty": 2, "hardcore": true, "maximum_players": 2},
	}
	if _, driftEngine, driftSession, driftErr := start(drifted, checkpoint.Participants); driftErr == nil {
		_ = driftSession.Close()
		_ = driftEngine.Close()
		t.Fatal("checkpoint accepted different immutable game rules")
	}
}

func TestAuthorityMonsterSpawnUsesCheckpointedLuaRandomStream(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	defer engine.Close()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	authority, err := Start(ctx, content.D2Legacy(), runtimeFixtureRecords{}, engine, session, 99)
	if err != nil {
		t.Fatal(err)
	}
	defer authority.Stop(ctx)
	payload, _ := json.Marshal(map[string]any{
		"spawn_id": "fallen-1", "seed": 123, "x": 10, "y": 20, "act": 1, "level_id": 2,
		"definition": map[string]any{
			"id": "fallen", "base_id": "fallen", "graphics_id": "fallen", "name_key": "Fallen",
			"ai": "fallen", "token": "FA", "weapon_class": "HTH", "components": map[string]string{},
			"life_min": 256, "life_max": 768, "level": 1, "defense": 0, "attack_rating": 0,
			"physical_min": 256, "physical_max": 256, "experience": 5, "treasure_class": "Act 1 H2H A",
			"collider_radius": 1, "select_radius": 1, "velocity": 5, "think_interval": 1,
			"aggro_radius": 20, "attack_range": 1,
		},
	})
	if err := session.Submit(simulation.Command{Tick: 1, Player: "population", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.monster.spawn", Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	stats, found := akara.GetDynamicStore(engine.World(), "d2legacy.monster.stats")
	if !found || stats.Len() != 1 {
		t.Fatalf("Lua spawn created %d monster stats", stats.Len())
	}
	value, _ := stats.Get(stats.Entities()[0])
	health, _ := value.Get("health")
	if health != int64(256) && health != int64(512) && health != int64(768) {
		t.Fatalf("spawned health = %v, want authored whole point", health)
	}
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	if len(replay.InitialParticipants) != 3 {
		t.Fatalf("participant states = %d, want identity, Lua state, and RNG", len(replay.InitialParticipants))
	}
}

func TestMonsterMeleeReachIncludesBothActorFootprints(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	defer engine.Close()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	authority, err := Start(ctx, content.D2Legacy(), runtimeFixtureRecords{}, engine, session, 17)
	if err != nil {
		t.Fatal(err)
	}
	defer authority.Stop(ctx)

	player, _ := json.Marshal(map[string]any{
		"character_id": "hero", "player": "alice", "name": "Hero", "class": "Amazon",
		"level": 1, "experience": 0, "dexterity": 20, "defense": 20,
		"health": 500, "max_health": 500, "mana": 20, "max_mana": 20,
		"expansion": true, "hardcore": false, "cof": "", "palette": "units",
		"direction": 0, "mode": "NU", "x": 10, "y": 10,
		"world_width": 100, "world_height": 100, "act": 1, "level_id": 1,
	})
	monster, _ := json.Marshal(map[string]any{
		"spawn_id": "spacing-fallen", "seed": 9, "x": 12.5, "y": 10, "act": 1, "level_id": 1,
		"definition": map[string]any{
			"id": "fallen", "base_id": "fallen", "graphics_id": "fallen", "name_key": "Fallen",
			"ai": "fallen", "token": "FA", "weapon_class": "HTH", "components": map[string]string{},
			"life_min": 256, "life_max": 256, "level": 1, "defense": 0, "attack_rating": 0,
			"physical_min": 256, "physical_max": 256, "experience": 5, "treasure_class": "",
			"collider_radius": 1, "select_radius": 1, "velocity": 5, "think_interval": 1,
			"aggro_radius": 20, "attack_range": 1,
		},
	})
	for _, command := range []simulation.Command{
		{Tick: 1, Player: "system", Authority: simulation.AuthoritySystem, Sequence: 1, Kind: "system.player.enter", Payload: player},
		{Tick: 1, Player: "population", Authority: simulation.AuthoritySystem, Sequence: 1, Kind: "system.monster.spawn", Payload: monster},
	} {
		if err := session.Submit(command); err != nil {
			t.Fatal(err)
		}
	}
	for range 3 {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
	}

	positions, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.position")
	selectables, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.selectable")
	var playerX, playerY, monsterX, monsterY float64
	for _, entity := range selectables.Entities() {
		selectable, _ := selectables.Get(entity)
		id, _ := selectable.Get("id")
		position, _ := positions.Get(entity)
		x, _ := position.Get("x")
		y, _ := position.Get("y")
		if id == "player:alice" {
			playerX, playerY = x.(float64), y.(float64)
		} else if id == "monster:spacing-fallen" {
			monsterX, monsterY = x.(float64), y.(float64)
		}
	}
	distance := math.Hypot(monsterX-playerX, monsterY-playerY)
	if distance < 2 {
		t.Fatalf("actor centers separated by %v, want at least combined collider radii 2", distance)
	}
	brains, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.ai")
	brain, _ := brains.Get(brains.Entities()[0])
	state, _ := brain.Get("state")
	if state != "attack" {
		t.Fatalf("monster state = %v at distance %v, want attack within footprint-aware reach 3", state, distance)
	}
}

type runtimeFixtureRecords struct{}

func (runtimeFixtureRecords) Invalidate(string)  {}
func (runtimeFixtureRecords) Loaded(string) bool { return true }
func (runtimeFixtureRecords) Load(path string) ([]map[string]string, error) {
	switch path {
	case "data/global/excel/charstats.txt":
		return []map[string]string{{"class": "Amazon", "StartSkill": "Fire Bolt"}}, nil
	case "data/global/excel/skilldesc.txt":
		return []map[string]string{{"skilldesc": "firebolt", "ListRow": "0", "IconCel": "0"}}, nil
	case "data/global/excel/skills.txt":
		return []map[string]string{{"Id": "36", "skill": "Fire Bolt", "skilldesc": "firebolt", "leftskill": "1", "general": "0", "passive": "0", "srvmissile": "firebolt", "etype": "fire", "interrupt": "1", "srvstfunc": "", "srvdofunc": "", "mana": "5", "manashift": "7", "emin": "3", "emax": "6", "HitShift": "8"}}, nil
	case "data/global/excel/Missiles.txt":
		return []map[string]string{{"Missile": "firebolt", "Skill": "Fire Bolt", "pSrvDoFunc": "1", "CollideType": "3", "CollideKill": "1", "Vel": "20", "Range": "40", "Size": "2", "CelFile": "firebolt", "AnimSpeed": "16", "NumDirections": "16", "LoopAnim": "1"}}, nil
	}
	return nil, nil
}

func TestAuthorityRestoresAllDeterministicParticipantsBeforeFirstTick(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	authority, err := Start(ctx, content.D2Legacy(), runtimeFixtureRecords{}, engine, session, 7)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := authority.Random.Uint64n("d2legacy.combat.missile.damage", 100); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	if err := authority.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	session.Close()

	checkpoint := replay.Checkpoints[0]
	restoredEngine, err := gameecs.RestoreSnapshot(*checkpoint.Snapshot)
	if err != nil {
		t.Fatal(err)
	}
	restoredSession, err := gamesession.New(restoredEngine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer restoredSession.Close()
	restored, err := StartWithConfig(ctx, content.D2Legacy(), runtimeFixtureRecords{}, restoredEngine, restoredSession, Config{Seed: 7, Restore: checkpoint.Participants})
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Stop(ctx)
	restoredReplay, err := restoredSession.Replay()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(checkpoint.Participants, restoredReplay.InitialParticipants) {
		t.Fatalf("restored participants differ\nwant: %#v\n got: %#v", checkpoint.Participants, restoredReplay.InitialParticipants)
	}
	assertParticipantIDs(t, restoredReplay.InitialParticipants,
		"engine.authoritative_rng/v1", "engine.authoritative_runtime/v1", "engine.authoritative_state/v1")
}

func TestAuthorityCheckpointRestoreContinuesWithIdenticalOutcome(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	authority, err := Start(ctx, content.D2Legacy(), runtimeFixtureRecords{}, engine, session, 77)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := replay.Checkpoints[len(replay.Checkpoints)-1]

	spawnPayload, _ := json.Marshal(map[string]any{"spawn_id": "continued-fallen", "seed": float64(9), "x": float64(8), "y": float64(9), "act": float64(1), "level_id": float64(2), "definition": map[string]any{
		"id": "fallen", "base_id": "fallen", "graphics_id": "fallen", "name_key": "Fallen", "ai": "fallen", "token": "FA", "weapon_class": "HTH", "components": map[string]string{},
		"life_min": float64(256), "life_max": float64(768), "level": float64(1), "defense": float64(0), "attack_rating": float64(0), "physical_min": float64(256), "physical_max": float64(256), "experience": float64(5), "treasure_class": "Act 1 H2H A", "collider_radius": float64(1), "select_radius": float64(1), "velocity": float64(5), "think_interval": float64(1), "aggro_radius": float64(20), "attack_range": float64(1)}})
	command := simulation.Command{Tick: 2, Player: "population", Authority: simulation.AuthoritySystem, Sequence: 1, Kind: "system.monster.spawn", Payload: spawnPayload}
	if err := session.Submit(command); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	originalReplay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	originalChecksum := originalReplay.Checkpoints[len(originalReplay.Checkpoints)-1].Checksum
	if err := authority.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	session.Close()

	restoredEngine, err := gameecs.RestoreSnapshot(*checkpoint.Snapshot)
	if err != nil {
		t.Fatal(err)
	}
	restoredSession, err := gamesession.New(restoredEngine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer restoredSession.Close()
	restored, err := StartWithConfig(ctx, content.D2Legacy(), runtimeFixtureRecords{}, restoredEngine, restoredSession, Config{Seed: 77, Restore: checkpoint.Participants})
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Stop(ctx)
	if err := restoredSession.Submit(command); err != nil {
		t.Fatal(err)
	}
	if err := restoredSession.Step(); err != nil {
		t.Fatal(err)
	}
	restoredReplay, err := restoredSession.Replay()
	if err != nil {
		t.Fatal(err)
	}
	restoredChecksum := restoredReplay.Checkpoints[len(restoredReplay.Checkpoints)-1].Checksum
	if restoredChecksum != originalChecksum {
		t.Fatalf("continued checksum = %s, want %s", restoredChecksum, originalChecksum)
	}
}

// TestStraightMissileCheckpointRestoreParity covers the complete migrated vertical
// path rather than a synthetic state mutation: Lua admits the player command,
// pays mana, advances cast timing, creates and moves a missile, resolves swept
// contact, applies damage, and emits the combat result. A newly constructed Lua
// runtime must continue the in-flight cast to the identical session checksum.
func TestStraightMissileCheckpointRestoreParity(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	authority, err := Start(ctx, content.D2Legacy(), runtimeFixtureRecords{}, engine, session, 123)
	if err != nil {
		t.Fatal(err)
	}

	playerPayload, _ := json.Marshal(map[string]any{
		"character_id": "hero", "player": "alice", "name": "Hero", "class": "Amazon",
		"level": 1, "experience": 0, "dexterity": 20, "defense": 0,
		"health": 50, "max_health": 50, "mana": 20, "max_mana": 20,
		"expansion": true, "hardcore": false, "cof": "", "palette": "units",
		"direction": 0, "mode": "NU", "x": 0, "y": 0,
		"world_width": 100, "world_height": 100, "act": 1, "level_id": 1,
	})
	monsterPayload, _ := json.Marshal(map[string]any{
		"spawn_id": "fallen-missile", "seed": 9, "x": 4, "y": 0, "act": 1, "level_id": 1,
		"definition": map[string]any{
			"id": "fallen", "base_id": "fallen", "graphics_id": "fallen", "name_key": "Fallen",
			"ai": "fallen", "token": "FA", "weapon_class": "HTH", "components": map[string]string{},
			"life_min": 4096, "life_max": 4096, "level": 1, "defense": 0, "attack_rating": 0,
			"physical_min": 256, "physical_max": 256, "experience": 5, "treasure_class": "",
			"collider_radius": 0.5, "select_radius": 0.5, "velocity": 0,
			"think_interval": 100, "aggro_radius": 0, "attack_range": 1,
		},
	})
	for _, command := range []simulation.Command{
		{Tick: 1, Player: "system", Authority: simulation.AuthoritySystem, Sequence: 1, Kind: "system.player.enter", Payload: playerPayload},
		{Tick: 1, Player: "population", Authority: simulation.AuthoritySystem, Sequence: 1, Kind: "system.monster.spawn", Payload: monsterPayload},
	} {
		if err := session.Submit(command); err != nil {
			t.Fatal(err)
		}
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	// No unit target is supplied: this is the ground-targeted missile vector.
	castPayload, _ := json.Marshal(map[string]any{"side": "left", "target_x": 8, "target_y": 0})
	if err := session.Submit(simulation.Command{Tick: 2, Player: "alice", Authority: simulation.AuthorityPlayer, Sequence: 1, Kind: "player.use_skill", Payload: castPayload}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := replay.Checkpoints[len(replay.Checkpoints)-1]

	for range 6 {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
	}
	originalReplay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	original := originalReplay.Checkpoints[len(originalReplay.Checkpoints)-1]
	if err := authority.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if err := session.Close(); err != nil {
		t.Fatal(err)
	}

	restoredEngine, err := gameecs.RestoreSnapshot(*checkpoint.Snapshot)
	if err != nil {
		t.Fatal(err)
	}
	defer restoredEngine.Close()
	restoredSession, err := gamesession.New(restoredEngine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer restoredSession.Close()
	restored, err := StartWithConfig(ctx, content.D2Legacy(), runtimeFixtureRecords{}, restoredEngine, restoredSession, Config{Seed: 123, Restore: checkpoint.Participants})
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Stop(ctx)
	for range 6 {
		if err := restoredSession.Step(); err != nil {
			t.Fatal(err)
		}
	}
	restoredReplay, err := restoredSession.Replay()
	if err != nil {
		t.Fatal(err)
	}
	continued := restoredReplay.Checkpoints[len(restoredReplay.Checkpoints)-1]
	if continued.Checksum != original.Checksum {
		t.Fatalf("straight-missile continuation checksum = %s, want %s", continued.Checksum, original.Checksum)
	}
}

func assertParticipantIDs(t *testing.T, states []simulation.ParticipantState, expected ...string) {
	t.Helper()
	if len(states) != len(expected) {
		t.Fatalf("participant count = %d, want %d", len(states), len(expected))
	}
	for index, id := range expected {
		if states[index].ID != id {
			t.Fatalf("participant %d = %q, want %q", index, states[index].ID, id)
		}
	}
}

func TestAuthorityBootsWithoutClientOrRenderer(t *testing.T) {
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	authority, err := Start(context.Background(), content.D2Legacy(), runtimeFixtureRecords{}, engine, session, 7)
	if err != nil {
		t.Fatal(err)
	}
	if err := authority.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
}
