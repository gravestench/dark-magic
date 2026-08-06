package loot

import (
	"errors"
	"fmt"
)

const qualityRollScale = 128

// Quality is an item's rolled quality, ordered from rarest to fallback.
type Quality string

const (
	QualityUnique   Quality = "unique"
	QualitySet      Quality = "set"
	QualityRare     Quality = "rare"
	QualityMagic    Quality = "magic"
	QualitySuperior Quality = "superior"
	QualityNormal   Quality = "normal"
	QualityLow      Quality = "low"
)

// QualityRule is one column family from ItemRatio.txt.
type QualityRule struct {
	Base    int `json:"base"`
	Divisor int `json:"divisor"`
	Minimum int `json:"minimum,omitempty"`
}

// QualityRatio contains one applicable ItemRatio row.
type QualityRatio struct {
	Version       int         `json:"version"`
	Uber          bool        `json:"uber"`
	ClassSpecific bool        `json:"classSpecific"`
	Unique        QualityRule `json:"unique"`
	Set           QualityRule `json:"set"`
	Rare          QualityRule `json:"rare"`
	Magic         QualityRule `json:"magic"`
	Superior      QualityRule `json:"superior"`
	Normal        QualityRule `json:"normal"`
}

// QualityContext supplies event-specific values to the ItemRatio calculation.
// MagicFind is the player's bonus percentage (zero means no bonus).
type QualityContext struct {
	MonsterLevel int              `json:"monsterLevel"`
	ItemLevel    int              `json:"itemLevel"`
	MagicFind    int              `json:"magicFind"`
	Modifiers    QualityModifiers `json:"treasureClassModifiers"`
}

// QualityChances exposes the final 128ths-based roll denominators for diagnostics.
type QualityChances struct {
	Unique   int `json:"unique"`
	Set      int `json:"set"`
	Rare     int `json:"rare"`
	Magic    int `json:"magic"`
	Superior int `json:"superior"`
	Normal   int `json:"normal"`
}

// CalculateQualityChances applies ItemRatio level, magic-find, minimum, and
// TreasureClass adjustments. A successful check rolls below 128.
func CalculateQualityChances(ratio QualityRatio, context QualityContext) (QualityChances, error) {
	if context.MonsterLevel < 0 || context.ItemLevel < 0 || context.MagicFind < 0 {
		return QualityChances{}, errors.New("loot: quality levels and magic find must not be negative")
	}
	modifiers := context.Modifiers
	if modifiers.Unique < 0 || modifiers.Unique > 1024 ||
		modifiers.Set < 0 || modifiers.Set > 1024 ||
		modifiers.Rare < 0 || modifiers.Rare > 1024 ||
		modifiers.Magic < 0 || modifiers.Magic > 1024 {
		return QualityChances{}, errors.New("loot: quality modifiers must be between 0 and 1024")
	}

	unique, err := qualityDenominator(ratio.Unique, context, modifiers.Unique, 250, true)
	if err != nil {
		return QualityChances{}, fmt.Errorf("loot: unique quality: %w", err)
	}
	set, err := qualityDenominator(ratio.Set, context, modifiers.Set, 500, true)
	if err != nil {
		return QualityChances{}, fmt.Errorf("loot: set quality: %w", err)
	}
	rare, err := qualityDenominator(ratio.Rare, context, modifiers.Rare, 600, true)
	if err != nil {
		return QualityChances{}, fmt.Errorf("loot: rare quality: %w", err)
	}
	magic, err := qualityDenominator(ratio.Magic, context, modifiers.Magic, 0, true)
	if err != nil {
		return QualityChances{}, fmt.Errorf("loot: magic quality: %w", err)
	}
	superior, err := qualityDenominator(ratio.Superior, context, 0, 0, false)
	if err != nil {
		return QualityChances{}, fmt.Errorf("loot: superior quality: %w", err)
	}
	normal, err := qualityDenominator(ratio.Normal, context, 0, 0, false)
	if err != nil {
		return QualityChances{}, fmt.Errorf("loot: normal quality: %w", err)
	}
	return QualityChances{Unique: unique, Set: set, Rare: rare, Magic: magic, Superior: superior, Normal: normal}, nil
}

// RollQuality performs the ordered Unique-to-Low checks with stable random state.
func RollQuality(ratio QualityRatio, context QualityContext, seed uint64) (Quality, QualityChances, error) {
	chances, err := CalculateQualityChances(ratio, context)
	if err != nil {
		return "", QualityChances{}, err
	}
	rng := splitMix64(seed)
	checks := []struct {
		quality     Quality
		denominator int
	}{
		{QualityUnique, chances.Unique},
		{QualitySet, chances.Set},
		{QualityRare, chances.Rare},
		{QualityMagic, chances.Magic},
		{QualitySuperior, chances.Superior},
		{QualityNormal, chances.Normal},
	}
	for _, check := range checks {
		if int(rng.next()%uint64(check.denominator)) < qualityRollScale {
			return check.quality, chances, nil
		}
	}
	return QualityLow, chances, nil
}

func qualityDenominator(rule QualityRule, context QualityContext, modifier, diminishing int, useMagicFind bool) (int, error) {
	if rule.Base <= 0 {
		return 0, errors.New("base must be positive")
	}
	if rule.Divisor < 0 || rule.Minimum < 0 {
		return 0, errors.New("divisor and minimum must not be negative")
	}
	adjustment := 0
	if rule.Divisor > 0 {
		adjustment = (context.MonsterLevel - context.ItemLevel) / rule.Divisor
	}
	denominator := (rule.Base - adjustment) * qualityRollScale
	if denominator < qualityRollScale {
		denominator = qualityRollScale
	}
	if useMagicFind && context.MagicFind > 0 {
		effectiveBonus := context.MagicFind
		if diminishing > 0 {
			effectiveBonus = context.MagicFind * diminishing / (context.MagicFind + diminishing)
		}
		denominator = denominator * 100 / (100 + effectiveBonus)
	}
	if rule.Minimum > 0 && denominator < rule.Minimum {
		denominator = rule.Minimum
	}
	denominator -= denominator * modifier / 1024
	if denominator < qualityRollScale {
		denominator = qualityRollScale
	}
	return denominator, nil
}
