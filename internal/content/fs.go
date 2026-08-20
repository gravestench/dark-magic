// Package content provides Dark Magic's layered game-content filesystem.
package content

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"sync"
)

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

// Close releases archive-backed layers. Directory and embedded layers are
// harmless no-ops. Callers that construct a long-lived environment stack own
// this lifecycle.
func (f *FS) Close() error {
	if f == nil {
		return nil
	}

	// Detach the layers before closing them so new lookups observe an empty stack and close callbacks never run while
	// the mount lock is held. Lookups that already own a snapshot retain the existing concurrent-close behavior.
	f.mu.Lock()
	layers := append([]Layer(nil), f.layers...)
	f.layers = nil
	f.mu.Unlock()

	var result error
	for _, layer := range layers {
		result = errors.Join(result, Close(layer.FS))
	}

	return result
}

// New constructs a layered filesystem. Earlier layers override later layers.
func New(layers ...Layer) (*FS, error) {
	result := &FS{subscribers: make(map[uint64]chan Change)}
	// MountFirst prepends, so reverse construction retains the caller's earlier-is-higher-priority ordering.
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

// snapshot returns a defensive copy so layer I/O never runs while the mount lock is held.
func (f *FS) snapshot() []Layer {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return append([]Layer(nil), f.layers...)
}
