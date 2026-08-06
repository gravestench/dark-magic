package loot

import (
	"strings"
	"testing"

	"github.com/gravestench/dark-magic/internal/game/data/model"
)

type staticTreasureClasses []models.TreasureClassEx

func (records staticTreasureClasses) TreasureClassRecords() ([]models.TreasureClassEx, error) {
	return records, nil
}

func TestCatalogFromTypedTreasureClassRecords(t *testing.T) {
	catalog, err := CatalogFromRecords(staticTreasureClasses{{
		TreasureClass: "Root", Picks: 2, NoDrop: 3, Unique: 4, Item1: "gold", Prob1: 5,
	}})
	if err != nil {
		t.Fatal(err)
	}
	class := catalog["Root"]
	if class.Picks != 2 || class.NoDrop != 3 || class.Quality.Unique != 4 || len(class.Entries) != 1 || class.Entries[0] != (Entry{Code: "gold", Weight: 5}) {
		t.Fatalf("class = %#v", class)
	}
}

func TestCatalogFromTypedTreasureClassRecordsRejectsLossyNumbers(t *testing.T) {
	_, err := CatalogFromRecords(staticTreasureClasses{{TreasureClass: "Root", Picks: 1, Item1: "gold", Prob1: 1.5}})
	if err == nil || !strings.Contains(err.Error(), "non-integral weight") {
		t.Fatalf("error = %v", err)
	}
}
