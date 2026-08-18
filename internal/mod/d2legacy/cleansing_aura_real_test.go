package d2legacy

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/localization"
)

func TestOwnedTargetCleansingAuraRecordsAndLocalizedIntent(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to the expansion 1.14d MPQ directory")
	}
	t.Setenv("MPQ_DIRECTORY", directory)
	assets, err := content.FromEnvironment(content.Layer{Name: "d2legacy", FS: content.D2Legacy()})
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
	cleansing := rowBy(skills, "Id", "109")
	if cleansing == nil {
		t.Fatal("owned expansion 1.14d Cleansing row is missing")
	}
	for field, want := range map[string]string{
		"skill": "Cleansing", "srvstfunc": "", "srvdofunc": "65", "aura": "1", "immediate": "",
		"leftskill": "", "range": "none", "InGame": "1", "InTown": "1", "aurafilter": "73731",
		"aurarangecalc": "ln12", "aurastate": "cleansing", "auratargetstate": "cleansing",
		"aurastat1": "item_poisonlengthresist", "aurastatcalc1": "100-dm34",
		"aurastat2": "hitpoints", "aurastatcalc2": "skill('Prayer'.edns)",
		"minmana": "0", "manashift": "8", "mana": "0", "lvlmana": "0",
		"Param1": "16", "Param2": "2", "Param3": "30", "Param4": "90", "perdelay": "50", "HitShift": "8",
	} {
		if cleansing[field] != want {
			t.Fatalf("owned expansion 1.14d Cleansing %s = %q, want %q", field, cleansing[field], want)
		}
	}
	states, err := store.Load("data/global/excel/states.txt")
	if err != nil {
		t.Fatal(err)
	}
	state := rowBy(states, "state", "cleansing")
	if state == nil {
		t.Fatal("owned expansion 1.14d Cleansing state is missing")
	}
	for field, want := range map[string]string{
		"id": "45", "aura": "1", "stat": "", "onsound": "paladin_aura_cleansing",
		"overlay1": "cleansingfront", "overlay2": "cleansingback", "castoverlay": "",
	} {
		if state[field] != want {
			t.Fatalf("owned expansion 1.14d Cleansing state %s = %q, want %q", field, state[field], want)
		}
	}
	for stateID, want := range map[string]map[string]string{
		"poison":        {"id": "2"},
		"amplifydamage": {"curse": "1", "curable": "1"},
		"battlecry":     {"curse": "1", "curable": ""},
		"shrine_armor":  {"id": "128", "curse": "1", "curable": ""},
	} {
		row := rowBy(states, "state", stateID)
		if row == nil {
			t.Fatalf("owned expansion 1.14d %s state is missing", stateID)
		}
		for field, value := range want {
			if row[field] != value {
				t.Fatalf("owned expansion 1.14d %s %s = %q, want %q", stateID, field, row[field], value)
			}
		}
	}
	descriptions, err := store.Load("data/global/excel/SkillDesc.txt")
	if err != nil {
		t.Fatal(err)
	}
	description := rowBy(descriptions, "skilldesc", "cleansing")
	if description == nil || description["desccalca1"] != "skill('Prayer'.edmn)" ||
		description["desccalca2"] != "dm34" || description["desccalca3"] != "ln12" ||
		description["dsc2calca1"] != "skill('Prayer'.edmn)" {
		t.Fatalf("owned expansion 1.14d Cleansing SkillDesc row = %#v", description)
	}
	locale := localization.New(assets, "English")
	for key, want := range map[string]string{
		"skillname109": "Cleansing",
		"skillsd109":   "aura - reduces poison and curse duration",
		"skillld109":   "will remain poisoned or cursed\nof time you and your party\nwhen active, aura reduces the length",
		"StrSkill50":   "Heals: ",
		"StrSkill74":   "Duration reduced by ",
		"StrSkill23":   " percent",
		"Healplev2":    "Life Healed Every 2 Seconds",
		"Sksyn":        "%s Receives Bonuses From:",
	} {
		text, _, resolveErr := locale.Resolve(key)
		if resolveErr != nil || text != want {
			t.Fatalf("owned English TBL %s = %q, %v; want %q", key, text, resolveErr, want)
		}
	}
}
