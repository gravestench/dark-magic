// Package gamedata loads the typed records that form Diablo II's authored game
// database. It deliberately owns data decoding rather than application service
// lifecycle; the host can build and publish immutable catalog snapshots.
package gamedata

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/gravestench/dark-magic/internal/recordstore"
)

// Load decodes one layered TSV table into the surviving csv-tagged record type.
// Unknown columns are retained by the generic record store and ignored here,
// allowing mods to extend tables without breaking typed engine consumers.
func Load[T any](store *recordstore.Store, path string) ([]T, error) {
	if store == nil {
		return nil, fmt.Errorf("gamedata: nil record store")
	}
	rows, err := store.Load(path)
	if err != nil {
		return nil, err
	}
	fields, err := recordFields[T]()
	if err != nil {
		return nil, fmt.Errorf("gamedata: %s: %w", path, err)
	}
	result := make([]T, len(rows))
	for rowIndex, row := range rows {
		value := reflect.ValueOf(&result[rowIndex]).Elem()
		for _, field := range fields {
			raw, exists := row[field.column]
			if !exists {
				continue
			}
			if err := assign(value.Field(field.index), raw); err != nil {
				return nil, fmt.Errorf("gamedata: %s row %d column %q field %s: %w", path, rowIndex+2, field.column, field.name, err)
			}
		}
	}
	return result, nil
}

type fieldBinding struct {
	index  int
	name   string
	column string
}

func recordFields[T any]() ([]fieldBinding, error) {
	typeOf := reflect.TypeOf((*T)(nil)).Elem()
	if typeOf.Kind() != reflect.Struct {
		return nil, fmt.Errorf("record type %s is not a struct", typeOf)
	}
	bindings := make([]fieldBinding, 0, typeOf.NumField())
	seen := make(map[string]string)
	for index := 0; index < typeOf.NumField(); {
		field := typeOf.Field(index)
		if !field.IsExported() {
			index++
			continue
		}
		tag := field.Tag.Get("csv")
		if tag == "" || tag == "-" {
			index++
			continue
		}
		// The historical schemas use grouped Go field declarations with one
		// comma-separated tag, for example `SizeX, SizeXN, SizeXH int
		// csv:"SizeX,SizeX(N),SizeX(H)"`. Reflection exposes three fields with
		// the same tag, so recover that convention by pairing the consecutive
		// fields and columns. This is distinct from unsupported csv options.
		groupEnd := index + 1
		for groupEnd < typeOf.NumField() && typeOf.Field(groupEnd).Tag.Get("csv") == tag {
			groupEnd++
		}
		columns := strings.Split(tag, ",")
		if groupEnd-index == 1 {
			columns = columns[:1]
		} else if len(columns) != groupEnd-index {
			return nil, fmt.Errorf("grouped csv tag %q has %d columns for %d fields", tag, len(columns), groupEnd-index)
		}
		for offset, column := range columns {
			groupField := typeOf.Field(index + offset)
			if previous, exists := seen[column]; exists {
				return nil, fmt.Errorf("duplicate csv column %q on fields %s and %s", column, previous, groupField.Name)
			}
			seen[column] = groupField.Name
			bindings = append(bindings, fieldBinding{index: index + offset, name: groupField.Name, column: column})
		}
		index = groupEnd
	}
	if len(bindings) == 0 {
		return nil, fmt.Errorf("record type %s has no csv-tagged fields", typeOf)
	}
	return bindings, nil
}

func assign(destination reflect.Value, raw string) error {
	if destination.Kind() == reflect.Pointer {
		if raw == "" {
			return nil
		}
		destination.Set(reflect.New(destination.Type().Elem()))
		return assign(destination.Elem(), raw)
	}
	switch destination.Kind() {
	case reflect.String:
		destination.SetString(raw)
	case reflect.Bool:
		value, err := parseBool(raw)
		if err != nil {
			return err
		}
		destination.SetBool(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if raw == "" {
			return nil
		}
		value, err := strconv.ParseInt(raw, 10, destination.Type().Bits())
		if err != nil {
			return err
		}
		destination.SetInt(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if raw == "" {
			return nil
		}
		value, err := strconv.ParseUint(raw, 10, destination.Type().Bits())
		if err != nil {
			return err
		}
		destination.SetUint(value)
	case reflect.Float32, reflect.Float64:
		if raw == "" {
			return nil
		}
		value, err := strconv.ParseFloat(raw, destination.Type().Bits())
		if err != nil {
			return err
		}
		destination.SetFloat(value)
	default:
		return fmt.Errorf("unsupported destination type %s", destination.Type())
	}
	return nil
}

func parseBool(raw string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "0", "false", "no":
		return false, nil
	case "1", "true", "yes":
		return true, nil
	default:
		return false, fmt.Errorf("invalid boolean %q", raw)
	}
}

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
// This matches the realities of shipped Diablo data without hiding ambiguity or
// making unused/sentinel duplicates fatal to otherwise usable tables.
func ObservedIndex[T any, K comparable](table string, records []T, key func(T) K) (map[K]T, []Issue, error) {
	if key == nil {
		return nil, nil, fmt.Errorf("gamedata: nil index key")
	}
	result := make(map[K]T, len(records))
	var issues []Issue
	for row, record := range records {
		value := key(record)
		if _, exists := result[value]; exists {
			issues = append(issues, Issue{Table: table, Row: row + 2, Kind: "duplicate-key", Message: fmt.Sprintf("duplicate key %v; lookup retains first occurrence", value)})
			continue
		}
		result[value] = record
	}
	return result, issues, nil
}
