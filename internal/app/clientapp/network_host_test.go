package clientapp

import (
	"reflect"
	"strings"
	"testing"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// TestListenDestinationRequiresCompleteTrustedWorldState verifies that host
// admission never falls back to partially prepared client-world data.
func TestListenDestinationRequiresCompleteTrustedWorldState(t *testing.T) {
	controller := newNetworkController(&application{activeWorldLevel: 1})

	_, err := controller.listenDestination()
	if err == nil || !strings.Contains(err.Error(), "trusted host destination") {
		t.Fatalf("incomplete destination error = %v", err)
	}
}

// TestListenDestinationUsesActiveWorldBounds verifies that profile admission
// receives the selected level's spawn, dimensions, act, and level identity.
func TestListenDestinationUsesActiveWorldBounds(t *testing.T) {
	world, err := gameworld.NewOpenMap(40, 60)
	if err != nil {
		t.Fatal(err)
	}

	zone, err := worldgen.NewZone(worldgen.Definition{
		Request: worldgen.Request{
			Version: worldgen.ContractVersion,
			Act:     2,
			LevelID: 41,
		},
		Kind:   worldgen.Kind("test"),
		Bounds: worldgen.Bounds{Width: 8, Height: 12},
	})
	if err != nil {
		t.Fatal(err)
	}

	app := &application{
		activeWorldLevel: 41,
		gameWorlds:       map[int]*gameworld.Map{41: world},
		gameWorldZones:   map[int]*worldgen.Zone{41: zone},
		gameWorldSpawns:  map[int][2]float64{41: {5, 7}},
	}
	controller := newNetworkController(app)

	destination, err := controller.listenDestination()
	if err != nil {
		t.Fatal(err)
	}

	want := playeradapter.Destination{
		X:       5,
		Y:       7,
		Width:   40,
		Height:  60,
		Act:     2,
		LevelID: 41,
	}
	if !reflect.DeepEqual(destination, want) {
		t.Fatalf("listen destination = %#v, want %#v", destination, want)
	}
}

// TestCommitListenHostRejectsStaleGeneration verifies that canceled startup
// cannot transfer ownership of its resources into a newer controller run.
func TestCommitListenHostRejectsStaleGeneration(t *testing.T) {
	controller := newNetworkController(&application{})
	controller.generation = 2
	controller.phase = "starting"

	if controller.commitListenHost(1, &listenHostRuntime{}) {
		t.Fatal("stale host generation committed resources")
	}
}
