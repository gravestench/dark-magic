// Package content provides Dark Magic's layered game-content filesystem.
package content

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

// FromEnvironment constructs the production content stack. User content wins,
// followed by the bundled d2legacy mod, Diablo II patches/expansion data, and base
// archives. Missing optional directories and archives are skipped.
func FromEnvironment() (*FS, error) {
	layers := make([]Layer, 0, 16)
	if mods := os.Getenv("DARK_MAGIC_MOD_DIRECTORY"); mods != "" {
		expanded, err := darkpaths.ExpandHost(mods)
		if err != nil {
			return nil, fmt.Errorf("content: expand mod directory %q: %w", mods, err)
		}
		mods = expanded
		info, err := os.Stat(mods)
		if err != nil {
			return nil, fmt.Errorf("content: inspect mod directory %q: %w", mods, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("content: mod path %q is not a directory", mods)
		}
		layers = append(layers, Layer{Name: "user-mods", FS: Directory(mods)})
	}
	layers = append(layers, Layer{Name: "darkmagic", FS: Shim()})
	if configured := os.Getenv("MPQ_DIRECTORY"); configured != "" {
		for index, entry := range strings.Split(configured, ",") {
			directory := strings.TrimSpace(entry)
			if directory == "" {
				return nil, fmt.Errorf("content: MPQ_DIRECTORY entry %d is empty", index+1)
			}
			expanded, err := darkpaths.ExpandHost(directory)
			if err != nil {
				return nil, fmt.Errorf("content: expand MPQ directory %q: %w", directory, err)
			}
			directory = expanded
			info, err := os.Stat(directory)
			if err != nil {
				return nil, fmt.Errorf("content: inspect MPQ directory %q: %w", directory, err)
			}
			if !info.IsDir() {
				return nil, fmt.Errorf("content: MPQ path %q is not a directory", directory)
			}
			prefix := fmt.Sprintf("mpq-%d", index)
			layers = append(layers, Layer{Name: prefix + "-directory", FS: Directory(directory)})
			priority := []string{
				"patch_d2.mpq", "d2exp.mpq", "d2data.mpq", "d2char.mpq",
				"d2music.mpq", "d2sfx.mpq", "d2speech.mpq", "d2video.mpq",
				"d2xmusic.mpq", "d2xtalk.mpq", "d2xvideo.mpq",
			}
			for _, name := range priority {
				fileName := filepath.Join(directory, name)
				if _, err := os.Stat(fileName); err != nil {
					if os.IsNotExist(err) {
						continue
					}
					return nil, fmt.Errorf("content: inspect archive %q: %w", fileName, err)
				}
				archive, err := MPQ(fileName)
				if err != nil {
					return nil, err
				}
				layers = append(layers, Layer{Name: prefix + "-" + name, FS: archive})
			}
		}
	}
	return New(layers...)
}

// Layer is one named content source. Earlier layers have higher priority.
type Layer struct {
	Name string
	FS   fs.FS
}

// Source identifies the layer that resolved a path.
type Source struct {
	Layer string
	Path  string
}

// FS overlays content layers in deterministic priority order.
type FS struct {
	mu          sync.RWMutex
	layers      []Layer
	generation  uint64
	subscribers map[uint64]chan Change
	nextSubID   uint64
}

// Change reports that content at Path may resolve differently. Generation is
// monotonically increasing for the lifetime of the layered filesystem.
type Change struct {
	Path       string
	Generation uint64
}

// Generation returns the latest invalidation generation.
func (f *FS) Generation() uint64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.generation
}

