package d2legacy

import (
	"io/fs"
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/localization"
)

// TestOwnedTargetRedemptionAuraRecordsAndLocalizedIntent pins corpse conversion
// behavior and its localized description to the owned data set.
func TestOwnedTargetRedemptionAuraRecordsAndLocalizedIntent(t *testing.T) {
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

	redemption := rowBy(skills, "Id", "124")
	if redemption == nil {
		t.Fatal("owned expansion 1.14d Redemption row is missing")
	}

	for field, want := range map[string]string{
		"skill": "Redemption", "srvstfunc": "", "srvdofunc": "82", "aura": "1", "immediate": "",
		"leftskill": "", "range": "none", "InGame": "1", "InTown": "1", "aurafilter": "4354",
		"aurarangecalc": "ln12", "aurastate": "redemption", "auratargetstate": "",
		"calc1": "dm34", "calc2": "ln56", "calc3": "ln56", "cltmissilea": "redemption",
		"minmana": "0", "manashift": "8", "mana": "0", "lvlmana": "0", "Param1": "16",
		"Param2": "0", "Param3": "10", "Param4": "100", "Param5": "25", "Param6": "5",
		"perdelay": "50", "HitShift": "8", "reqlevel": "30", "reqskill1": "Vigor",
	} {
		if redemption[field] != want {
			t.Fatalf("owned expansion 1.14d Redemption %s = %q, want %q", field, redemption[field], want)
		}
	}

	states, err := store.Load("data/global/excel/states.txt")
	if err != nil {
		t.Fatal(err)
	}

	auraState := rowBy(states, "state", "redemption")
	redeemedState := rowBy(states, "state", "redeemed")

	if auraState == nil || auraState["id"] != "50" || auraState["aura"] != "1" ||
		auraState["onsound"] != "paladin_aura_redemption" || auraState["overlay1"] != "redemptionfront" ||
		auraState["overlay2"] != "redemptionback" {
		t.Fatalf("owned expansion 1.14d Redemption aura state = %#v", auraState)
	}

	if redeemedState == nil || redeemedState["id"] != "99" || redeemedState["udead"] != "1" ||
		redeemedState["monstaydeath"] != "1" || redeemedState["bossstaydeath"] != "1" ||
		redeemedState["hide"] != "1" || redeemedState["setfunc"] != "10" ||
		redeemedState["onsound"] != "paladin_redeemed_soul" {
		t.Fatalf("owned expansion 1.14d redeemed corpse state = %#v", redeemedState)
	}

	descriptions, err := store.Load("data/global/excel/SkillDesc.txt")
	if err != nil {
		t.Fatal(err)
	}

	description := rowBy(descriptions, "skilldesc", "redemption")
	if description == nil || description["desccalca1"] != "ln56" || description["desccalca2"] != "dm34" ||
		description["desccalca3"] != "ln12" || description["desctexta1"] != "StrSkill107" ||
		description["desctexta2"] != "StrSkill109" {
		t.Fatalf("owned expansion 1.14d Redemption SkillDesc row = %#v", description)
	}

	overlays, err := store.Load("data/global/excel/Overlay.txt")
	if err != nil {
		t.Fatal(err)
	}

	front := rowBy(overlays, "overlay", "redemptionfront")

	back := rowBy(overlays, "overlay", "redemptionback")
	if front == nil || front["Filename"] != "null" || front["Frames"] != "2" || front["Trans"] != "3" ||
		front["PreDraw"] != "0" || back == nil || back["Filename"] != "Redemption" ||
		back["Frames"] != "12" || back["AnimRate"] != "16" || back["Trans"] != "3" ||
		back["PreDraw"] != "1" {
		t.Fatalf("owned expansion 1.14d Redemption overlays = front %#v back %#v", front, back)
	}

	missiles, err := store.Load("data/global/excel/Missiles.txt")
	if err != nil {
		t.Fatal(err)
	}

	success := rowBy(missiles, "Missile", "redemption")

	failure := rowBy(missiles, "Missile", "redemptionfail")
	if success == nil || success["Id"] != "173" || success["CelFile"] != "RedemptionGhost" ||
		success["AnimLen"] != "16" || success["Range"] != "18" || success["Trans"] != "1" ||
		failure == nil || failure["Id"] != "174" || failure["CelFile"] != "CorpseExplodeGuts" ||
		failure["AnimLen"] != "13" || failure["Range"] != "13" || failure["NumDirections"] != "4" {
		t.Fatalf("owned expansion 1.14d Redemption missiles = success %#v failure %#v", success, failure)
	}

	for _, path := range []string{
		"data/global/overlays/Redemption.dcc",
		"data/global/missiles/RedemptionGhost.dcc",
		"data/global/missiles/CorpseExplodeGuts.dcc",
	} {
		if _, err := fs.Stat(assets, path); err != nil {
			t.Fatalf("owned expansion 1.14d Redemption asset %s: %v", path, err)
		}
	}

	locale := localization.New(assets, "English")
	for key, want := range map[string]string{
		"skillname124": "Redemption",
		"skillsd124":   "aura - redeems the dead for mana and life",
		"skillxld124":  "you life and mana\nthe souls of slain enemies to give\nwhen active, aura attempts to redeem",
		"StrSkill107":  "Life/Mana Recovered: ",
		"StrSkill108":  " points",
		"StrSkill109":  "Chance to redeem soul: ",
		"StrSkill18":   "Radius: ",
		"StrSkill23":   " percent",
	} {
		text, _, resolveErr := locale.Resolve(key)
		if resolveErr != nil || text != want {
			t.Fatalf("owned English TBL %s = %q, %v; want %q", key, text, resolveErr, want)
		}
	}
}
