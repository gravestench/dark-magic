package loot

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// ParseItemRatiosTSV reads itemratio.txt rows used by the quality calculator.
func ParseItemRatiosTSV(input io.Reader) ([]QualityRatio, error) {
	reader := csv.NewReader(input)
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("loot: read item ratio header: %w", err)
	}
	columns := make(map[string]int, len(header))
	for index, name := range header {
		columns[strings.TrimPrefix(name, "\ufeff")] = index
	}
	required := []string{
		"Version", "Uber", "Class Specific",
		"Unique", "UniqueDivisor", "UniqueMin",
		"Set", "SetDivisor", "SetMin",
		"Rare", "RareDivisor", "RareMin",
		"Magic", "MagicDivisor", "MagicMin",
		"HiQuality", "HiQualityDivisor", "Normal", "NormalDivisor",
	}
	for _, name := range required {
		if _, ok := columns[name]; !ok {
			return nil, fmt.Errorf("loot: item ratio table missing required column %q", name)
		}
	}

	var ratios []QualityRatio
	for rowNumber := 2; ; rowNumber++ {
		row, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("loot: read item ratio row %d: %w", rowNumber, readErr)
		}
		if strings.TrimSpace(strings.Join(row, "")) == "" {
			continue
		}
		values := make(map[string]int, len(required))
		for _, name := range required {
			value, parseErr := integerField(row, columns, name, rowNumber)
			if parseErr != nil {
				return nil, parseErr
			}
			values[name] = value
		}
		ratio := QualityRatio{
			Version: values["Version"], Uber: values["Uber"] != 0, ClassSpecific: values["Class Specific"] != 0,
			Unique:   QualityRule{Base: values["Unique"], Divisor: values["UniqueDivisor"], Minimum: values["UniqueMin"]},
			Set:      QualityRule{Base: values["Set"], Divisor: values["SetDivisor"], Minimum: values["SetMin"]},
			Rare:     QualityRule{Base: values["Rare"], Divisor: values["RareDivisor"], Minimum: values["RareMin"]},
			Magic:    QualityRule{Base: values["Magic"], Divisor: values["MagicDivisor"], Minimum: values["MagicMin"]},
			Superior: QualityRule{Base: values["HiQuality"], Divisor: values["HiQualityDivisor"]},
			Normal:   QualityRule{Base: values["Normal"], Divisor: values["NormalDivisor"]},
		}
		ratios = append(ratios, ratio)
	}
	return ratios, nil
}

// SelectQualityRatio returns the exact row for an item's game mode and flags.
func SelectQualityRatio(ratios []QualityRatio, version int, uber, classSpecific bool) (QualityRatio, error) {
	var selected QualityRatio
	found := false
	for _, ratio := range ratios {
		if ratio.Version != version || ratio.Uber != uber || ratio.ClassSpecific != classSpecific {
			continue
		}
		if found {
			return QualityRatio{}, fmt.Errorf("loot: multiple item ratio rows match version=%d uber=%t classSpecific=%t", version, uber, classSpecific)
		}
		selected, found = ratio, true
	}
	if !found {
		return QualityRatio{}, fmt.Errorf("loot: no item ratio matches version=%d uber=%t classSpecific=%t", version, uber, classSpecific)
	}
	return selected, nil
}
