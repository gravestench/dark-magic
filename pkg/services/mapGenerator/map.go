package mapGenerator

import (
	"github.com/gravestench/ds1"
	"github.com/gravestench/dt1"
	lua "github.com/yuin/gopher-lua"

	"github.com/gravestench/dark-magic/pkg/models"
)

type act = uint

func NewWorldMap() *WorldMap {
	m := &WorldMap{
		Levels: make(map[act][]Level),
	}

	return m
}

type WorldMap struct {
	Levels map[act][]Level
}

func (wm *WorldMap) LuaTable(state *lua.LState) *lua.LTable {
	return nil
}

type Level struct {
	Name             string
	Data             models.LevelData
	Group            models.LevelGroup
	MazeData         models.LevelMazeData
	Preset           models.LevelPreset
	SubstitutionData models.LevelSubstitutionData
	Type             models.LevelType
	Warp             models.LevelWarp

	TileSets   []dt1.DT1
	TileStamps []ds1.DS1

	Act int
}
