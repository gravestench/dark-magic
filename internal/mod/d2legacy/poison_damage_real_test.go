package d2legacy

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
)

// TestOwnedPoisonDamageRecords pins the stat and presentation identities used
// by the poison path to the configured, privately owned 1.14d archives.
func TestOwnedPoisonDamageRecords(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to the expansion 1.14d MPQ directory")
	}
	t.Setenv("MPQ_DIRECTORY", directory)
	assets, err := content.FromEnvironment(content.Layer{Name: "d2legacy", FS: content.D2Legacy()})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = assets.Close() }()
	store := recordstore.New(assets)
	store.SetLogger(nil)
	stats, err := store.Load("data/global/excel/ItemStatCost.txt")
	if err != nil {
		t.Fatal(err)
	}
	for stat, id := range map[string]string{
		"poisonmindam": "57",
		"poisonmaxdam": "58",
		"poisonlength": "59",
		"hpregen":      "74",
	} {
		row := rowBy(stats, "Stat", stat)
		if row == nil || row["ID"] != id {
			t.Errorf("owned ItemStatCost %s = %#v, want ID %s", stat, row, id)
		}
	}
	states, err := store.Load("data/global/excel/States.txt")
	if err != nil {
		t.Fatal(err)
	}
	poison := rowBy(states, "state", "poison")
	if poison == nil || poison["id"] != "2" || poison["green"] != "1" || poison["castoverlay"] != "poisonhit" {
		t.Fatalf("owned poison state = %#v, want id=2 green=1 castoverlay=poisonhit", poison)
	}
}
