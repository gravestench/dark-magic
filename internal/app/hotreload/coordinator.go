// Package hotreload coordinates VFS invalidation and transactional Lua
// component replacement for development-time source changes.
package hotreload

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/runtime/lua"
)

// ErrAuthoritativeReload explains why gameplay scripts are deliberately not
// swapped in the middle of a running deterministic session. A future explicit
// safe-boundary migration may replace this guard; ordinary presentation and
// lab scripts remain reloadable today.
var ErrAuthoritativeReload = errors.New("hotreload: authoritative d2legacy changes require a new session")

// Invalidator is the narrow generic-record cache seam used after VFS changes.
type Invalidator interface {
	Invalidate(string)
}

// Coordinator translates one changed virtual path into ordered invalidation and
// transactional component replacement. It never edits source files itself.
type Coordinator struct {
	content *content.FS
	source  fs.FS
	runtime *modruntime.Runtime
	manager *host.Manager
	records Invalidator

	mu          sync.Mutex
	definitions map[string]string
}

// knownDefinition pairs a component ID with the source that must continue to define it during a bulk reload.
type knownDefinition struct {
	id     string
	source string
}

// New indexes known definitions without starting watchers or reload work, leaving lifecycle ownership with the caller.
func New(
	contentFS *content.FS,
	runtime *modruntime.Runtime,
	manager *host.Manager,
	records Invalidator,
	definitions []modruntime.Definition,
) *Coordinator {
	sources := make(map[string]string, len(definitions))
	for _, definition := range definitions {
		sources[definition.ID] = definition.Source
	}

	return &Coordinator{
		content:     contentFS,
		source:      contentFS,
		runtime:     runtime,
		manager:     manager,
		records:     records,
		definitions: sources,
	}
}

// Reload invalidates derived content and transactionally refreshes affected
// managed definitions. Non-Lua files only emit VFS/record invalidation.
func (c *Coordinator) Reload(ctx context.Context, changed string) error {
	clean, err := content.Normalize(changed)
	if err != nil {
		return err
	}

	if authoritativePath(clean) {
		return fmt.Errorf("%w: %s", ErrAuthoritativeReload, clean)
	}

	if err := c.invalidateDerivedContent(clean); err != nil {
		return err
	}

	if definitionPath(clean) {
		return c.reloadDefinition(ctx, clean)
	}

	module, isModule := moduleName(clean)
	if !isModule {
		return nil
	}

	return c.reloadModuleConsumers(ctx, module)
}

// invalidateDerivedContent evicts caches in dependency order so record consumers never observe stale VFS data.
func (c *Coordinator) invalidateDerivedContent(clean string) error {
	if _, err := c.content.Invalidate(clean); err != nil {
		return err
	}

	if c.records != nil {
		c.records.Invalidate(clean)
	}

	return nil
}

// authoritativePath identifies deterministic gameplay sources that cannot be replaced within an active session.
func authoritativePath(clean string) bool {
	return clean == "components/d2legacy.lua" || strings.HasPrefix(clean, "lua/d2legacy/")
}

// definitionPath identifies Lua sources that directly declare managed component definitions.
func definitionPath(clean string) bool {
	return clean == "boot.lua" || strings.HasPrefix(clean, "components/") && path.Ext(clean) == ".lua"
}

// moduleName translates a reloadable Lua source path into the dotted name stored in the runtime module cache.
func moduleName(clean string) (string, bool) {
	if !strings.HasPrefix(clean, "lua/") || path.Ext(clean) != ".lua" {
		return "", false
	}

	module := strings.TrimSuffix(strings.TrimPrefix(clean, "lua/"), ".lua")

	return strings.ReplaceAll(module, "/", "."), true
}

// reloadModuleConsumers clears the changed module before rebuilding every definition that may require it.
func (c *Coordinator) reloadModuleConsumers(ctx context.Context, module string) error {
	if err := c.runtime.InvalidateModule(ctx, module); err != nil {
		return err
	}

	return c.reloadAll(ctx)
}

// reloadDefinition loads one definition and registers new IDs while transactionally replacing IDs already in use.
func (c *Coordinator) reloadDefinition(ctx context.Context, source string) error {
	definition, err := modruntime.LoadDefinition(ctx, c.runtime, c.source, source)
	if err != nil {
		return err
	}

	if err := c.installDefinition(ctx, definition); err != nil {
		return err
	}

	c.rememberDefinition(definition.ID, source)

	return nil
}

// installDefinition selects registration or replacement without publishing a failed definition to the source index.
func (c *Coordinator) installDefinition(ctx context.Context, definition modruntime.Definition) error {
	if c.knowsDefinition(definition.ID) {
		return c.manager.Replace(ctx, definition.Managed())
	}

	return c.manager.Register(definition.Managed())
}

// knowsDefinition reads the source index under its mutex so reload requests may safely arrive concurrently.
func (c *Coordinator) knowsDefinition(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, known := c.definitions[id]

	return known
}

// rememberDefinition updates the source index only after the manager accepts the loaded definition.
func (c *Coordinator) rememberDefinition(id, source string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.definitions[id] = source
}

// reloadAll replaces a stable, ID-sorted snapshot so map iteration cannot make reload order nondeterministic.
func (c *Coordinator) reloadAll(ctx context.Context) error {
	var errs []error

	for _, known := range c.definitionSnapshot() {
		if err := c.reloadKnownDefinition(ctx, known); err != nil {
			errs = append(errs, fmt.Errorf("hotreload: reload %q: %w", known.id, err))
		}
	}

	return errors.Join(errs...)
}

// definitionSnapshot copies the mutable source index while locked and returns definitions in deterministic ID order.
func (c *Coordinator) definitionSnapshot() []knownDefinition {
	c.mu.Lock()
	defer c.mu.Unlock()

	definitions := make([]knownDefinition, 0, len(c.definitions))
	for id, source := range c.definitions {
		definitions = append(definitions, knownDefinition{id: id, source: source})
	}
	// Sorting under the same lock keeps the returned snapshot internally consistent with one index state.
	sort.Slice(definitions, func(i, j int) bool {
		return definitions[i].id < definitions[j].id
	})

	return definitions
}

// reloadKnownDefinition enforces stable component identity before replacing an indexed definition.
func (c *Coordinator) reloadKnownDefinition(ctx context.Context, known knownDefinition) error {
	definition, err := modruntime.LoadDefinition(ctx, c.runtime, c.source, known.source)
	if err != nil {
		return err
	}

	if definition.ID != known.id {
		return fmt.Errorf(
			"hotreload: %q changed component ID from %q to %q",
			known.source,
			known.id,
			definition.ID,
		)
	}

	return c.manager.Replace(ctx, definition.Managed())
}
