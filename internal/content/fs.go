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
)

// FromEnvironment constructs the production content stack. User content wins,
// followed by the Dark Magic shim, Diablo II patches/expansion data, and base
// archives. Missing optional directories and archives are skipped.
func FromEnvironment() (*FS, error) {
	layers := make([]Layer, 0, 16)
	if mods := os.Getenv("DARK_MAGIC_MOD_DIRECTORY"); mods != "" {
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
	if directory := os.Getenv("MPQ_DIRECTORY"); directory != "" {
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
			layers = append(layers, Layer{Name: name, FS: archive})
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
	mu     sync.RWMutex
	layers []Layer
}

// New constructs a layered filesystem. Earlier layers override later layers.
func New(layers ...Layer) (*FS, error) {
	result := &FS{}
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
