package d2legacy

import (
	"io/fs"
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/localization"
)

// TestOwnedTargetMeditationAuraRecordsAndLocalizedIntent ensures regeneration
// policy and player-facing text remain grounded in the same archive record.
func TestOwnedTargetMeditationAuraRecordsAndLocalizedIntent(t *testing.T) {
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

	meditation := rowBy(skills, "Id", "120")
	if meditation == nil {
		t.Fatal("owned expansion 1.14d Meditation row is missing")
	}

	for field, want := range map[string]string{
		"skill": "Meditation", "srvstfunc": "", "srvdofunc": "65", "aura": "1", "immediate": "",
		"leftskill": "", "range": "none", "InGame": "1", "InTown": "1", "aurafilter": "73729",
		"aurarangecalc": "ln12", "aurastate": "meditation", "auratargetstate": "meditation",
		"aurastat1": "manarecoverybonus", "aurastatcalc1": "ln34", "aurastat2": "hitpoints",
		"aurastatcalc2": "skill('Prayer'.edns)", "minmana": "0", "manashift": "8", "mana": "0",
		"lvlmana": "0", "Param1": "16", "Param2": "2", "Param3": "300", "Param4": "25",
		"perdelay": "50", "HitShift": "8", "reqlevel": "24", "reqskill1": "Cleansing",
	} {
		if meditation[field] != want {
			t.Fatalf("owned expansion 1.14d Meditation %s = %q, want %q", field, meditation[field], want)
		}
	}

	charStats, err := store.Load("data/global/excel/charstats.txt")
	if err != nil {
		t.Fatal(err)
	}

	paladin := rowBy(charStats, "class", "Paladin")
	if paladin == nil || paladin["ManaRegen"] != "120" {
		t.Fatalf("owned expansion 1.14d Paladin ManaRegen = %#v, want 120", paladin)
	}

	states, err := store.Load("data/global/excel/states.txt")
	if err != nil {
		t.Fatal(err)
	}

	state := rowBy(states, "state", "meditation")
	if state == nil {
		t.Fatal("owned expansion 1.14d Meditation state is missing")
	}

	for field, want := range map[string]string{
		"id": "48", "aura": "1", "immed": "1", "stat": "",
		"onsound": "paladin_aura_meditation", "overlay1": "meditationfront", "overlay2": "meditationback",
	} {
		if state[field] != want {
			t.Fatalf("owned expansion 1.14d Meditation state %s = %q, want %q", field, state[field], want)
		}
	}

	descriptions, err := store.Load("data/global/excel/SkillDesc.txt")
	if err != nil {
		t.Fatal(err)
	}

	description := rowBy(descriptions, "skilldesc", "meditation")
	if description == nil || description["desccalca1"] != "skill('Prayer'.edmn)" ||
		description["desccalca2"] != "ln34" || description["desccalca3"] != "ln12" ||
		description["dsc2calca1"] != "skill('Prayer'.edmn)" {
		t.Fatalf("owned expansion 1.14d Meditation SkillDesc row = %#v", description)
	}

	overlays, err := store.Load("data/global/excel/Overlay.txt")
	if err != nil {
		t.Fatal(err)
	}

	front := rowBy(overlays, "overlay", "meditationfront")

	back := rowBy(overlays, "overlay", "meditationback")
	if front == nil || front["Filename"] != "Meditation" || front["Frames"] != "13" ||
		front["AnimRate"] != "16" || front["Trans"] != "3" || front["PreDraw"] != "0" ||
		back == nil || back["Filename"] != "null" || back["Frames"] != "2" || back["PreDraw"] != "1" {
		t.Fatalf("owned expansion 1.14d Meditation overlays = front %#v back %#v", front, back)
	}

	if _, err := fs.Stat(assets, "data/global/overlays/Meditation.dcc"); err != nil {
		t.Fatalf("owned expansion 1.14d Meditation DCC: %v", err)
	}

	locale := localization.New(assets, "English")
	for key, want := range map[string]string{
		"skillname120": "Meditation",
		"skillsd120":   "aura - increases mana recovery",
		"skillld120":   "for you and your party\nwhen active, aura increases mana recovery",
		"StrSkill50":   "Heals: ",
		"StrSkill88":   "Mana Recovery Rate: ",
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
