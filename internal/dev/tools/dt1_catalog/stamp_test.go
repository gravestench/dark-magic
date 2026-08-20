package main

import (
	"bytes"
	"testing"

	ds1 "github.com/gravestench/ds1/pkg"
)

// TestWriteStampTileIdentities preserves coordinate and layer ordering while filtering unusable records.
func TestWriteStampTileIdentities(t *testing.T) {
	tiles := [][]ds1.TileRecord{
		{
			{
				Floors: []ds1.FloorShadowRecord{
					{Prop1: 1, Style: 2, Sequence: 3},
					{Prop1: 1, Style: 4, Sequence: 5, Hidden: true},
					{Style: 6, Sequence: 7},
				},
				Walls: []ds1.WallRecord{
					{Prop1: 1, Type: 8, Style: 9, Sequence: 10},
					{Prop1: 1, Type: 11, Style: 12, Sequence: 13, Hidden: true},
				},
			},
			{
				Floors: []ds1.FloorShadowRecord{{Prop1: 1, Style: 14, Sequence: 15}},
			},
		},
	}

	var output bytes.Buffer
	writeStampTileIdentities(&output, tiles)

	want := "" +
		"    0,  0 floor[0] main=  2 sub=  3\n" +
		"    0,  0 wall[0] orientation= 8 main=  9 sub= 10\n" +
		"    1,  0 floor[0] main= 14 sub= 15\n"
	if got := output.String(); got != want {
		t.Fatalf("writeStampTileIdentities() output = %q, want %q", got, want)
	}
}
