package modruntime

import (
	"github.com/gravestench/dark-magic/internal/savecore"
	lua "github.com/yuin/gopher-lua"
)

// SaveModule exposes character metadata and selection without exposing save
// file ownership or mutable Go objects to Lua.
func SaveModule(store *savecore.Store) Module {
	return Module{Name: "dm.save/v1", Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"create_named": func(state *lua.LState) int {
				character, err := store.CreateNamed(state.CheckString(1), state.CheckString(2))
				if err != nil {
					state.Push(lua.LNil)
					state.Push(lua.LString(err.Error()))
					return 2
				}
				state.Push(lua.LString(character.ID))
				return 1
			},
			"create": func(state *lua.LState) int {
				err := store.Create(savecore.Character{ID: state.CheckString(1), Name: state.CheckString(2), Class: state.CheckString(3), Level: 1})
				if err != nil {
					state.Push(lua.LNil)
					state.Push(lua.LString(err.Error()))
					return 2
				}
				state.Push(lua.LTrue)
				return 1
			},
			"characters": func(state *lua.LState) int {
				result := state.NewTable()
				for _, character := range store.Characters() {
					entry := state.NewTable()
					entry.RawSetString("id", lua.LString(character.ID))
					entry.RawSetString("name", lua.LString(character.Name))
					entry.RawSetString("class", lua.LString(character.Class))
					entry.RawSetString("level", lua.LNumber(character.Level))
					result.Append(entry)
				}
				state.Push(result)
				return 1
			},
			"select": func(state *lua.LState) int {
				if err := store.Select(state.CheckString(1)); err != nil {
					state.Push(lua.LNil)
					state.Push(lua.LString(err.Error()))
					return 2
				}
				state.Push(lua.LTrue)
				return 1
			},
			"selected": func(state *lua.LState) int {
				character, ok := store.Selected()
				if !ok {
					state.Push(lua.LNil)
					return 1
				}
				result := state.NewTable()
				result.RawSetString("id", lua.LString(character.ID))
				result.RawSetString("name", lua.LString(character.Name))
				result.RawSetString("class", lua.LString(character.Class))
				result.RawSetString("level", lua.LNumber(character.Level))
				state.Push(result)
				return 1
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}
