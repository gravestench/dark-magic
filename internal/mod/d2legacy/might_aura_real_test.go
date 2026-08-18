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

func TestOwnedTargetDefianceAuraRecordsAndLocalizedIntent(t *testing.T) {
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
	defiance := rowBy(skills, "Id", "104")
	if defiance == nil {
		t.Fatal("owned expansion 1.14d Defiance row is missing")
	}
	for field, want := range map[string]string{
		"skill": "Defiance", "srvstfunc": "", "srvdofunc": "65", "aura": "1", "immediate": "1",
		"leftskill": "", "range": "none", "InGame": "1", "aurafilter": "73731",
		"aurarangecalc": "ln12", "aurastate": "defiance", "auratargetstate": "defiance",
		"aurastat1": "skill_armor_percent", "aurastatcalc1": "ln34", "mana": "0", "lvlmana": "0",
		"Param1": "16", "Param2": "2", "Param3": "70", "Param4": "10", "perdelay": "50",
	} {
		if defiance[field] != want {
			t.Fatalf("owned expansion 1.14d Defiance %s = %q, want %q", field, defiance[field], want)
		}
	}
	states, err := store.Load("data/global/excel/states.txt")
	if err != nil {
		t.Fatal(err)
	}
	state := rowBy(states, "state", "defiance")
	if state == nil {
		t.Fatal("owned expansion 1.14d Defiance state row is missing")
	}
	for field, want := range map[string]string{
		"id": "37", "aura": "1", "stat": "skill_armor_percent", "onsound": "paladin_aura_defiance",
		"overlay1": "aura_defiance_front", "overlay2": "aura_defiance_back",
	} {
		if state[field] != want {
			t.Fatalf("owned expansion 1.14d Defiance state %s = %q, want %q", field, state[field], want)
		}
	}
	descriptions, err := store.Load("data/global/excel/SkillDesc.txt")
	if err != nil {
		t.Fatal(err)
	}
	description := rowBy(descriptions, "skilldesc", "defiance")
	if description == nil || description["desccalca1"] != "ln34" || description["desccalca2"] != "ln12" ||
		description["desctexta1"] != "StrSkill31" || description["desctexta2"] != "StrSkill18" {
		t.Fatalf("owned expansion 1.14d Defiance SkillDesc row = %#v", description)
	}
	locale := localization.New(assets, "English")
	for key, want := range map[string]string{
		"skillname104": "Defiance",
		"skillsd104":   "aura - increases defense",
		"skillld104":   "of you and your party\nwhen active, aura increases the defense rating",
		"StrSkill31":   "Defense Bonus: ",
		"StrSkill18":   "Radius: ",
	} {
		text, _, resolveErr := locale.Resolve(key)
		if resolveErr != nil || text != want {
			t.Fatalf("owned English TBL %s = %q, %v; want %q", key, text, resolveErr, want)
		}
	}
}

