package d2legacy

import (
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/localization"
)

// TestOwnedTargetColdArmorFamilyRecordsAndLocalizedIntent locks the generic
// decoder to the user's Expansion 1.14d records. SkillDesc and English TBL
// assertions preserve the human-authored modifier intent that is otherwise
// easy to miss when reading only server function numbers.
func TestOwnedTargetColdArmorFamilyRecordsAndLocalizedIntent(t *testing.T) {
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
	wants := map[string]map[string]string{
		"40": {
			"skill": "Frozen Armor", "srvdofunc": "18", "aurastate": "frozenarmor",
			"auraevent1": "damagedinmelee", "auraeventfunc1": "2", "Param1": "30", "Param2": "5",
			"Param3": "3000", "Param4": "300", "Param5": "30", "Param6": "3", "Param7": "250",
			"Param8": "5", "mana": "7", "manashift": "8", "cltoverlaya": "frozenarmor_hit",
		},
		"50": {
			"skill": "Shiver Armor", "srvdofunc": "18", "aurastate": "shiverarmor",
			"auraevent1": "attackedinmelee", "auraeventfunc1": "3", "Param1": "45", "Param2": "6",
			"Param3": "3000", "Param4": "300", "Param7": "250", "Param8": "9", "EType": "cold",
			"EMin": "12", "EMax": "16", "ELen": "100", "mana": "11", "manashift": "8",
			"cltoverlaya": "shiverarmor_hit",
		},
		"60": {
			"skill": "Chilling Armor", "srvdofunc": "18", "aurastate": "chillingarmor",
			"auraevent1": "hitbymissile", "auraeventfunc1": "1", "srvmissilea": "chillingarmorbolt",
			"cltmissilea": "chillingarmorbolt", "Param1": "45", "Param2": "5", "Param3": "3600",
			"Param4": "150", "Param7": "250", "Param8": "7", "EType": "cold", "EMin": "8",
			"EMax": "12", "ELen": "100", "mana": "17", "manashift": "8",
		},
	}
	for id, fields := range wants {
		row := rowBy(skills, "Id", id)
		if row == nil {
			t.Fatalf("owned Expansion 1.14d cold armor skill %s is missing", id)
		}
		for field, want := range fields {
			if row[field] != want {
				t.Fatalf("owned Expansion 1.14d %s %s = %q, want %q", row["skill"], field, row[field], want)
			}
		}
		if !strings.Contains(row["auralencalc"], ".blvl") {
			t.Fatalf("owned Expansion 1.14d %s duration no longer uses hard-point synergies: %q", row["skill"], row["auralencalc"])
		}
	}

	states, err := store.Load("data/global/excel/states.txt")
	if err != nil {
		t.Fatal(err)
	}
	for state, overlay := range map[string]string{
		"frozenarmor": "frozenarmor", "shiverarmor": "shiverarmor", "chillingarmor": "chillarmor",
	} {
		row := rowBy(states, "state", state)
		if row == nil || row["group"] != "1" || row["overlay1"] != overlay {
			t.Fatalf("owned Expansion 1.14d state %s = %#v, want exclusive group 1 overlay %s", state, row, overlay)
		}
	}

	missiles, err := store.Load("data/global/excel/Missiles.txt")
	if err != nil {
		t.Fatal(err)
	}
	bolt := rowBy(missiles, "Missile", "chillingarmorbolt")
	for field, want := range map[string]string{
		"Skill": "Chilling Armor", "CelFile": "IceBolt", "pSrvDoFunc": "1", "CollideType": "3",
		"CollideKill": "1", "Vel": "18", "Range": "25", "Size": "1", "NumDirections": "16",
		"CanSlow": "1", "HitShift": "8", "Trans": "1", "LoopAnim": "1",
		"TravelSound": "sorceress_icebolt_1", "HitSound": "impact_cold_1",
	} {
		if bolt == nil || bolt[field] != want {
			t.Fatalf("owned Expansion 1.14d chillingarmorbolt %s = %#v, want %q", field, bolt, want)
		}
	}
	if _, err := fs.Stat(assets, "data/global/missiles/IceBolt.dcc"); err != nil {
		t.Fatalf("owned Expansion 1.14d Chilling Armor missile DCC: %v", err)
	}

	descriptions, err := store.Load("data/global/excel/SkillDesc.txt")
	if err != nil {
		t.Fatal(err)
	}
	locale := localization.New(assets, "English")
	for skillID, name := range map[string]string{
		"40": "Frozen Armor", "50": "Shiver Armor", "60": "Chilling Armor",
	} {
		skill := rowBy(skills, "Id", skillID)
		descriptionID := skill["skilldesc"]
		description := rowBy(descriptions, "skilldesc", descriptionID)
		if description == nil {
			t.Fatalf("owned Expansion 1.14d SkillDesc %s is missing", descriptionID)
		}
		localizedName, _, resolveErr := locale.Resolve(description["str name"])
		if resolveErr != nil || localizedName != name {
			t.Fatalf("owned English TBL %s name = %q, %v; want %q", descriptionID, localizedName, resolveErr, name)
		}
		long, _, resolveErr := locale.Resolve(description["str long"])
		if resolveErr != nil || !strings.Contains(strings.ToLower(long), "defense") {
			t.Fatalf("owned English TBL %s long description = %q, %v; want defense intent", descriptionID, long, resolveErr)
		}
		if description["dsc3texta1"] == "" || description["dsc3texta2"] == "" {
			t.Fatalf("owned Expansion 1.14d SkillDesc %s no longer declares both family synergy labels", descriptionID)
		}
		for _, field := range []string{"dsc3texta1", "dsc3texta2"} {
			if text, _, resolveErr := locale.Resolve(description[field]); resolveErr != nil || text == "" {
				t.Fatalf("owned English TBL %s %s = %q, %v", descriptionID, field, text, resolveErr)
			}
		}
	}
	if text, _, resolveErr := locale.Resolve("Sksyn"); resolveErr != nil || text != "%s Receives Bonuses From:" {
		t.Fatalf("owned English TBL Sksyn = %q, %v; want replacement-token synergy heading", text, resolveErr)
	}
}
