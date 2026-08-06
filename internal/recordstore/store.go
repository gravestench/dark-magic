// Package recordstore loads Diablo tabular records directly from layered content.
package recordstore

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"sync"
)

// Store caches immutable generic TSV records by normalized content path.
type Store struct {
	source fs.FS
	mu     sync.RWMutex
	cache  map[string][]map[string]string
}

// New constructs a record store over source.
func New(source fs.FS) *Store {
	return &Store{source: source, cache: make(map[string][]map[string]string)}
}

// Read returns original layered table bytes for a format codec. Generic rows
// and typed decoding share VFS ownership without duplicating source discovery.
func (s *Store) Read(path string) ([]byte, error) {
	if s == nil || s.source == nil {
		return nil, fmt.Errorf("recordstore: no content source")
	}
	data, err := fs.ReadFile(s.source, path)
	if err != nil {
		return nil, fmt.Errorf("recordstore: read %q: %w", path, err)
	}
	return bytes.Clone(data), nil
}

// Load returns a defensive copy of a TSV table.
func (s *Store) Load(path string) ([]map[string]string, error) {
	s.mu.RLock()
	cached, exists := s.cache[path]
	s.mu.RUnlock()
	if exists {
		return cloneRows(cached), nil
	}
	file, err := s.source.Open(path)
	if err != nil {
		return nil, fmt.Errorf("recordstore: open %q: %w", path, err)
	}
	defer file.Close()
	rows, err := parseTSV(file)
	if err != nil {
		return nil, fmt.Errorf("recordstore: parse %q: %w", path, err)
	}
	s.mu.Lock()
	if existing, loaded := s.cache[path]; loaded {
		rows = existing
	} else {
		s.cache[path] = rows
	}
	s.mu.Unlock()
	return cloneRows(rows), nil
}

// Invalidate removes one cached table so its next access reloads layered content.
func (s *Store) Invalidate(path string) {
	s.mu.Lock()
	delete(s.cache, path)
	s.mu.Unlock()
}

// Loaded reports whether path is cached.
func (s *Store) Loaded(path string) bool {
	s.mu.RLock()
	_, exists := s.cache[path]
	s.mu.RUnlock()
	return exists
}

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
	header[0] = strings.TrimPrefix(header[0], "\ufeff")
	seen := make(map[string]bool, len(header))
	for _, column := range header {
		if column == "" {
			return nil, fmt.Errorf("empty column name")
		}
		if seen[column] {
			return nil, fmt.Errorf("duplicate column %q", column)
		}
		seen[column] = true
	}
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
