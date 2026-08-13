package player

import "testing"

func TestDiffWorldViewIsDeterministicAndIdempotent(t *testing.T) {
	health, changed := int64(8), int64(7)
	previous := WorldView{Version: 1, Tick: 10, Entities: []WorldEntity{{ID: "b", Kind: "monster", Health: &health}, {ID: "a", Kind: "object"}}}
	current := WorldView{Version: 1, Tick: 11, Entities: []WorldEntity{{ID: "b", Kind: "monster", Health: &changed}, {ID: "c", Kind: "object"}}}
	delta := DiffWorldView(previous, current)
	if delta.Reset || len(delta.Upserts) != 2 || delta.Upserts[0].ID != "b" || delta.Upserts[1].ID != "c" || len(delta.Removed) != 1 || delta.Removed[0] != "a" {
		t.Fatalf("delta = %#v", delta)
	}
	unchanged := DiffWorldView(current, current)
	if len(unchanged.Upserts) != 0 || len(unchanged.Removed) != 0 {
		t.Fatalf("idempotent delta = %#v", unchanged)
	}
}

func TestDiffWorldViewResetsFromTruncatedProjection(t *testing.T) {
	delta := DiffWorldView(WorldView{Tick: 1, Truncated: true}, WorldView{Tick: 2, Entities: []WorldEntity{{ID: "a", Kind: "object"}}})
	if !delta.Reset || len(delta.Upserts) != 1 || len(delta.Removed) != 0 {
		t.Fatalf("reset delta = %#v", delta)
	}
}
