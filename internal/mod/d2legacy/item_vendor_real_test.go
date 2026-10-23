package d2legacy

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	lua "github.com/yuin/gopher-lua"
)

// TestOwnedAkaraConsumablePurchases runs the ordinary production purchase case
// with mounted vendor/item rows rather than the suite's synthetic price records.
func TestOwnedAkaraConsumablePurchases(t *testing.T) {
	runOwnedConsumableCase(t, "lua/d2legacy/commands/item_vendor_test.lua",
		"higher_level_purchases_keep_distinct_items_and_restore_supply")
}

// TestOwnedMinorHealingPotion uses the mounted hp1 record for purchase, healing and continuation.
func TestOwnedMinorHealingPotion(t *testing.T) {
	runOwnedConsumableCase(t, "lua/d2legacy/systems/potions_test.lua",
		"purchased_identity_heals_combat_damage_and_retires_once")
}

func runOwnedConsumableCase(t *testing.T, path, name string) {
	t.Helper()

	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to the owned expansion MPQ directory")
	}

	t.Setenv("MPQ_DIRECTORY", directory)

	assets, err := content.FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = assets.Close() })

	store, generation, err := recordstore.Pin(assets)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("mounted record generation %s", generation.ID)

	for _, file := range generation.Files {
		switch file.Path {
		case "data/global/excel/misc.txt", "data/global/excel/npc.txt", "data/global/excel/ItemStatCost.txt":
			t.Logf("record provenance: %s source=%s digest=%s", file.Path, file.Source, file.Digest)
		}
	}

	source := content.D2Legacy()

	fixture := newLuaSuiteFixture(t, source)
	config := luaSuiteConfig{seed: 42, records: fixtureRecords{}}

	if err := fixture.authority.Runtime.Run(t.Context(), func(state *lua.LState) error {
		suite, err := loadLuaSuite(state, source, path)
		if err != nil {
			return err
		}

		return config.read(suite)
	}); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"data/global/excel/npc.txt", "data/global/excel/misc.txt", "data/global/excel/itemstatcost.txt",
	} {
		rows, err := store.Load(path)
		if err != nil {
			t.Fatal(err)
		}

		values := make([]map[string]string, len(rows))
		copy(values, rows)

		config.records[path] = values
		if path == "data/global/excel/itemstatcost.txt" {
			regen := rowBy(values, "Stat", "hpregen")
			if regen == nil || (regen["ValShift"] != "" && regen["ValShift"] != "0") {
				t.Fatal("hpregen requires the qualified zero ValShift")
			}

			t.Logf("mounted hpregen ValShift=%q (blank compiles as zero)", regen["ValShift"])
		}

		if path == "data/global/excel/misc.txt" {
			for _, row := range values {
				if row["code"] == "hp1" {
					t.Logf("mounted hp1: pSpell=%s state=%s stat1=%s calc1=%s len=%s cost=%s",
						row["pSpell"], row["state"], row["stat1"], row["calc1"], row["len"], row["cost"])
				}
			}
		}
	}

	runLuaCase(t, source, path, name, config)
}
