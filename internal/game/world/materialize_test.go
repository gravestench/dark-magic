package world

import (
	"context"
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/game/mapgen"
)

func materializerFixture(t *testing.T) *Materializer {
	t.Helper()
	zone, err := mapgen.NewZone(mapgen.Definition{
		Request: mapgen.Request{Version: mapgen.ContractVersion, Seed: 1, Act: 1, LevelID: 9}, Kind: mapgen.Maze,
		Bounds: mapgen.Bounds{Width: 4, Height: 2},
		Stamps: []mapgen.Stamp{{ID: 1, Role: "previous-level", Width: 2, Height: 2, DS1Path: "one.ds1"}, {ID: 2, Role: "next-level", X: 2, Width: 2, Height: 2, DS1Path: "two.ds1"}},
		Rooms:  []mapgen.Room{{ID: 1, Width: 2, Height: 2, StampID: 1}, {ID: 2, X: 2, Width: 2, Height: 2, StampID: 2}},
		Links:  []mapgen.Link{{From: 1, To: 2}},
	})
	if err != nil {
		t.Fatal(err)
	}
	materializer, err := NewMaterializer(fstest.MapFS{}, zone)
	if err != nil {
		t.Fatal(err)
	}
	materializer.load = func(_ fs.FS, name string, _ []string, _ ...ObjectResolver) (*Map, error) {
		result := &Map{WidthTiles: 2, HeightTiles: 2, WidthSubtiles: 10, HeightSubtiles: 10, Act: 1, flags: make([]Flags, 100)}
		result.Tiles = []TilePlacement{{X: 0, Y: 1}}
		result.Objects = []Object{{X: 2, Y: 3}}
		if name == "two.ds1" {
			result.flags[3*10+4].BlockWalk = true
		}
		return result, nil
	}
	return materializer
}

func TestMaterializerClipsDS1SharedTerminalEdge(t *testing.T) {
	materializer := materializerFixture(t)
	materializer.load = func(_ fs.FS, _ string, _ []string, _ ...ObjectResolver) (*Map, error) {
		result := &Map{WidthTiles: 3, HeightTiles: 3, WidthSubtiles: 15, HeightSubtiles: 15, flags: make([]Flags, 225)}
		result.Tiles = []TilePlacement{{X: 1, Y: 1}, {X: 2, Y: 1}, {X: 1, Y: 2}}
		result.flags[14*15+14].BlockWalk = true
		return result, nil
	}
	if err := materializer.Step(t.Context()); err != nil {
		t.Fatal(err)
	}
	if len(materializer.assembled.Tiles) != 1 {
		t.Fatalf("terminal tiles were retained: %#v", materializer.assembled.Tiles)
	}
	if flags, _ := materializer.assembled.FlagsAt(9, 9); flags.BlockWalk {
		t.Fatal("terminal collision leaked into room footprint")
	}
}

func TestMaterializerPublishesOnlyCompleteShiftedMap(t *testing.T) {
	materializer := materializerFixture(t)
	if _, err := materializer.Result(); !errors.Is(err, ErrMaterializationIncomplete) {
		t.Fatalf("early result error = %v", err)
	}
	if got := materializer.Progress(); got.Completed != 0 || got.Total != 2 || got.CurrentRole != "previous-level" {
		t.Fatalf("initial progress = %#v", got)
	}
	if err := materializer.Step(t.Context()); err != nil {
		t.Fatal(err)
	}
	if _, err := materializer.Result(); !errors.Is(err, ErrMaterializationIncomplete) {
		t.Fatalf("partial result error = %v", err)
	}
	if err := materializer.Step(t.Context()); err != nil {
		t.Fatal(err)
	}
	result, err := materializer.Result()
	if err != nil {
		t.Fatal(err)
	}
	if result.WidthTiles != 4 || len(result.Tiles) != 2 || result.Tiles[1].X != 2 || result.Objects[1].X != 12 {
		t.Fatalf("materialized map = %#v", result)
	}
	flags, ok := result.FlagsAt(14, 3)
	if !ok || !flags.BlockWalk {
		t.Fatalf("shifted collision = %#v, %v", flags, ok)
	}
}

func TestMaterializerHonorsCancellationWithoutAdvancing(t *testing.T) {
	materializer := materializerFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := materializer.Step(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v", err)
	}
	if materializer.Progress().Completed != 0 {
		t.Fatal("canceled step advanced progress")
	}
}

func TestMaterializerOverlayReplacesMatchingTileLayer(t *testing.T) {
	zone, err := mapgen.NewZone(mapgen.Definition{
		Request: mapgen.Request{Version: mapgen.ContractVersion, Seed: 1, Act: 1, LevelID: 2}, Kind: mapgen.Outdoor,
		Bounds: mapgen.Bounds{Width: 1, Height: 1},
		Stamps: []mapgen.Stamp{
			{ID: 1, Width: 1, Height: 1, DS1Path: "fill.ds1"},
			{ID: 2, Width: 1, Height: 1, DS1Path: "river.ds1", Overlay: true},
		},
		Rooms: []mapgen.Room{{ID: 1, Width: 1, Height: 1, StampID: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	materializer, err := NewMaterializer(fstest.MapFS{}, zone)
	if err != nil {
		t.Fatal(err)
	}
	materializer.load = func(_ fs.FS, name string, _ []string, _ ...ObjectResolver) (*Map, error) {
		identity := TileIdentity{MainIndex: 1}
		if name == "river.ds1" {
			identity.MainIndex = 2
		}
		return &Map{WidthTiles: 1, HeightTiles: 1, WidthSubtiles: 5, HeightSubtiles: 5,
			flags: make([]Flags, 25), Tiles: []TilePlacement{{Layer: LayerFloor, Identity: identity}}}, nil
	}
	if err := materializer.Step(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := materializer.Step(t.Context()); err != nil {
		t.Fatal(err)
	}
	result, err := materializer.Result()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Tiles) != 1 || result.Tiles[0].Identity.MainIndex != 2 {
		t.Fatalf("overlay tiles = %#v", result.Tiles)
	}
}
