package d2legacy

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// Both cave endpoints reserve their interaction footprint in addition to
// ordinary terrain and entity collision checks.
func TestDenPopulationSkipsEndpointFootprints(t *testing.T) {
	authority, engine, session := coldPopulationSession(t, 0, denPopulationRecords(), nil)
	denMap, err := gameworld.NewOpenMap(100, 100)
	if err != nil {
		t.Fatal(err)
	}
	if err := modruntime.SetWorldMapForLevel(t.Context(), authority.Runtime,
		"d2legacy.gameplay.systems.init", "set_collision", 8, denMap); err != nil {
		t.Fatal(err)
	}
	zone := denPopulationZone(0)
	zone["exclusions"] = []any{
		map[string]any{"x": float64(10), "y": float64(10), "radius": float64(1)},
		map[string]any{"x": float64(20), "y": float64(10), "radius": float64(3.5)},
	}
	coldBootstrap(t, session, 1, coldZone(2, 0), coldZone(3, 0), coldZone(4, 0), zone)
	coldEnter(t, session, 8, "alice", 1)
	stepPartySession(t, session)
	entity := monsterEntity(t, engine, "level:8:room:same:monster:1")
	positions, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.position")
	position, _ := positions.Get(entity)
	x, _ := position.Get("x")
	y, _ := position.Get("y")
	if x != float64(15) || y != float64(10) {
		t.Fatalf("Den resident at (%v,%v), want only safe center (15,10)", x, y)
	}
	if got := monsterCount(engine); got != 1 {
		t.Fatalf("Den created %d residents, want one away from cave endpoints", got)
	}
}

// An eligible but fully blocked room reports why it produced no resident,
// without creating a monster on the clicked return or arrival position.
func TestDenPopulationExplainsNoSafePoint(t *testing.T) {
	authority, engine, session := coldPopulationSession(t, 0, denPopulationRecords(), nil)
	denMap, err := gameworld.NewOpenMap(100, 100)
	if err != nil {
		t.Fatal(err)
	}
	if err := modruntime.SetWorldMapForLevel(t.Context(), authority.Runtime,
		"d2legacy.gameplay.systems.init", "set_collision", 8, denMap); err != nil {
		t.Fatal(err)
	}
	zone := denPopulationZone(0)
	zone["rooms"].([]any)[0].(map[string]any)["points"] = []any{
		map[string]any{"x": float64(10), "y": float64(10)},
		map[string]any{"x": float64(20), "y": float64(10)},
	}
	zone["exclusions"] = []any{
		map[string]any{"x": float64(10), "y": float64(10), "radius": float64(1)},
		map[string]any{"x": float64(20), "y": float64(10), "radius": float64(3.5)},
	}
	coldBootstrap(t, session, 1, coldZone(2, 0), coldZone(3, 0), coldZone(4, 0), zone)
	coldEnter(t, session, 8, "alice", 1)
	stepPartySession(t, session)
	if got := monsterCount(engine); got != 0 {
		t.Fatalf("unsafe Den points created %d residents", got)
	}
	rooms := coldPlan(t, authority)[3]["rooms"].([]any)
	if reason := rooms[0].(map[string]any)["population_skip_reason"]; reason != "no_safe_point" {
		t.Fatalf("Den room skip reason=%v, want no_safe_point", reason)
	}
}

// denPopulationRecords supplies a minimal level-8 row to isolate Lua admission and activation policy.
func denPopulationRecords() fixtureRecords {
	records := stonyPopulationRecords()
	row := map[string]string{}
	for key, value := range records["data/global/excel/levels.txt"][0] {
		row[key] = value
	}
	for key, value := range map[string]string{
		"Id": "8", "NumMon": "1", "mon1": "cold", "nmon1": "cold",
		"MonDen": "100000", "MonDen(N)": "100000", "MonDen(H)": "100000",
	} {
		row[key] = value
	}
	records["data/global/excel/levels.txt"] = append(records["data/global/excel/levels.txt"], row)
	return records
}

// denPopulationZone adds the required cave envelope to the ordinary zone fixture.
func denPopulationZone(difficulty int) map[string]any {
	zone := coldZone(8, difficulty)
	zone["exclusions"] = []any{
		map[string]any{"x": float64(30), "y": float64(30), "radius": float64(1)},
		map[string]any{"x": float64(40), "y": float64(40), "radius": float64(3.5)},
	}
	return zone
}

// A single system bootstrap owns all four zones; only the level-8 player can
// activate ordinary Den residents.
func TestDenPopulationAdmitsAndSpawns(t *testing.T) {
	for difficulty := 0; difficulty < 3; difficulty++ {
		t.Run(fmt.Sprint(difficulty), func(t *testing.T) {
			authority, engine, session := coldPopulationSession(t, difficulty, denPopulationRecords(), nil)
			denMap, err := gameworld.NewOpenMap(100, 100)
			if err != nil {
				t.Fatal(err)
			}
			if err := modruntime.SetWorldMapForLevel(t.Context(), authority.Runtime,
				"d2legacy.gameplay.systems.init", "set_collision", 8, denMap); err != nil {
				t.Fatal(err)
			}
			coldBootstrap(t, session, 1,
				coldZone(2, difficulty), coldZone(3, difficulty), coldZone(4, difficulty), denPopulationZone(difficulty))
			coldEnter(t, session, 8, "alice", 1)
			stepPartySession(t, session)
			zones := coldPlan(t, authority)
			if len(zones) != 4 || zones[3]["level_id"] != float64(8) {
				t.Fatalf("installed zones=%d, want final Den zone", len(zones))
			}
			monsterEntity(t, engine, "level:8:room:same:monster:1")
		})
	}
}

