package d2legacy

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	lua "github.com/yuin/gopher-lua"
)

func TestOwnedMonsterClassification(t *testing.T) {
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
	rows, err := store.Load("data/global/excel/monstats.txt")
	if err != nil {
		t.Fatal(err)
	}
	vectors := []struct {
		id, demon, lowUndead, highUndead string
	}{
		{"bighead1", "1", "", ""},
		{"skeleton1", "", "1", ""},
		{"unraveler1", "", "", "1"},
		{"foulcrow1", "", "", ""},
	}
	for _, vector := range vectors {
		row := rowBy(rows, "Id", vector.id)
		if row == nil {
			t.Fatalf("owned MonStats row %q is missing", vector.id)
		}
		for column, want := range map[string]string{
			"demon":   vector.demon,
			"lUndead": vector.lowUndead,
			"hUndead": vector.highUndead,
		} {
			if row[column] != want {
				t.Errorf("owned MonStats %s %s = %q, want %q", vector.id, column, row[column], want)
			}
		}
	}
	graphics, err := store.Load("data/global/excel/monstats2.txt")
	if err != nil {
		t.Fatal(err)
	}
	levels, err := store.Load("data/global/excel/monlvl.txt")
	if err != nil {
		t.Fatal(err)
	}
	fixture := newLuaSuiteFixture(t, content.D2Legacy(), luaSuiteConfig{
		profile: "authority",
		seed:    42,
		records: fixtureRecords{
			"data/global/excel/monstats.txt":  rows,
			"data/global/excel/monstats2.txt": graphics,
			"data/global/excel/monlvl.txt":    levels,
		},
	})
	err = fixture.authority.Runtime.Run(t.Context(), func(state *lua.LState) error {
		module, err := requireLuaModule(state, "d2legacy.tests.support.monster_classification_real")
		if err != nil {
			return err
		}
		check := module.RawGetString("check")
		for _, vector := range vectors {
			expected := state.NewTable()
			expected.RawSetString("demon", lua.LBool(vector.demon == "1"))
			expected.RawSetString("low_undead", lua.LBool(vector.lowUndead == "1"))
			expected.RawSetString("high_undead", lua.LBool(vector.highUndead == "1"))
			if err := state.CallByParam(lua.P{Fn: check, Protect: true}, lua.LString(vector.id), expected); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
