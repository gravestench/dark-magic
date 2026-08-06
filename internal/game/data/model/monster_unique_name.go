package models

type MonsterUniqueName = MonsterUniqueAppellation

// MonsterUniqueAppellation represents the data for unique monster name suffixes
type MonsterUniqueAppellation struct {
	Name string `csv:"Name"` // A string key, which is used as a potential selection for generating a unique monster’s name
}
