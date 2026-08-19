package recordstore

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// parseTSV decodes generic record tables without imposing typed schema rules. Flexible field counts and lazy quotes
// preserve compatibility with shipped Diablo II tables and mod-authored variants.
func parseTSV(input io.Reader) ([]map[string]string, error) {
	reader := csv.NewReader(input)
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	header, err := reader.Read()
	if err != nil {
		return nil, err
	}

	if len(header) == 0 {
		return nil, fmt.Errorf("empty header")
	}

	normalizeHeader(header)

	return readRows(reader, header)
}

// normalizeHeader gives every authored column a stable map key. The first duplicate keeps its original name for typed
// compatibility, while later duplicates and unnamed columns remain available to generic consumers.
func normalizeHeader(header []string) {
	header[0] = strings.TrimPrefix(header[0], "\ufeff")
	seen := make(map[string]int, len(header))

	for index, column := range header {
		if column == "" {
			header[index] = fmt.Sprintf("#unnamed-%d", index+1)
			continue
		}

		seen[column]++
		if seen[column] > 1 {
			// Shipped tables contain duplicate headers, so suffix only later occurrences to avoid discarding cells.
			header[index] = fmt.Sprintf("%s#%d", column, seen[column])
		}
	}
}

// readRows maps variable-width records onto the normalized header. Missing trailing cells become empty strings and
// surplus cells remain ignored, matching the package's established generic-table behavior.
func readRows(reader *csv.Reader, header []string) ([]map[string]string, error) {
	var result []map[string]string

	for rowNumber := 2; ; rowNumber++ {
		values, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("row %d: %w", rowNumber, err)
		}

		row := make(map[string]string, len(header))
		for index, column := range header {
			if index < len(values) {
				row[column] = values[index]
			} else {
				row[column] = ""
			}
		}

		result = append(result, row)
	}

	return result, nil
}

// cloneRows returns full slice-and-map ownership to the caller so cached records remain immutable across API calls.
func cloneRows(rows []map[string]string) []map[string]string {
	result := make([]map[string]string, len(rows))
	for index, row := range rows {
		copyRow := make(map[string]string, len(row))
		for key, value := range row {
			copyRow[key] = value
		}

		result[index] = copyRow
	}

	return result
}
