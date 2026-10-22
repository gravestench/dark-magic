package d2legacy

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	lua "github.com/yuin/gopher-lua"
)

// TestOwnedMonsterResistances pins the MonStats values consumed by monster spawns.
func TestOwnedMonsterResistances(t *testing.T) {
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
		id, column, normal, nightmare, hell string
	}{
		{"zombie4", "ResDm", "20", "50", "50"},
		{"skeleton5", "ResMa", "20", "40", "70"},
		{"skeleton4", "ResFi", "50", "75", "100"},
		{"skeleton5", "ResLi", "60", "80", "120"},
		{"skeleton3", "ResCo", "40", "60", "110"},
		{"skeleton2", "ResPo", "50", "75", "100"},
	}
	for _, vector := range vectors {
		row := rowBy(rows, "Id", vector.id)
		if row == nil {
			t.Fatalf("owned MonStats row %q is missing", vector.id)
		}
		for column, want := range map[string]string{
			vector.column:         vector.normal,
			vector.column + "(N)": vector.nightmare,
			vector.column + "(H)": vector.hell,
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
	for difficulty, wantPoison := range []int{50, 75, 100} {
		t.Run([]string{"Normal", "Nightmare", "Hell"}[difficulty], func(t *testing.T) {
			wantPhysical := []int{0, 0, 33}[difficulty]
			wantMagic := []int{20, 40, 70}[difficulty]
			wantLightning := []int{60, 80, 120}[difficulty]
			fixture := newLuaSuiteFixture(t, content.D2Legacy(), luaSuiteConfig{
				profile: "authority",
				seed:    42,
				initialData: map[string]any{
					"d2legacy.game_rules": map[string]any{"difficulty": difficulty},
				},
				records: fixtureRecords{
					"data/global/excel/monstats.txt":  rows,
					"data/global/excel/monstats2.txt": graphics,
					"data/global/excel/monlvl.txt":    levels,
				},
			})
			err := fixture.authority.Runtime.Run(t.Context(), func(state *lua.LState) error {
				module, err := requireLuaModule(state, "d2legacy.tests.support.monster_resistance_real")
				if err != nil {
					return err
				}
				check := module.RawGetString("check")
				for _, sample := range []struct {
					id     string
					values map[string]int
				}{
					{"skeleton2", map[string]int{
						"poison_resist":      wantPoison,
						"base_poison_resist": wantPoison,
						"physical_resist":    wantPhysical,
					}},
					{"skeleton5", map[string]int{
						"magic_resist":     wantMagic,
						"lightning_resist": wantLightning,
					}},
				} {
					expected := state.NewTable()
					for name, value := range sample.values {
						expected.RawSetString(name, lua.LNumber(value))
					}
					if err := state.CallByParam(lua.P{Fn: check, Protect: true}, lua.LString(sample.id), expected); err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
