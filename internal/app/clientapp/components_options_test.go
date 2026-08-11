package clientapp

import (
	"reflect"
	"testing"
)

func TestRequestedOverlaysNormalizesCaptureFixtureList(t *testing.T) {
	want := []string{"inventory", "character"}
	if got := requestedOverlays(" inventory, ,character "); !reflect.DeepEqual(got, want) {
		t.Fatalf("requested overlays = %v, want %v", got, want)
	}
}
