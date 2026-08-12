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

// TestGeneratedHostileLifecycleRestoresIdentically closes the last M21.14.7
// acceptance gap. The monster is selected and materialized by the production
// Blood Moor population policy, acquires the player through production AI, and
// dies through the production cast/projectile/damage/death systems. A newly
// constructed Lua runtime must continue that live hostile to the same corpse,
// credit, XP, loot-event surface, and authoritative checksum.
func TestGeneratedHostileLifecycleRestoresIdentically(t *testing.T) {
	ctx := context.Background()
	records := generatedHostileRecords()
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	authority, err := Start(ctx, content.D2Legacy(), records, engine, session, 314)
	if err != nil {
		t.Fatal(err)
	}

	playerPayload, _ := json.Marshal(map[string]any{
		"character_id": "hero", "player": "alice", "name": "Hero", "class": "Amazon",
		"level": 1, "experience": 0, "dexterity": 20, "defense": 0,
		"health": 50, "max_health": 50, "mana": 20, "max_mana": 20,
		"expansion": true, "hardcore": false, "cof": "", "palette": "units",
		"direction": 0, "mode": "NU", "x": 0, "y": 0,
		"world_width": 100, "world_height": 100, "act": 1, "level_id": 2,
	})
	populationPayload, _ := json.Marshal(map[string]any{
		"act": 1, "level_id": 2, "difficulty": 0,
		"rooms": []map[string]any{{
			"id": "blood-moor-a", "populate": true,
			"points": []map[string]any{{"x": 4, "y": 0}},
		}},
	})
	for _, command := range []simulation.Command{
		{Tick: 1, Player: "system", Authority: simulation.AuthoritySystem, Sequence: 1, Kind: "system.player.enter", Payload: playerPayload},
		{Tick: 1, Player: "population", Authority: simulation.AuthoritySystem, Sequence: 1, Kind: "system.population.bootstrap", Payload: populationPayload},
	} {
		if err := session.Submit(command); err != nil {
			t.Fatal(err)
		}
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := replay.Checkpoints[len(replay.Checkpoints)-1]

	castPayload, _ := json.Marshal(map[string]any{
		"side": "left", "target_x": 4, "target_y": 0,
		"target_id": "monster:level:2:room:blood-moor-a:monster:1",
	})
	cast := simulation.Command{Tick: 2, Player: "alice", Authority: simulation.AuthorityPlayer, Sequence: 1, Kind: "player.use_skill", Payload: castPayload}
	if err := session.Submit(cast); err != nil {
		t.Fatal(err)
	}
	stepSession(t, session, 8)
	originalReplay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	original := originalReplay.Checkpoints[len(originalReplay.Checkpoints)-1]
	assertCompletedHostileLifecycle(t, engine)
	if err := authority.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if err := session.Close(); err != nil {
		t.Fatal(err)
	}
	if err := engine.Close(); err != nil {
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
	restored, err := StartWithConfig(ctx, content.D2Legacy(), records, restoredEngine, restoredSession, Config{Seed: 314, Restore: checkpoint.Participants})
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Stop(ctx)
	if err := restoredSession.Submit(cast); err != nil {
		t.Fatal(err)
	}
	stepSession(t, restoredSession, 8)
	restoredReplay, err := restoredSession.Replay()
	if err != nil {
		t.Fatal(err)
	}
	continued := restoredReplay.Checkpoints[len(restoredReplay.Checkpoints)-1]
	if continued.Checksum != original.Checksum {
		t.Fatalf("restored hostile lifecycle checksum = %s, want %s", continued.Checksum, original.Checksum)
	}
	assertCompletedHostileLifecycle(t, restoredEngine)
}

func stepSession(t *testing.T, session *gamesession.Session, count int) {
	t.Helper()
	for range count {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
	}
}

func assertCompletedHostileLifecycle(t *testing.T, engine *gameecs.Engine) {
	t.Helper()
	identities, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.identity")
	deaths, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.death")
	events, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.death_event")
	selectables, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.selectable")
	brains, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.ai")
	progress, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.progress")
	if identities.Len() != 1 || deaths.Len() != 1 {
		t.Fatalf("monster identities/deaths = %d/%d, want 1/1", identities.Len(), deaths.Len())
	}
	monster := identities.Entities()[0]
	death, present := deaths.Get(monster)
	if !present {
		t.Fatal("generated monster has no durable death state")
	}
	credited, _ := death.Get("credited_id")
	active, _ := death.Get("active")
	corpse, _ := death.Get("corpse_usable")
	if credited != "player:alice" || active != false || corpse != true {
		t.Fatalf("death credit/active/corpse = %v/%v/%v", credited, active, corpse)
	}
	if events.Len() != 4 {
		t.Fatalf("death events = %d, want kill, loot, quest, and presentation", events.Len())
	}
	if _, present := brains.Get(monster); present {
		t.Fatal("dead monster retained active AI")
	}
	for _, entity := range selectables.Entities() {
		value, _ := selectables.Get(entity)
		id, _ := value.Get("id")
		if id == "monster:level:2:room:blood-moor-a:monster:1" {
			t.Fatal("dead monster remained targetable")
		}
	}
	if progress.Len() != 1 {
		t.Fatalf("player progress records = %d, want 1", progress.Len())
	}
	value, _ := progress.Get(progress.Entities()[0])
	experience, _ := value.Get("experience")
	if experience != int64(5) {
		t.Fatalf("credited experience = %v, want 5", experience)
	}
}

func generatedHostileRecords() fixtureRecords {
	records := fixtureRecords{}
	records["data/global/excel/levels.txt"] = []map[string]string{{
		"Id": "2", "MonDen": "100000", "NumMon": "1", "mon1": "fallen",
	}}
	records["data/global/excel/monstats.txt"] = []map[string]string{{
		"Id": "fallen", "BaseId": "fallen", "NameStr": "Fallen", "AI": "fallen", "Code": "FA",
		"enabled": "1", "isSpawn": "1", "npc": "0", "noRatio": "1", "Level": "1",
		"minHP": "3", "maxHP": "3", "AC": "0", "A1TH": "0", "A1MinD": "1", "A1MaxD": "1",
		"Exp": "5", "Velocity": "0", "aidel": "1", "aidist": "20", "MinGrp": "1", "MaxGrp": "1",
		"Rarity": "1", "TreasureClass1": "",
	}}
	records["data/global/excel/monstats2.txt"] = []map[string]string{{
		"Id": "fallen", "BaseW": "HTH", "SizeX": "1", "SizeY": "1", "MeleeRng": "1",
	}}
	records["data/global/excel/monlvl.txt"] = []map[string]string{{"Level": "1"}}
	return records
}
