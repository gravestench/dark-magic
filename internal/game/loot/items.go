package loot

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ItemKind identifies the base record table containing an item.
type ItemKind string

const (
	ItemWeapon ItemKind = "weapon"
	ItemArmor  ItemKind = "armor"
	ItemMisc   ItemKind = "misc"
)

// BaseItem is the common subset needed after a TreasureClass selects a code.
// More specialized item construction can retain and parse the source record later.
type BaseItem struct {
	Code       string   `json:"code"`
	Kind       ItemKind `json:"kind"`
	NameKey    string   `json:"nameKey,omitempty"`
	Type       string   `json:"type,omitempty"`
	Type2      string   `json:"type2,omitempty"`
	Level      int      `json:"level,omitempty"`
	LevelReq   int      `json:"levelRequirement,omitempty"`
	MagicLevel int      `json:"magicLevel,omitempty"`
	InvFile    string   `json:"inventoryFile,omitempty"`
	FlippyFile string   `json:"flippyFile,omitempty"`
}

// ItemCatalog indexes weapons.txt, armor.txt, and misc.txt records by item code.
type ItemCatalog map[string]BaseItem

// ResolvedDrop joins a treasure roll to its base item record. Unresolved codes
// remain explicit because TreasureClass also permits dynamic type and gold codes.
type ResolvedDrop struct {
	Drop
	Item     *BaseItem `json:"item,omitempty"`
	Resolved bool      `json:"resolved"`
}

// ParseBaseItemsTSV reads the common columns from one base item table.
func ParseBaseItemsTSV(input io.Reader, kind ItemKind) (ItemCatalog, error) {
	if kind != ItemWeapon && kind != ItemArmor && kind != ItemMisc {
		return nil, fmt.Errorf("loot: unsupported item kind %q", kind)
	}
	reader := csv.NewReader(input)
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("loot: read %s header: %w", kind, err)
	}
	columns := make(map[string]int, len(header))
	for index, name := range header {
		columns[strings.TrimPrefix(name, "\ufeff")] = index
	}
	if _, ok := columns["code"]; !ok {
		return nil, fmt.Errorf("loot: %s table missing required column %q", kind, "code")
	}

	items := make(ItemCatalog)
	for rowNumber := 2; ; rowNumber++ {
		row, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("loot: read %s row %d: %w", kind, rowNumber, readErr)
		}
		code := field(row, columns, "code")
		if code == "" {
			continue
		}
		if _, duplicate := items[code]; duplicate {
			return nil, fmt.Errorf("loot: duplicate %s item code %q at row %d", kind, code, rowNumber)
		}
		level, err := optionalItemInteger(row, columns, "level", rowNumber, kind)
		if err != nil {
			return nil, err
		}
		levelReq, err := optionalItemInteger(row, columns, "levelreq", rowNumber, kind)
		if err != nil {
			return nil, err
		}
		magicLevel, err := optionalItemInteger(row, columns, "magic lvl", rowNumber, kind)
		if err != nil {
			return nil, err
		}
		items[code] = BaseItem{
			Code: code, Kind: kind, NameKey: field(row, columns, "namestr"),
			Type: field(row, columns, "type"), Type2: field(row, columns, "type2"), Level: level, LevelReq: levelReq, MagicLevel: magicLevel,
			InvFile: field(row, columns, "invfile"), FlippyFile: field(row, columns, "flippyfile"),
		}
	}
	return items, nil
}

// MergeItemCatalogs combines base tables and rejects ambiguous codes.
func MergeItemCatalogs(catalogs ...ItemCatalog) (ItemCatalog, error) {
	merged := make(ItemCatalog)
	for _, catalog := range catalogs {
		for code, item := range catalog {
			if existing, duplicate := merged[code]; duplicate {
				return nil, fmt.Errorf("loot: item code %q exists in both %s and %s tables", code, existing.Kind, item.Kind)
			}
			merged[code] = item
		}
	}
	return merged, nil
}

// ResolveBaseItems associates direct terminal codes with base item records.
func ResolveBaseItems(drops []Drop, catalog ItemCatalog) []ResolvedDrop {
	resolved := make([]ResolvedDrop, len(drops))
	for index, drop := range drops {
		resolved[index].Drop = drop
		if item, ok := catalog[drop.Code]; ok {
			itemCopy := item
			resolved[index].Item = &itemCopy
			resolved[index].Resolved = true
		}
	}
	return resolved
}

func optionalItemInteger(row []string, columns map[string]int, name string, rowNumber int, kind ItemKind) (int, error) {
	value := field(row, columns, name)
	if value == "" {
		return 0, nil
	}
	result, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("loot: %s row %d column %q: expected integer, got %q", kind, rowNumber, name, value)
	}
	return result, nil
}
