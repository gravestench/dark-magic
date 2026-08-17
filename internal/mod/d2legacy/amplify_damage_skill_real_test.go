package d2legacy

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
)

func TestOwnedPointAreaAmplifyDamageRecords(t *testing.T) {
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
	amplify := rowBy(skills, "Id", "66")
	if amplify == nil {
		t.Fatal("owned expansion 1.14d Amplify Damage row is missing")
	}
	for field, want := range map[string]string{
		"skill": "Amplify Damage", "srvstfunc": "", "srvdofunc": "30", "aurafilter": "3",
		"auratargetstate": "amplifydamage", "auralencalc": "ln34", "aurarangecalc": "ln12",
		"aurastat1": "damageresist", "aurastatcalc1": "-par5", "range": "none", "anim": "SC",
		"LineOfSight": "4", "mana": "4", "lvlmana": "0", "minmana": "1", "manashift": "8",
		"interrupt": "1", "Param1": "3", "Param2": "1", "Param3": "200", "Param4": "75",
		"Param5": "100", "InGame": "1",
	} {
		if amplify[field] != want {
			t.Fatalf("owned expansion 1.14d Amplify Damage %s = %q, want %q", field, amplify[field], want)
		}
	}
	states, err := store.Load("data/global/excel/states.txt")
	if err != nil {
		t.Fatal(err)
	}
	state := rowBy(states, "state", amplify["auratargetstate"])
	if state == nil {
		t.Fatal("owned expansion 1.14d Amplify Damage state row is missing")
	}
	for field, want := range map[string]string{
		"state": "amplifydamage", "curse": "1", "curable": "1", "overlay1": "curseamplifydamage",
		"castoverlay": "curse_hit", "stat": "damageresist",
	} {
		if state[field] != want {
			t.Fatalf("owned expansion 1.14d Amplify Damage state %s = %q, want %q", field, state[field], want)
		}
	}
	descriptions, err := store.Load("data/global/excel/SkillDesc.txt")
	if err != nil {
		t.Fatal(err)
	}
	description := rowBy(descriptions, "skilldesc", "amplify damage")
	if description == nil {
		t.Fatal("owned expansion 1.14d Amplify Damage SkillDesc row is missing")
	}
	for field, want := range map[string]string{
		"desccalca2": "ln34", "desccalca3": "ln12", "dsc2calca1": "par5",
	} {
		if description[field] != want {
			t.Fatalf("owned expansion 1.14d Amplify Damage SkillDesc %s = %q, want %q", field, description[field], want)
		}
	}
}
