package models

// ObjectType represents the data from the "object types.txt" file.
type ObjectType struct {
	// Name is the name of the object type.
	Name string `csv:"Name"`

	// Token is the token associated with the object type.
	Token string `csv:"Token"`

	// Beta is a flag indicating whether the object type is related to the beta version.
	Beta int `csv:"Beta"`
}
