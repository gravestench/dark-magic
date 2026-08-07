package world

import (
	"testing"

	"github.com/gravestench/dt1"
)

func TestApplyCombinesSubtileFlags(t *testing.T) {
	m := &Map{WidthSubtiles: 10, HeightSubtiles: 5, flags: make([]Flags, 50)}
	first := &dt1.Tile{}
	first.SubTileFlags[0].BlockLOS = true
	first.SubTileFlags[24].BlockWalk = true
	second := &dt1.Tile{}
	second.SubTileFlags[0].BlockPlayerWalk = true

	m.apply(1, 0, first)
	m.apply(1, 0, second)

	got, ok := m.FlagsAt(5, 0)
	if !ok || !got.BlockLOS || !got.BlockPlayerWalk || !got.Blocked() {
		t.Fatalf("FlagsAt(5, 0) = %#v, %v", got, ok)
	}
	got, ok = m.FlagsAt(9, 4)
	if !ok || !got.BlockWalk || !got.Blocked() {
		t.Fatalf("FlagsAt(9, 4) = %#v, %v", got, ok)
	}
}

func TestFlagsAtRejectsOutsideMap(t *testing.T) {
	m := &Map{WidthSubtiles: 5, HeightSubtiles: 5, flags: make([]Flags, 25)}
	for _, point := range [][2]int{{-1, 0}, {0, -1}, {5, 0}, {0, 5}} {
		if _, ok := m.FlagsAt(point[0], point[1]); ok {
			t.Fatalf("FlagsAt(%d, %d) unexpectedly succeeded", point[0], point[1])
		}
	}
}

func TestChooseIsDeterministicAndHandlesZeroWeight(t *testing.T) {
	first := &dt1.Tile{RarityFrameIndex: 0}
	second := &dt1.Tile{RarityFrameIndex: 0}
	if got := choose([]*dt1.Tile{first, second}, 7, 11); got != first {
		t.Fatalf("zero-weight choose returned %p, want first tile %p", got, first)
	}
	first.RarityFrameIndex, second.RarityFrameIndex = 1, 10
	want := choose([]*dt1.Tile{first, second}, 7, 11)
	for range 10 {
		if got := choose([]*dt1.Tile{first, second}, 7, 11); got != want {
			t.Fatalf("choose was not deterministic: got %p, want %p", got, want)
		}
	}
}

func TestApplyIgnoresTileOutsideMap(t *testing.T) {
	m := &Map{WidthSubtiles: 5, HeightSubtiles: 5, flags: make([]Flags, 25)}
	tile := &dt1.Tile{}
	tile.SubTileFlags[0].BlockWalk = true
	m.apply(-1, 0, tile)
	m.apply(1, 0, tile)
}
