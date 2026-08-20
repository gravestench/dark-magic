// Package typed binds one immutable generic record table to a caller-selected
// Go schema. It does not choose tables, join domains, or define a game catalog.
package typed

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/tsv"
)

// Load decodes one layered TSV table into the surviving csv-tagged record type.
// Unknown columns remain available through the generic store but do not break
// typed consumers, which lets mods extend tables independently of engine schemas.
func Load[T any](store *recordstore.Store, path string) ([]T, error) {
	if store == nil {
		return nil, fmt.Errorf("gamedata: nil record store")
	}

	records, decodeErr := decodeTSV[T](store, path)
	if decodeErr != nil && !isBareQuoteError(decodeErr) {
		return nil, fmt.Errorf("gamedata: %s: %w", path, decodeErr)
	}

	rows, err := store.Load(path)
	if err != nil {
		return nil, err
	}

	if decodeErr != nil || shouldBindGenericRows(records, rows) {
		// The generic store accepts the shipped tables' literal quotes with
		// LazyQuotes, so its rows provide the compatibility path for that data.
		return bindRows[T](path, rows)
	}

	if len(records) != len(rows) {
		// The strict codec filters malformed-width rows such as shipped one-cell
		// expansion sentinels, while generic access intentionally preserves them.
		return records, nil
	}

	if err := reconcileDecodedRows(path, records, rows); err != nil {
		return nil, err
	}

	return records, nil
}

// decodeTSV owns the strict decoder's file lifetime and gives decode failures
// precedence over close failures, preserving the most useful source diagnostic.
func decodeTSV[T any](store *recordstore.Store, path string) ([]T, error) {
	file, err := store.Open(path)
	if err != nil {
		return nil, err
	}

	var records []T

	decodeErr := tsv.Decode(file, &records)

	closeErr := file.Close()
	if decodeErr == nil {
		decodeErr = closeErr
	}

	return records, decodeErr
}

// isBareQuoteError limits tolerant fallback to the one known shipped-data
// incompatibility; all other codec errors retain their original failure path.
func isBareQuoteError(err error) bool {
	return strings.Contains(err.Error(), `bare " in non-quoted-field`)
}

// shouldBindGenericRows detects a strict-decoder false empty result without
// treating a genuinely empty table as malformed.
func shouldBindGenericRows[T any](records []T, rows []map[string]string) bool {
	return len(records) == 0 && len(rows) > 0
}

// reconcileDecodedRows repairs only schema conventions the strict decoder
// cannot represent, leaving its normal scalar conversions and errors intact.
func reconcileDecodedRows[T any](path string, records []T, rows []map[string]string) error {
	fields, err := recordFields[T]()
	if err != nil {
		return fmt.Errorf("gamedata: %s: %w", path, err)
	}

	for rowIndex, row := range rows {
		record := recordValue(&records[rowIndex])
		if err := reconcileDecodedRow(path, rowIndex, record, row, fields); err != nil {
			return err
		}
	}

	return nil
}

// reconcileDecodedRow restores nil optional values and historical grouped
// fields so callers observe the schema conventions used by the surviving models.
func reconcileDecodedRow(
	path string,
	rowIndex int,
	record reflect.Value,
	row map[string]string,
	fields []fieldBinding,
) error {
	for _, field := range fields {
		raw, exists := row[field.column]
		if !exists {
			continue
		}

		destination := field.destination(record)
		if destination.Kind() == reflect.Pointer && raw == "" {
			// The strict decoder allocates a zero pointer for a blank cell, but
			// typed schemas use nil to distinguish absent values from authored 0.
			destination.SetZero()
			continue
		}

		if !field.grouped {
			continue
		}

		if err := assignBoundValue(path, rowIndex, destination, field, raw); err != nil {
			return err
		}
	}

	return nil
}
