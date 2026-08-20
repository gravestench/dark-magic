package movement

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
)

// TestOwnedExpansion114dMovementAndKnockbackRecords pins the external MPQ facts that movement formulas assume.
// The opt-in boundary keeps ordinary tests hermetic while detecting incompatible source data during asset validation.
func TestOwnedExpansion114dMovementAndKnockbackRecords(t *testing.T) {
	store, catalog := loadOwnedMovementCatalog(t)

	assertOwnedClassRates(t, catalog)
	assertOwnedMovementStats(t, mustLoadMovementRows(t, store, "data/global/excel/ItemStatCost.txt"))
	assertOwnedMovementProperties(t, mustLoadMovementRows(t, store, "data/global/excel/Properties.txt"))
	assertOwnedKnockbackTargets(t, mustLoadMovementRows(t, store, "data/global/excel/monstats2.txt"))
	assertOwnedArmorPenalties(t, mustLoadMovementRows(t, store, "data/global/excel/armor.txt"))
}

// loadOwnedMovementCatalog opens the opt-in MPQ source and registers cleanup with the owning test.
func loadOwnedMovementCatalog(t *testing.T) (*recordstore.Store, Catalog) {
	t.Helper()

	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to the expansion 1.14d MPQ directory")
	}

	t.Setenv("MPQ_DIRECTORY", directory)

	assets, err := content.FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}

	// Cleanup errors cannot affect compatibility assertions after the owned source has been read.
	t.Cleanup(func() {
		_ = assets.Close()
	})

	store, _, err := recordstore.Pin(assets)
	if err != nil {
		t.Fatal(err)
	}

	catalog, err := LoadCatalog(store)
	if err != nil {
		t.Fatal(err)
	}

	return store, catalog
}

// mustLoadMovementRows fails at the source boundary so downstream assertions can focus on record meaning.
func mustLoadMovementRows(t *testing.T, store *recordstore.Store, path string) []map[string]string {
	t.Helper()

	rows, err := store.Load(path)
	if err != nil {
		t.Fatal(err)
	}

	return rows
}

// assertOwnedClassRates verifies every playable class fact consumed by the immutable movement catalog.
func assertOwnedClassRates(t *testing.T, catalog Catalog) {
	t.Helper()

	want := map[string]ClassRates{
		"Amazon":      ownedClassRates(20, 84, 20, 4, 4),
		"Sorceress":   ownedClassRates(10, 74, 20, 4, 4),
		"Necromancer": ownedClassRates(15, 79, 20, 4, 4),
		"Paladin":     ownedClassRates(25, 89, 20, 4, 4),
		"Barbarian":   ownedClassRates(25, 92, 20, 4, 4),
		"Druid":       ownedClassRates(25, 84, 20, 4, 4),
		"Assassin":    ownedClassRates(20, 95, 15, 5, 5),
	}
	for class, expected := range want {
		rates, found := catalog.Rates(class)
		if !found || rates != expected {
			t.Fatalf("owned expansion 1.14d %s movement rates = %+v, %v", class, rates, found)
		}
	}
}

// ownedClassRates keeps the universally pinned walk/run rates visible while naming class-specific stamina facts.
func ownedClassRates(vitality, stamina, runDrain, perLevel, perVitality int64) ClassRates {
	return ClassRates{
		Walk:               6,
		Run:                9,
		StartingVitality:   vitality,
		StartingStamina:    stamina,
		RunDrain:           runDrain,
		StaminaPerLevel:    perLevel,
		StaminaPerVitality: perVitality,
	}
}

// assertOwnedMovementStats pins callback IDs and dependency operands that define movement and stamina behavior.
func assertOwnedMovementStats(t *testing.T, rows []map[string]string) {
	t.Helper()

	for stat, expected := range map[string]map[string]string{
		"item_knockback": {
			"ID":             "81",
			"fCallback":      "1",
			"itemevent1":     "domeleedamage",
			"itemeventfunc1": "7",
			"itemevent2":     "domissiledamage",
			"itemeventfunc2": "7",
		},
		"maxstamina": {
			"ID": "11", "ValShift": "8", "fCallback": "1",
		},
		"vitality": {
			"ID": "3", "op": "9", "op stat2": "maxstamina",
		},
		"skill_staminapercent": {
			"ID": "162", "op": "1", "op stat1": "maxstamina",
		},
		"skill_passive_staminapercent": {
			"ID": "163", "op": "1", "op stat1": "maxstamina",
		},
		"item_stamina_perlevel": {
			"ID": "242", "op": "2", "op base": "level", "op param": "3", "op stat1": "maxstamina",
		},
		"item_stamina_bytime": {
			"ID": "295", "op": "6", "Signed": "1", "Divide": "1024", "op stat1": "maxstamina",
		},
	} {
		assertMovementRowFields(t, movementRowBy(rows, "Stat", stat), stat, expected)
	}

	assertVelocityStatChannels(t, rows)
}

