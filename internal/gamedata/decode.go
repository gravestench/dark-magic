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
	for index := 0; index < typeOf.NumField(); index++ {
		field := typeOf.Field(index)
		if !field.IsExported() {
			continue
		}
		column := strings.Split(field.Tag.Get("csv"), ",")[0]
		if column == "" || column == "-" {
			continue
		}
		if previous, exists := seen[column]; exists {
			return nil, fmt.Errorf("duplicate csv column %q on fields %s and %s", column, previous, field.Name)
		}
		seen[column] = field.Name
		bindings = append(bindings, fieldBinding{index: index, name: field.Name, column: column})
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
