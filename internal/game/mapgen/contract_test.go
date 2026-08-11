package mapgen

import (
	"errors"
	"slices"
	"testing"
)

func validDefinition() Definition {
	return Definition{
		Request: Request{Version: ContractVersion, Seed: 42, Act: 1, LevelID: 1, Difficulty: Normal},
		Kind:    Preset,
		Bounds:  Bounds{Width: 16, Height: 16},
		Stamps:  []Stamp{{ID: 2, Width: 8, Height: 8, DS1Path: "second.ds1", TilePaths: []string{"b.dt1", "a.dt1"}}, {ID: 1, Width: 8, Height: 8, DS1Path: "first.ds1"}},
		Rooms:   []Room{{ID: 2, X: 8, Width: 8, Height: 8, StampID: 2}, {ID: 1, Width: 8, Height: 8, StampID: 1}},
		Links:   []Link{{From: 2, To: 1}},
		Trace:   []string{"preset variant stream selected file 1"},
	}
}

func TestZoneCanonicalizesAndCopiesDefinition(t *testing.T) {
	definition := validDefinition()
	zone, err := NewZone(definition)
	if err != nil {
		t.Fatal(err)
	}
	definition.Stamps[0].DS1Path = "mutated.ds1"
	definition.Stamps[0].TilePaths[0] = "mutated.dt1"
	stamps := zone.Stamps()
	if stamps[0].ID != 1 || stamps[1].DS1Path != "second.ds1" || !slices.Equal(stamps[1].TilePaths, []string{"a.dt1", "b.dt1"}) {
		t.Fatalf("canonical stamps = %#v", stamps)
	}
	stamps[1].TilePaths[0] = "consumer-mutated.dt1"
	if zone.Stamps()[1].TilePaths[0] != "a.dt1" {
		t.Fatal("stamp accessor leaked mutable storage")
	}
	if got := zone.Links()[0]; got != (Link{From: 1, To: 2}) {
		t.Fatalf("canonical link = %#v", got)
	}
}

func TestEquivalentZonesHaveSameChecksum(t *testing.T) {
	left, err := NewZone(validDefinition())
	if err != nil {
		t.Fatal(err)
	}
	rightDefinition := validDefinition()
	slices.Reverse(rightDefinition.Stamps)
	slices.Reverse(rightDefinition.Rooms)
	rightDefinition.Links[0] = Link{From: 1, To: 2}
	right, err := NewZone(rightDefinition)
	if err != nil {
		t.Fatal(err)
	}
	leftSum, _ := left.Checksum()
	rightSum, _ := right.Checksum()
	if leftSum != rightSum {
		t.Fatalf("equivalent zones differ: %s != %s", leftSum, rightSum)
	}
}

func TestZoneRejectsDanglingRoomStamp(t *testing.T) {
	definition := validDefinition()
	definition.Rooms[0].StampID = 99
	_, err := NewZone(definition)
	if !errors.Is(err, ErrZone) {
		t.Fatalf("error = %v", err)
	}
}

func TestZoneCopiesAndValidatesPathTiles(t *testing.T) {
	definition := validDefinition()
	definition.Paths = []PathTile{{X: 1, Y: 2}}
	zone, err := NewZone(definition)
	if err != nil {
		t.Fatal(err)
	}
	paths := zone.Paths()
	paths[0].X = 99
	if zone.Paths()[0].X != 1 {
		t.Fatal("path accessor leaked mutable state")
	}
	definition.Paths = []PathTile{{X: -1, Y: 0}}
	if _, err := NewZone(definition); err == nil {
		t.Fatal("accepted out-of-bounds path tile")
	}
}

func TestRequestRejectsImplicitOrUnknownInputs(t *testing.T) {
	for _, request := range []Request{
		{Act: 1, LevelID: 1},
		{Version: ContractVersion, Act: 0, LevelID: 1},
		{Version: ContractVersion, Act: 1, LevelID: 0},
		{Version: ContractVersion, Act: 1, LevelID: 1, Difficulty: 3},
	} {
		if !errors.Is(request.Validate(), ErrRequest) {
			t.Fatalf("request %#v was accepted", request)
		}
	}
}

func TestPurposeStreamsDoNotPerturbEachOther(t *testing.T) {
	streams := NewStreams(42)
	topology := streams.For("topology")
	want := topology.Uint64()
	_ = streams.For("preset-variant").Uint64()
	if got := streams.For("topology").Uint64(); got != want {
		t.Fatalf("topology stream was perturbed: %d != %d", got, want)
	}
}
