package models

type MonsterUniqueName = MonsterUniqueAppellation

// MonsterUniqueAppellation represents the data for unique monster name suffixes
type MonsterUniqueAppellation struct {
	// Name is a string-table key considered when generating a unique monster's name.
	Name string `csv:"Name"`
}
