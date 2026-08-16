// Package recordstore loads Diablo tabular records directly from layered content.
package recordstore

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"strings"
	"sync"
)

// Store caches immutable generic TSV records by normalized content path.
type Store struct {
	source       fs.FS
	logger       *slog.Logger
	mu           sync.RWMutex
	cache        map[string][]map[string]string
	generationID string
	provenance   map[string]Provenance
}

type Provenance struct {
	Layer string
	Path  string
}

// Source reports the winning immutable source for a pinned table.
func (s *Store) Source(path string) (Provenance, bool) {
	if s == nil {
		return Provenance{}, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, found := s.provenance[path]
	return value, found
}

// GenerationID identifies an immutable pinned authoritative view. Ordinary
// development stores return an empty ID and must not be attached to a Session.
func (s *Store) GenerationID() string {
	if s == nil {
		return ""
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.generationID
}

// New constructs a record store over source.
func New(source fs.FS) *Store {
	return &Store{source: source, logger: slog.Default(), cache: make(map[string][]map[string]string)}
}

// SetLogger configures record-load diagnostics. A nil logger disables them.
func (s *Store) SetLogger(logger *slog.Logger) {
	s.mu.Lock()
	s.logger = logger
	s.mu.Unlock()
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

// Open returns the original layered table as a stream for format codecs.
// The caller owns the returned file and must close it.
func (s *Store) Open(path string) (fs.File, error) {
	if s == nil || s.source == nil {
		return nil, fmt.Errorf("recordstore: no content source")
	}
	file, err := s.source.Open(path)
	if err != nil {
		return nil, fmt.Errorf("recordstore: open %q: %w", path, err)
	}
	return file, nil
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
	loaded := false
	if existing, cached := s.cache[path]; cached {
		rows = existing
	} else {
		s.cache[path] = rows
		loaded = true
	}
	logger := s.logger
	s.mu.Unlock()
	if loaded && logger != nil {
		sourceLayer, sourcePath := s.resolveSource(path)
		logger.Info("loaded records", "table", path, "records", len(rows), "source", sourceLayer, "source_path", sourcePath)
	}
	return cloneRows(rows), nil
}

func (s *Store) resolveSource(path string) (string, string) {
	resolver, ok := s.source.(interface {
		ResolveSource(string) (layer string, path string, err error)
	})
	if !ok {
		return "filesystem", path
	}
	layer, resolvedPath, err := resolver.ResolveSource(path)
	if err != nil {
		return "unresolved", path
	}
	return layer, resolvedPath
}

// Invalidate removes one cached table so its next access reloads layered content.
func (s *Store) Invalidate(path string) {
	s.mu.Lock()
	delete(s.cache, path)
	s.mu.Unlock()
}

// InvalidateAll clears every derived table after the mounted package recipe
// changes. The immutable bytes stay in the VFS and reload lazily on demand.
func (s *Store) InvalidateAll() {
	s.mu.Lock()
	s.cache = make(map[string][]map[string]string)
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
	seen := make(map[string]int, len(header))
	for index, column := range header {
		if column == "" {
			header[index] = fmt.Sprintf("#unnamed-%d", index+1)
			continue
		}
		seen[column]++
		if seen[column] > 1 {
			// Shipped Diablo II tables contain duplicate headers. Preserve every
			// cell for generic/mod consumers while keeping the first occurrence
			// at its authored name for typed compatibility.
			header[index] = fmt.Sprintf("%s#%d", column, seen[column])
		}
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
