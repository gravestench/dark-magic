package entryworld

import (
	"fmt"
	"math"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
)

// Removing level 4 or publishing an unsafe arrival must fail before the
// prepared topology can support a visible reciprocal crossing.
func TestStonyFieldPreparedSeamReal(t *testing.T) {
	source := realEntryWorldContent(t)
	records := recordstore.New(source)
	for difficulty := 0; difficulty <= 2; difficulty++ {
		t.Run(fmt.Sprintf("difficulty-%d", difficulty), func(t *testing.T) {
			prepared, err := Build(t.Context(), source, content.D2Legacy(), records,
				entryWorldObjectResolver{}, 17, difficulty)
			if err != nil {
				t.Fatal(err)
			}
			if len(prepared.Worlds) != 4 || len(prepared.Zones) != 4 || len(prepared.Seams) != 3 {
				t.Fatalf("prepared maps/zones/seams=%d/%d/%d",
					len(prepared.Worlds), len(prepared.Zones), len(prepared.Seams))
			}
			for index, seam := range prepared.Seams {
				if seam.First.LevelID != index+1 || seam.Second.LevelID != index+2 {
					t.Fatalf("seam %d binds wrong levels: %#v", index, seam)
				}
				for _, endpoint := range []gametransition.SeamEndpoint{seam.First, seam.Second} {
					assertOwnedSeamPoint(t, prepared.Worlds[endpoint.LevelID], endpoint.X, endpoint.Y)
					assertOwnedSeamPoint(t, prepared.Worlds[endpoint.LevelID], endpoint.ArrivalX, endpoint.ArrivalY)
				}
			}
			seam := prepared.Seams[2]
			if seam.TriggerRadius != 2 || seam.First.TriggerStrip != nil || seam.Second.TriggerStrip != nil {
				t.Fatalf("generated seam trigger=%#v", seam)
			}
			opposite := map[string]string{"north": "south", "south": "north", "east": "west", "west": "east"}
			if opposite[seam.First.Direction] != seam.Second.Direction || seam.Second.Direction == "" {
				t.Fatalf("seam directions are not reciprocal: %#v", seam)
			}
			for _, endpoint := range []gametransition.SeamEndpoint{seam.First, seam.Second} {
				world := prepared.Worlds[endpoint.LevelID]
				if endpoint.Width != float64(world.WidthSubtiles) || endpoint.Height != float64(world.HeightSubtiles) {
					t.Fatalf("level %d endpoint dimensions differ from collision", endpoint.LevelID)
				}
				arrival := gameworld.Point{X: endpoint.ArrivalX, Y: endpoint.ArrivalY}
				edge := gameworld.Point{X: endpoint.X, Y: endpoint.Y}
				if math.Hypot(arrival.X-edge.X, arrival.Y-edge.Y) <= seam.TriggerRadius {
					t.Fatalf("level %d return arrival overlaps reverse trigger", endpoint.LevelID)
				}
				inward := map[string]float64{
					"north": arrival.Y - edge.Y, "east": edge.X - arrival.X,
					"south": edge.Y - arrival.Y, "west": arrival.X - edge.X,
				}[endpoint.Direction]
				if inward <= seam.TriggerRadius {
					t.Fatalf("level %d return arrival is not inset from its edge", endpoint.LevelID)
				}
				x, y, found := world.OpenPointNearCenter()
				if !found || x < float64(world.WidthSubtiles)/4 || x >= float64(world.WidthSubtiles)*3/4 ||
					y < float64(world.HeightSubtiles)/4 || y >= float64(world.HeightSubtiles)*3/4 {
					t.Fatalf("level %d has no open interior", endpoint.LevelID)
				}
				for name, request := range map[string]gameworld.PathRequest{
					"interior": {Start: arrival, Goal: gameworld.Point{X: x, Y: y}, Radius: 1},
					"return":   {Start: arrival, Goal: edge, Radius: 1, StopRadius: seam.TriggerRadius},
				} {
					if _, err := world.FindPath(request); err != nil {
						t.Fatalf("level %d has no player-sized %s route: %v", endpoint.LevelID, name, err)
					}
				}
			}
			assertOwnedSeamAlignment(t, prepared)
			population, err := prepared.PopulationData(0)
			if err != nil {
				t.Fatal(err)
			}
			zones := population["zones"].([]any)
			if len(zones) != 3 {
				t.Fatalf("population zones=%d, want three wildernesses", len(zones))
			}
			for index, raw := range zones {
				zone := raw.(map[string]any)
				if zone["level_id"] != float64(index+2) || zone["difficulty"] != float64(difficulty) {
					t.Fatalf("population zone order/identity=%#v", zone)
				}
				rooms := zone["rooms"].([]any)
				if len(rooms) == 0 || len(rooms) != len(prepared.Zones[index+2].Rooms()) {
					t.Fatalf("level %d missing population geometry", index+2)
				}
			}

			t.Logf("owned seed 17 difficulty %d: four maps, three seams, safe aligned 3/4 routes",
				difficulty)
		})
	}
}

