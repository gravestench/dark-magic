package loot

import "fmt"

// MaterializeItem assembles a portable item instance with concrete modifier values.
func MaterializeItem(base BaseItem, quality Quality, special *SpecialItem, prefixes, suffixes []Affix, seed uint64) (GeneratedItem, error) {
	if base.Code == "" {
		return GeneratedItem{}, fmt.Errorf("loot: cannot materialize an item with an empty base code")
	}
	if !knownQuality(quality) {
		return GeneratedItem{}, fmt.Errorf("loot: unsupported item quality %q", quality)
	}
	if quality == QualityUnique || quality == QualitySet {
		if special == nil {
			return GeneratedItem{}, fmt.Errorf("loot: %s item %q requires a concrete special record", quality, base.Code)
		}
		if special.BaseCode != base.Code {
			return GeneratedItem{}, fmt.Errorf("loot: special item %q targets %q, not %q", special.Name, special.BaseCode, base.Code)
		}
	} else if special != nil {
		return GeneratedItem{}, fmt.Errorf("loot: %s item %q must not have special record %q", quality, base.Code, special.Name)
	}
	if quality != QualityMagic && quality != QualityRare && (len(prefixes) > 0 || len(suffixes) > 0) {
		return GeneratedItem{}, fmt.Errorf("loot: %s item %q must not have magic affixes", quality, base.Code)
	}

	rng := splitMix64(seed)
	rolledPrefixes, prefixReq, err := rollAffixValues(prefixes, &rng)
	if err != nil {
		return GeneratedItem{}, err
	}
	rolledSuffixes, suffixReq, err := rollAffixValues(suffixes, &rng)
	if err != nil {
		return GeneratedItem{}, err
	}
	levelReq := max(base.LevelReq, prefixReq, suffixReq)
	if special != nil {
		levelReq = max(levelReq, special.LevelReq)
	}
	item := GeneratedItem{
		Base: base, Quality: quality,
		LevelRequirement: levelReq,
		Prefixes:         rolledPrefixes, Suffixes: rolledSuffixes,
	}
	if special != nil {
		specialCopy := *special
		item.Special = &specialCopy
	}
	return item, nil
}

func knownQuality(quality Quality) bool {
	switch quality {
	case QualityUnique, QualitySet, QualityRare, QualityMagic, QualitySuperior, QualityNormal, QualityLow:
		return true
	default:
		return false
	}
}

func rollAffixValues(affixes []Affix, rng *splitMix64) ([]RolledAffix, int, error) {
	rolled := make([]RolledAffix, 0, len(affixes))
	levelReq := 0
	for _, affix := range affixes {
		if affix.Name == "" {
			return nil, 0, fmt.Errorf("loot: cannot materialize an unnamed %s", affix.Kind)
		}
		levelReq = max(levelReq, affix.LevelReq)
		result := RolledAffix{Name: affix.Name, Kind: affix.Kind, Group: affix.Group}
		for _, modifier := range affix.Modifiers {
			if modifier.Code == "" {
				return nil, 0, fmt.Errorf("loot: affix %q has an empty modifier code", affix.Name)
			}
			if modifier.Minimum > modifier.Maximum {
				return nil, 0, fmt.Errorf("loot: affix %q modifier %q has minimum %d above maximum %d", affix.Name, modifier.Code, modifier.Minimum, modifier.Maximum)
			}
			width := uint64(modifier.Maximum-modifier.Minimum) + 1
			value := modifier.Minimum + int(rng.next()%width)
			result.Modifiers = append(result.Modifiers, RolledModifier{Code: modifier.Code, Parameter: modifier.Parameter, Value: value})
		}
		rolled = append(rolled, result)
	}
	return rolled, levelReq, nil
}
