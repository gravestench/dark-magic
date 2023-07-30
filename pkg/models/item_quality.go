package models

// Quality represents the possible item qualities.
type Quality int

const (
	QualityNormal Quality = iota
	QualityExceptional
	QualityElite
	QualitySocketed
	QualityEthereal
	QualityRuneword
)

// ItemQualityLevel represents the item quality level
type ItemQualityLevel int

const (
	// AnyQuality represents any quality level (Used for a random quality)
	AnyQuality ItemQualityLevel = iota
	// LowQuality represents Low Quality level (Ex: "Crude")
	LowQuality
	// NormalQuality represents Normal Quality level (Default value if the value is empty)
	NormalQuality
	// HighQuality represents High Quality level (Superior)
	HighQuality
	// MagicQuality represents Magic Quality level (Uses Magic Prefixes and Suffixes)
	MagicQuality
	// SetItem represents Set Item quality level
	SetItem
	// RareQuality represents Rare Quality level
	RareQuality
	// UniqueQuality represents Unique Quality level (Predetermined stats)
	UniqueQuality
	// CraftedQuality represents Crafted Quality level
	CraftedQuality
	// TemperedQuality represents Tempered Quality level
	TemperedQuality
)
