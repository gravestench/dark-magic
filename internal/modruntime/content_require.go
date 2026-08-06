package modruntime

import (
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// ContentRequire installs a Lua require searcher backed by source. A module
// such as darkmagic.screens.loading resolves to
// <root>/darkmagic/screens/loading.lua.
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
			state.Push(function)
			return 1
		}))
		return nil
	}}
}
