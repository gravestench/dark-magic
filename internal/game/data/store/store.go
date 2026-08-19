// Package recordstore loads Diablo tabular records directly from layered content.
package recordstore

import (
	"bytes"
	"fmt"
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
	canonical    map[string]string
	generationID string
	provenance   map[string]Provenance
}

// canonicalPath resolves case-insensitive pinned paths while leaving ordinary filesystem paths unchanged. The read
// lock keeps concurrent cache invalidation and lookups independent of generation setup.
func (s *Store) canonicalPath(path string) string {
	if s == nil {
		return path
	}
	s.mu.RLock()
	canonical := s.canonical[strings.ToLower(path)]
	s.mu.RUnlock()
	if canonical != "" {
		return canonical
	}
	return path
}

// Provenance identifies the winning content layer and its source-relative path for an immutable pinned table.
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
	value, found := s.provenance[s.canonicalPathLocked(path)]
	return value, found
}

// canonicalPathLocked resolves a path while the caller already holds either store lock, avoiding recursive locking in
// operations that must read provenance atomically.
func (s *Store) canonicalPathLocked(path string) string {
	if canonical := s.canonical[strings.ToLower(path)]; canonical != "" {
		return canonical
	}
	return path
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
	return &Store{
		source:    source,
		logger:    slog.Default(),
		cache:     make(map[string][]map[string]string),
		canonical: make(map[string]string),
	}
}

// SetLogger configures record-load diagnostics. Synchronizing the replacement makes it safe to change diagnostics
// while other goroutines load tables; a nil logger deliberately disables them.
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
	requested := path
	path = s.canonicalPath(path)
	data, err := fs.ReadFile(s.source, path)
	if err != nil {
		return nil, fmt.Errorf("recordstore: read %q: %w", requested, err)
	}
	return bytes.Clone(data), nil
}

// Open returns the original layered table as a stream for format codecs.
// The caller owns the returned file and must close it.
func (s *Store) Open(path string) (fs.File, error) {
	if s == nil || s.source == nil {
		return nil, fmt.Errorf("recordstore: no content source")
	}
	requested := path
	path = s.canonicalPath(path)
	file, err := s.source.Open(path)
	if err != nil {
		return nil, fmt.Errorf("recordstore: open %q: %w", requested, err)
	}
	return file, nil
}

// Load returns a defensive copy of a TSV table.
func (s *Store) Load(path string) ([]map[string]string, error) {
	requested := path
	path = s.canonicalPath(path)

	if cached, exists := s.cachedRows(path); exists {
		return cached, nil
	}

	rows, err := s.loadUncachedRows(requested, path)
	if err != nil {
		return nil, err
	}

	rows, loaded, logger := s.cacheLoadedRows(path, rows)
	if loaded && logger != nil {
		s.logLoadedRows(logger, path, len(rows))
	}

	return cloneRows(rows), nil
}

// cachedRows returns an owned copy on cache hits so callers cannot mutate the shared immutable snapshot.
func (s *Store) cachedRows(path string) ([]map[string]string, bool) {
	s.mu.RLock()
	cached, exists := s.cache[path]
	s.mu.RUnlock()

	if !exists {
		return nil, false
	}

	return cloneRows(cached), true
}

// loadUncachedRows reads and parses one table without holding the cache lock, allowing independent tables to load in
// parallel. The requested spelling remains in open errors, while parse errors use the canonical pinned path.
func (s *Store) loadUncachedRows(requested, path string) ([]map[string]string, error) {
	file, err := s.source.Open(path)
	if err != nil {
		return nil, fmt.Errorf("recordstore: open %q: %w", requested, err)
	}
	defer file.Close()

	rows, err := parseTSV(file)
	if err != nil {
		return nil, fmt.Errorf("recordstore: parse %q: %w", path, err)
	}

	return rows, nil
}

// cacheLoadedRows elects the first completed load as the cached snapshot. Concurrent losers reuse that winner so all
// callers observe identical rows and only the winning load emits diagnostics.
func (s *Store) cacheLoadedRows(path string, rows []map[string]string) ([]map[string]string, bool, *slog.Logger) {
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

	return rows, loaded, logger
}

// logLoadedRows resolves provenance after caching so slow or failing source metadata cannot block other cache users.
func (s *Store) logLoadedRows(logger *slog.Logger, path string, count int) {
	sourceLayer, sourcePath := s.resolveSource(path)
	logger.Info(
		"loaded records",
		"table", path,
		"records", count,
		"source", sourceLayer,
		"source_path", sourcePath,
	)
}

// resolveSource converts optional layered-filesystem metadata into stable diagnostic fields. Resolution failures are
// reported as metadata rather than failing an otherwise successful record load.
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
	path = s.canonicalPath(path)
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
	path = s.canonicalPath(path)
	s.mu.RLock()
	_, exists := s.cache[path]
	s.mu.RUnlock()
	return exists
}