func TestOwnedTargetBlessedAimAuraAndLearnedPassiveRecords(t *testing.T) {
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
	blessedAim := rowBy(skills, "Id", "108")
	if blessedAim == nil {
		t.Fatal("owned expansion 1.14d Blessed Aim row is missing")
	}
	for field, want := range map[string]string{
		"skill": "Blessed Aim", "srvstfunc": "", "srvdofunc": "65", "aura": "1", "immediate": "1",
		"leftskill": "", "range": "none", "InGame": "1", "aurafilter": "73731",
		"aurarangecalc": "ln12", "aurastate": "blessedaim", "auratargetstate": "blessedaim",
		"aurastat1": "item_tohit_percent", "aurastatcalc1": "ln34", "mana": "0", "lvlmana": "0",
		"Param1": "16", "Param2": "2", "Param3": "75", "Param4": "15", "perdelay": "50",
		"passivestate": "penetrate", "passivestat1": "item_tohit_percent",
		"passivecalc1": "skill('Blessed Aim'.blvl) * par8", "Param8": "5",
	} {
		if blessedAim[field] != want {
			t.Fatalf("owned expansion 1.14d Blessed Aim %s = %q, want %q", field, blessedAim[field], want)
		}
	}
	states, err := store.Load("data/global/excel/states.txt")
	if err != nil {
		t.Fatal(err)
	}
	for name, fields := range map[string]map[string]string{
		"blessedaim": {
			"id": "40", "aura": "1", "stat": "item_tohit_percent", "onsound": "paladin_aura_blessedaim",
			"overlay1": "blessedaimfront", "overlay2": "blessedaimback",
		},
		"penetrate": {"id": "67"},
	} {
		state := rowBy(states, "state", name)
		if state == nil {
			t.Fatalf("owned expansion 1.14d Blessed Aim state %q is missing", name)
		}
		for field, want := range fields {
			if state[field] != want {
				t.Fatalf("owned expansion 1.14d state %s %s = %q, want %q", name, field, state[field], want)
			}
		}
	}
	descriptions, err := store.Load("data/global/excel/SkillDesc.txt")
	if err != nil {
		t.Fatal(err)
	}
	description := rowBy(descriptions, "skilldesc", "blessed aim")
	if description == nil || description["desccalca1"] != "ln34" || description["desccalca2"] != "ln12" ||
		description["desctexta1"] != "StrSkill22" || description["desctexta2"] != "StrSkill18" {
		t.Fatalf("owned expansion 1.14d Blessed Aim SkillDesc row = %#v", description)
	}
	locale := localization.New(assets, "English")
	for key, want := range map[string]string{
		"skillname108": "Blessed Aim",
		"skillsd108":   "aura - increases your attack rating",
		"skillld108":   "for you and your party\nwhen active, aura increases the attack rating",
		"StrSkill22":   "Attack: ",
		"StrSkill23":   " percent",
		"StrSkill18":   "Radius: ",
	} {
		text, _, resolveErr := locale.Resolve(key)
		if resolveErr != nil || text != want {
			t.Fatalf("owned English TBL %s = %q, %v; want %q", key, text, resolveErr, want)
		}
	}
}

func TestOwnedTargetResistFireAuraAndHardPointPassiveRecords(t *testing.T) {
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
	resistFire := rowBy(skills, "Id", "100")
	if resistFire == nil {
		t.Fatal("owned expansion 1.14d Resist Fire row is missing")
	}
	for field, want := range map[string]string{
		"skill": "Resist Fire", "srvstfunc": "", "srvdofunc": "65", "aura": "1", "immediate": "1",
		"leftskill": "", "range": "none", "InGame": "1", "aurafilter": "73731",
		"aurarangecalc": "ln12", "aurastate": "resistfire", "auratargetstate": "resistfire",
		"aurastat1": "fireresist", "aurastatcalc1": "dm34",
		"aurastat2": "maxfireresist", "aurastatcalc2": "skill('Resist Fire'.blvl)",
		"passivestate": "passive_resistfire", "passivestat1": "maxfireresist",
		"passivecalc1": "skill('Resist Fire'.blvl)/2", "mana": "0", "lvlmana": "0",
		"Param1": "16", "Param2": "2", "Param3": "35", "Param4": "150", "perdelay": "50",
	} {
		if resistFire[field] != want {
			t.Fatalf("owned expansion 1.14d Resist Fire %s = %q, want %q", field, resistFire[field], want)
		}
	}
	states, err := store.Load("data/global/excel/states.txt")
	if err != nil {
		t.Fatal(err)
	}
	for name, fields := range map[string]map[string]string{
		"resistfire": {
			"id": "3", "aura": "1", "stat": "fireresist", "onsound": "paladin_aura_resistfire",
			"overlay1": "aura_resistfire", "castoverlay": "cast_resistfire",
		},
		"passive_resistfire": {"id": "181"},
	} {
		state := rowBy(states, "state", name)
		if state == nil {
			t.Fatalf("owned expansion 1.14d Resist Fire state %q is missing", name)
		}
		for field, want := range fields {
			if state[field] != want {
				t.Fatalf("owned expansion 1.14d state %s %s = %q, want %q", name, field, state[field], want)
			}
		}
	}
	descriptions, err := store.Load("data/global/excel/SkillDesc.txt")
	if err != nil {
		t.Fatal(err)
	}
	description := rowBy(descriptions, "skilldesc", "resist fire")
	if description == nil || description["desccalca1"] != "dm34" || description["desccalca2"] != "ln12" ||
		description["desctexta1"] != "StrSkill51" || description["desctexta2"] != "StrSkill18" {
		t.Fatalf("owned expansion 1.14d Resist Fire SkillDesc row = %#v", description)
	}
	locale := localization.New(assets, "English")
	for key, want := range map[string]string{
		"skillname100": "Resist Fire",
		"skillsd100":   "aura - protects against fire damage",
		"skillld100":   "done to you and your party\nwhen active, aura decreases fire damage",
		"StrSkill51":   "Resist Fire: ",
		"StrSkill23":   " percent",
		"StrSkill18":   "Radius: ",
	} {
		text, _, resolveErr := locale.Resolve(key)
		if resolveErr != nil || text != want {
			t.Fatalf("owned English TBL %s = %q, %v; want %q", key, text, resolveErr, want)
		}
	}
}

