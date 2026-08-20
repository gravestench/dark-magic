package movement

import (
	"fmt"
	"strconv"
	"strings"
)

const charStatsPath = "data/global/excel/charstats.txt"

// recordsGateway supplies the generation-pinned rows needed to build authoritative movement facts.
type recordsGateway interface {
	Load(string) ([]map[string]string, error)
}

// ClassRates are authoritative world-subtile units per second. CharStats.txt
// names these fields as velocities; Dark Magic's world coordinate and fixed-
// tick integrator use the same unit directly.
type ClassRates struct {
	Walk               float64
	Run                float64
	StartingVitality   int64
	StartingStamina    int64
	RunDrain           int64
	StaminaPerLevel    int64
	StaminaPerVitality int64
}

// Catalog is one immutable, case-insensitive projection of pinned CharStats.
// Authority and client prediction receive copies built from the same runtime-
// identified record generation.
type Catalog struct {
	byClass map[string]ClassRates
}

// LoadCatalog projects the pinned CharStats record into immutable movement facts keyed without case sensitivity.
// Failing the whole load on malformed or duplicate classes prevents authority and prediction from using partial data.
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

		rates, valid := classRatesFromRow(row)
		if !valid {
			return Catalog{}, fmt.Errorf(
				"d2legacy movement: %s row %d class %q has invalid movement/stamina facts",
				charStatsPath,
				index+2,
				class,
			)
		}

		key := strings.ToLower(class)
		if _, duplicate := result.byClass[key]; duplicate {
			return Catalog{}, fmt.Errorf("d2legacy movement: %s contains duplicate class %q", charStatsPath, class)
		}

		result.byClass[key] = rates
	}

	if len(result.byClass) == 0 {
		return Catalog{}, fmt.Errorf("d2legacy movement: %s contains no playable class velocities", charStatsPath)
	}

	return result, nil
}

// classRatesFromRow parses all movement and stamina fields together so a class is accepted or rejected atomically.
func classRatesFromRow(row map[string]string) (ClassRates, bool) {
	walk, walkErr := positiveNumber(row["WalkVelocity"])
	run, runErr := positiveNumber(row["RunVelocity"])
	stamina, staminaErr := positiveInteger(row["stamina"])
	vitality, baseVitalityErr := positiveInteger(row["vit"])
	runDrain, runDrainErr := positiveInteger(row["RunDrain"])
	staminaPerLevel, levelErr := nonNegativeInteger(row["StaminaPerLevel"])
	staminaPerVitality, vitalityErr := nonNegativeInteger(row["StaminaPerVitality"])

	valid := walkErr == nil && runErr == nil && staminaErr == nil && baseVitalityErr == nil && runDrainErr == nil &&
		levelErr == nil && vitalityErr == nil && run >= walk
	if !valid {
		return ClassRates{}, false
	}

	return ClassRates{
		Walk:               walk,
		Run:                run,
		StartingVitality:   vitality,
		StartingStamina:    stamina,
		RunDrain:           runDrain,
		StaminaPerLevel:    staminaPerLevel,
		StaminaPerVitality: staminaPerVitality,
	}, true
}

// positiveNumber parses a required positive velocity while rejecting empty, zero, and negative record values.
func positiveNumber(raw string) (float64, error) {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("invalid velocity %q", raw)
	}

	return value, nil
}

// positiveInteger parses record fields whose zero value would invalidate the pinned class movement model.
func positiveInteger(raw string) (int64, error) {
	value, err := nonNegativeInteger(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("invalid positive integer %q", raw)
	}

	return value, nil
}

// nonNegativeInteger parses progression fields where zero explicitly means the class receives no increment.
func nonNegativeInteger(raw string) (int64, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("invalid non-negative integer %q", raw)
	}

	return value, nil
}

// Rates returns a copied class record so callers cannot mutate the catalog's authoritative map entry.
func (catalog Catalog) Rates(class string) (ClassRates, bool) {
	rates, found := catalog.byClass[strings.ToLower(strings.TrimSpace(class))]
	return rates, found
}
