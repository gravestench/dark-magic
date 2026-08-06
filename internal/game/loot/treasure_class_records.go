package loot

import (
	"fmt"
	"math"

	"github.com/gravestench/dark-magic/internal/game/data/model"
)

// TreasureClassRecords is the narrow typed game-data view required for rolls.
type TreasureClassRecords interface {
	TreasureClassRecords() ([]models.TreasureClassEx, error)
}

// CatalogFromRecords converts admitted game-data records into the smaller
// deterministic simulation model owned by this package.
func CatalogFromRecords(source TreasureClassRecords) (Catalog, error) {
	records, err := source.TreasureClassRecords()
	if err != nil {
		return nil, fmt.Errorf("loot: load treasure classes: %w", err)
	}
	catalog := make(Catalog, len(records))
	for index, record := range records {
		if record.TreasureClass == "" {
			continue
		}
		if _, exists := catalog[record.TreasureClass]; exists {
			return nil, fmt.Errorf("loot: duplicate treasure class %q at record %d", record.TreasureClass, index+1)
		}
		values := []float64{record.NoDrop, record.Unique, record.Set, record.Rare, record.Magic}
		for _, value := range values {
			if math.Trunc(value) != value {
				return nil, fmt.Errorf("loot: treasure class %q contains non-integral roll value %v", record.TreasureClass, value)
			}
		}
		class := Class{
			Name: record.TreasureClass, Picks: record.Picks, NoDrop: int(record.NoDrop),
			Quality: QualityModifiers{Unique: int(record.Unique), Set: int(record.Set), Rare: int(record.Rare), Magic: int(record.Magic)},
		}
		items := [...]string{record.Item1, record.Item2, record.Item3, record.Item4, record.Item5, record.Item6, record.Item7, record.Item8, record.Item9, record.Item10}
		weights := [...]float64{record.Prob1, record.Prob2, record.Prob3, record.Prob4, record.Prob5, record.Prob6, record.Prob7, record.Prob8, record.Prob9, record.Prob10}
		for entryIndex, code := range items {
			if code == "" {
				continue
			}
			if math.Trunc(weights[entryIndex]) != weights[entryIndex] {
				return nil, fmt.Errorf("loot: treasure class %q item %d contains non-integral weight %v", record.TreasureClass, entryIndex+1, weights[entryIndex])
			}
			class.Entries = append(class.Entries, Entry{Code: code, Weight: int(weights[entryIndex])})
		}
		catalog[class.Name] = class
	}
	return catalog, nil
}
