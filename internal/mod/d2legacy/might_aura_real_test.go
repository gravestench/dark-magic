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

func TestOwnedTargetSalvationAuraRecordsAndLocalizedIntent(t *testing.T) {
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
	salvation := rowBy(skills, "Id", "125")
	if salvation == nil {
		t.Fatal("owned expansion 1.14d Salvation row is missing")
	}
	for field, want := range map[string]string{
		"skill": "Salvation", "srvstfunc": "", "srvdofunc": "65", "aura": "1", "immediate": "1",
		"leftskill": "", "range": "none", "InGame": "1", "aurafilter": "73731",
		"aurarangecalc": "ln12", "aurastate": "resistall", "auratargetstate": "resistall",
		"aurastat1": "fireresist", "aurastatcalc1": "dm34",
		"aurastat2": "coldresist", "aurastatcalc2": "dm34",
		"aurastat3": "lightresist", "aurastatcalc3": "dm34",
		"passivestate": "", "passivestat1": "", "passivecalc1": "", "mana": "0", "lvlmana": "0",
		"Param1": "16", "Param2": "2", "Param3": "50", "Param4": "120", "perdelay": "50",
	} {
		if salvation[field] != want {
			t.Fatalf("owned expansion 1.14d Salvation %s = %q, want %q", field, salvation[field], want)
		}
	}
	states, err := store.Load("data/global/excel/states.txt")
	if err != nil {
		t.Fatal(err)
	}
	state := rowBy(states, "state", "resistall")
	if state == nil {
		t.Fatal("owned expansion 1.14d Salvation state is missing")
	}
	for field, want := range map[string]string{
		"id": "8", "aura": "1", "stat": "lightresist", "onsound": "paladin_aura_salvation",
		"overlay1": "aura_resistall_front", "overlay2": "aura_resistall_back", "castoverlay": "cast_resistall",
	} {
		if state[field] != want {
			t.Fatalf("owned expansion 1.14d Salvation state %s = %q, want %q", field, state[field], want)
		}
	}
	descriptions, err := store.Load("data/global/excel/SkillDesc.txt")
	if err != nil {
		t.Fatal(err)
	}
	description := rowBy(descriptions, "skilldesc", "salvation")
	if description == nil || description["desccalca1"] != "dm34" || description["desccalca2"] != "ln12" ||
		description["desctexta1"] != "StrSkill54" || description["desctexta2"] != "StrSkill18" {
		t.Fatalf("owned expansion 1.14d Salvation SkillDesc row = %#v", description)
	}
	locale := localization.New(assets, "English")
	for key, want := range map[string]string{
		"skillname125": "Salvation",
		"skillsd125":   "aura - protects against elemental damage",
		"skillld125":   "done to you and your party\nwhen active, aura decreases fire, cold and lightning damage",
		"StrSkill54":   "Resist All: ",
		"StrSkill23":   " percent",
		"StrSkill18":   "Radius: ",
	} {
		text, _, resolveErr := locale.Resolve(key)
		if resolveErr != nil || text != want {
			t.Fatalf("owned English TBL %s = %q, %v; want %q", key, text, resolveErr, want)
		}
	}
}

func TestOwnedTargetVigorAuraRecordsAndLocalizedIntent(t *testing.T) {
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
	vigor := rowBy(skills, "Id", "115")
	if vigor == nil {
		t.Fatal("owned expansion 1.14d Vigor row is missing")
	}
	for field, want := range map[string]string{
		"skill": "Vigor", "srvstfunc": "", "srvdofunc": "65", "aura": "1", "immediate": "1",
		"leftskill": "", "range": "none", "InGame": "1", "aurafilter": "73731",
		"aurarangecalc": "ln12", "aurastate": "stamina", "auratargetstate": "stamina",
		"aurastat1": "staminarecoverybonus", "aurastatcalc1": "ln34",
		"aurastat2": "skill_staminapercent", "aurastatcalc2": "ln34",
		"aurastat3": "velocitypercent", "aurastatcalc3": "dm56",
		"passivestate": "", "passivestat1": "", "passivecalc1": "", "mana": "0", "lvlmana": "0",
		"Param1": "16", "Param2": "3", "Param3": "50", "Param4": "25",
		"Param5": "7", "Param6": "50", "perdelay": "50",
	} {
		if vigor[field] != want {
			t.Fatalf("owned expansion 1.14d Vigor %s = %q, want %q", field, vigor[field], want)
		}
	}
	states, err := store.Load("data/global/excel/states.txt")
	if err != nil {
		t.Fatal(err)
	}
	state := rowBy(states, "state", "stamina")
	if state == nil {
		t.Fatal("owned expansion 1.14d Vigor state is missing")
	}
	for field, want := range map[string]string{
		"id": "41", "aura": "1", "stat": "maxstamina", "onsound": "paladin_aura_stamina",
		"overlay1": "staminafront", "overlay2": "staminaback", "castoverlay": "",
	} {
		if state[field] != want {
			t.Fatalf("owned expansion 1.14d Vigor state %s = %q, want %q", field, state[field], want)
		}
	}
	itemStats, err := store.Load("data/global/excel/ItemStatCost.txt")
	if err != nil {
		t.Fatal(err)
	}
	staminaPercent := rowBy(itemStats, "Stat", "skill_staminapercent")
	if staminaPercent == nil || staminaPercent["op"] != "1" || staminaPercent["op stat1"] != "maxstamina" {
		t.Fatalf("owned expansion 1.14d skill_staminapercent ItemStatCost row = %#v", staminaPercent)
	}
	descriptions, err := store.Load("data/global/excel/SkillDesc.txt")
	if err != nil {
		t.Fatal(err)
	}
	description := rowBy(descriptions, "skilldesc", "vigor")
	if description == nil || description["desccalca1"] != "ln34" || description["desccalca2"] != "ln34" ||
		description["desccalca3"] != "dm56" || description["desccalca4"] != "ln12" ||
		description["desctexta1"] != "StrSkill69" || description["desctexta2"] != "StrSkill71" ||
		description["desctexta3"] != "StrSkill70" || description["desctexta4"] != "StrSkill18" {
		t.Fatalf("owned expansion 1.14d Vigor SkillDesc row = %#v", description)
	}
	locale := localization.New(assets, "English")
	for key, want := range map[string]string{
		"skillname115": "Vigor",
		"skillsd115":   "aura - increases speed and stamina recovery",
		"skillld115":   "and movement speed for you and your party\nwhen active, aura increases stamina recovery rate, maximum stamina",
		"StrSkill69":   "Stamina Recovery Rate: ",
		"StrSkill71":   "Stamina Bonus: ",
		"StrSkill70":   "Velocity: ",
		"StrSkill23":   " percent",
		"StrSkill18":   "Radius: ",
	} {
		text, _, resolveErr := locale.Resolve(key)
		if resolveErr != nil || text != want {
			t.Fatalf("owned English TBL %s = %q, %v; want %q", key, text, resolveErr, want)
		}
	}
}

