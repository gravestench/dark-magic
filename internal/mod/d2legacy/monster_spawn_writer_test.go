package d2legacy

import (
	"encoding/json"
	"testing"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func TestGeneratedBossSpawnWriter(t *testing.T) {
	records := generatedHostileRecords()
	records["data/global/excel/monstats.txt"][0]["boss"] = "1"
	records["data/global/excel/monstats.txt"][0]["primeevil"] = "1"

	engine := gameecs.New()
	defer func() { _ = engine.Close() }()
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = session.Close() }()
	authority, err := Start(t.Context(), content.D2Legacy(), records, engine, session, 314)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = authority.Stop(t.Context()) }()

	population, _ := json.Marshal(map[string]any{
		"act": 1, "level_id": 2, "difficulty": 0,
		"links": []map[string]any{},
		"rooms": []map[string]any{{
			"id": "boss-room", "populate": true,
			"x": 0, "y": 0, "width": 10, "height": 10,
			"points": []map[string]any{{"x": 4, "y": 0}},
		}},
	})
	if err := session.Submit(simulation.Command{Tick: 1, Player: "population", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.population.bootstrap", Payload: population}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}

	player, _ := json.Marshal(map[string]any{
		"character_id": "hero", "player": "alice", "name": "Hero", "class": "Amazon",
		"level": 1, "experience": 0, "dexterity": 20, "vitality": 20, "defense": 0,
		"health": 50, "max_health": 50, "mana": 20, "max_mana": 20,
		"stamina": 84, "max_stamina": 84,
		"expansion": true, "hardcore": false, "cof": "", "palette": "units",
		"direction": 0, "mode": "NU", "x": 0, "y": 0,
		"world_width": 100, "world_height": 100, "act": 1, "level_id": 2,
	})
	if err := session.Submit(simulation.Command{Tick: 2, Player: "system", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.player.enter", Payload: player}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}

	assertGeneratedBossMarkers(t, engine)
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := replay.Checkpoints[len(replay.Checkpoints)-1]
	restored, err := gameecs.RestoreSnapshot(*checkpoint.Snapshot)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = restored.Close() }()
	assertGeneratedBossMarkers(t, restored)
}

func assertGeneratedBossMarkers(t *testing.T, engine *gameecs.Engine) {
	t.Helper()
	for _, name := range []string{
		"d2legacy.monster.boss",
		"d2legacy.monster.freeze_immune",
		"d2legacy.monster.prime_evil",
	} {
		store, ok := akara.GetDynamicStore(engine.World(), name)
		if !ok {
			t.Fatalf("missing component store %s", name)
		}
		if store.Len() != 1 {
			t.Fatalf("%s count = %d, want 1", name, store.Len())
		}
	}
}
