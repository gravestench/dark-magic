package modruntime

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

// ContentRequire installs a Lua require searcher backed by source. A module
// such as example.screens.loading resolves beneath the configured content root
// using ordinary dotted Lua module names.
func ContentRequire(source fs.FS, root string) Installer {
	return Installer{Name: "content.require", Install: func(state *lua.LState) error {
		packageValue := state.GetGlobal("package")
		packageTable, ok := packageValue.(*lua.LTable)
		if !ok {
			return fmt.Errorf("Lua package table is unavailable")
		}
		loadersValue := packageTable.RawGetString("loaders")
		loaders, ok := loadersValue.(*lua.LTable)
		if !ok {
			return fmt.Errorf("Lua package.loaders table is unavailable")
		}
		loaders.Append(state.NewFunction(func(state *lua.LState) int {
			module := state.CheckString(1)
			fileName := path.Join(root, strings.ReplaceAll(module, ".", "/")+".lua")
			data, err := fs.ReadFile(source, fileName)
			if err != nil {
				state.Push(lua.LString(fmt.Sprintf("\n\tno content module %q: %v", fileName, err)))
				return 1
			}
			function, err := state.Load(bytes.NewReader(data), "@"+fileName)
			if err != nil {
				state.RaiseError("compiling content module %q: %v", fileName, err)
				return 0
			}
			state.SetFEnv(function, isolatedEnvironment(state))
			state.Push(function)
			return 1
		}))
		return nil
	}}
}

// PackageRequire resolves a module only inside the namespace of the package
// whose ID prefixes it. A package cannot replace another package's private Lua
// merely by shipping the same module-relative path.
func PackageRequire(source fs.FS, packageIDs []string) Installer {
	registry := NewPackageRegistry(packageIDs)
	return PackageRequireRegistry(source, registry)
}

type PackageRegistry struct {
	mu  sync.RWMutex
	ids []string
}

func NewPackageRegistry(packageIDs []string) *PackageRegistry {
	registry := &PackageRegistry{}
	registry.Replace(packageIDs)
	return registry
}

func (registry *PackageRegistry) Replace(packageIDs []string) {
	ids := append([]string(nil), packageIDs...)
	sort.Slice(ids, func(first, second int) bool { return len(ids[first]) > len(ids[second]) })
	registry.mu.Lock()
	registry.ids = ids
	registry.mu.Unlock()
}

func (registry *PackageRegistry) owner(module string) string {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	for _, id := range registry.ids {
		if module == id || strings.HasPrefix(module, id+".") {
			return id
		}
	}
	return ""
}

func PackageRequireRegistry(source fs.FS, registry *PackageRegistry) Installer {
	return Installer{Name: "content.require", Install: func(state *lua.LState) error {
		if registry == nil {
			return fmt.Errorf("Lua package registry is required")
		}
		packageValue := state.GetGlobal("package")
		packageTable, ok := packageValue.(*lua.LTable)
		if !ok {
			return fmt.Errorf("Lua package table is unavailable")
		}
		loaders, ok := packageTable.RawGetString("loaders").(*lua.LTable)
		if !ok {
			return fmt.Errorf("Lua package.loaders table is unavailable")
		}
		loaders.Append(state.NewFunction(func(state *lua.LState) int {
			module := state.CheckString(1)
			owner := registry.owner(module)
			if owner == "" {
				state.Push(lua.LString(fmt.Sprintf("\n\tno package owns Lua module %q", module)))
				return 1
			}
			fileName := path.Join("mods", owner, "lua", strings.ReplaceAll(module, ".", "/")+".lua")
			data, err := fs.ReadFile(source, fileName)
			if err != nil {
				state.Push(lua.LString(fmt.Sprintf("\n\tno package module %q: %v", fileName, err)))
				return 1
			}
			function, err := state.Load(bytes.NewReader(data), "@"+fileName)
			if err != nil {
				state.RaiseError("compiling package module %q: %v", fileName, err)
				return 0
			}
			state.SetFEnv(function, isolatedEnvironment(state))
			state.Push(function)
			return 1
		}))
		return nil
	}}
}

// InvalidatePackageModules removes cached require results owned by packageIDs.
// A package update or session-lock change must not keep executing modules from
// an archive that is no longer mounted merely because Lua required it earlier.
func InvalidatePackageModules(ctx context.Context, runtime *Runtime, packageIDs ...string) error {
	if runtime == nil {
		return fmt.Errorf("modruntime: Lua runtime is required")
	}
	owners := NewPackageRegistry(packageIDs)
	return runtime.Run(ctx, func(state *lua.LState) error {
		packageTable, ok := state.GetGlobal("package").(*lua.LTable)
		if !ok {
			return fmt.Errorf("Lua package table is unavailable")
		}
		loaded, ok := packageTable.RawGetString("loaded").(*lua.LTable)
		if !ok {
			return fmt.Errorf("Lua package.loaded table is unavailable")
		}
		var remove []lua.LValue
		loaded.ForEach(func(key, _ lua.LValue) {
			if name, ok := key.(lua.LString); ok && owners.owner(string(name)) != "" {
				remove = append(remove, key)
			}
		})
		for _, key := range remove {
			loaded.RawSet(key, lua.LNil)
		}
		return nil
	})
}
