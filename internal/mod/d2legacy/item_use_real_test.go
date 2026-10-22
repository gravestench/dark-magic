package d2legacy

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
)

// TestOwnedTownPortalScrollSpellBinding grounds the supported item code in the owned spell dispatch rows.
func TestOwnedTownPortalScrollSpellBinding(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to the expansion 1.14d MPQ directory")
	}

	t.Setenv("MPQ_DIRECTORY", directory)

	assets, err := content.FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = assets.Close() }()

	pinned, generation, err := recordstore.Pin(assets)
	if err != nil {
		t.Fatal(err)
	}

	misc, err := pinned.Load("data/global/excel/misc.txt")
	if err != nil {
		t.Fatal(err)
	}

	books, err := pinned.Load("data/global/excel/books.txt")
	if err != nil {
		t.Fatal(err)
	}

	scroll := rowBy(misc, "code", "tsc")

	book := rowBy(books, "ScrollSpellCode", "tsc")
	if scroll == nil || scroll["type"] != "scro" || scroll["useable"] != "1" || scroll["pSpell"] != "2" ||
		book == nil || book["pSpell"] != "2" {
		t.Fatal("Town Portal scroll is not a usable scroll bound to item-spell handler 2")
	}

	t.Logf("owned generation %s: Misc tsc is usable scro; Misc/Books bind tsc to pSpell=2", generation.ID)
}
