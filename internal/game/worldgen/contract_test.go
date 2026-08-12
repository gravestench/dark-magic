package worldgen

import (
	"errors"
	"slices"
	"testing"
)

// validDefinition intentionally uses vocabulary the engine does not know. A
// reusable host must validate shape and references without restricting a mod to
// Diablo's acts, difficulties, generation families, or structure names.
func validDefinition() Definition {
	return Definition{
		Request: Request{Version: ContractVersion, Seed: 42, Act: 9, LevelID: 1, Difficulty: 7},
		Kind:    "custom-mod-layout",
		Bounds:  Bounds{Width: 16, Height: 16},
		Stamps: []Stamp{
			{ID: 2, Width: 8, Height: 8, DS1Path: "second.stamp", TilePaths: []string{"b.tiles", "a.tiles"}},
			{ID: 1, Width: 8, Height: 8, DS1Path: "first.stamp"},
		},
		Rooms: []Room{
			{ID: 2, X: 8, Width: 8, Height: 8, StampID: 2},
			{ID: 1, Width: 8, Height: 8, StampID: 1},
		},
		Links: []Link{{From: 2, To: 1}},
		Trace: []string{"mod policy selected variant 1"},
	}
}

func TestZoneCanonicalizesAndCopiesDefinition(t *testing.T) {
	definition := validDefinition()
	zone, err := NewZone(definition)
	if err != nil {
		t.Fatal(err)
	}
	definition.Stamps[0].DS1Path = "mutated.stamp"
	definition.Stamps[0].TilePaths[0] = "mutated.tiles"
	stamps := zone.Stamps()
	if stamps[0].ID != 1 || stamps[1].DS1Path != "second.stamp" || !slices.Equal(stamps[1].TilePaths, []string{"a.tiles", "b.tiles"}) {
		t.Fatalf("canonical stamps = %#v", stamps)
	}
	stamps[1].TilePaths[0] = "consumer-mutated.tiles"
	if zone.Stamps()[1].TilePaths[0] != "a.tiles" {
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

func TestZoneAcceptsOpaqueModStructureKinds(t *testing.T) {
	definition := validDefinition()
	definition.Structures = []StructureTile{
		{X: 1, Y: 2, Kind: "force-field", Passable: true},
		{X: 1, Y: 3, Kind: "lava", Passable: false},
	}
	zone, err := NewZone(definition)
	if err != nil {
		t.Fatal(err)
	}
	definition.Structures[0].Kind = "mutated"
	if got := zone.Structures()[0].Kind; got != "force-field" {
		t.Fatalf("immutable structure kind = %q", got)
	}
}

func TestRequestValidatesOnlyGenericReplayIdentity(t *testing.T) {
	for _, request := range []Request{
		{Act: 1, LevelID: 1},
		{Version: ContractVersion, Act: 1, LevelID: 0},
	} {
		if !errors.Is(request.Validate(), ErrRequest) {
			t.Fatalf("request %#v was accepted", request)
		}
	}
	if err := (Request{Version: ContractVersion, Act: 0, LevelID: 1, Difficulty: 255}).Validate(); err != nil {
		t.Fatalf("engine rejected opaque mod dimensions: %v", err)
	}
}
