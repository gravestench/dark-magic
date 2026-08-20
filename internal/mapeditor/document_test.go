package mapeditor

import (
	"bytes"
	"testing"

	"github.com/gravestench/ds1"
)

// TestDocumentPaintUndoRedoAndEncode exercises one complete authored edit and codec round trip.
func TestDocumentPaintUndoRedoAndEncode(t *testing.T) {
	document, err := New(NewConfig{Width: 3, Height: 2, Act: 2, Files: []string{"floor.dt1"}})
	if err != nil {
		t.Fatal(err)
	}
	brush := Brush{Identity: Identity{Style: 7, Sequence: 11}}
	if err := document.BeginStroke(LayerFloor, 0, brush); err != nil {
		t.Fatal(err)
	}
	if changed, err := document.Paint(Point{X: 1, Y: 1}, brush); err != nil || !changed {
		t.Fatalf("paint = %v, %v", changed, err)
	}
	if !document.EndStroke() {
		t.Fatal("stroke was not committed")
	}
	tile, ok := document.TileAt(Point{X: 1, Y: 1})
	if !ok || tile.Floors[0].Prop1 != 2 || tile.Floors[0].Style != 7 || tile.Floors[0].Sequence != 11 {
		t.Fatalf("painted floor = %#v", tile)
	}
	if !document.Undo() {
		t.Fatal("undo unavailable")
	}
	tile, _ = document.TileAt(Point{X: 1, Y: 1})
	if tile.Floors[0] != (ds1.FloorShadowRecord{}) {
		t.Fatalf("undo floor = %#v", tile.Floors[0])
	}
	if !document.Redo() {
		t.Fatal("redo unavailable")
	}
	encoded, err := document.Encode()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := ds1.FromBytes(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if got := decoded.Tiles[1][1].Floors[0]; got.Prop1 != 2 || got.Style != 7 || got.Sequence != 11 {
		t.Fatalf("encoded floor = %#v", got)
	}
}

// TestDocumentStrokeCoalescesRepeatedCells keeps a drag across the same cell as one undoable mutation.
func TestDocumentStrokeCoalescesRepeatedCells(t *testing.T) {
	document, err := New(NewConfig{Width: 2, Height: 2})
	if err != nil {
		t.Fatal(err)
	}
	first := Brush{Identity: Identity{Style: 1, Sequence: 2}}
	second := Brush{Identity: Identity{Style: 3, Sequence: 4}}
	if err := document.BeginStroke(LayerFloor, 0, first); err != nil {
		t.Fatal(err)
	}
	if _, err := document.Paint(Point{X: 0, Y: 0}, first); err != nil {
		t.Fatal(err)
	}
	if _, err := document.Paint(Point{X: 0, Y: 0}, second); err != nil {
		t.Fatal(err)
	}
	document.EndStroke()
	if !document.Undo() {
		t.Fatal("undo unavailable")
	}
	tile, _ := document.TileAt(Point{X: 0, Y: 0})
	if tile.Floors[0] != (ds1.FloorShadowRecord{}) {
		t.Fatalf("undo did not restore original value: %#v", tile.Floors[0])
	}
}

// TestDocumentReportsDirtyPointsForStrokeUndoAndRedo protects the renderer's local invalidation contract.
func TestDocumentReportsDirtyPointsForStrokeUndoAndRedo(t *testing.T) {
	document, err := New(NewConfig{Width: 3, Height: 2})
	if err != nil {
		t.Fatal(err)
	}
	brush := Brush{Identity: Identity{Style: 2, Sequence: 3}}
	if err := document.BeginStroke(LayerFloor, 0, brush); err != nil {
		t.Fatal(err)
	}
	for _, point := range []Point{{X: 0, Y: 0}, {X: 2, Y: 1}} {
		if _, err := document.Paint(point, brush); err != nil {
			t.Fatal(err)
		}
	}
	want := []Point{{X: 0, Y: 0}, {X: 2, Y: 1}}
	if got := document.EndStrokePoints(); !pointsEqual(got, want) {
		t.Fatalf("end dirty points = %#v, want %#v", got, want)
	}
	if got := document.UndoPoints(); !pointsEqual(got, want) {
		t.Fatalf("undo dirty points = %#v, want %#v", got, want)
	}
	if got := document.RedoPoints(); !pointsEqual(got, want) {
		t.Fatalf("redo dirty points = %#v, want %#v", got, want)
	}
}

// pointsEqual compares the ordered dirty-cell lists used by history assertions.
func pointsEqual(left, right []Point) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

// TestOpenPreservesUntouchedDocumentBytes ensures opening and saving an unedited map is lossless.
func TestOpenPreservesUntouchedDocumentBytes(t *testing.T) {
	original, err := New(NewConfig{Width: 1, Height: 1})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := original.Encode()
	if err != nil {
		t.Fatal(err)
	}
	document, err := Open(encoded)
	if err != nil {
		t.Fatal(err)
	}
	reencoded, err := document.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, reencoded) {
		t.Fatalf("open/encode changed bytes\n got %x\nwant %x", reencoded, encoded)
	}
	if document.Summary().Dirty {
		t.Fatal("opened document is dirty")
	}
}

// TestNewRejectsDimensionsTheDS1CodecCannotEncode rejects dangerous allocations before codec invocation.
func TestNewRejectsDimensionsTheDS1CodecCannotEncode(t *testing.T) {
	_, err := New(NewConfig{Width: 4096, Height: 4096})
	if err == nil {
		t.Fatal("New accepted an unencodable map size")
	}
}
