package loot

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const treasureEntryCount = 10

// ParseTreasureClassTSV converts TreasureClass.txt or TreasureClassEx.txt data
// into a roll catalog. Unused quality and ladder columns are safely ignored.
func ParseTreasureClassTSV(input io.Reader) (Catalog, error) {
	reader := csv.NewReader(input)
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1
	reader.ReuseRecord = true

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("loot: read treasure-class header: %w", err)
	}
	columns := make(map[string]int, len(header))
	for index, name := range header {
		columns[strings.TrimPrefix(name, "\ufeff")] = index
	}
	for _, required := range []string{"Treasure Class", "Picks", "NoDrop"} {
		if _, ok := columns[required]; !ok {
			return nil, fmt.Errorf("loot: missing required column %q", required)
		}
	}

	catalog := make(Catalog)
	for rowNumber := 2; ; rowNumber++ {
		row, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("loot: read treasure-class row %d: %w", rowNumber, readErr)
		}
		name := field(row, columns, "Treasure Class")
		if name == "" {
			continue
		}
		if _, duplicate := catalog[name]; duplicate {
			return nil, fmt.Errorf("loot: duplicate treasure class %q at row %d", name, rowNumber)
		}
		picks, err := integerField(row, columns, "Picks", rowNumber)
		if err != nil {
			return nil, err
		}
		noDrop, err := integerField(row, columns, "NoDrop", rowNumber)
		if err != nil {
			return nil, err
		}
		quality := QualityModifiers{}
		qualityFields := []struct {
			name   string
			target *int
		}{
			{name: "Unique", target: &quality.Unique},
			{name: "Set", target: &quality.Set},
			{name: "Rare", target: &quality.Rare},
			{name: "Magic", target: &quality.Magic},
		}
		for _, qualityField := range qualityFields {
			*qualityField.target, err = integerField(row, columns, qualityField.name, rowNumber)
			if err != nil {
				return nil, err
			}
		}
		class := Class{Name: name, Picks: picks, NoDrop: noDrop, Quality: quality}
		for index := 1; index <= treasureEntryCount; index++ {
			code := field(row, columns, fmt.Sprintf("Item%d", index))
			if code == "" {
				continue
			}
			weight, err := integerField(row, columns, fmt.Sprintf("Prob%d", index), rowNumber)
			if err != nil {
				return nil, err
			}
			class.Entries = append(class.Entries, Entry{Code: code, Weight: weight})
		}
		catalog[name] = class
	}

	return catalog, nil
}

func field(row []string, columns map[string]int, name string) string {
	index, exists := columns[name]
	if !exists || index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

func integerField(row []string, columns map[string]int, name string, rowNumber int) (int, error) {
	value := field(row, columns, name)
	if value == "" {
		return 0, nil
	}
	result, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("loot: row %d column %q: expected integer, got %q", rowNumber, name, value)
	}
	return result, nil
}
