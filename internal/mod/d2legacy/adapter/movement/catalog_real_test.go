package movement

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
)

func TestOwnedExpansion114dClassMovementRates(t *testing.T) {
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
	pinned, _, err := recordstore.Pin(assets)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCatalog(pinned)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]ClassRates{
		"Amazon":      {Walk: 6, Run: 9, StartingVitality: 20, StartingStamina: 84, RunDrain: 20, StaminaPerLevel: 4, StaminaPerVitality: 4},
		"Sorceress":   {Walk: 6, Run: 9, StartingVitality: 10, StartingStamina: 74, RunDrain: 20, StaminaPerLevel: 4, StaminaPerVitality: 4},
		"Necromancer": {Walk: 6, Run: 9, StartingVitality: 15, StartingStamina: 79, RunDrain: 20, StaminaPerLevel: 4, StaminaPerVitality: 4},
		"Paladin":     {Walk: 6, Run: 9, StartingVitality: 25, StartingStamina: 89, RunDrain: 20, StaminaPerLevel: 4, StaminaPerVitality: 4},
		"Barbarian":   {Walk: 6, Run: 9, StartingVitality: 25, StartingStamina: 92, RunDrain: 20, StaminaPerLevel: 4, StaminaPerVitality: 4},
		"Druid":       {Walk: 6, Run: 9, StartingVitality: 25, StartingStamina: 84, RunDrain: 20, StaminaPerLevel: 4, StaminaPerVitality: 4},
		"Assassin":    {Walk: 6, Run: 9, StartingVitality: 20, StartingStamina: 95, RunDrain: 15, StaminaPerLevel: 5, StaminaPerVitality: 5},
	}
	for class, expected := range want {
		rates, found := catalog.Rates(class)
		if !found || rates != expected {
			t.Fatalf("owned expansion 1.14d %s movement rates = %+v, %v", class, rates, found)
		}
	}

	stats, err := pinned.Load("data/global/excel/ItemStatCost.txt")
	if err != nil {
		t.Fatal(err)
	}
	velocity, itemFRW := movementRowBy(stats, "Stat", "velocitypercent"), movementRowBy(stats, "Stat", "item_fastermovevelocity")
	for stat, expected := range map[string]map[string]string{
		"maxstamina":                   {"ID": "11", "ValShift": "8", "fCallback": "1"},
		"vitality":                     {"ID": "3", "op": "9", "op stat2": "maxstamina"},
		"skill_staminapercent":         {"ID": "162", "op": "1", "op stat1": "maxstamina"},
		"skill_passive_staminapercent": {"ID": "163", "op": "1", "op stat1": "maxstamina"},
		"item_stamina_perlevel":        {"ID": "242", "op": "2", "op base": "level", "op param": "3", "op stat1": "maxstamina"},
		"item_stamina_bytime":          {"ID": "295", "op": "6", "Signed": "1", "Divide": "1024", "op stat1": "maxstamina"},
	} {
		row := movementRowBy(stats, "Stat", stat)
		for field, value := range expected {
			if row == nil || row[field] != value {
				t.Fatalf("owned stamina stat %s field %s = %#v", stat, field, row)
			}
		}
	}
	if velocity == nil || velocity["ID"] != "67" || velocity["Signed"] != "1" || velocity["UpdateAnimRate"] != "1" ||
		itemFRW == nil || itemFRW["ID"] != "96" || itemFRW["Signed"] != "1" || itemFRW["Multiply"] != "156" || itemFRW["Add"] != "4083" {
		t.Fatalf("owned movement stats velocity=%#v item_frw=%#v", velocity, itemFRW)
	}
	properties, err := pinned.Load("data/global/excel/Properties.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, code := range []string{"move1", "move2", "move3"} {
		row := movementRowBy(properties, "code", code)
		if row == nil || row["func1"] != "8" || row["stat1"] != "item_fastermovevelocity" {
			t.Fatalf("owned movement property %s = %#v", code, row)
		}
	}
	for code, expected := range map[string]string{"stam": "maxstamina", "stam/lvl": "item_stamina_perlevel"} {
		row := movementRowBy(properties, "code", code)
		if row == nil || row["stat1"] != expected {
			t.Fatalf("owned stamina property %s = %#v", code, row)
		}
	}
	byTime := movementRowBy(properties, "code", "stam/time")
	if byTime == nil || byTime["func1"] != "18" || byTime["stat1"] != "item_stamina_bytime" ||
		byTime["*param"] != "center period" || byTime["*notes"] != "max at center period, min at opposite period, linear progression" {
		t.Fatalf("owned stamina bytime property = %#v", byTime)
	}
	armor, err := pinned.Load("data/global/excel/armor.txt")
	if err != nil {
		t.Fatal(err)
	}
	for code, want := range map[string]struct{ speed, itemType string }{
		"hla": {speed: "0", itemType: "tors"},
		"plt": {speed: "10", itemType: "tors"},
		"lrg": {speed: "5", itemType: "shie"},
		"tow": {speed: "10", itemType: "shie"},
	} {
		row := movementRowBy(armor, "code", code)
		if row["speed"] != want.speed || row["type"] != want.itemType {
			t.Fatalf("owned armor %s speed/type = %q/%q, want %q/%q", code, row["speed"], row["type"], want.speed, want.itemType)
		}
	}
}

func movementRowBy(rows []map[string]string, column, value string) map[string]string {
	for _, row := range rows {
		if row[column] == value {
			return row
		}
	}
	return nil
}
