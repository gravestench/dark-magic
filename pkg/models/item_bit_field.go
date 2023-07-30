package models

// FieldBitField represents the bit fields for controlling different flags that can affect the item.
type FieldBitField int

const (
	FieldBitFieldMagicQuality FieldBitField = 1 << iota
	FieldBitFieldMetal
	FieldBitFieldSpellcaster
	FieldBitFieldSkillBased
)