func TestOwnedTargetResistColdAuraAndHardPointPassiveRecords(t *testing.T) {
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
	resistCold := rowBy(skills, "Id", "105")
	if resistCold == nil {
		t.Fatal("owned expansion 1.14d Resist Cold row is missing")
	}
	for field, want := range map[string]string{
		"skill": "Resist Cold", "srvstfunc": "", "srvdofunc": "65", "aura": "1", "immediate": "1",
		"leftskill": "", "range": "none", "InGame": "1", "aurafilter": "73731",
		"aurarangecalc": "ln12", "aurastate": "resistcold", "auratargetstate": "resistcold",
		"aurastat1": "coldresist", "aurastatcalc1": "dm34",
		"aurastat2": "maxcoldresist", "aurastatcalc2": "skill('Resist Cold'.blvl)",
		"passivestate": "passive_resistcold", "passivestat1": "maxcoldresist",
		"passivecalc1": "skill('Resist Cold'.blvl)/2", "mana": "0", "lvlmana": "0",
		"Param1": "16", "Param2": "2", "Param3": "35", "Param4": "150", "perdelay": "50",
	} {
		if resistCold[field] != want {
			t.Fatalf("owned expansion 1.14d Resist Cold %s = %q, want %q", field, resistCold[field], want)
		}
	}
	states, err := store.Load("data/global/excel/states.txt")
	if err != nil {
		t.Fatal(err)
	}
	for name, fields := range map[string]map[string]string{
		"resistcold": {
			"id": "4", "aura": "1", "stat": "coldresist", "onsound": "paladin_aura_resistcold",
			"overlay1": "aura_resistcold", "castoverlay": "cast_resistcold",
		},
		"passive_resistcold": {"id": "182"},
	} {
		state := rowBy(states, "state", name)
		if state == nil {
			t.Fatalf("owned expansion 1.14d Resist Cold state %q is missing", name)
		}
		for field, want := range fields {
			if state[field] != want {
				t.Fatalf("owned expansion 1.14d state %s %s = %q, want %q", name, field, state[field], want)
			}
		}
	}
	descriptions, err := store.Load("data/global/excel/SkillDesc.txt")
	if err != nil {
		t.Fatal(err)
	}
	description := rowBy(descriptions, "skilldesc", "resist cold")
	if description == nil || description["desccalca1"] != "dm34" || description["desccalca2"] != "ln12" ||
		description["desctexta1"] != "StrSkill52" || description["desctexta2"] != "StrSkill18" {
		t.Fatalf("owned expansion 1.14d Resist Cold SkillDesc row = %#v", description)
	}
	locale := localization.New(assets, "English")
	for key, want := range map[string]string{
		"skillname105": "Resist Cold",
		"skillsd105":   "aura - protects against cold damage",
		"skillld105":   "done to you and your party\nwhen active, aura decreases cold damage",
		"StrSkill52":   "Resist Cold: ",
		"StrSkill23":   " percent",
		"StrSkill18":   "Radius: ",
	} {
		text, _, resolveErr := locale.Resolve(key)
		if resolveErr != nil || text != want {
			t.Fatalf("owned English TBL %s = %q, %v; want %q", key, text, resolveErr, want)
		}
	}
}

