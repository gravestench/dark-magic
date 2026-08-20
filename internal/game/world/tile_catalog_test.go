package world

import "testing"

// TestTileCatalogGroupsCandidatesAndReturnsCopies protects authored order and catalog ownership from caller mutation.
func TestTileCatalogGroupsCandidatesAndReturnsCopies(t *testing.T) {
	identity := TileIdentity{Orientation: 1, MainIndex: 2, SubIndex: 3}
	catalog := NewTileCatalog([]TileReference{{Identity: identity, Index: 4}, {Identity: identity, Index: 5}})

	candidates := catalog.Candidates(identity)
	if len(candidates) != 2 || candidates[0].Index != 4 || candidates[1].Index != 5 {
		t.Fatalf("candidates = %#v", candidates)
	}

	candidates[0].Index = 99
	if got := catalog.Candidates(identity)[0].Index; got != 4 {
		t.Fatalf("mutated catalog through candidate copy: %d", got)
	}
}

// TestTileCatalogGroupsPhysicalRecordsBySourcePath verifies the editor can display each declared DT1 independently.
func TestTileCatalogGroupsPhysicalRecordsBySourcePath(t *testing.T) {
	first := TileIdentity{MainIndex: 1}
	second := TileIdentity{MainIndex: 2}
	catalog := NewTileCatalog([]TileReference{
		{Identity: first, Path: "floor.dt1", Index: 0},
		{Identity: second, Path: "wall.dt1", Index: 1},
		{Identity: second, Path: "floor.dt1", Index: 2},
	})
	references := catalog.References("floor.dt1")
	if len(references) != 2 || references[0].Index != 0 || references[1].Index != 2 {
		t.Fatalf("path references = %#v", references)
	}
	references[0].Index = 99
	if got := catalog.References("floor.dt1")[0].Index; got != 0 {
		t.Fatalf("mutated catalog through path copy: %d", got)
	}
}

// TestTileCatalogIdentitiesAreSorted keeps logical tile pickers deterministic across catalog construction order.
func TestTileCatalogIdentitiesAreSorted(t *testing.T) {
	catalog := NewTileCatalog([]TileReference{
		{Identity: TileIdentity{Orientation: 13, MainIndex: 2, SubIndex: 1}},
		{Identity: TileIdentity{Orientation: 1, MainIndex: 8, SubIndex: 2}},
		{Identity: TileIdentity{Orientation: 1, MainIndex: 8, SubIndex: 1}},
	})
	got := catalog.Identities()
	want := []TileIdentity{
		{Orientation: 1, MainIndex: 8, SubIndex: 1},
		{Orientation: 1, MainIndex: 8, SubIndex: 2},
		{Orientation: 13, MainIndex: 2, SubIndex: 1},
	}
	if len(got) != len(want) {
		t.Fatalf("identities = %#v", got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("identity %d = %#v, want %#v", index, got[index], want[index])
		}
	}
}

// TestTileCatalogSelectionIsDeterministicAndSkipsZeroRarity pins weighted selection for repeatable map materialization.
func TestTileCatalogSelectionIsDeterministicAndSkipsZeroRarity(t *testing.T) {
	identity := TileIdentity{Orientation: 1, MainIndex: 2, SubIndex: 3}
	catalog := NewTileCatalog([]TileReference{
		{Identity: identity, Index: 0, Rarity: 0},
		{Identity: identity, Index: 1, Rarity: 1},
		{Identity: identity, Index: 2, Rarity: 10},
	})

	want, ok := catalog.Select(identity, 7, 11, 42)
	if !ok || want.Index == 0 {
		t.Fatalf("selection = %#v, %v", want, ok)
	}

	for range 10 {
		got, found := catalog.Select(identity, 7, 11, 42)
		if !found || got.Index != want.Index {
			t.Fatalf("selection changed: %#v, %v; want %#v", got, found, want)
		}
	}
}

// TestTileCatalogAllZeroRarityUsesFirstAuthoredRecord preserves compatibility with unweighted legacy DT1 groups.
func TestTileCatalogAllZeroRarityUsesFirstAuthoredRecord(t *testing.T) {
	identity := TileIdentity{Orientation: 13, MainIndex: 8, SubIndex: 5}
	catalog := NewTileCatalog([]TileReference{
		{Identity: identity, Index: 8},
		{Identity: identity, Index: 9},
	})

	got, ok := catalog.Select(identity, 0, 0, 0)
	if !ok || got.Index != 8 {
		t.Fatalf("selection = %#v, %v", got, ok)
	}
}

// TestTileRandomDistinguishesZeroAxes guards against coordinate mixing that collapses either world axis at zero.
func TestTileRandomDistinguishesZeroAxes(t *testing.T) {
	identity := TileIdentity{Orientation: 0, MainIndex: 1, SubIndex: 2}

	values := map[uint64]bool{}
	for coordinate := 0; coordinate < 8; coordinate++ {
		values[tileRandom(0, identity, coordinate, 0)] = true
		values[tileRandom(0, identity, 0, coordinate)] = true
	}

	if len(values) < 8 {
		t.Fatalf("zero-axis coordinates collapsed to %d random values", len(values))
	}
}

// TestTileCatalogMissingIdentityDoesNotSelect ensures loaders skip absent physical tiles instead of inventing defaults.
func TestTileCatalogMissingIdentityDoesNotSelect(t *testing.T) {
	catalog := NewTileCatalog(nil)
	if _, ok := catalog.Select(TileIdentity{Orientation: 99}, 0, 0, 0); ok {
		t.Fatal("missing identity unexpectedly selected")
	}
}
