package movement

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
)

func TestOwnedExpansion114dClassMovementRates(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to the expansion 1.14d MPQ directory")
	}
	t.Setenv("MPQ_DIRECTORY", directory)
	assets, err := content.FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	defer assets.Close()
	pinned, _, err := recordstore.Pin(assets)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCatalog(pinned)
	if err != nil {
		t.Fatal(err)
	}
	for _, class := range []string{"Amazon", "Sorceress", "Necromancer", "Paladin", "Barbarian", "Druid", "Assassin"} {
		rates, found := catalog.Rates(class)
		if !found || rates != (ClassRates{Walk: 6, Run: 9}) {
			t.Fatalf("owned expansion 1.14d %s movement rates = %+v, %v", class, rates, found)
		}
	}
}
