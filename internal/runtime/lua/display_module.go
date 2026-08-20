package modruntime

import lua "github.com/yuin/gopher-lua"

// DisplayModule exposes the current renderable surface without exposing a
// platform window handle. Responsive scenes use it to lay out native-pixel
// workspaces while input remains in the same coordinate space.
func DisplayModule(size func() (width, height int)) Module {
	help := documentedModule("Inspect the current logical or native render surface.", map[string]CommandHelp{
		"size": commandHelp("engine.display.size()", "Return the current render surface width and height."),
	})

	return Module{Name: "engine.display/v1", Help: help, Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"size": func(state *lua.LState) int {
				width, height := 0, 0
				if size != nil {
					width, height = size()
				}
				state.Push(lua.LNumber(width))
				state.Push(lua.LNumber(height))
				return 2
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}
