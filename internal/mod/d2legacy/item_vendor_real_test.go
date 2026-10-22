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

	store := recordstore.New(assets)
	store.SetLogger(nil)

	source := content.D2Legacy()

	const path = "lua/d2legacy/commands/item_vendor_test.lua"

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

	for _, path := range []string{"data/global/excel/npc.txt", "data/global/excel/misc.txt"} {
		rows, err := store.Load(path)
		if err != nil {
			t.Fatal(err)
		}

		values := make([]map[string]string, len(rows))
		copy(values, rows)

		config.records[path] = values
	}

	runLuaCase(t, source, path, "higher_level_purchases_keep_distinct_items_and_restore_supply", config)
}
