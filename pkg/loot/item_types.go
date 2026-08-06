package loot

import (
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const dynamicLevelRange = 3

var dynamicCodePattern = regexp.MustCompile(`^(.*?)(\d+)$`)

// ItemType captures the equivalence edges used by automatic treasure classes.
type ItemType struct {
	Code   string `json:"code"`
	Equiv1 string `json:"equiv1,omitempty"`
	Equiv2 string `json:"equiv2,omitempty"`
	Magic  bool   `json:"magic,omitempty"`
	Rare   bool   `json:"rare,omitempty"`
	Normal bool   `json:"normal,omitempty"`
}

type ItemTypes map[string]ItemType

// ItemIndex precomputes type-hierarchy membership once for repeated gameplay
// event rolls.
type ItemIndex struct {
	catalog    ItemCatalog
	types      ItemTypes
	candidates map[string][]BaseItem
	levels     map[string]map[int][]BaseItem
}

func NewItemIndex(catalog ItemCatalog, types ItemTypes) (*ItemIndex, error) {
	index := &ItemIndex{catalog: catalog, types: types, candidates: make(map[string][]BaseItem, len(types)), levels: make(map[string]map[int][]BaseItem, len(types))}
	for requested := range types {
		matches, err := matchingItems(catalog, types, requested, 0, false)
		if err != nil {
			return nil, fmt.Errorf("loot: index item type %q: %w", requested, err)
		}
		index.candidates[requested] = matches
		byLevel := make(map[int][]BaseItem)
		maximum := 0
		for _, item := range matches {
			if item.Level > maximum {
				maximum = item.Level
			}
		}
		for minimum := 0; minimum <= maximum; minimum++ {
			for _, item := range matches {
				if item.Level >= minimum && item.Level < minimum+dynamicLevelRange {
					byLevel[minimum] = append(byLevel[minimum], item)
				}
			}
		}
		index.levels[requested] = byLevel
	}
	return index, nil
}

// ParseItemTypesTSV reads the ItemTypes.txt hierarchy needed for dynamic codes.
func ParseItemTypesTSV(input io.Reader) (ItemTypes, error) {
	reader := csv.NewReader(input)
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("loot: read item types header: %w", err)
	}
	columns := make(map[string]int, len(header))
	for index, name := range header {
		columns[strings.TrimPrefix(name, "\ufeff")] = index
	}
	if _, ok := columns["Code"]; !ok {
		return nil, fmt.Errorf("loot: item types table missing required column %q", "Code")
	}
	types := make(ItemTypes)
	for rowNumber := 2; ; rowNumber++ {
		row, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("loot: read item types row %d: %w", rowNumber, readErr)
		}
		code := field(row, columns, "Code")
		if code == "" {
			continue
		}
		if _, duplicate := types[code]; duplicate {
			return nil, fmt.Errorf("loot: duplicate item type %q at row %d", code, rowNumber)
		}
		types[code] = ItemType{
			Code: code, Equiv1: field(row, columns, "Equiv1"), Equiv2: field(row, columns, "Equiv2"),
			Magic: booleanField(row, columns, "Magic"), Rare: booleanField(row, columns, "Rare"),
			Normal: booleanField(row, columns, "Normal"),
		}
	}
	return types, nil
}

// ResolveItems resolves direct, generic type, and level-suffixed dynamic codes.
// Selection is reproducible for a given input, item catalog, hierarchy, and seed.
func ResolveItems(drops []Drop, catalog ItemCatalog, types ItemTypes, seed uint64) ([]ResolvedDrop, error) {
	index, err := NewItemIndex(catalog, types)
	if err != nil {
		return nil, err
	}
	return index.Resolve(drops, seed)
}

// Resolve selects direct and dynamic items using pre-indexed type candidates.
func (i *ItemIndex) Resolve(drops []Drop, seed uint64) ([]ResolvedDrop, error) {
	rng := splitMix64(seed)
	result := make([]ResolvedDrop, len(drops))
	for index, drop := range drops {
		result[index].Drop = drop
		if item, ok := i.catalog[drop.Code]; ok {
			result[index].Item = copyItem(item)
			result[index].Resolved = true
			continue
		}

		typeCode, minLevel, dynamic := splitDynamicCode(drop.Code)
		levelQualified := dynamic
		if _, exactType := i.types[drop.Code]; exactType {
			typeCode, dynamic = drop.Code, true
			minLevel = 0
			levelQualified = false
		}
		if !dynamic {
			continue
		}
		if _, knownType := i.types[typeCode]; !knownType {
			continue
		}
		candidates := i.candidates[typeCode]
		if levelQualified {
			candidates = i.levels[typeCode][minLevel]
		}
		if len(candidates) == 0 {
			continue
		}
		selected := candidates[int(rng.next()%uint64(len(candidates)))]
		result[index].Item = copyItem(selected)
		result[index].Resolved = true
	}
	return result, nil
}

func matchingItems(catalog ItemCatalog, types ItemTypes, requested string, minLevel int, levelQualified bool) ([]BaseItem, error) {
	candidates := make([]BaseItem, 0)
	for _, item := range catalog {
		matches, err := itemMatchesType(item, types, requested)
		if err != nil {
			return nil, err
		}
		if !matches || (levelQualified && (item.Level < minLevel || item.Level >= minLevel+dynamicLevelRange)) {
			continue
		}
		candidates = append(candidates, item)
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].Code < candidates[j].Code })
	return candidates, nil
}

func itemMatchesType(item BaseItem, types ItemTypes, requested string) (bool, error) {
	for _, itemType := range []string{item.Type, item.Type2} {
		if itemType == "" {
			continue
		}
		matches, err := typeIncludes(types, itemType, requested, make(map[string]bool))
		if err != nil || matches {
			return matches, err
		}
	}
	return false, nil
}

func typeIncludes(types ItemTypes, current, requested string, active map[string]bool) (bool, error) {
	if current == requested {
		return true, nil
	}
	if active[current] {
		return false, fmt.Errorf("item type cycle at %q", current)
	}
	itemType, ok := types[current]
	if !ok {
		return false, nil
	}
	active[current] = true
	defer delete(active, current)
	for _, parent := range []string{itemType.Equiv1, itemType.Equiv2} {
		if parent == "" {
			continue
		}
		matches, err := typeIncludes(types, parent, requested, active)
		if err != nil || matches {
			return matches, err
		}
	}
	return false, nil
}

func splitDynamicCode(code string) (string, int, bool) {
	parts := dynamicCodePattern.FindStringSubmatch(code)
	if len(parts) != 3 || parts[1] == "" {
		return "", 0, false
	}
	level, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", 0, false
	}
	return parts[1], level, true
}

func copyItem(item BaseItem) *BaseItem {
	result := item
	return &result
}

func booleanField(row []string, columns map[string]int, name string) bool {
	value := strings.ToLower(field(row, columns, name))
	return value == "1" || value == "true" || value == "yes"
}
