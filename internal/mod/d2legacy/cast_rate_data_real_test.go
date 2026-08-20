package d2legacy

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/localization"
)

// TestOwnedTargetArchivesPinFasterCastRateEvidence anchors cast breakpoints to
// real character and skill records rather than synthetic assumptions.
func TestOwnedTargetArchivesPinFasterCastRateEvidence(t *testing.T) {
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

	stats, err := pinned.Load("data/global/excel/ItemStatCost.txt")
	if err != nil {
		t.Fatal(err)
	}

	fasterCast := rowBy(stats, "Stat", "item_fastercastrate")
	if generation.ID == "" || fasterCast == nil || fasterCast["ID"] != "105" || fasterCast["Signed"] != "1" ||
		fasterCast["descstrpos"] != "ModStr4v" || fasterCast["descstrneg"] != "ModStr4v" {
		t.Fatalf("generation=%q item_fastercastrate=%#v", generation.ID, fasterCast)
	}

	properties, err := pinned.Load("data/global/excel/Properties.txt")
	if err != nil {
		t.Fatal(err)
	}

	for _, code := range []string{"cast1", "cast2", "cast3"} {
		row := rowBy(properties, "code", code)
		if row == nil || row["func1"] != "8" || row["stat1"] != "item_fastercastrate" {
			t.Fatalf("property %s = %#v", code, row)
		}
	}

	text, source, err := localization.New(assets, "English").Resolve("ModStr4v")
	if err != nil {
		t.Fatal(err)
	}

	if text != "Faster Cast Rate" || source != "data/local/lng/eng/string.tbl" {
		t.Fatalf("ModStr4v = %q from %q", text, source)
	}

	skills, err := pinned.Load("data/global/excel/Skills.txt")
	if err != nil {
		t.Fatal(err)
	}

	fireBolt, lightning := rowBy(skills, "Id", "36"), rowBy(skills, "Id", "49")
	if fireBolt == nil || fireBolt["skill"] != "Fire Bolt" || fireBolt["anim"] != "SC" || fireBolt["seqtrans"] != "SC" ||
		lightning == nil || lightning["skill"] != "Lightning" || lightning["anim"] != "SQ" ||
		lightning["seqtrans"] != "SC" || lightning["seqnum"] != "12" || lightning["UseAttackRate"] != "1" {
		t.Fatalf("Fire Bolt=%#v Lightning=%#v", fireBolt, lightning)
	}
}
