package modruntime

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/gravestench/dark-magic/internal/assets/decode"
	"github.com/gravestench/dark-magic/internal/content"
	lua "github.com/yuin/gopher-lua"
)

// VFSModule exposes read-only layered content access to Lua.
func VFSModule(source *content.FS) Module {
	return Module{Name: "dm.vfs/v1", Help: documentedModule("Read assets from the layered virtual filesystem.", map[string]CommandHelp{
		"read_text": commandHelp("dm.vfs.read_text(path)", "Read an asset as text."),
		"read":      commandHelp("dm.vfs.read(path)", "Read an asset as binary data."),
		"source":    commandHelp("dm.vfs.source(path)", "Describe the layer and resolved path supplying an asset."),
		"list":      commandHelp("dm.vfs.list(prefix[, suffix])", "List mounted asset paths below a prefix, optionally filtered by case-insensitive suffix."),
	}), Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"read_text": func(state *lua.LState) int {
				name := state.CheckString(1)
				data, err := fs.ReadFile(source, name)
				if err == nil {
					var text string
					text, err = assetdecode.Text(data)
					if err == nil {
						state.Push(lua.LString(text))
						return 1
					}
				}
				state.Push(lua.LNil)
				state.Push(lua.LString(err.Error()))
				return 2
			},
			"read": func(state *lua.LState) int {
				name := state.CheckString(1)
				data, err := fs.ReadFile(source, name)
				if err != nil {
					state.Push(lua.LNil)
					state.Push(lua.LString(err.Error()))
					return 2
				}
				state.Push(lua.LString(string(data)))
				return 1
			},
			"source": func(state *lua.LState) int {
				name := state.CheckString(1)
				resolved, err := source.Resolve(name)
				if err != nil {
					state.Push(lua.LNil)
					state.Push(lua.LString(err.Error()))
					return 2
				}
				result := state.NewTable()
				result.RawSetString("layer", lua.LString(resolved.Layer))
				result.RawSetString("path", lua.LString(resolved.Path))
				state.Push(result)
				return 1
			},
			"list": func(state *lua.LState) int {
				prefix := strings.TrimSpace(state.CheckString(1))
				suffix := strings.ToLower(strings.TrimSpace(state.OptString(2, "")))
				paths, err := source.List(prefix, suffix)
				if err != nil {
					state.Push(lua.LNil)
					state.Push(lua.LString(err.Error()))
					return 2
				}
				result := state.NewTable()
				for _, path := range paths {
					result.Append(lua.LString(path))
				}
				state.Push(result)
				return 1
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		module.RawSetString("name", lua.LString(fmt.Sprintf("%s", "dm.vfs/v1")))
		state.Push(module)
		return 1
	}}
}
