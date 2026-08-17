package skillevidence

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/localization"
)

func TestOwnedTargetArchivesJoinLocalizedSkillEvidence(t *testing.T) {
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
	store := recordstore.New(assets)
	store.SetLogger(nil)
	skills, err := store.Load("data/global/excel/skills.txt")
	if err != nil {
		t.Fatal(err)
	}
	descriptions, err := store.Load("data/global/excel/SkillDesc.txt")
	if err != nil {
		t.Fatal(err)
	}
	report, err := Build([]int{0, 36, 40}, skills, descriptions, localization.New(assets, "English"))
	if err != nil {
		t.Fatal(err)
	}
	if report.Skills[0].Localization[0].Text != "Attack" || len(report.Skills[0].CrossSkillModifiers) != 0 {
		t.Fatalf("Attack evidence = %#v", report.Skills[0])
	}
	if modifiers := report.Skills[1].CrossSkillModifiers; len(modifiers) != 2 ||
		modifiers[0].ReferencedID != 47 || modifiers[1].ReferencedID != 56 {
		t.Fatalf("Fire Bolt modifiers = %#v", modifiers)
	}
	if modifiers := report.Skills[2].CrossSkillModifiers; len(modifiers) != 8 ||
		modifiers[0].ReferencedID != 50 || modifiers[1].ReferencedID != 60 {
		t.Fatalf("Frozen Armor modifiers = %#v", modifiers)
	}
}
