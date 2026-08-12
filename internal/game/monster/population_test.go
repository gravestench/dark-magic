package monster

import (
	"testing"

	"github.com/gravestench/akara"
	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
	models "github.com/gravestench/dark-magic/internal/game/data/model"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/mapgen"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
)

type openPlacement struct{}

func (openPlacement) OpenPointNearSubtile(x, y float64) (float64, float64, bool) { return x, y, true }

func TestBloodMoorPopulationIsDeterministicInspectableAndMaterialized(t *testing.T) {
	zone := populationZone(t, 42)
	snapshot := populationSnapshot()
	left, err := BuildBloodMoorPopulation(zone, openPlacement{}, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	right, err := BuildBloodMoorPopulation(zone, openPlacement{}, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	leftChecksum, _ := left.Checksum()
	rightChecksum, _ := right.Checksum()
	if leftChecksum != rightChecksum {
		t.Fatal("same zone and records changed population")
	}
	if left.Policy != PopulationPolicy || left.Stream != populationStream || len(left.Spawns) != 4 || len(left.Trace) != 3 {
		t.Fatalf("plan = %#v", left)
	}
	for _, spawn := range left.Spawns {
		if spawn.Definition.ID != "fallen" || spawn.LevelID != 2 || spawn.Act != 1 {
			t.Fatalf("spawn = %#v", spawn)
		}
	}

	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := Register(session); err != nil {
		t.Fatal(err)
	}
	if err := SubmitPopulation(session, left, "population", 1); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	identities, found := akara.GetDynamicStore(engine.World(), "d2legacy.monster.identity")
	if !found || identities.Len() != len(left.Spawns) {
		t.Fatalf("materialized=%d found=%v", identities.Len(), found)
	}
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	if len(replay.Commands) != len(left.Spawns) {
		t.Fatalf("replay commands=%d", len(replay.Commands))
	}
}

func TestBloodMoorPopulationChangesWithZoneSeed(t *testing.T) {
	left, _ := BuildBloodMoorPopulation(populationZone(t, 42), openPlacement{}, populationSnapshot())
	right, _ := BuildBloodMoorPopulation(populationZone(t, 43), openPlacement{}, populationSnapshot())
	a, _ := left.Checksum()
	b, _ := right.Checksum()
	if a == b {
		t.Fatal("different zone seeds produced the same canonical plan")
	}
}

func populationZone(t *testing.T, seed uint64) *mapgen.Zone {
	t.Helper()
	zone, err := mapgen.NewZone(mapgen.Definition{
		Request: mapgen.Request{Version: mapgen.ContractVersion, Seed: seed, Act: 1, LevelID: 2, Difficulty: mapgen.Normal}, Kind: mapgen.Outdoor,
		Bounds: mapgen.Bounds{Width: 24, Height: 8},
		Stamps: []mapgen.Stamp{{ID: 1, Width: 8, Height: 8, DS1Path: "a.ds1", Populate: true}, {ID: 2, X: 8, Width: 8, Height: 8, DS1Path: "b.ds1", Populate: false}, {ID: 3, X: 16, Width: 8, Height: 8, DS1Path: "c.ds1", Populate: true}},
		Rooms:  []mapgen.Room{{ID: 1, Width: 8, Height: 8, StampID: 1}, {ID: 2, X: 8, Width: 8, Height: 8, StampID: 2}, {ID: 3, X: 16, Width: 8, Height: 8, StampID: 3}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return zone
}

func populationSnapshot() gamedata.Snapshot {
	stats, graphics, levelStats := ordinaryFixture()
	stats.MinGrp, stats.MaxGrp, stats.Rarity = "2", "2", "1"
	level := models.LevelData{Id: 2, Act: 0, NumMon: 1, Mon1: "fallen", MonDen: 100000}
	return gamedata.Snapshot{
		LevelsByID: map[int]models.LevelData{2: level}, MonstersByID: map[string]models.MonsterStats{"fallen": stats},
		MonsterGfxByID: map[string]models.MonsterStats2{"fallen2": graphics}, MonsterLevelByLevel: map[int]models.MonsterLevelStats{2: levelStats},
	}
}
