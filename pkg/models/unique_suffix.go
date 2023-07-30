package models

// UniqueSuffix represents the data for unique monster name suffixes
type UniqueSuffix struct {
	Name string `csv:"name"` // A string key, which is used as a potential selection for generating a unique monster’s name
}
