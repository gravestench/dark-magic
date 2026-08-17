package d2legacy

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
)

func rowBy(rows []map[string]string, column, value string) map[string]string {
	for _, row := range rows {
		if row[column] == value {
			return row
		}
	}
	return nil
}

func TestOwnedTargetArchivesPinAttackRateData(t *testing.T) {
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
	pinned, generation, err := recordstore.Pin(assets)
	if err != nil {
		t.Fatal(err)
	}

	stats, err := pinned.Load("data/global/excel/ItemStatCost.txt")
	if err != nil {
		t.Fatal(err)
	}
	attackrate := rowBy(stats, "Stat", "attackrate")
	itemRate := rowBy(stats, "Stat", "item_fasterattackrate")
	if generation.ID == "" || attackrate == nil || attackrate["ID"] != "68" || attackrate["Signed"] != "1" ||
		attackrate["UpdateAnimRate"] != "1" || itemRate == nil || itemRate["ID"] != "93" || itemRate["Signed"] != "1" {
		t.Fatalf("generation=%q attackrate=%#v item_fasterattackrate=%#v", generation.ID, attackrate, itemRate)
	}

	weapons, err := pinned.Load("data/global/excel/weapons.txt")
	if err != nil {
		t.Fatal(err)
	}
	phaseBlade, warPike := rowBy(weapons, "code", "7cr"), rowBy(weapons, "code", "7p7")
	if phaseBlade == nil || phaseBlade["version"] != "100" || phaseBlade["speed"] != "-30" ||
		warPike == nil || warPike["version"] != "100" || warPike["speed"] != "20" {
		t.Fatalf("phase blade=%#v war pike=%#v", phaseBlade, warPike)
	}

	properties, err := pinned.Load("data/global/excel/Properties.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, code := range []string{"swing1", "swing2", "swing3"} {
		row := rowBy(properties, "code", code)
		if row == nil || row["func1"] != "8" || row["stat1"] != "item_fasterattackrate" {
			t.Fatalf("property %s = %#v", code, row)
		}
	}
}
