package content

import (
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

// indexedFS names archive members without pretending that a flat archive implements directory traversal.
type indexedFS interface {
	Paths() []string
}

// mountedPathCollector accumulates normalized, case-insensitively unique paths for one List operation.
// The first spelling wins, which preserves the mounted layer priority before the final lexical sort.
type mountedPathCollector struct {
	root   string
	suffix string
	seen   map[string]struct{}
	paths  []string
}

// ReadDir returns the union of a directory across all layers, with higher-priority entries shadowing equal names.
func (f *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	clean, err := Normalize(name)
	if err != nil {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: err}
	}

	entries := make(map[string]fs.DirEntry)
	found := false

	for _, layer := range f.snapshot() {
		layerEntries, err := fs.ReadDir(layer.FS, clean)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}

			return nil, fmt.Errorf("content: read directory %q from layer %q: %w", clean, layer.Name, err)
		}

		found = true

		for _, entry := range layerEntries {
			if _, exists := entries[entry.Name()]; !exists {
				entries[entry.Name()] = entry
			}
		}
	}

	if !found {
		return nil, &fs.PathError{Op: "readdir", Path: clean, Err: fs.ErrNotExist}
	}

	result := make([]fs.DirEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry)
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Name() < result[j].Name() })

	return result, nil
}

// Exists reports whether a normalized content path currently resolves and closes the temporary probe immediately.
func (f *FS) Exists(name string) bool {
	file, err := f.Open(name)
	if err != nil {
		return false
	}

	_ = file.Close()

	return true
}

// Walk enumerates the merged layered view in lexical order through the standard io/fs traversal contract.
func (f *FS) Walk(root string, visit fs.WalkDirFunc) error {
	return fs.WalkDir(f, root, visit)
}

// List returns mounted files below root, optionally filtered by a case-insensitive suffix.
// Higher-priority layers determine the retained spelling when the same logical path appears more than once.
func (f *FS) List(root, suffix string) ([]string, error) {
	cleanRoot, err := Normalize(root)
	if err != nil {
		return nil, err
	}

	collector := mountedPathCollector{
		root:   cleanRoot,
		suffix: strings.ToLower(strings.TrimSpace(suffix)),
		seen:   make(map[string]struct{}),
		paths:  make([]string, 0),
	}

	for _, layer := range f.snapshot() {
		if indexed, ok := layer.FS.(indexedFS); ok {
			collector.addAll(indexed.Paths())
			continue
		}

		walkErr := fs.WalkDir(layer.FS, cleanRoot, collector.visit)
		if walkErr != nil && !errors.Is(walkErr, fs.ErrNotExist) {
			return nil, fmt.Errorf("content: list %q from layer %q: %w", cleanRoot, layer.Name, walkErr)
		}
	}

	sort.Strings(collector.paths)

	return collector.paths, nil
}

// addAll feeds a flat archive index through the same normalization and priority rules as walked files.
func (c *mountedPathCollector) addAll(paths []string) {
	for _, name := range paths {
		c.add(name)
	}
}

// visit adapts an ordinary directory walk to the collector while preserving traversal failures from the source.
func (c *mountedPathCollector) visit(name string, entry fs.DirEntry, visitErr error) error {
	if visitErr != nil {
		return visitErr
	}

	if !entry.IsDir() {
		c.add(name)
	}

	return nil
}

// add admits one normalized in-root path unless an earlier layer already supplied its case-insensitive identity.
func (c *mountedPathCollector) add(name string) {
	clean, err := Normalize(name)
	if err != nil || !pathBelow(c.root, clean) {
		return
	}

	if c.suffix != "" && !strings.HasSuffix(strings.ToLower(clean), c.suffix) {
		return
	}

	key := strings.ToLower(clean)
	if _, exists := c.seen[key]; exists {
		return
	}

	c.seen[key] = struct{}{}
	c.paths = append(c.paths, clean)
}

// pathBelow reports whether name is root itself or a descendant without confusing sibling prefixes for children.
func pathBelow(root, name string) bool {
	return root == "." || name == root || strings.HasPrefix(name, root+"/")
}
