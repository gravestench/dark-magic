package modruntime

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// ValidatePackageSyntax compiles every Lua file in a package without executing
// it. Network composition uses this side-effect-free gate before it stops a
// live component or changes the VFS. Full definition and lifecycle validation
// still happens in the production runtime afterward.
func ValidatePackageSyntax(source fs.FS) error {
	if source == nil {
		return errors.New("modruntime: package source is required")
	}

	var names []string

	if err := fs.WalkDir(source, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if !entry.IsDir() && path.Ext(name) == ".lua" {
			names = append(names, name)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("modruntime: inspect package Lua: %w", err)
	}

	sort.Strings(names)

	state := lua.NewState(lua.Options{SkipOpenLibs: true})
	defer state.Close()

	for _, name := range names {
		data, err := fs.ReadFile(source, name)
		if err != nil {
			return fmt.Errorf("modruntime: read Lua %q: %w", name, err)
		}

		if _, err := state.Load(bytes.NewReader(data), "@"+name); err != nil {
			return fmt.Errorf("modruntime: compile Lua %q: %w", name, err)
		}
	}

	return nil
}

// DiscoverDefinitions loads boot.lua and every Lua file below components/.
// Merely discovering a definition never starts it; desired state is reconciled
// separately by host.Manager.
func DiscoverDefinitions(
	ctx context.Context,
	runtime *Runtime,
	source fs.FS,
) ([]Definition, error) {
	names := []string{"boot.lua"}

	err := fs.WalkDir(
		source,
		"components",
		func(name string, entry fs.DirEntry, walkErr error) error {
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
		},
	)
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
			return nil, fmt.Errorf(
				"modruntime: duplicate component %q in %q and %q",
				definition.ID,
				previous,
				name,
			)
		}

		seen[definition.ID] = name
		definitions = append(definitions, definition)
	}

	return definitions, nil
}

// DiscoverOwnedDefinitions keeps lifecycle identity aligned with package
// namespace ownership. Shared VFS precedence is an asset mechanism; it must
// never let one package impersonate another package's managed component.
func DiscoverOwnedDefinitions(
	ctx context.Context,
	runtime *Runtime,
	source fs.FS,
	owner string,
) ([]Definition, error) {
	if owner == "" {
		return nil, errors.New("modruntime: package owner is required")
	}

	definitions, err := DiscoverDefinitions(ctx, runtime, source)
	if err != nil {
		return nil, err
	}

	prefix := owner + "."
	for _, definition := range definitions {
		if !strings.HasPrefix(definition.ID, prefix) {
			return nil, fmt.Errorf(
				"modruntime: package %q does not own component %q",
				owner,
				definition.ID,
			)
		}
	}

	return definitions, nil
}

// ValidateDefinitionDependencies prevents a package from gaining an undeclared
// lifecycle dependency on another package merely by guessing its component ID.
func ValidateDefinitionDependencies(
	definitions []Definition,
	owner string,
	dependencies []string,
) error {
	allowed := append([]string{owner}, dependencies...)

	for _, definition := range definitions {
		for _, dependency := range definition.DependsOn {
			owned := false

			for _, packageID := range allowed {
				if strings.HasPrefix(dependency, packageID+".") {
					owned = true
					break
				}
			}

			if !owned {
				return fmt.Errorf(
					"modruntime: component %q depends on undeclared package component %q",
					definition.ID,
					dependency,
				)
			}
		}
	}

	return nil
}

// ValidateDefinitionEntrypoints proves that every lifecycle ID advertised by a
// package manifest is actually defined by that package. A recipe must not pass
// distribution checks only to fail later while the host enables components.
func ValidateDefinitionEntrypoints(definitions []Definition, entrypoints ...[]string) error {
	available := make(map[string]bool, len(definitions))
	for _, definition := range definitions {
		available[definition.ID] = true
	}

	for _, group := range entrypoints {
		for _, id := range group {
			if !available[id] {
				return fmt.Errorf("modruntime: manifest entrypoint %q is not defined", id)
			}
		}
	}

	return nil
}

// ValidateDefinitionDomains rejects dependency closures that smuggle an
// opposite-domain manifest entrypoint into activation. Internal helper
// components may be shared, but a declared client entrypoint can never cause a
// declared authority entrypoint to run (or vice versa).
func ValidateDefinitionDomains(
	definitions []Definition,
	clientEntrypoints, authorityEntrypoints []string,
) error {
	byID := make(map[string]Definition, len(definitions))
	for _, definition := range definitions {
		byID[definition.ID] = definition
	}

	clients := stringSet(clientEntrypoints)
	authorities := stringSet(authorityEntrypoints)

	validate := func(root string, forbidden map[string]bool, domain string) error {
		visiting := make(map[string]bool)
		visited := make(map[string]bool)

		var walk func(string) error

		walk = func(id string) error {
			if forbidden[id] {
				return fmt.Errorf(
					"modruntime: %s entrypoint %q depends on opposite-domain entrypoint %q",
					domain,
					root,
					id,
				)
			}

			if visited[id] {
				return nil
			}

			if visiting[id] {
				return fmt.Errorf("modruntime: component dependency cycle includes %q", id)
			}

			definition, found := byID[id]
			if !found {
				return nil
			}

			visiting[id] = true

			for _, dependency := range definition.DependsOn {
				if err := walk(dependency); err != nil {
					return err
				}
			}

			delete(visiting, id)
			visited[id] = true

			return nil
		}

		return walk(root)
	}
	for _, root := range clientEntrypoints {
		if err := validate(root, authorities, "client"); err != nil {
			return err
		}
	}

	for _, root := range authorityEntrypoints {
		if err := validate(root, clients, "authority"); err != nil {
			return err
		}
	}

	return nil
}

// stringSet owns the string set step at this boundary, keeping its side effects and failure point explicit to
// callers.
func stringSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}

	return result
}
