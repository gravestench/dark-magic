package models

// UniquePrefix represents the data for unique monster name prefixes
type UniquePrefix struct {
	Name string `csv:"name"` // A string key, which is used as a potential selection for generating a unique monster’s name
}
