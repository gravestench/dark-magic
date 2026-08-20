package d2legacy

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
)

// TestOwnedTargetEnchantRecords pins the buff and damage fields that make
// Enchant affect both the target and its attacks.
func TestOwnedTargetEnchantRecords(t *testing.T) {
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

	enchant := rowBy(skills, "Id", "52")
	if enchant == nil {
		t.Fatal("owned expansion 1.14d Enchant row is missing")
	}

	for field, want := range map[string]string{
		"skill": "Enchant", "srvstfunc": "", "srvdofunc": "25", "anim": "SC", "range": "none",
		"TargetPet": "1", "TargetAlly": "1", "mana": "25", "lvlmana": "1", "manashift": "8", "minmana": "1",
		"aurastate": "enchant", "auralencalc": "ln12", "aurastat1": "firemindam", "aurastatcalc1": "enma",
		"aurastat2": "firemaxdam", "aurastatcalc2": "exma", "aurastat3": "item_tohit_percent",
		"aurastatcalc3": "toht", "EType": "fire", "EMin": "16", "EMax": "20", "HitShift": "7",
		"EDmgSymPerCalc": "(skill('Warmth'.blvl))*par8", "Param1": "3600", "Param2": "600", "Param3": "33",
		"Param8": "9", "EMinLev1": "3", "EMinLev2": "7", "EMinLev3": "11", "EMinLev4": "15",
		"EMinLev5": "19", "EMaxLev1": "5", "EMaxLev2": "9", "EMaxLev3": "13", "EMaxLev4": "17",
		"EMaxLev5": "21", "ToHit": "20", "LevToHit": "9", "ToHitCalc": "", "leftskill": "",
		"general": "", "InGame": "1",
	} {
		if enchant[field] != want {
			t.Fatalf("owned expansion 1.14d Enchant %s = %q, want %q", field, enchant[field], want)
		}
	}

	states, err := store.Load("data/global/excel/states.txt")
	if err != nil {
		t.Fatal(err)
	}

	state := rowBy(states, "state", enchant["aurastate"])
	if state == nil {
		t.Fatal("owned expansion 1.14d Enchant state row is missing")
	}

	for field, want := range map[string]string{"state": "enchant", "group": "", "castoverlay": "enchant"} {
		if state[field] != want {
			t.Fatalf("owned expansion 1.14d Enchant state %s = %q, want %q", field, state[field], want)
		}
	}

	descriptions, err := store.Load("data/global/excel/SkillDesc.txt")
	if err != nil {
		t.Fatal(err)
	}

	description := rowBy(descriptions, "skilldesc", "enchant")
	if description == nil {
		t.Fatal("owned expansion 1.14d Enchant SkillDesc row is missing")
	}

	for field, want := range map[string]string{
		"desccalca2": "toht", "desccalca4": "ln12", "dsc3calca1": "2", "dsc3calca2": "par8",
	} {
		if description[field] != want {
			t.Fatalf("owned expansion 1.14d Enchant SkillDesc %s = %q, want %q", field, description[field], want)
		}
	}

	calculations, err := store.Load("data/global/excel/skillcalc.txt")
	if err != nil {
		t.Fatal(err)
	}

	if len(calculations) <= 20 || calculations[20]["code"] != "toht" || calculations[20]["*desc"] != "to hit" {
		t.Fatalf("owned expansion 1.14d skill calculation 20 = %#v, want toht", calculations)
	}
}
