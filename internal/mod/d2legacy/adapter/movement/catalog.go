package movement

import (
	"fmt"
	"strconv"
	"strings"
)

const charStatsPath = "data/global/excel/charstats.txt"

type recordsGateway interface {
	Load(string) ([]map[string]string, error)
}

// ClassRates are authoritative world-subtile units per second. CharStats.txt
// names these fields as velocities; Dark Magic's world coordinate and fixed-
// tick integrator use the same unit directly.
type ClassRates struct {
	Walk float64
	Run  float64
}

// Catalog is one immutable, case-insensitive projection of pinned CharStats.
// Authority and client prediction receive copies built from the same runtime-
// identified record generation.
type Catalog struct {
	byClass map[string]ClassRates
}

func LoadCatalog(records recordsGateway) (Catalog, error) {
	if records == nil {
		return Catalog{}, fmt.Errorf("d2legacy movement: records are required")
	}
	rows, err := records.Load(charStatsPath)
	if err != nil {
		return Catalog{}, fmt.Errorf("d2legacy movement: load %s: %w", charStatsPath, err)
	}
	result := Catalog{byClass: make(map[string]ClassRates)}
	for index, row := range rows {
		class := strings.TrimSpace(row["class"])
		if class == "" || strings.EqualFold(class, "Expansion") {
			continue
		}
		walk, walkErr := positiveVelocity(row["WalkVelocity"])
		run, runErr := positiveVelocity(row["RunVelocity"])
		if walkErr != nil || runErr != nil || run < walk {
			return Catalog{}, fmt.Errorf("d2legacy movement: %s row %d class %q has invalid walk/run velocity", charStatsPath, index+2, class)
		}
		key := strings.ToLower(class)
		if _, duplicate := result.byClass[key]; duplicate {
			return Catalog{}, fmt.Errorf("d2legacy movement: %s contains duplicate class %q", charStatsPath, class)
		}
		result.byClass[key] = ClassRates{Walk: walk, Run: run}
	}
	if len(result.byClass) == 0 {
		return Catalog{}, fmt.Errorf("d2legacy movement: %s contains no playable class velocities", charStatsPath)
	}
	return result, nil
}

func positiveVelocity(raw string) (float64, error) {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("invalid velocity %q", raw)
	}
	return value, nil
}

func (catalog Catalog) Rates(class string) (ClassRates, bool) {
	rates, found := catalog.byClass[strings.ToLower(strings.TrimSpace(class))]
	return rates, found
}
