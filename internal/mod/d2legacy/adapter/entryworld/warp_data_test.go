package entryworld

import (
	"reflect"
	"testing"
)

// Changing the publication to an empty or one-sided list must break this contract.
func TestDenOfEvilPreparedWarpDataPublishesReciprocalStaticEndpoints(t *testing.T) {
	prepared := &Prepared{CaveWarp: &CaveWarpPair{
		First: CaveWarpEndpoint{ID: "static-warp:2:3", PairedID: "static-warp:8:1", LevelID: 2, TargetLevel: 8,
			X: 11, Y: 12, ArrivalX: 21, ArrivalY: 22, Width: 120, Height: 120},
		Second: CaveWarpEndpoint{ID: "static-warp:8:1", PairedID: "static-warp:2:3", LevelID: 8, TargetLevel: 2,
			X: 23, Y: 24, ArrivalX: 13, ArrivalY: 14, Width: 640, Height: 480},
	}}
	want := map[string]any{"endpoints": []any{
		map[string]any{"id": "static-warp:2:3", "pair_id": "static-warp:8:1", "static": true,
			"label": "DEN OF EVIL", "level_id": 2, "x": float64(11), "y": float64(12), "radius": float64(3.5),
			"destination_level": 8, "destination_x": float64(21), "destination_y": float64(22),
			"destination_width": float64(120), "destination_height": float64(120)},
		map[string]any{"id": "static-warp:8:1", "pair_id": "static-warp:2:3", "static": true,
			"label": "BLOOD MOOR", "level_id": 8, "x": float64(23), "y": float64(24), "radius": float64(3.5),
			"destination_level": 2, "destination_x": float64(13), "destination_y": float64(14),
			"destination_width": float64(640), "destination_height": float64(480)},
	}}
	if got := prepared.WarpData(); !reflect.DeepEqual(got, want) {
		t.Fatalf("warp publication = %#v, want %#v", got, want)
	}
}
