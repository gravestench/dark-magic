package clientapp

import (
	"reflect"
	"testing"
)

// TestRequestedOverlaysNormalizesCaptureFixtureList keeps empty and padded CLI values out of capture policy.
func TestRequestedOverlaysNormalizesCaptureFixtureList(t *testing.T) {
	want := []string{"inventory", "character"}
	if got := requestedOverlays(" inventory, ,character "); !reflect.DeepEqual(got, want) {
		t.Fatalf("requested overlays = %v, want %v", got, want)
	}
}
