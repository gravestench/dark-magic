package typed

import "fmt"

// Index constructs a deterministic primary-key lookup and rejects ambiguous
// table data rather than silently choosing one of two authored records.
func Index[T any, K comparable](records []T, key func(T) K) (map[K]T, error) {
	if key == nil {
		return nil, fmt.Errorf("gamedata: nil index key")
	}

	result := make(map[K]T, len(records))
	for row, record := range records {
		value := key(record)
		if _, exists := result[value]; exists {
			return nil, fmt.Errorf("gamedata: duplicate key %v at row %d", value, row+2)
		}

		result[value] = record
	}

	return result, nil
}

// Issue describes a tolerated source-data problem. The complete row remains in
// its typed slice; lookup indexes use a documented deterministic winner.
type Issue struct {
	Table   string
	Row     int
	Kind    string
	Message string
}

// ObservedIndex builds a first-record-wins index while reporting duplicates.
// This matches shipped Diablo data without hiding ambiguity or making unused
// sentinel duplicates fatal to otherwise usable tables.
func ObservedIndex[T any, K comparable](
	table string,
	records []T,
	key func(T) K,
) (map[K]T, []Issue, error) {
	if key == nil {
		return nil, nil, fmt.Errorf("gamedata: nil index key")
	}

	result := make(map[K]T, len(records))

	var issues []Issue

	for row, record := range records {
		value := key(record)
		if _, exists := result[value]; exists {
			issues = append(issues, Issue{
				Table:   table,
				Row:     row + 2,
				Kind:    "duplicate-key",
				Message: fmt.Sprintf("duplicate key %v; lookup retains first occurrence", value),
			})

			continue
		}

		result[value] = record
	}

	return result, issues, nil
}
