package modruntime

import (
	"context"
	"testing"

	"github.com/gravestench/dark-magic/internal/game/data/catalog"
	"github.com/gravestench/dark-magic/internal/game/data/model"
	lua "github.com/yuin/gopher-lua"
)

type staticGameData struct {
	snapshot gamedata.Snapshot
	err      error
}

func (data staticGameData) Snapshot() (gamedata.Snapshot, error) { return data.snapshot, data.err }

func TestGameDataModuleExposesTypedCopies(t *testing.T) {
	runtime := New()
	data := staticGameData{snapshot: gamedata.Snapshot{
		CharStatsByClass: map[string]models.CharStats{
			"Amazon": {Class: "Amazon", Strength: 20, Dexterity: 25, Intelligence: 15, Vitality: 20, Stamina: 84},
		},
		UniqueTitles: []models.UniqueTitle{{Name: "unused", Namco: "Judge"}, {Name: "unused", Namco: "Countess"}},
	}}
	if err := runtime.RegisterModule(GameDataModule(data)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())

	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		return state.DoString(`
			local data = require("dm.game_data/v1")
			local amazon, err = data.character_class("amazon")
			assert(err == nil and amazon.class == "Amazon")
			assert(amazon.strength == 20 and amazon.dexterity == 25)
			assert(amazon.intelligence == 15 and amazon.vitality == 20 and amazon.stamina == 84)
			local titles = data.unique_titles()
			assert(#titles == 2 and titles[1].title == "Judge" and titles[2].title == "Countess")
			titles[1].title = "changed"
			assert(data.unique_titles()[1].title == "Judge")
			local missing, message = data.character_class("missing")
			assert(missing == nil and message == "unknown character class: missing")
		`)
	}); err != nil {
		t.Fatal(err)
	}
}
