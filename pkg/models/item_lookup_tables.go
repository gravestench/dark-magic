package models

// LowQualityItemName is one authored name option for low-quality items.
type LowQualityItemName struct {
	Name string `csv:"Name"`
}

// BodyLocation maps a display label to an equipment-location code.
type BodyLocation struct {
	Name string           `csv:"Body Location"`
	Code ItemBodyLocation `csv:"Code"`
}

// StorePage maps a display label to an NPC shop page code.
type StorePage struct {
	Name string        `csv:"Store Page"`
	Code ItemStorePage `csv:"Code"`
}

// CompositeComponent maps an equipment animation component to its COF token.
type CompositeComponent struct {
	Name  string `csv:"Name"`
	Token string `csv:"Token"`
}

// HitClass maps a weapon hit-class label to its impact-sound code.
type HitClass struct {
	Name string `csv:"Hit Class"`
	Code string `csv:"Code"`
}
