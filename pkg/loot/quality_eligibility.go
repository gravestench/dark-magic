package loot

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// SpecialItem is the availability subset shared by UniqueItems and SetItems.
type SpecialItem struct {
	BaseCode string `json:"baseCode"`
	Level    int    `json:"level"`
	Version  int    `json:"version"`
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

func hasAvailableSpecial(items []SpecialItem, context EligibilityContext) bool {
	for _, item := range items {
		if !item.Enabled || item.Level > context.DropLevel {
			continue
		}
		if context.Version < 100 && item.Version >= 100 {
			continue
		}
		return true
	}
	return false
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
	for _, required := range []string{codeColumn, "lvl"} {
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
		items[code] = append(items[code], SpecialItem{BaseCode: code, Level: level, Version: version, Enabled: enabled})
	}
	return items, nil
}