// assertOwnedSeamPoint checks that a trigger or arrival lies on walkable materialized collision.
func assertOwnedSeamPoint(t *testing.T, world *gameworld.Map, x, y float64) {
	t.Helper()
	flags, found := world.FlagsAt(gameworld.CollisionCell(x), gameworld.CollisionCell(y))
	if !found || flags.Blocked() || x < 0 || y < 0 ||
		x >= float64(world.WidthSubtiles) || y >= float64(world.HeightSubtiles) {
		t.Fatalf("seam point (%v,%v) is blocked or outside materialized map", x, y)
	}
}

// One translation must align each source with its destination arrival in both
// directions. This is the geometry contract consumed by outdoor_seams.layout.
func assertOwnedSeamAlignment(t *testing.T, prepared *Prepared) {
	t.Helper()
	directed := TransitionData(prepared.Seams)["seams"].([]any)
	if len(directed) != 6 {
		t.Fatalf("directed transitions=%d, want six", len(directed))
	}
	for index, seam := range prepared.Seams {
		forward := directed[index*2].(map[string]any)
		reverse := directed[index*2+1].(map[string]any)
		pairs := [][2]gametransition.SeamEndpoint{
			{seam.First, seam.Second}, {seam.Second, seam.First},
		}
		for side, endpoints := range pairs {
			data := directed[index*2+side].(map[string]any)
			source, destination := endpoints[0], endpoints[1]
			if data["source_level"] != float64(source.LevelID) ||
				data["destination_level"] != float64(destination.LevelID) ||
				data["source_x"] != source.X || data["source_y"] != source.Y ||
				data["source_direction"] != source.Direction ||
				data["arrival_x"] != destination.ArrivalX || data["arrival_y"] != destination.ArrivalY ||
				data["world_width"] != destination.Width || data["world_height"] != destination.Height {
				t.Fatalf("seam %d directed endpoint %d differs from prepared collision: %#v", index, side, data)
			}
		}
		first, second := prepared.Worlds[seam.First.LevelID], prepared.Worlds[seam.Second.LevelID]
		fx, fy := first.SubtileToPixel(forward["source_x"].(float64), forward["source_y"].(float64))
		sx, sy := second.SubtileToPixel(forward["arrival_x"].(float64), forward["arrival_y"].(float64))
		rx, ry := second.SubtileToPixel(reverse["source_x"].(float64), reverse["source_y"].(float64))
		ax, ay := first.SubtileToPixel(reverse["arrival_x"].(float64), reverse["arrival_y"].(float64))
		if math.Abs((fx-sx)+(rx-ax)) > 0.5 || math.Abs((fy-sy)+(ry-ay)) > 0.5 {
			t.Fatalf("seam %d reciprocal visual alignment residual=(%v,%v)", index, (fx-sx)+(rx-ax), (fy-sy)+(ry-ay))
		}
	}
}
