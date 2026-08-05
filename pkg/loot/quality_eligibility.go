package loot

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// SpecialItem is the availability subset shared by UniqueItems and SetItems.
type SpecialItem struct {
	Name     string `json:"name"`
	BaseCode string `json:"baseCode"`
	Level    int    `json:"level"`
	Version  int    `json:"version"`
	Rarity   int    `json:"rarity"`
	Enabled  bool   `json:"enabled"`
}

// SpecialItems indexes possible unique or set variants by their base item code.
type SpecialItems map[string][]SpecialItem

// EligibilityContext controls special-item availability for a quality fallback.
type EligibilityContext struct {
	Version   int `json:"version"`
	DropLevel int `json:"dropLevel"`
}

// ParseUniqueItemsTSV reads availability from UniqueItems.txt.
func ParseUniqueItemsTSV(input io.Reader) (SpecialItems, error) {
	return parseSpecialItemsTSV(input, "unique", "code", true)
}

// ParseSetItemsTSV reads availability from SetItems.txt.
func ParseSetItemsTSV(input io.Reader) (SpecialItems, error) {
	return parseSpecialItemsTSV(input, "set", "item", false)
}

// ApplyQualityEligibility sanitizes a rolled quality for the selected base item.
func ApplyQualityEligibility(rolled Quality, item BaseItem, types ItemTypes, uniques, sets SpecialItems, context EligibilityContext) (Quality, error) {
	itemType, ok := types[item.Type]
	if !ok {
		return "", fmt.Errorf("loot: item %q references unknown type %q", item.Code, item.Type)
	}
	if itemType.Normal {
		return QualityNormal, nil
	}
	if itemType.Magic {
		return QualityMagic, nil
	}

	quality := rolled
	if quality == QualityUnique && !hasAvailableSpecial(uniques[item.Code], context) {
		quality = QualityRare
	}
	if quality == QualitySet && !hasAvailableSpecial(sets[item.Code], context) {
		quality = QualityRare
	}
	if quality == QualityRare && !itemType.Rare {
		quality = QualityMagic
	}
	return quality, nil
}

// SelectSpecialItem chooses an eligible UniqueItems or SetItems record by its
// rarity weight. Sorting by stable record fields prevents map/input order drift.
func SelectSpecialItem(items []SpecialItem, context EligibilityContext, seed uint64) (SpecialItem, error) {
	candidates := make([]SpecialItem, 0, len(items))
	total := 0
	for _, item := range items {
		if !specialAvailable(item, context) || item.Rarity <= 0 {
			continue
		}
		candidates = append(candidates, item)
		total += item.Rarity
	}
	if len(candidates) == 0 {
		return SpecialItem{}, fmt.Errorf("loot: no eligible weighted special item at level %d for version %d", context.DropLevel, context.Version)
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Name == candidates[j].Name {
			return candidates[i].BaseCode < candidates[j].BaseCode
		}
		return candidates[i].Name < candidates[j].Name
	})
	rng := splitMix64(seed)
	roll := int(rng.next() % uint64(total))
	for _, candidate := range candidates {
		if roll < candidate.Rarity {
			return candidate, nil
		}
		roll -= candidate.Rarity
	}
	panic("unreachable special-item selection")
}

func hasAvailableSpecial(items []SpecialItem, context EligibilityContext) bool {
	for _, item := range items {
		if specialAvailable(item, context) && item.Rarity > 0 {
			return true
		}
	}
	return false
}

func specialAvailable(item SpecialItem, context EligibilityContext) bool {
	return item.Enabled && item.Level <= context.DropLevel && !(context.Version < 100 && item.Version >= 100)
}

func parseSpecialItemsTSV(input io.Reader, label, codeColumn string, hasEnabled bool) (SpecialItems, error) {
	reader := csv.NewReader(input)
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("loot: read %s items header: %w", label, err)
	}
	columns := make(map[string]int, len(header))
	for index, name := range header {
		columns[strings.TrimPrefix(name, "\ufeff")] = index
	}
	for _, required := range []string{"index", codeColumn, "lvl", "rarity"} {
		if _, ok := columns[required]; !ok {
			return nil, fmt.Errorf("loot: %s items table missing required column %q", label, required)
		}
	}
	if hasEnabled {
		for _, required := range []string{"version", "enabled"} {
			if _, ok := columns[required]; !ok {
				return nil, fmt.Errorf("loot: %s items table missing required column %q", label, required)
			}
		}
	}
	items := make(SpecialItems)
	for rowNumber := 2; ; rowNumber++ {
		row, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("loot: read %s items row %d: %w", label, rowNumber, readErr)
		}
		code := field(row, columns, codeColumn)
		if code == "" {
			continue
		}
		level, err := integerField(row, columns, "lvl", rowNumber)
		if err != nil {
			return nil, err
		}
		version, err := integerField(row, columns, "version", rowNumber)
		if err != nil {
			return nil, err
		}
		enabled := true
		if hasEnabled {
			enabled = booleanField(row, columns, "enabled")
		}
		rarityValue := field(row, columns, "rarity")
		rarity, err := strconv.Atoi(rarityValue)
		if err != nil {
			return nil, fmt.Errorf("loot: %s items row %d column %q: expected integer, got %q", label, rowNumber, "rarity", rarityValue)
		}
		if rarity < 0 {
			return nil, fmt.Errorf("loot: %s items row %d column %q: must not be negative", label, rowNumber, "rarity")
		}
		items[code] = append(items[code], SpecialItem{
			Name: field(row, columns, "index"), BaseCode: code, Level: level,
			Version: version, Rarity: rarity, Enabled: enabled,
		})
	}
	return items, nil
}
