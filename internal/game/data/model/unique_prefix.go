package models

// UniquePrefix represents the data for unique monster name prefixes
type UniquePrefix struct {
	// Name is a string-table key considered when generating a unique monster's name.
	Name string `csv:"name"`
}
