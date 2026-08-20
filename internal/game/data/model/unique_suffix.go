package models

// UniqueSuffix represents the data for unique monster name suffixes
type UniqueSuffix struct {
	// Name is a string-table key considered when generating a unique monster's name.
	Name string `csv:"name"`
}
