package modruntime

import (
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

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
	ids := append([]string(nil), packageIDs...)
	sort.Slice(ids, func(first, second int) bool { return len(ids[first]) > len(ids[second]) })
	return Installer{Name: "content.require", Install: func(state *lua.LState) error {
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
			owner := ""
			for _, id := range ids {
				if module == id || strings.HasPrefix(module, id+".") {
					owner = id
					break
				}
			}
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
