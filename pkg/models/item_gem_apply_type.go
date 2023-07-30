package models

// GemApplyType represents the types of effects from a gem or rune when socketed into an item.
type GemApplyType int

const (
	GemApplyTypeWeapon GemApplyType = iota
	GemApplyTypeArmorHelmet
	GemApplyTypeShield
)
