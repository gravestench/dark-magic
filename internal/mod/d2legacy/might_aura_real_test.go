package d2legacy

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/localization"
)

func TestOwnedTargetMightAuraRecordsAndLocalizedIntent(t *testing.T) {
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
	might := rowBy(skills, "Id", "98")
	if might == nil {
		t.Fatal("owned expansion 1.14d Might row is missing")
	}
	for field, want := range map[string]string{
		"skill": "Might", "srvstfunc": "", "srvdofunc": "65", "aura": "1", "immediate": "1",
		"leftskill": "", "range": "none", "InGame": "1", "aurafilter": "73731",
		"aurarangecalc": "ln12", "aurastate": "might", "auratargetstate": "might",
		"aurastat1": "damagepercent", "aurastatcalc1": "ln34", "mana": "0", "lvlmana": "0",
		"Param1": "16", "Param2": "2", "Param3": "40", "Param4": "10", "perdelay": "50",
	} {
		if might[field] != want {
			t.Fatalf("owned expansion 1.14d Might %s = %q, want %q", field, might[field], want)
		}
	}
	states, err := store.Load("data/global/excel/states.txt")
	if err != nil {
		t.Fatal(err)
	}
	state := rowBy(states, "state", "might")
	if state == nil {
		t.Fatal("owned expansion 1.14d Might state row is missing")
	}
	for field, want := range map[string]string{
		"id": "33", "aura": "1", "stat": "damagepercent", "onsound": "paladin_aura_might",
		"overlay1": "aura_might_front", "overlay2": "aura_might_back",
	} {
		if state[field] != want {
			t.Fatalf("owned expansion 1.14d Might state %s = %q, want %q", field, state[field], want)
		}
	}
	descriptions, err := store.Load("data/global/excel/SkillDesc.txt")
	if err != nil {
		t.Fatal(err)
	}
	description := rowBy(descriptions, "skilldesc", "might")
	if description == nil || description["desccalca1"] != "ln34" || description["desccalca2"] != "ln12" ||
		description["desctexta1"] != "StrSkill4" || description["desctexta2"] != "StrSkill18" {
		t.Fatalf("owned expansion 1.14d Might SkillDesc row = %#v", description)
	}
	locale := localization.New(assets, "English")
	for key, want := range map[string]string{
		"skillname98": "Might",
		"skillsd98":   "aura - increases damage",
		"skillld98":   "done by you and your party\nwhen active, aura increases the damage",
		"StrSkill4":   "Damage: ",
		"StrSkill18":  "Radius: ",
	} {
		text, _, resolveErr := locale.Resolve(key)
		if resolveErr != nil || text != want {
			t.Fatalf("owned English TBL %s = %q, %v; want %q", key, text, resolveErr, want)
		}
	}
}