// New constructs a layered filesystem. Earlier layers override later layers.
func New(layers ...Layer) (*FS, error) {
	result := &FS{subscribers: make(map[uint64]chan Change)}
	for i := len(layers) - 1; i >= 0; i-- {
		if err := result.MountFirst(layers[i]); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// MountFirst mounts a layer at the highest priority.
func (f *FS) MountFirst(layer Layer) error {
	if layer.Name == "" {
		return errors.New("content: layer name is required")
	}
	if layer.FS == nil {
		return fmt.Errorf("content: layer %q has a nil filesystem", layer.Name)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, existing := range f.layers {
		if existing.Name == layer.Name {
			return fmt.Errorf("content: layer %q is already mounted", layer.Name)
		}
	}
	f.layers = append([]Layer{layer}, f.layers...)
	return nil
}

// Unmount removes a named layer.
func (f *FS) Unmount(name string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i, layer := range f.layers {
		if layer.Name == name {
			f.layers = append(f.layers[:i], f.layers[i+1:]...)
			return true
		}
	}
	return false
}

// Layers returns mounted layer names in lookup order.
func (f *FS) Layers() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	result := make([]string, len(f.layers))
	for i, layer := range f.layers {
		result[i] = layer.Name
	}
	return result
}

// Open implements fs.FS.
func (f *FS) Open(name string) (fs.File, error) {
	clean, err := Normalize(name)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}
	for _, layer := range f.snapshot() {
		file, err := layer.FS.Open(clean)
		if err == nil {
			return file, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("content: open %q from layer %q: %w", clean, layer.Name, err)
		}
	}
	return nil, &fs.PathError{Op: "open", Path: clean, Err: fs.ErrNotExist}
}

// Resolve identifies the highest-priority layer containing name.
func (f *FS) Resolve(name string) (Source, error) {
	clean, err := Normalize(name)
	if err != nil {
		return Source{}, err
	}
	for _, layer := range f.snapshot() {
		file, err := layer.FS.Open(clean)
		if err == nil {
			_ = file.Close()
			return Source{Layer: layer.Name, Path: clean}, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return Source{}, fmt.Errorf("content: resolve %q from layer %q: %w", clean, layer.Name, err)
		}
	}
	return Source{}, &fs.PathError{Op: "resolve", Path: clean, Err: fs.ErrNotExist}
}

// ResolveSource exposes provenance to consumers that should not depend on the
// content package's concrete Source type.
func (f *FS) ResolveSource(name string) (layer string, resolvedPath string, err error) {
	resolved, err := f.Resolve(name)
	if err != nil {
		return "", "", err
	}
	return resolved.Layer, resolved.Path, nil
}

// ReadDir returns the union of a directory across all layers. A
// higher-priority entry shadows a lower-priority entry with the same name.
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

// Exists reports whether a normalized content path currently resolves.
func (f *FS) Exists(name string) bool {
	file, err := f.Open(name)
	if err != nil {
		return false
	}
	_ = file.Close()
	return true
}

// Walk enumerates the merged layered view in lexical order.
func (f *FS) Walk(root string, visit fs.WalkDirFunc) error {
	return fs.WalkDir(f, root, visit)
}

// List returns every mounted file below root, optionally filtered by a
// case-insensitive suffix. Ordinary filesystems are walked normally. Archive
// adapters may expose a flat path index when their format has no ReadDir API.
// Higher-priority layers win when the same path appears more than once.
func (f *FS) List(root, suffix string) ([]string, error) {
	cleanRoot, err := Normalize(root)
	if err != nil {
		return nil, err
	}
	lowerSuffix := strings.ToLower(strings.TrimSpace(suffix))
	seen := make(map[string]struct{})
	result := make([]string, 0)
	add := func(name string) {
		clean, normalizeErr := Normalize(name)
		if normalizeErr != nil || !pathBelow(cleanRoot, clean) {
			return
		}
		if lowerSuffix != "" && !strings.HasSuffix(strings.ToLower(clean), lowerSuffix) {
			return
		}
		key := strings.ToLower(clean)
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		result = append(result, clean)
	}

	for _, layer := range f.snapshot() {
		if indexed, ok := layer.FS.(interface{ Paths() []string }); ok {
			for _, name := range indexed.Paths() {
				add(name)
			}
			continue
		}
		walkErr := fs.WalkDir(layer.FS, cleanRoot, func(name string, entry fs.DirEntry, visitErr error) error {
			if visitErr != nil {
				return visitErr
			}
			if !entry.IsDir() {
				add(name)
			}
			return nil
		})
		if walkErr != nil && !errors.Is(walkErr, fs.ErrNotExist) {
			return nil, fmt.Errorf("content: list %q from layer %q: %w", cleanRoot, layer.Name, walkErr)
		}
	}
	sort.Strings(result)
	return result, nil
}

func pathBelow(root, name string) bool {
	return root == "." || name == root || strings.HasPrefix(name, root+"/")
}

// Invalidate publishes a normalized development-time content change. The FS
// does not cache bytes itself; consumers use this signal to invalidate decoded
// records, required Lua modules, and other derived resources.
func (f *FS) Invalidate(name string) (Change, error) {
	clean, err := Normalize(name)
	if err != nil {
		return Change{}, err
	}
	f.mu.Lock()
	f.generation++
	change := Change{Path: clean, Generation: f.generation}
	for _, subscriber := range f.subscribers {
		select {
		case subscriber <- change:
		default:
		}
	}
	f.mu.Unlock()
	return change, nil
}

// Subscribe returns a best-effort diagnostic change stream. Consumers must
// treat the latest generation as authoritative because slow subscribers may
// coalesce changes.
func (f *FS) Subscribe(buffer int) (<-chan Change, func()) {
	if buffer < 1 {
		buffer = 1
	}
	f.mu.Lock()
	id := f.nextSubID
	f.nextSubID++
	changes := make(chan Change, buffer)
	f.subscribers[id] = changes
	f.mu.Unlock()
	var once sync.Once
	return changes, func() {
		once.Do(func() {
			f.mu.Lock()
			delete(f.subscribers, id)
			close(changes)
			f.mu.Unlock()
		})
	}
}

// Normalize converts Diablo-style paths to valid fs.FS paths and rejects
// traversal outside a content root.
func Normalize(name string) (string, error) {
	name = strings.ReplaceAll(name, `\`, "/")
	name = strings.TrimLeft(name, "/")
	clean := path.Clean(name)
	if clean == "" {
		clean = "."
	}
	if clean == ".." || strings.HasPrefix(clean, "../") || !fs.ValidPath(clean) {
		return "", fmt.Errorf("content: invalid path %q", name)
	}
	return clean, nil
}

func (f *FS) snapshot() []Layer {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return append([]Layer(nil), f.layers...)
}