// Reject malformed cave envelopes before installing a plan or creating any ECS resident.
func TestDenPopulationRejectsIncompleteCave(t *testing.T) {
	for _, test := range []struct {
		name string
		make func() []map[string]any
	}{
		{"missing level 3", func() []map[string]any { return []map[string]any{coldZone(2, 0), denPopulationZone(0)} }},
		{"missing level 4", func() []map[string]any {
			return []map[string]any{coldZone(2, 0), coldZone(3, 0), denPopulationZone(0)}
		}},
		{"duplicate level 8", func() []map[string]any {
			return []map[string]any{coldZone(2, 0), coldZone(3, 0), coldZone(4, 0), denPopulationZone(0), denPopulationZone(0)}
		}},
		{"missing exclusions", func() []map[string]any {
			zone := denPopulationZone(0)
			delete(zone, "exclusions")
			return []map[string]any{coldZone(2, 0), coldZone(3, 0), coldZone(4, 0), zone}
		}},
		{"missing return exclusion", func() []map[string]any {
			zone := denPopulationZone(0)
			zone["exclusions"] = zone["exclusions"].([]any)[:1]
			return []map[string]any{coldZone(2, 0), coldZone(3, 0), coldZone(4, 0), zone}
		}},
		{"non-number coordinate", func() []map[string]any {
			zone := denPopulationZone(0)
			zone["exclusions"].([]any)[0].(map[string]any)["x"] = "NaN"
			return []map[string]any{coldZone(2, 0), coldZone(3, 0), coldZone(4, 0), zone}
		}},
		{"outside room", func() []map[string]any {
			zone := denPopulationZone(0)
			zone["exclusions"].([]any)[0].(map[string]any)["x"] = float64(60)
			return []map[string]any{coldZone(2, 0), coldZone(3, 0), coldZone(4, 0), zone}
		}},
		{"zero arrival radius", func() []map[string]any {
			zone := denPopulationZone(0)
			zone["exclusions"].([]any)[0].(map[string]any)["radius"] = float64(0)
			return []map[string]any{coldZone(2, 0), coldZone(3, 0), coldZone(4, 0), zone}
		}},
		{"small return radius", func() []map[string]any {
			zone := denPopulationZone(0)
			zone["exclusions"].([]any)[1].(map[string]any)["radius"] = float64(3)
			return []map[string]any{coldZone(2, 0), coldZone(3, 0), coldZone(4, 0), zone}
		}},
		{"wrong difficulty", func() []map[string]any {
			return []map[string]any{coldZone(2, 0), coldZone(3, 0), coldZone(4, 0), denPopulationZone(1)}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			authority, engine, session := coldPopulationSession(t, 0, denPopulationRecords(), nil)
			before, _ := authority.State.Read("d2legacy.population.plan")
			payload, err := json.Marshal(map[string]any{"zones": test.make()})
			if err != nil {
				t.Fatal(err)
			}
			err = session.Submit(simulation.Command{Tick: 1, Player: "population", Authority: simulation.AuthoritySystem,
				Sequence: 1, Kind: "system.population.bootstrap", Payload: payload})
			if err == nil {
				t.Fatal("invalid Den envelope was admitted")
			}
			after, _ := authority.State.Read("d2legacy.population.plan")
			if !reflect.DeepEqual(before, after) || monsterCount(engine) != 0 {
				t.Fatal("rejected Den envelope mutated authority")
			}
		})
	}
}

// The existing monster interpreter must reject an area whose referenced
// family is not an eligible ordinary spawn, including in the new cave zone.
func TestDenPopulationRejectsInvalidCandidates(t *testing.T) {
	records := denPopulationRecords()
	records["data/global/excel/levels.txt"][3]["mon1"] = "missing-den-family"
	authority, engine, session := coldPopulationSession(t, 0, records, nil)
	denMap, err := gameworld.NewOpenMap(100, 100)
	if err != nil {
		t.Fatal(err)
	}
	if err := modruntime.SetWorldMapForLevel(t.Context(), authority.Runtime,
		"d2legacy.gameplay.systems.init", "set_collision", 8, denMap); err != nil {
		t.Fatal(err)
	}
	coldBootstrap(t, session, 1, coldZone(2, 0), coldZone(3, 0), coldZone(4, 0), denPopulationZone(0))
	coldEnter(t, session, 8, "alice", 1)
	if err := session.Step(); err == nil || !strings.Contains(err.Error(), "no valid monster candidates") {
		t.Fatalf("invalid Den candidate error=%v", err)
	}
	if monsterCount(engine) != 0 {
		t.Fatal("invalid candidate created a Den resident")
	}
}
