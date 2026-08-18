package d2legacy

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	dcc "github.com/gravestench/dcc/pkg"
)

func ownedDCCBounds(assets fs.FS, path string, direction int) (string, error) {
	file, err := assets.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	encoded, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	asset, err := dcc.OpenBytes(encoded)
	if err != nil {
		return "", err
	}
	decoded, err := asset.DecodeDirection(direction)
	if err != nil {
		return "", err
	}
	frames := decoded.Frames()
	if len(frames) == 0 {
		return "", fmt.Errorf("no frames")
	}
	return frames[0].Bounds().String(), nil
}

func TestOwnedStateOverlayPresentationRecordsAndAssets(t *testing.T) {
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
	overlays, err := store.Load("data/global/excel/Overlay.txt")
	if err != nil {
		t.Fatal(err)
	}
	monsterGraphics, err := store.Load("data/global/excel/monstats2.txt")
	if err != nil {
		t.Fatal(err)
	}
	heightCategories := map[string]bool{}
	for _, row := range monsterGraphics {
		value := row["OverlayHeight"]
		if value != "" && value != "0" && value != "1" && value != "2" && value != "3" && value != "4" {
			t.Fatalf("owned Expansion 1.14d MonStats2 OverlayHeight %q is outside 0..4", value)
		}
		heightCategories[value] = true
	}
	for _, value := range []string{"1", "2", "3", "4"} {
		if !heightCategories[value] {
			t.Fatalf("owned Expansion 1.14d MonStats2 has no OverlayHeight category %s", value)
		}
	}
	want := map[string]map[string]string{
		"staminafront":         {"Filename": "Stamina_front", "Frames": "16", "AnimRate": "8", "Trans": "3", "PreDraw": "0", "Character": "Paladin", "Red": "255", "Green": "255", "Blue": "255"},
		"staminaback":          {"Filename": "Stamina_back", "Frames": "16", "AnimRate": "8", "Trans": "3", "PreDraw": "1", "Character": "Paladin", "Red": "255", "Green": "255", "Blue": "255"},
		"aura_resistall_front": {"Filename": "AuraResistAllFront", "Frames": "20", "AnimRate": "16", "Trans": "3", "PreDraw": "0", "Character": "Paladin", "InitRadius": "6", "Radius": "6", "Red": "211", "Green": "148", "Blue": "255"},
		"aura_resistall_back":  {"Filename": "AuraResistAllBack", "Frames": "20", "AnimRate": "16", "Trans": "3", "PreDraw": "1", "Character": "Paladin", "InitRadius": "6", "Radius": "6", "Red": "211", "Green": "148", "Blue": "255"},
		"cast_resistall":       {"Filename": "AuraResistAllCast", "Frames": "20", "AnimRate": "16", "Trans": "3", "PreDraw": "1", "Character": "Paladin", "InitRadius": "1", "Radius": "10", "Red": "211", "Green": "148", "Blue": "255"},
		"aura_resistlight":     {"Filename": "AuraResistLightningBack", "Frames": "20", "AnimRate": "16", "Trans": "3", "PreDraw": "1", "InitRadius": "6", "Radius": "6", "Red": "255", "Green": "255", "Blue": "200"},
		"aura_resistcold":      {"Filename": "AuraResistColdBack", "Frames": "20", "AnimRate": "16", "Trans": "3", "PreDraw": "1", "InitRadius": "6", "Radius": "6", "Red": "210", "Green": "210", "Blue": "255"},
		"aura_resistfire":      {"Filename": "AuraResistFireFront", "Frames": "20", "AnimRate": "16", "Trans": "3", "PreDraw": "1", "InitRadius": "6", "Radius": "6", "Red": "255", "Green": "193", "Blue": "103"},
		"blessedaimback":       {"Filename": "BlessedAim_back", "Frames": "10", "AnimRate": "8", "Trans": "3"},
		"blessedaimfront":      {"Filename": "BlessedAim_front", "Frames": "10", "AnimRate": "8", "Trans": "3"},
		"cast_resistfire":      {"Filename": "AuraResistFireFrontCast", "Frames": "11", "AnimRate": "16", "Trans": "3", "PreDraw": "1", "InitRadius": "1", "Radius": "10", "Red": "255", "Green": "193", "Blue": "103"},
		"cast_resistlight":     {"Filename": "AuraResistLightningBackCast", "Frames": "11", "AnimRate": "16", "Trans": "3", "PreDraw": "0", "InitRadius": "1", "Radius": "10", "Red": "255", "Green": "255", "Blue": "200"},
		"cast_resistcold":      {"Filename": "AuraResistColdBackCast", "Frames": "11", "AnimRate": "16", "Trans": "3", "PreDraw": "0", "InitRadius": "1", "Radius": "10", "Red": "210", "Green": "210", "Blue": "255"},
		"aura_defiance_back":   {"Filename": "AuraDefianceBack", "Frames": "15", "AnimRate": "8", "Trans": "3"},
		"aura_defiance_front":  {"Filename": "AuraDefianceFront", "Frames": "15", "AnimRate": "8", "Trans": "3"},
		"aura_might_back":      {"Filename": "AuraMightBack"},
		"aura_might_front":     {"Filename": "AuraMightFront"},
		"frozenarmor":          {"Filename": "FrozenArmor", "Frames": "24", "AnimRate": "16", "Trans": "3"},
		"curse_hit":            {"Filename": "CurseHit", "Frames": "10", "AnimRate": "16", "Trans": "3"},
		"curseamplifydamage":   {"Filename": "CurseAmplifyDamageEffect", "Frames": "24", "AnimRate": "16", "Trans": "3", "Character": "all", "1ofN": "1", "Height1": "14", "Height2": "0", "Height3": "-14", "Height4": "-60", "InitRadius": "1", "Radius": "6", "Red": "255", "Green": "64", "Blue": "64"},
		"curseweaken":          {"Filename": "CurseWeakenEffect", "Frames": "24", "AnimRate": "16", "Trans": "3", "Character": "all", "1ofN": "1", "Height1": "14", "Height2": "0", "Height3": "-14", "Height4": "-60", "InitRadius": "1", "Radius": "6", "Red": "255", "Green": "210", "Blue": "210"},
		"enchant":              {"Filename": "FireEnchant", "Frames": "17", "AnimRate": "16", "Trans": "3"},
		"fire_cast_1":          {"Filename": "FireCast_for_Sorceress", "Frames": "14", "AnimRate": "16", "NumDirections": "1", "Trans": "3"},
		"fire_cast_2":          {"Filename": "FireCast2", "Frames": "16", "AnimRate": "16", "NumDirections": "1", "Trans": "3"},
		"ice_cast_1":           {"Filename": "IceCastNew01", "Frames": "15", "AnimRate": "16", "NumDirections": "1", "Trans": "3"},
		"ice_cast_2":           {"Filename": "IceCastNew02", "Frames": "15", "AnimRate": "16", "NumDirections": "1", "Trans": "3"},
		"light_cast_1":         {"Filename": "LightningCast", "Frames": "10", "AnimRate": "16", "NumDirections": "1", "Trans": "3"},
		"teleport":             {"Filename": "Teleport", "Frames": "18", "AnimRate": "16", "NumDirections": "1", "Trans": "3"},
	}
	for id, fields := range want {
		row := rowBy(overlays, "overlay", id)
		if row == nil {
			t.Fatalf("owned expansion 1.14d overlay %q is missing", id)
		}
		for field, value := range fields {
			if row[field] != value {
				t.Fatalf("owned expansion 1.14d overlay %s %s = %q, want %q", id, field, row[field], value)
			}
		}
		path := "data/global/overlays/" + row["Filename"] + ".dcc"
		if _, err := fs.Stat(assets, path); err != nil {
			t.Fatalf("owned expansion 1.14d overlay %s asset %q: %v", id, path, err)
		}
	}
	for _, target := range []struct {
		path      string
		direction int
		bounds    string
	}{
		{path: "data/global/overlays/FireCast2.dcc", bounds: "(-74,-89)-(71,44)"},
		{path: "data/global/missiles/Fireball.dcc", direction: 4, bounds: "(-17,-81)-(17,-26)"},
	} {
		bounds, err := ownedDCCBounds(assets, target.path, target.direction)
		if err != nil {
			t.Fatalf("owned Expansion 1.14d DCC bounds %s: %v", target.path, err)
		}
		if bounds != target.bounds {
			t.Fatalf("owned Expansion 1.14d DCC %s direction %d canvas = %s, want %s", target.path, target.direction, bounds, target.bounds)
		}
	}
}
