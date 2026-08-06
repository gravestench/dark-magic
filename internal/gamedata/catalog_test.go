package gamedata

import (
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/recordstore"
)

func TestCatalogBuildsClonedTypedSnapshotAndIndexes(t *testing.T) {
	t.Parallel()

	source := fstest.MapFS{
		CharStatsTable: &fstest.MapFile{Data: []byte("class\tstr\nAmazon\t20\n")},
		SoundsTable:    &fstest.MapFile{Data: []byte("Sound\tFileName\tLoop\nmenu_music\tmusic.wav\t1\n")},
	}
	catalog := New(recordstore.New(source))
	first, err := catalog.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if first.CharStatsByClass["Amazon"].Strength != 20 || first.SoundsByName["menu_music"].Loop != 1 {
		t.Fatalf("typed snapshot = %#v", first)
	}
	delete(first.CharStatsByClass, "Amazon")
	first.CharStats[0].Strength = 1
	second, err := catalog.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if second.CharStatsByClass["Amazon"].Strength != 20 || second.CharStats[0].Strength != 20 {
		t.Fatal("catalog exposed mutable snapshot ownership")
	}
}

func TestCatalogInvalidationAtomicallyRebuildsTypedData(t *testing.T) {
	t.Parallel()

	source := fstest.MapFS{
		CharStatsTable: &fstest.MapFile{Data: []byte("class\tstr\nAmazon\t20\n")},
		SoundsTable:    &fstest.MapFile{Data: []byte("Sound\tFileName\nmenu_music\tmusic.wav\n")},
	}
	catalog := New(recordstore.New(source))
	if _, err := catalog.Snapshot(); err != nil {
		t.Fatal(err)
	}
	source[CharStatsTable].Data = []byte("class\tstr\nAmazon\t25\n")
	catalog.Invalidate(CharStatsTable)
	updated, err := catalog.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if updated.CharStatsByClass["Amazon"].Strength != 25 {
		t.Fatalf("reloaded strength = %d, want 25", updated.CharStatsByClass["Amazon"].Strength)
	}
}
