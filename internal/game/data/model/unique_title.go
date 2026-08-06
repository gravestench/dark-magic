package models

// UniqueTitle is one ordered title fragment from UniqueTitle.txt. The shipped
// Name column is the repeated sentinel "unused", so neither column is treated
// as a stable record key.
type UniqueTitle struct {
	Name  string `csv:"Name"`
	Namco string `csv:"Namco"`
}
