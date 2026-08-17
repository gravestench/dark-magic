package skillevidence

import (
	"errors"
	"testing"
)

type fixtureLocale map[string]string

func (locale fixtureLocale) Resolve(key string) (string, string, error) {
	value, found := locale[key]
	if !found {
		return key, "", errors.New("missing")
	}
	return value, "patchstring.tbl", nil
}

func TestBuildJoinsTooltipTextTokensAndHardLevelModifiers(t *testing.T) {
	skills := []map[string]string{
		{"Id": "40", "skill": "Frozen Armor", "skilldesc": "frozen armor", "auralencalc": "skill('Shiver Armor'.blvl)"},
		{"Id": "50", "skill": "Shiver Armor", "skilldesc": "shiver armor"},
		{"Id": "60", "skill": "Chilling Armor", "skilldesc": "chilling armor"},
	}
	descriptions := []map[string]string{{
		"skilldesc": "frozen armor", "str name": "skillname40", "dsc3texta2": "skillname50",
		"dsc3textb2": "Secplev2", "dsc3calca2": "skill('Chilling Armor'.blvl) * par7",
	}}
	report, err := Build([]int{40}, skills, descriptions, fixtureLocale{
		"skillname40": "Frozen Armor", "skillname50": "Shiver Armor", "Secplev2": "+%d seconds per level",
	})
	if err != nil {
		t.Fatal(err)
	}
	skill := report.Skills[0]
	if len(skill.Localization) != 3 || skill.Localization[0].Text != "Shiver Armor" ||
		len(skill.Localization[1].ReplacementTokens) != 1 || skill.Localization[1].ReplacementTokens[0] != "%d" {
		t.Fatalf("localization = %#v", skill.Localization)
	}
	if len(skill.CrossSkillModifiers) != 2 || skill.CrossSkillModifiers[0].ReferencedID != 50 ||
		skill.CrossSkillModifiers[1].ReferencedID != 60 || skill.CrossSkillModifiers[1].LevelSelector != "blvl" {
		t.Fatalf("modifiers = %#v", skill.CrossSkillModifiers)
	}
}

func TestBuildKeepsMissingLocalizationVisible(t *testing.T) {
	report, err := Build([]int{0}, []map[string]string{{"Id": "0", "skill": "Attack", "skilldesc": "attack"}},
		[]map[string]string{{"skilldesc": "attack", "str long": "missing"}}, fixtureLocale{})
	if err != nil {
		t.Fatal(err)
	}
	if !report.Skills[0].Localization[0].Missing {
		t.Fatal("missing localization key was hidden")
	}
}