// assertMovementRowFields reports the exact missing record field instead of obscuring it in a whole-map comparison.
func assertMovementRowFields(t *testing.T, row map[string]string, name string, expected map[string]string) {
	t.Helper()

	for field, value := range expected {
		if row == nil || row[field] != value {
			t.Fatalf("owned movement record %s field %s = %#v", name, field, row)
		}
	}
}

// assertVelocityStatChannels protects the two signed channels and item-FRW encoding used by movement rate rules.
func assertVelocityStatChannels(t *testing.T, rows []map[string]string) {
	t.Helper()

	velocity := movementRowBy(rows, "Stat", "velocitypercent")
	itemFRW := movementRowBy(rows, "Stat", "item_fastermovevelocity")
	velocityValid := velocity != nil && velocity["ID"] == "67" && velocity["Signed"] == "1" &&
		velocity["UpdateAnimRate"] == "1"

	itemFRWValid := itemFRW != nil && itemFRW["ID"] == "96" && itemFRW["Signed"] == "1" &&
		itemFRW["Multiply"] == "156" && itemFRW["Add"] == "4083"
	if !velocityValid || !itemFRWValid {
		t.Fatalf("owned movement stats velocity=%#v item_frw=%#v", velocity, itemFRW)
	}
}

// assertOwnedMovementProperties pins property-to-stat wiring, including the recovered by-time operand description.
func assertOwnedMovementProperties(t *testing.T, rows []map[string]string) {
	t.Helper()

	for _, code := range []string{"move1", "move2", "move3"} {
		row := movementRowBy(rows, "code", code)
		if row == nil || row["func1"] != "8" || row["stat1"] != "item_fastermovevelocity" {
			t.Fatalf("owned movement property %s = %#v", code, row)
		}
	}

	for code, expected := range map[string]string{
		"stam":     "maxstamina",
		"stam/lvl": "item_stamina_perlevel",
	} {
		row := movementRowBy(rows, "code", code)
		if row == nil || row["stat1"] != expected {
			t.Fatalf("owned stamina property %s = %#v", code, row)
		}
	}

	assertOwnedByTimeProperty(t, rows)

	knockback := movementRowBy(rows, "code", "knock")
	if knockback == nil || knockback["func1"] != "1" || knockback["stat1"] != "item_knockback" {
		t.Fatalf("owned knockback property = %#v", knockback)
	}
}

// assertOwnedByTimeProperty protects both executable wiring and documentation recovered from the source record.
func assertOwnedByTimeProperty(t *testing.T, rows []map[string]string) {
	t.Helper()

	byTime := movementRowBy(rows, "code", "stam/time")

	valid := byTime != nil && byTime["func1"] == "18" && byTime["stat1"] == "item_stamina_bytime" &&
		byTime["*param"] == "center period" &&
		byTime["*notes"] == "max at center period, min at opposite period, linear progression"
	if !valid {
		t.Fatalf("owned stamina bytime property = %#v", byTime)
	}
}

// assertOwnedKnockbackTargets pins monster size and knockback eligibility facts used by collision outcomes.
func assertOwnedKnockbackTargets(t *testing.T, rows []map[string]string) {
	t.Helper()

	for id, expected := range map[string]map[string]string{
		"fallen1":  {"mKB": "1", "dKB": "8", "small": "1", "large": ""},
		"bighead1": {"mKB": "1", "dKB": "8", "small": "", "large": "1"},
		"gorgon1":  {"mKB": "", "dKB": "8", "small": "", "large": ""},
	} {
		assertMovementRowFields(t, movementRowBy(rows, "Id", id), id, expected)
	}
}

// assertOwnedArmorPenalties pins armor and shield speed values that feed authoritative stamina drain.
func assertOwnedArmorPenalties(t *testing.T, rows []map[string]string) {
	t.Helper()

	for code, expected := range map[string]struct {
		speed    string
		itemType string
	}{
		"hla": {speed: "0", itemType: "tors"},
		"plt": {speed: "10", itemType: "tors"},
		"lrg": {speed: "5", itemType: "shie"},
		"tow": {speed: "10", itemType: "shie"},
	} {
		row := movementRowBy(rows, "code", code)
		if row["speed"] != expected.speed || row["type"] != expected.itemType {
			t.Fatalf(
				"owned armor %s speed/type = %q/%q, want %q/%q",
				code,
				row["speed"],
				row["type"],
				expected.speed,
				expected.itemType,
			)
		}
	}
}

// movementRowBy returns the first matching pinned record because source-table uniqueness is asserted elsewhere.
func movementRowBy(rows []map[string]string, column, value string) map[string]string {
	for _, row := range rows {
		if row[column] == value {
			return row
		}
	}

	return nil
}
