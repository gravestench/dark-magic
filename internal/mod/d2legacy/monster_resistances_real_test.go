package d2legacy

import (
	"fmt"
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
				return state.DoString(fmt.Sprintf(`
                    local definition = require("d2legacy.data.monster").load("skeleton2")
                    local command = { tick = 1, payload = {
                        spawn_id = "skeleton2", seed = 1, x = 0, y = 0, act = 1, level_id = 1,
                        definition = definition,
                    } }
                    local components = require("d2legacy.commands.spawn_monster").components(command, 1)
                    local defense = components["d2legacy.combat.defense"]
                    assert(defense.poison_resist == %d, "spawned poison resistance differs from owned MonStats")
                    assert(defense.base_poison_resist == %d, "spawned base poison resistance differs from owned MonStats")
                    assert(defense.physical_resist == %d, "spawned physical resistance differs from owned MonStats")

                    command.payload.spawn_id = "skeleton5"
                    command.payload.definition = require("d2legacy.data.monster").load("skeleton5")
                    local second = require("d2legacy.commands.spawn_monster").components(command, 1)["d2legacy.combat.defense"]
                    assert(second.magic_resist == %d, "spawned magic resistance differs from owned MonStats")
                    assert(second.lightning_resist == %d, "spawned lightning resistance differs from owned MonStats")
                `, wantPoison, wantPoison, wantPhysical, wantMagic, wantLightning))
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