func TestOwnedTargetThornsAuraRecordsAndLocalizedIntent(t *testing.T) {
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
	thorns := rowBy(skills, "Id", "103")
	if thorns == nil {
		t.Fatal("owned expansion 1.14d Thorns row is missing")
	}
	for field, want := range map[string]string{
		"skill": "Thorns", "srvstfunc": "", "srvdofunc": "65", "aura": "1", "immediate": "",
		"leftskill": "", "range": "none", "InGame": "1", "aurafilter": "73731",
		"aurarangecalc": "ln12", "aurastate": "thorns", "auratargetstate": "thorns",
		"aurastat1": "thorns_percent", "aurastatcalc1": "ln34",
		"passivestate": "", "passivestat1": "", "passivecalc1": "", "mana": "0", "lvlmana": "0",
		"Param1": "16", "Param2": "2", "Param3": "250", "Param4": "40", "perdelay": "50",
	} {
		if thorns[field] != want {
			t.Fatalf("owned expansion 1.14d Thorns %s = %q, want %q", field, thorns[field], want)
		}
	}
	states, err := store.Load("data/global/excel/states.txt")
	if err != nil {
		t.Fatal(err)
	}
	state := rowBy(states, "state", "thorns")
	if state == nil {
		t.Fatal("owned expansion 1.14d Thorns state is missing")
	}
	for field, want := range map[string]string{
		"id": "36", "aura": "1", "stat": "", "onsound": "paladin_aura_thorns",
		"overlay1": "aura_thorns_front", "overlay2": "aura_thorns_back", "castoverlay": "",
	} {
		if state[field] != want {
			t.Fatalf("owned expansion 1.14d Thorns state %s = %q, want %q", field, state[field], want)
		}
	}
	itemStats, err := store.Load("data/global/excel/ItemStatCost.txt")
	if err != nil {
		t.Fatal(err)
	}
	returnedPercent := rowBy(itemStats, "Stat", "thorns_percent")
	if returnedPercent == nil || returnedPercent["ID"] != "131" || returnedPercent["op"] != "" {
		t.Fatalf("owned expansion 1.14d thorns_percent ItemStatCost row = %#v", returnedPercent)
	}
	descriptions, err := store.Load("data/global/excel/SkillDesc.txt")
	if err != nil {
		t.Fatal(err)
	}
	description := rowBy(descriptions, "skilldesc", "thorns")
	if description == nil || description["desccalca1"] != "ln34" || description["desccalca2"] != "ln12" ||
		description["desctexta1"] != "StrSkill55" || description["desctexta2"] != "StrSkill18" {
		t.Fatalf("owned expansion 1.14d Thorns SkillDesc row = %#v", description)
	}
	locale := localization.New(assets, "English")
	for key, want := range map[string]string{
		"skillname103": "Thorns",
		"skillsd103":   "aura - reflects damage back at enemies",
		"skillld103":   "back at your attacker\nwhen active, aura reflects damage done to you",
		"StrSkill55":   " percent damage returned",
		"StrSkill18":   "Radius: ",
	} {
		text, _, resolveErr := locale.Resolve(key)
		if resolveErr != nil || text != want {
			t.Fatalf("owned English TBL %s = %q, %v; want %q", key, text, resolveErr, want)
		}
	}
}
