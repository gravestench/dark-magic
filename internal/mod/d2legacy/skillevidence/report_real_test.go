package skillevidence

import (
	"os"
	"strings"
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
	report, err := Build([]int{0, 36, 40, 54, 55}, skills, descriptions, localization.New(assets, "English"))
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
	teleport := report.Skills[3]
	if len(teleport.CrossSkillModifiers) != 0 {
		t.Fatalf("Teleport modifiers = %#v", teleport.CrossSkillModifiers)
	}
	teleportLocalized := map[string]LocalizationReference{}
	for _, evidence := range teleport.Localization {
		teleportLocalized[evidence.Column] = evidence
	}
	if teleportLocalized["str name"].Text != "Teleport" ||
		teleportLocalized["str long"].Key != "skillld54" ||
		teleportLocalized["str long"].Text != "instantly moves to a destination within your line of sight" ||
		teleportLocalized["str long"].Source != "data/local/lng/eng/string.tbl" {
		t.Fatalf("Teleport localization = %#v", teleport.Localization)
	}
	glacial := report.Skills[4]
	if modifiers := glacial.CrossSkillModifiers; len(modifiers) != 5 ||
		modifiers[0].ReferencedID != 39 || modifiers[1].ReferencedID != 45 ||
		modifiers[2].ReferencedID != 64 || modifiers[3].ReferencedID != 59 ||
		modifiers[4].ReferencedID != 59 {
		t.Fatalf("Glacial Spike modifiers = %#v", modifiers)
	}
	localized := map[string]LocalizationReference{}
	for _, evidence := range glacial.Localization {
		localized[evidence.Column] = evidence
	}
	if localized["dsc2texta1"].Text != "Radius: " ||
		localized["desctexta2"].Text != "Freezes for " ||
		!strings.Contains(localized["str long"].Text, "nearby enemies") ||
		localized["dsc3texta1"].Source != "data/local/lng/eng/patchstring.tbl" ||
		len(localized["dsc3texta1"].ReplacementTokens) != 1 ||
		localized["dsc3texta1"].ReplacementTokens[0] != "%s" {
		t.Fatalf("Glacial Spike localization = %#v", glacial.Localization)
	}
}
