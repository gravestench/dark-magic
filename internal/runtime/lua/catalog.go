package modruntime

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"
)

// DiscoverDefinitions loads boot.lua and every Lua file below components/.
// Merely discovering a definition never starts it; desired state is reconciled
// separately by host.Manager.
func DiscoverDefinitions(ctx context.Context, runtime *Runtime, source fs.FS) ([]Definition, error) {
	names := []string{"boot.lua"}
	err := fs.WalkDir(source, "components", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if name == "components" && errors.Is(walkErr, fs.ErrNotExist) {
				return fs.SkipDir
			}
			return walkErr
		}
		if !entry.IsDir() && path.Ext(name) == ".lua" {
			names = append(names, name)
		}
		return nil
	})
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("modruntime: discover definitions: %w", err)
	}
	sort.Strings(names[1:])

	definitions := make([]Definition, 0, len(names))
	seen := make(map[string]string, len(names))
	for _, name := range names {
		definition, err := LoadDefinition(ctx, runtime, source, name)
		if err != nil {
			return nil, err
		}
		if previous, exists := seen[definition.ID]; exists {
			return nil, fmt.Errorf("modruntime: duplicate component %q in %q and %q", definition.ID, previous, name)
		}
		seen[definition.ID] = name
		definitions = append(definitions, definition)
	}
	return definitions, nil
}
