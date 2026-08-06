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

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/host"
	"github.com/gravestench/dark-magic/internal/modruntime"
)

type Invalidator interface {
	Invalidate(string)
}

type Coordinator struct {
	content *content.FS
	source  fs.FS
	runtime *modruntime.Runtime
	manager *host.Manager
	records Invalidator

	mu          sync.Mutex
	definitions map[string]string
}

func New(contentFS *content.FS, runtime *modruntime.Runtime, manager *host.Manager, records Invalidator, definitions []modruntime.Definition) *Coordinator {
	sources := make(map[string]string, len(definitions))
	for _, definition := range definitions {
		sources[definition.ID] = definition.Source
	}
	return &Coordinator{content: contentFS, source: contentFS, runtime: runtime, manager: manager, records: records, definitions: sources}
}

// Reload invalidates derived content and transactionally refreshes affected
// managed definitions. Non-Lua files only emit VFS/record invalidation.
func (c *Coordinator) Reload(ctx context.Context, changed string) error {
	clean, err := content.Normalize(changed)
	if err != nil {
		return err
	}
	if _, err := c.content.Invalidate(clean); err != nil {
		return err
	}
	if c.records != nil {
		c.records.Invalidate(clean)
	}
	if clean == "boot.lua" || strings.HasPrefix(clean, "components/") && path.Ext(clean) == ".lua" {
		return c.reloadDefinition(ctx, clean)
	}
	if strings.HasPrefix(clean, "lua/") && path.Ext(clean) == ".lua" {
		module := strings.TrimSuffix(strings.TrimPrefix(clean, "lua/"), ".lua")
		module = strings.ReplaceAll(module, "/", ".")
		if err := c.runtime.InvalidateModule(ctx, module); err != nil {
			return err
		}
		return c.reloadAll(ctx)
	}
	return nil
}

func (c *Coordinator) reloadDefinition(ctx context.Context, source string) error {
	definition, err := modruntime.LoadDefinition(ctx, c.runtime, c.source, source)
	if err != nil {
		return err
	}
	c.mu.Lock()
	_, known := c.definitions[definition.ID]
	c.mu.Unlock()
	if !known {
		if err := c.manager.Register(definition.Managed()); err != nil {
			return err
		}
		c.mu.Lock()
		c.definitions[definition.ID] = source
		c.mu.Unlock()
		return nil
	}
	if err := c.manager.Replace(ctx, definition.Managed()); err != nil {
		return err
	}
	c.mu.Lock()
	c.definitions[definition.ID] = source
	c.mu.Unlock()
	return nil
}

func (c *Coordinator) reloadAll(ctx context.Context) error {
	c.mu.Lock()
	ids := make([]string, 0, len(c.definitions))
	for id := range c.definitions {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	sources := make(map[string]string, len(c.definitions))
	for id, source := range c.definitions {
		sources[id] = source
	}
	c.mu.Unlock()
	var errs []error
	for _, id := range ids {
		definition, err := modruntime.LoadDefinition(ctx, c.runtime, c.source, sources[id])
		if err == nil && definition.ID != id {
			err = fmt.Errorf("hotreload: %q changed component ID from %q to %q", sources[id], id, definition.ID)
		}
		if err == nil {
			err = c.manager.Replace(ctx, definition.Managed())
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("hotreload: reload %q: %w", id, err))
		}
	}
	return errors.Join(errs...)
}
