package models

// Transparency represents the types of transparency for items.
type Transparency int

const (
	Transparency25 Transparency = iota
	Transparency50
	Transparency75
	TransparencyBlackAlpha
	TransparencyWhiteAlpha
	TransparencyNo
	TransparencyDark
	TransparencyHighlight
	TransparencyBlended
)
