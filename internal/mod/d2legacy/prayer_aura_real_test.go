package d2legacy

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/localization"
)

// TestOwnedTargetPrayerAuraRecordsAndLocalizedIntent ties healing cadence and
// localized intent to the same owned aura data.
func TestOwnedTargetPrayerAuraRecordsAndLocalizedIntent(t *testing.T) {
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

	skills, err := store.Load("data/global/excel/skills.txt")
	if err != nil {
		t.Fatal(err)
	}

	prayer := rowBy(skills, "Id", "99")
	if prayer == nil {
		t.Fatal("owned expansion 1.14d Prayer row is missing")
	}

	for field, want := range map[string]string{
		"skill": "Prayer", "srvstfunc": "", "srvdofunc": "65", "aura": "1", "immediate": "",
		"leftskill": "", "range": "none", "InGame": "1", "InTown": "1", "aurafilter": "73731",
		"aurarangecalc": "ln12", "aurastate": "prayer", "auratargetstate": "prayer",
		"aurastat1": "hitpoints", "aurastatcalc1": "edns", "minmana": "1", "manashift": "4",
		"mana": "16", "lvlmana": "3", "Param1": "16", "Param2": "2", "perdelay": "50",
		"HitShift": "8", "EMin": "2", "EMinLev1": "1", "EMinLev2": "1", "EMinLev3": "2",
		"EMinLev4": "2", "EMinLev5": "3",
	} {
		if prayer[field] != want {
			t.Fatalf("owned expansion 1.14d Prayer %s = %q, want %q", field, prayer[field], want)
		}
	}

	states, err := store.Load("data/global/excel/states.txt")
	if err != nil {
		t.Fatal(err)
	}

	state := rowBy(states, "state", "prayer")
	if state == nil {
		t.Fatal("owned expansion 1.14d Prayer state is missing")
	}

	for field, want := range map[string]string{
		"id": "34", "aura": "1", "stat": "", "onsound": "paladin_aura_prayer",
		"overlay1": "aura_prayer_front", "overlay2": "aura_prayer_back", "castoverlay": "",
	} {
		if state[field] != want {
			t.Fatalf("owned expansion 1.14d Prayer state %s = %q, want %q", field, state[field], want)
		}
	}

	descriptions, err := store.Load("data/global/excel/SkillDesc.txt")
	if err != nil {
		t.Fatal(err)
	}

	description := rowBy(descriptions, "skilldesc", "prayer")
	if description == nil || description["desccalca2"] != "edmn" || description["desccalca3"] != "ln12" ||
		description["desctexta2"] != "StrSkill50" || description["desctexta3"] != "StrSkill18" {
		t.Fatalf("owned expansion 1.14d Prayer SkillDesc row = %#v", description)
	}

	locale := localization.New(assets, "English")
	for key, want := range map[string]string{
		"skillname99": "Prayer",
		"skillsd99":   "aura - regenerates life",
		"skillld99":   "the life of you and your party\nwhen active, aura slowly regenerates",
		"StrSkill50":  "Heals: ",
		"StrSkill18":  "Radius: ",
		"StrSkill3":   "Mana Cost: ",
	} {
		text, _, resolveErr := locale.Resolve(key)
		if resolveErr != nil || text != want {
			t.Fatalf("owned English TBL %s = %q, %v; want %q", key, text, resolveErr, want)
		}
	}
}