func TestOwnedTargetResistLightningAuraAndHardPointPassiveRecords(t *testing.T) {
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
	resistLightning := rowBy(skills, "Id", "110")
	if resistLightning == nil {
		t.Fatal("owned expansion 1.14d Resist Lightning row is missing")
	}
	for field, want := range map[string]string{
		"skill": "Resist Lightning", "srvstfunc": "", "srvdofunc": "65", "aura": "1", "immediate": "1",
		"leftskill": "", "range": "none", "InGame": "1", "aurafilter": "73731",
		"aurarangecalc": "ln12", "aurastate": "resistlight", "auratargetstate": "resistlight",
		"aurastat1": "lightresist", "aurastatcalc1": "dm34",
		"aurastat2": "maxlightresist", "aurastatcalc2": "skill('Resist Lightning'.blvl)",
		"passivestate": "passive_resistltng", "passivestat1": "maxlightresist",
		"passivecalc1": "skill('Resist Lightning'.blvl)/2", "mana": "0", "lvlmana": "0",
		"Param1": "16", "Param2": "2", "Param3": "35", "Param4": "150", "perdelay": "50",
	} {
		if resistLightning[field] != want {
			t.Fatalf("owned expansion 1.14d Resist Lightning %s = %q, want %q", field, resistLightning[field], want)
		}
	}
	states, err := store.Load("data/global/excel/states.txt")
	if err != nil {
		t.Fatal(err)
	}
	for name, fields := range map[string]map[string]string{
		"resistlight": {
			"id": "5", "aura": "1", "stat": "lightresist", "onsound": "paladin_aura_resistlightning",
			"overlay1": "aura_resistlight", "castoverlay": "cast_resistlight",
		},
		"passive_resistltng": {"id": "183"},
	} {
		state := rowBy(states, "state", name)
		if state == nil {
			t.Fatalf("owned expansion 1.14d Resist Lightning state %q is missing", name)
		}
		for field, want := range fields {
			if state[field] != want {
				t.Fatalf("owned expansion 1.14d state %s %s = %q, want %q", name, field, state[field], want)
			}
		}
	}
	descriptions, err := store.Load("data/global/excel/SkillDesc.txt")
	if err != nil {
		t.Fatal(err)
	}
	description := rowBy(descriptions, "skilldesc", "resist lightning")
	if description == nil || description["desccalca1"] != "dm34" || description["desccalca2"] != "ln12" ||
		description["desctexta1"] != "StrSkill53" || description["desctexta2"] != "StrSkill18" {
		t.Fatalf("owned expansion 1.14d Resist Lightning SkillDesc row = %#v", description)
	}
	locale := localization.New(assets, "English")
	for key, want := range map[string]string{
		"skillname110": "Resist Lightning",
		"skillsd110":   "aura - protects against lightning damage",
		"skillld110":   "done to you and your party\nwhen active, aura decreases lightning damage",
		"StrSkill53":   "Resist Lightning: ",
		"StrSkill23":   " percent",
		"StrSkill18":   "Radius: ",
	} {
		text, _, resolveErr := locale.Resolve(key)
		if resolveErr != nil || text != want {
			t.Fatalf("owned English TBL %s = %q, %v; want %q", key, text, resolveErr, want)
		}
	}
}
