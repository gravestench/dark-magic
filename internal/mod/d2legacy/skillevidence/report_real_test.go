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
	report, err := Build([]int{0, 36, 40, 52, 54, 55, 66, 72, 99, 109, 120, 69, 70, 89, 80, 95,
		75, 85, 90, 94}, skills, descriptions, localization.New(assets, "English"))
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
	enchant := report.Skills[3]
	if modifiers := enchant.CrossSkillModifiers; len(modifiers) != 1 || modifiers[0].ReferencedID != 37 {
		t.Fatalf("Enchant modifiers = %#v", modifiers)
	}
	enchantLocalized := map[string]LocalizationReference{}
	for _, evidence := range enchant.Localization {
		enchantLocalized[evidence.Column] = evidence
	}
	if enchantLocalized["str long"].Key != "skillld52" ||
		!strings.Contains(enchantLocalized["str long"].Text, "targeted character or minion") ||
		enchantLocalized["dsc3texta1"].Source != "data/local/lng/eng/patchstring.tbl" ||
		len(enchantLocalized["dsc3texta1"].ReplacementTokens) != 1 ||
		enchantLocalized["dsc3texta1"].ReplacementTokens[0] != "%s" {
		t.Fatalf("Enchant localization = %#v", enchant.Localization)
	}
	teleport := report.Skills[4]
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
	glacial := report.Skills[5]
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
	amplify := report.Skills[6]
	if len(amplify.CrossSkillModifiers) != 0 {
		t.Fatalf("Amplify Damage modifiers = %#v", amplify.CrossSkillModifiers)
	}
	amplifyLocalized := map[string]LocalizationReference{}
	for _, evidence := range amplify.Localization {
		amplifyLocalized[evidence.Column] = evidence
	}
	if amplifyLocalized["str name"].Text != "Amplify Damage" ||
		!strings.Contains(amplifyLocalized["str long"].Text, "non-magic damage") ||
		amplifyLocalized["dsc2texta1"].Text != "Damage Taken: " ||
		amplifyLocalized["desctexta2"].Text != "Duration: " ||
		amplifyLocalized["desctexta3"].Text != "Radius: " {
		t.Fatalf("Amplify Damage localization = %#v", amplify.Localization)
	}
	weaken := report.Skills[7]
	if len(weaken.CrossSkillModifiers) != 0 {
		t.Fatalf("Weaken modifiers = %#v", weaken.CrossSkillModifiers)
	}
	weakenLocalized := map[string]LocalizationReference{}
	for _, evidence := range weaken.Localization {
		weakenLocalized[evidence.Column] = evidence
	}
	if weakenLocalized["str name"].Text != "Weaken" ||
		!strings.Contains(weakenLocalized["str long"].Text, "reducing the amount of damage") ||
		weakenLocalized["dsc2texta1"].Text != "Target's Damage: " ||
		weakenLocalized["desctexta2"].Text != "Duration: " ||
		weakenLocalized["desctexta3"].Text != "Radius: " ||
		weakenLocalized["dsc2textb1"].Text != " percent" {
		t.Fatalf("Weaken localization = %#v", weaken.Localization)
	}
	if len(report.Skills[8].CrossSkillModifiers) != 0 {
		t.Fatalf("Prayer modifiers = %#v", report.Skills[8].CrossSkillModifiers)
	}
	for _, skill := range report.Skills[9:11] {
		modifiers := skill.CrossSkillModifiers
		if len(modifiers) != 3 || modifiers[0].ReferencedID != 99 || modifiers[0].Selector != "edns" ||
			modifiers[1].ReferencedID != 99 || modifiers[1].Selector != "edmn" ||
			modifiers[2].ReferencedID != 99 || modifiers[2].Selector != "edmn" {
			t.Fatalf("%s Prayer references = %#v", skill.Name, modifiers)
		}
	}
	raise := report.Skills[12]
	if modifiers := raise.CrossSkillModifiers; len(modifiers) != 10 {
		t.Fatalf("Raise Skeleton modifiers = %#v", modifiers)
	} else {
		for _, modifier := range modifiers {
			if modifier.ReferencedID != 69 || modifier.Referenced != "Skeleton Mastery" {
				t.Fatalf("Raise Skeleton modifier = %#v", modifier)
			}
		}
	}
	raiseLocalized := map[string]LocalizationReference{}
	for _, evidence := range raise.Localization {
		raiseLocalized[evidence.Column] = evidence
	}
	if raiseLocalized["str name"].Text != "Raise Skeleton" ||
		!strings.Contains(strings.ToLower(raiseLocalized["str long"].Text), "corpse") ||
		!strings.Contains(strings.ToLower(raiseLocalized["str long"].Text), "skeleton warrior") ||
		!strings.Contains(raiseLocalized["desctexta2"].Text, "skeleton total") ||
		raiseLocalized["desctexta4"].Text != "Defense: " ||
		raiseLocalized["desctexta5"].Text != "Attack: " ||
		raiseLocalized["desctexta6"].Text != "Damage: " ||
		raiseLocalized["str mana"].Text != "Mana Cost: " ||
		raiseLocalized["dsc3texta1"].ReplacementTokens[0] != "%s" {
		t.Fatalf("Raise Skeleton localization = %#v", raise.Localization)
	}
	modifierNames := map[int]string{}
	for _, index := range []int{11, 13} {
		for _, evidence := range report.Skills[index].Localization {
			if evidence.Column == "str name" {
				modifierNames[index] = evidence.Text
			}
		}
	}
	if modifierNames[11] != "Skeleton Mastery" || modifierNames[13] != "Summon Resist" {
		t.Fatalf("summon modifier localization = mastery %#v resist %#v", report.Skills[11], report.Skills[13])
	}
	for _, check := range []struct {
		index       int
		name        string
		longNeedle  string
		modifierLen int
	}{
		{index: 14, name: "Raise Skeletal Mage", longNeedle: "corpse", modifierLen: 7},
		{index: 15, name: "Revive", longNeedle: "fight by your side", modifierLen: 8},
	} {
		skill := report.Skills[check.index]
		if len(skill.CrossSkillModifiers) != check.modifierLen {
			t.Fatalf("%s modifiers = %#v", check.name, skill.CrossSkillModifiers)
		}
		for _, modifier := range skill.CrossSkillModifiers {
			if modifier.ReferencedID != 69 || modifier.Referenced != "Skeleton Mastery" {
				t.Fatalf("%s modifier = %#v", check.name, modifier)
			}
		}
		localized := map[string]LocalizationReference{}
		for _, evidence := range skill.Localization {
			localized[evidence.Column] = evidence
		}
		if localized["str name"].Text != check.name ||
			!strings.Contains(strings.ToLower(localized["str long"].Text), check.longNeedle) ||
			localized["dsc3texta2"].Text != "Skeleton Mastery" ||
			localized["dsc3texta3"].Text != "Summon Resist" ||
			len(localized["dsc3texta1"].ReplacementTokens) != 1 ||
			localized["dsc3texta1"].ReplacementTokens[0] != "%s" {
			t.Fatalf("%s localization = %#v", check.name, skill.Localization)
		}
	}
	for _, check := range []struct {
		index, modifierLen int
		name, longNeedle   string
	}{
		{16, 18, "Clay Golem", "earth"},
		{17, 18, "Blood Golem", "shares with you the life"},
		{18, 18, "Iron Golem", "metallic item"},
		{19, 20, "Fire Golem", "fire into life"},
	} {
		skill := report.Skills[check.index]
		if len(skill.CrossSkillModifiers) != check.modifierLen {
			t.Fatalf("%s modifiers = %#v", check.name, skill.CrossSkillModifiers)
		}
		localized := map[string]LocalizationReference{}
		for _, evidence := range skill.Localization {
			localized[evidence.Column] = evidence
		}
		if localized["str name"].Text != check.name ||
			!strings.Contains(strings.ToLower(localized["str long"].Text), check.longNeedle) ||
			localized["dsc3texta2"].Text != "Golem Mastery" ||
			localized["dsc3texta3"].Text != "Summon Resist" ||
			len(localized["dsc3texta1"].ReplacementTokens) != 1 ||
			localized["dsc3texta1"].ReplacementTokens[0] != "%s" {
			t.Fatalf("%s localization = %#v", check.name, skill.Localization)
		}
	}
}
