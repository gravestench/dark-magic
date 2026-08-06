package audio

import (
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/game/data/store"
)

func TestCatalogResolvesDeterministicSoundRecord(t *testing.T) {
	t.Parallel()
	source := fstest.MapFS{
		"data/global/excel/Sounds.txt":     &fstest.MapFile{Data: []byte("Sound\tRedirect\tFileName\tIsLocal\tIsMusic\tIsAmbientScene\tIsAmbientEvent\tIsUI\tVolume Min\tVolume Max\tGroup Size\tGroup Weight\tLoop\tStream\nmenu\t\tui/one.wav\t0\t0\t0\t0\t1\t128\t255\t2\t1\t0\t0\nmenu_alt\t\tui/two.wav\t0\t0\t0\t0\t1\t128\t255\t0\t3\t0\t0\ntown\t\tmusic/town.wav\t0\t1\t0\t0\t0\t255\t255\t0\t0\t1\t1\n")},
		"data/global/sfx/ui/one.wav":       &fstest.MapFile{Data: []byte("one")},
		"data/global/sfx/ui/two.wav":       &fstest.MapFile{Data: []byte("two")},
		"data/global/music/music/town.wav": &fstest.MapFile{Data: []byte("town")},
	}
	records := recordstore.New(source)
	catalog := NewCatalog(source, records)
	menu, err := catalog.Resolve("menu", 2)
	if err != nil {
		t.Fatal(err)
	}
	if menu.Name != "menu_alt" || menu.Options.Bus != "ui" || string(menu.Data) != "two" {
		t.Fatalf("menu = %#v", menu)
	}
	music, err := catalog.Resolve("town", 0)
	if err != nil {
		t.Fatal(err)
	}
	if !music.Options.Stream || !music.Options.Loop || music.Options.Bus != "music" {
		t.Fatalf("music = %#v", music)
	}
	source["data/global/excel/Sounds.txt"].Data = []byte("Sound\tFileName\tIsUI\tVolume Min\tVolume Max\nmenu\tui/one.wav\t1\t255\t255\n")
	records.Invalidate(soundsTable)
	reloaded, err := catalog.Resolve("menu", 0)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Name != "menu" || string(reloaded.Data) != "one" {
		t.Fatalf("reloaded menu = %#v", reloaded)
	}
}
