package mapGenerator

import (
	"github.com/gravestench/ds1"
	"github.com/gravestench/dt1"

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
	LevelData             models.LevelData
	LevelGroup            models.LevelGroup
	LevelName             models.LevelName
	LevelMazeData         models.LevelMazeData
	LevelPreset           models.LevelPreset
	LevelSubstitutionData models.LevelSubstitutionData
	LevelType             models.LevelType
	LevelWarp             models.LevelWarp

	Levels map[act][]Level
}

type Level struct {
	Name string

	tileSets   []dt1.DT1
	tileStamps []ds1.DS1

	Act int
}
