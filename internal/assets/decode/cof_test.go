package assetdecode

import (
	"testing"
	"testing/fstest"

	cof "github.com/gravestench/cof"
)

// TestCOFReadsCompositionMetadata verifies that layer, frame, and priority
// metadata retain their authored nesting through the filesystem boundary.
func TestCOFReadsCompositionMetadata(t *testing.T) {
	input := cof.New()
	input.NumberOfDirections = 1
	input.FramesPerDirection = 1
	input.NumberOfLayers = 1
	input.Speed = 128
	input.CofLayers = []cof.CofLayer{{
		Type:        0,
		Selectable:  true,
		WeaponClass: cof.WeaponClassFromString("hth"),
	}}
	input.AnimationFrames = []cof.FrameEvent{1}
	input.Priority = [][][]cof.CompositeType{{{0}}}

	decoded, err := COF(fstest.MapFS{
		"unit.cof": &fstest.MapFile{Data: cof.Marshal(input)},
	}, "unit.cof")
	if err != nil {
		t.Fatal(err)
	}

	if decoded.NumberOfLayers != 1 ||
		decoded.FramesPerDirection != 1 ||
		decoded.Priority[0][0][0] != 0 {
		t.Fatalf("COF = %#v", decoded)
	}
}
