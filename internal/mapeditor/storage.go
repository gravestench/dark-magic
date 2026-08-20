package mapeditor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

// Storage confines editor writes to one explicitly configured project root.
// The mounted game-content VFS remains read-only and never becomes a write
// authority merely because a map was opened from it.
type Storage struct {
	root          string
	readOnlyRoots []string
}

// NewStorage validates and creates one user-selected output directory.
// Read-only roots name mounted source-data directories. Save rejects a
// destination below one, even if a caller misconfigures the output root.
func NewStorage(root string, readOnlyRoots ...string) (*Storage, error) {
	if strings.TrimSpace(root) == "" {
		return nil, fmt.Errorf("map editor: an output directory is required to save")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("map editor: resolve output directory: %w", err)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, fmt.Errorf("map editor: create output directory: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, fmt.Errorf("map editor: resolve output directory links: %w", err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return nil, fmt.Errorf("map editor: inspect output directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("map editor: output path %q is not a directory", root)
	}
	protected, err := resolveRoots(readOnlyRoots)
	if err != nil {
		return nil, err
	}
	for _, protectedRoot := range protected {
		if samePath(resolved, protectedRoot) {
			return nil, fmt.Errorf("map editor: output directory %q is a mounted read-only source", root)
		}
	}
	return &Storage{root: resolved, readOnlyRoots: protected}, nil
}

// Save writes one encoded DS1 atomically below the configured project root.
// Relative paths may organize a mod's data tree but may never escape it.
func (storage *Storage) Save(relative string, encoded []byte) (string, error) {
	if storage == nil {
		return "", fmt.Errorf("map editor: saving is unavailable")
	}
	clean, err := cleanSavePath(relative)
	if err != nil {
		return "", err
	}
	destination := filepath.Join(storage.root, clean)
	if storage.isReadOnlyDestination(destination) {
		return "", fmt.Errorf("map editor: save path %q targets mounted read-only game data", relative)
	}
	parent := filepath.Dir(destination)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", fmt.Errorf("map editor: create save directory: %w", err)
	}
	resolvedParent, err := filepath.EvalSymlinks(parent)
	if err != nil {
		return "", fmt.Errorf("map editor: resolve save directory links: %w", err)
	}
	if !within(storage.root, resolvedParent) {
		return "", fmt.Errorf("map editor: save path %q escapes configured output directory", relative)
	}

	temporary, err := os.CreateTemp(resolvedParent, ".map-editor-*.tmp")
	if err != nil {
		return "", fmt.Errorf("map editor: create temporary save: %w", err)
	}
	temporaryName := temporary.Name()
	defer func() { _ = os.Remove(temporaryName) }()
	if _, err := temporary.Write(encoded); err != nil {
		_ = temporary.Close()
		return "", fmt.Errorf("map editor: write temporary save: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return "", fmt.Errorf("map editor: sync temporary save: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return "", fmt.Errorf("map editor: close temporary save: %w", err)
	}
	if err := os.Rename(temporaryName, destination); err != nil {
		return "", fmt.Errorf("map editor: replace save: %w", err)
	}
	if err := darkpaths.SyncDirectory(resolvedParent); err != nil {
		return "", fmt.Errorf("map editor: sync save directory: %w", err)
	}
	return destination, nil
}

// isReadOnlyDestination performs the final defense against an output path that aliases mounted source data.
func (storage *Storage) isReadOnlyDestination(destination string) bool {
	for _, root := range storage.readOnlyRoots {
		if within(root, destination) {
			return true
		}
	}
	return false
}

// resolveRoots canonicalizes protected paths so symlink aliases cannot bypass containment checks.
func resolveRoots(roots []string) ([]string, error) {
	result := make([]string, 0, len(roots))
	for _, root := range roots {
		if strings.TrimSpace(root) == "" {
			continue
		}
		absolute, err := filepath.Abs(root)
		if err != nil {
			return nil, fmt.Errorf("map editor: resolve read-only source %q: %w", root, err)
		}
		resolved, err := filepath.EvalSymlinks(absolute)
		if err != nil {
			return nil, fmt.Errorf("map editor: resolve read-only source %q links: %w", root, err)
		}
		result = append(result, resolved)
	}
	return result, nil
}

// samePath compares canonical paths without requiring either path to name a child.
func samePath(left, right string) bool {
	return filepath.Clean(left) == filepath.Clean(right)
}

// cleanSavePath accepts only relative DS1 destinations below the configured project root.
func cleanSavePath(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("map editor: a DS1 file name is required")
	}
	if filepath.IsAbs(value) {
		return "", fmt.Errorf("map editor: save path must be relative")
	}
	clean := filepath.Clean(value)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("map editor: save path %q escapes output directory", value)
	}
	if !strings.EqualFold(filepath.Ext(clean), ".ds1") {
		return "", fmt.Errorf("map editor: save path %q must end in .ds1", value)
	}
	return clean, nil
}

// within performs a path-component-aware containment check instead of a vulnerable string-prefix comparison.
func within(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
