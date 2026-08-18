package d2legacy

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
)

func TestOwnedTargetCorpseSummonFamilyRecordsDefineMaterializationAndModifiers(t *testing.T) {
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
	raise := rowBy(skills, "Id", "70")
	if raise == nil {
		t.Fatal("owned Expansion 1.14d Raise Skeleton row is missing")
	}
	for field, want := range map[string]string{
		"srvstfunc": "15", "srvdofunc": "31", "TargetCorpse": "1", "SelectProc": "2",
		"leftskill": "", "InTown": "", "range": "none", "summon": "necroskeleton",
		"pettype": "skeleton", "summode": "S1", "mana": "6", "lvlmana": "1",
		"petmax":        "(lvl < 4) ?lvl:(2+lvl/3)",
		"calc1":         "(lvl < 4) ? 0 : (par2 * (lvl - 3))",
		"aurastatcalc1": "((lvl < 4) ? 0 : ((lvl-3)*par3))",
		"aurastatcalc2": "(lvl+skill('Skeleton Mastery'.lvl))*par4",
		"aurastatcalc3": "(lvl+skill('Skeleton Mastery'.lvl))*par5",
		"passivecalc1":  "skill('Skeleton Mastery'.lvl) * skill('Skeleton Mastery'.par1) * 256",
		"passivecalc2":  "skill('Skeleton Mastery'.lvl) * skill('Skeleton Mastery'.par2) + edmn",
	} {
		if raise[field] != want {
			t.Fatalf("owned Expansion 1.14d Raise Skeleton %s = %q, want %q", field, raise[field], want)
		}
	}
	mage := rowBy(skills, "Id", "80")
	if mage == nil {
		t.Fatal("owned Expansion 1.14d Raise Skeletal Mage row is missing")
	}
	for field, want := range map[string]string{
		"srvstfunc": "15", "srvdofunc": "31", "TargetCorpse": "1", "SelectProc": "2",
		"leftskill": "", "InTown": "", "range": "none", "summon": "necromage",
		"pettype": "skeletonmage", "summode": "S1", "mana": "8", "lvlmana": "1",
		"petmax":        "(lvl < 4) ?lvl:(2+lvl/3)",
		"calc1":         "(lvl < 4) ? 0 : (par2 * (lvl - 3))",
		"aurastatcalc1": "(lvl+skill('Skeleton Mastery'.lvl))*par5",
		"passivecalc1":  "skill('Skeleton Mastery'.lvl) * skill('Skeleton Mastery'.par1) * 256",
		"sumskill1":     "NecromageMissile",
		"sumsk1calc":    "skill('Skeleton Mastery'.lvl) + ((lvl < 4)?0:((lvl-2)/2))",
	} {
		if mage[field] != want {
			t.Fatalf("owned Expansion 1.14d Raise Skeletal Mage %s = %q, want %q", field, mage[field], want)
		}
	}
	revive := rowBy(skills, "Id", "95")
	if revive == nil {
		t.Fatal("owned Expansion 1.14d Revive row is missing")
	}
	for field, want := range map[string]string{
		"srvstfunc": "21", "srvdofunc": "58", "TargetCorpse": "1", "SelectProc": "3",
		"leftskill": "", "InTown": "", "range": "none", "summon": "",
		"pettype": "revive", "petmax": "lvl", "summode": "NU", "mana": "45", "lvlmana": "0",
		"calc1":         "par1+skill('Skeleton Mastery'.lvl) * skill('Skeleton Mastery'.par3)",
		"calc2":         "ln34",
		"aurastatcalc1": "skill('Skeleton Mastery'.lvl) * skill('Skeleton Mastery'.par4)",
		"passivecalc1":  "par5",
		"Param1":        "200",
		"Param3":        "4500",
		"Param4":        "0",
		"Param5":        "50",
	} {
		if revive[field] != want {
			t.Fatalf("owned Expansion 1.14d Revive %s = %q, want %q", field, revive[field], want)
		}
	}
	mastery, resist := rowBy(skills, "Id", "69"), rowBy(skills, "Id", "89")
	if mastery == nil || mastery["Param1"] != "8" || mastery["Param2"] != "2" ||
		mastery["Param3"] != "5" || mastery["Param4"] != "10" {
		t.Fatalf("Skeleton Mastery record = %#v", mastery)
	}
	if resist == nil || resist["passivestat1"] != "passive_summon_resist" ||
		resist["passivecalc1"] != "dm12" || resist["Param1"] != "20" || resist["Param2"] != "75" {
		t.Fatalf("Summon Resist record = %#v", resist)
	}

	pets, err := store.Load("data/global/excel/PetType.txt")
	if err != nil {
		t.Fatal(err)
	}
	skeleton := rowBy(pets, "pet type", "skeleton")
	if skeleton == nil || skeleton["basemax"] != "0" || skeleton["drawhp"] != "1" ||
		skeleton["automap"] != "1" || skeleton["icontype"] != "2" || skeleton["unsummon"] != "1" ||
		skeleton["warp"] != "1" || skeleton["partysend"] != "1" || skeleton["baseicon"] != "skeletonicon" {
		t.Fatalf("skeleton PetType record = %#v", skeleton)
	}
	for _, petType := range []string{"skeletonmage", "revive"} {
		pet := rowBy(pets, "pet type", petType)
		if pet == nil || pet["basemax"] != "0" || pet["drawhp"] != "1" ||
			pet["automap"] != "1" || pet["icontype"] != "2" || pet["unsummon"] != "1" ||
			pet["warp"] != "1" || pet["partysend"] != "1" {
			t.Fatalf("%s PetType record = %#v", petType, pet)
		}
	}

	monsters, err := store.Load("data/global/excel/monstats.txt")
	if err != nil {
		t.Fatal(err)
	}
	pet := rowBy(monsters, "Id", "necroskeleton")
	if pet == nil || pet["AI"] != "NecroPet" || pet["Align"] != "1" || pet["enabled"] != "1" ||
		pet["isSpawn"] != "" || pet["noRatio"] != "1" || pet["minHP"] != "21" ||
		pet["MinHP(N)"] != "30" || pet["MinHP(H)"] != "42" || pet["A1MinD"] != "1" ||
		pet["A1MaxD"] != "2" || pet["AC"] != "5" || pet["A1TH"] != "5" {
		t.Fatalf("necroskeleton MonStats record = %#v", pet)
	}
	magePet := rowBy(monsters, "Id", "necromage")
	if magePet == nil || magePet["AI"] != "NecroPet" || magePet["Align"] != "1" ||
		magePet["enabled"] != "1" || magePet["isSpawn"] != "" || magePet["noRatio"] != "1" ||
		magePet["minHP"] != "61" || magePet["maxHP"] != "61" || magePet["AC"] != "24" ||
		magePet["Skill1"] != "NecromageMissile" {
		t.Fatalf("necromage MonStats record = %#v", magePet)
	}
	graphics, err := store.Load("data/global/excel/MonStats2.txt")
	if err != nil {
		t.Fatal(err)
	}
	petGraphics := rowBy(graphics, "Id", "necroskeleton")
	if petGraphics == nil || petGraphics["BaseW"] != "1hs" || petGraphics["SizeX"] != "2" ||
		petGraphics["SizeY"] != "2" || petGraphics["OverlayHeight"] != "2" || petGraphics["isSel"] != "1" {
		t.Fatalf("necroskeleton MonStats2 record = %#v", petGraphics)
	}
	mageGraphics := rowBy(graphics, "Id", "necromage")
	if mageGraphics == nil || mageGraphics["BaseW"] != "hth" || mageGraphics["SizeX"] != "2" ||
		mageGraphics["SizeY"] != "2" || mageGraphics["OverlayHeight"] != "2" || mageGraphics["isSel"] != "1" {
		t.Fatalf("necromage MonStats2 record = %#v", mageGraphics)
	}
	fallenGraphics := rowBy(graphics, "Id", "fallen1")
	if fallenGraphics == nil || fallenGraphics["corpseSel"] != "1" || fallenGraphics["revive"] != "1" {
		t.Fatalf("fallen1 revive capability = %#v", fallenGraphics)
	}
	levels, err := store.Load("data/global/excel/MonLvl.txt")
	if err != nil {
		t.Fatal(err)
	}
	levelOne := rowBy(levels, "Level", "1")
	if levelOne == nil || levelOne["L-AC"] != "6" || levelOne["L-TH"] != "8" {
		t.Fatalf("level-one summon baselines = %#v", levelOne)
	}
}
