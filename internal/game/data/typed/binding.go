package typed

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// fieldBinding records how one authored TSV column reaches a Go field or array
// element, including historical grouped declarations that need reconciliation.
type fieldBinding struct {
	index      int
	name       string
	column     string
	grouped    bool
	arrayIndex int
}

// destination resolves a binding against an addressable record. Keeping array
// traversal here prevents the two binding paths from drifting apart.
func (binding fieldBinding) destination(record reflect.Value) reflect.Value {
	destination := record.Field(binding.index)
	if binding.arrayIndex >= 0 {
		destination = destination.Index(binding.arrayIndex)
	}

	return destination
}

// bindRows converts lossless generic rows when the strict decoder cannot accept
// shipped input, preserving source row numbers in all conversion errors.
func bindRows[T any](path string, rows []map[string]string) ([]T, error) {
	fields, err := recordFields[T]()
	if err != nil {
		return nil, fmt.Errorf("gamedata: %s: %w", path, err)
	}

	records := make([]T, len(rows))
	for rowIndex, row := range rows {
		record := recordValue(&records[rowIndex])
		if err := bindRow(path, rowIndex, record, row, fields); err != nil {
			return nil, err
		}
	}

	return records, nil
}

// bindRow assigns every known authored column while ignoring mod-owned columns
// that do not belong to the caller-selected schema.
func bindRow(
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
		if err := assignBoundValue(path, rowIndex, destination, field, raw); err != nil {
			return err
		}
	}

	return nil
}

// assignBoundValue adds stable table coordinates to conversion failures so bad
// authored data can be located without exposing reflection details to callers.
func assignBoundValue(
	path string,
	rowIndex int,
	destination reflect.Value,
	field fieldBinding,
	raw string,
) error {
	if err := assign(destination, raw); err != nil {
		return fmt.Errorf(
			"gamedata: %s row %d column %q field %s: %w",
			path,
			rowIndex+2,
			field.column,
			field.name,
			err,
		)
	}

	return nil
}

// recordValue returns the addressable struct behind a record pointer, which is
// required for reflection assignments to update the caller-visible slice.
func recordValue[T any](record *T) reflect.Value {
	return reflect.ValueOf(record).Elem()
}

// recordFields validates a record schema and expands its csv tags into ordered
// bindings, retaining declaration order for deterministic conversion failures.
func recordFields[T any]() ([]fieldBinding, error) {
	recordType := reflect.TypeOf((*T)(nil)).Elem()
	if recordType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("record type %s is not a struct", recordType)
	}

	bindings := make([]fieldBinding, 0, recordType.NumField())
	seenColumns := make(map[string]string)

	for index := 0; index < recordType.NumField(); {
		field := recordType.Field(index)

		tag := field.Tag.Get("csv")
		if !field.IsExported() || tag == "" || tag == "-" {
			index++
			continue
		}

		groupEnd := csvFieldGroupEnd(recordType, index, tag)

		columns := strings.Split(tag, ",")
		if field.Type.Kind() == reflect.Array {
			arrayBindings, err := arrayFieldBindings(field, index, columns)
			if err != nil {
				return nil, err
			}

			// Arrays historically expand independently of scalar duplicate-tag
			// tracking, so retain that compatibility behavior here.
			bindings = append(bindings, arrayBindings...)
			index++

			continue
		}

		groupBindings, err := groupedFieldBindings(recordType, index, groupEnd, columns, seenColumns)
		if err != nil {
			return nil, err
		}

		bindings = append(bindings, groupBindings...)
		index = groupEnd
	}

	if len(bindings) == 0 {
		return nil, fmt.Errorf("record type %s has no csv-tagged fields", recordType)
	}

	return bindings, nil
}

// csvFieldGroupEnd finds consecutive fields sharing one historical grouped tag;
// stopping at the first different tag preserves declaration boundaries.
func csvFieldGroupEnd(recordType reflect.Type, start int, tag string) int {
	end := start + 1
	for end < recordType.NumField() && recordType.Field(end).Tag.Get("csv") == tag {
		end++
	}

	return end
}

// arrayFieldBindings maps a multi-column tag onto one fixed-size array and
// rejects width mismatches before any row can be partially assigned.
func arrayFieldBindings(field reflect.StructField, index int, columns []string) ([]fieldBinding, error) {
	if len(columns) != field.Type.Len() {
		return nil, fmt.Errorf(
			"array csv tag %q has %d columns for %d elements",
			field.Tag.Get("csv"),
			len(columns),
			field.Type.Len(),
		)
	}

	bindings := make([]fieldBinding, 0, len(columns))
	for arrayIndex, column := range columns {
		bindings = append(bindings, fieldBinding{
			index:      index,
			name:       field.Name,
			column:     column,
			grouped:    true,
			arrayIndex: arrayIndex,
		})
	}

	return bindings, nil
}

// groupedFieldBindings pairs consecutive fields with their authored columns
// and rejects ambiguous scalar destinations before row conversion begins.
func groupedFieldBindings(
	recordType reflect.Type,
	start int,
	end int,
	columns []string,
	seenColumns map[string]string,
) ([]fieldBinding, error) {
	fieldCount := end - start
	if fieldCount == 1 {
		// Singleton fields historically use only the first tag segment; retain
		// that convention rather than treating later segments as destinations.
		columns = columns[:1]
	} else if len(columns) != fieldCount {
		return nil, fmt.Errorf(
			"grouped csv tag %q has %d columns for %d fields",
			recordType.Field(start).Tag.Get("csv"),
			len(columns),
			fieldCount,
		)
	}

	bindings := make([]fieldBinding, 0, len(columns))
	grouped := fieldCount > 1

	for offset, column := range columns {
		field := recordType.Field(start + offset)
		if previous, exists := seenColumns[column]; exists {
			return nil, fmt.Errorf(
				"duplicate csv column %q on fields %s and %s",
				column,
				previous,
				field.Name,
			)
		}

		seenColumns[column] = field.Name
		bindings = append(bindings, fieldBinding{
			index:      start + offset,
			name:       field.Name,
			column:     column,
			grouped:    grouped,
			arrayIndex: -1,
		})
	}

	return bindings, nil
}

// assign converts one source cell according to its exact destination type.
// Blank numeric cells retain zero values, while blank pointers remain absent.
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

// parseBool accepts the numeric and textual forms authored across Diablo data
// while rejecting unknown tokens instead of silently interpreting them as false.
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
