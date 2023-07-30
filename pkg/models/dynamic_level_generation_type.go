package models

type DynamicRandomLevelGenerationType int
type DrlgType = DynamicRandomLevelGenerationType

const (
	None DrlgType = iota
	Maze
	Preset
	Outdoor
)
