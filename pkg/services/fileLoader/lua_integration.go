package fileLoader

import (
	lua "github.com/yuin/gopher-lua"

	luaService "github.com/gravestench/dark-magic/pkg/services/luaManager"
)

const LuaApiKey = "loader"

var _ luaService.LuaPlugin = &Service{}

func (s *Service) ExportToLua(state *lua.LState, rootTable *lua.LTable) {
	table := state.NewTable()

	s.bindMethods(state, table)

	rootTable.RawSetString(LuaApiKey, table)
}

func (s *Service) bindMethods(state *lua.LState, table *lua.LTable) {
	fnMap := map[string]lua.LGFunction{
		"Groups":                s.luaGroupsGet,
		"AddSource":             s.luaSourceAdd,
		"AddSourceToGroup":      s.luaAddSourceToGroup,
		"RemoveSource":          s.luaRemoveSource,
		"RemoveSourceFromGroup": s.luaRemoveSourceFromGroup,
		"RemoveGroup":           s.luaRemoveGroup,
		// TODO: expose Open method of the composite fs loader service in lua
		//"Open": s.luaOpen,
	}

	for key, fn := range fnMap {
		table.RawSetString(key, state.NewFunction(fn))
	}
}

// TODO: figure out clean way of exposing the Open method of fs.FS in lua
//func convertFileHandleToLuaTable(f fs.File, state *lua.LState) (*lua.LTable, error) {
//	table := state.NewTable()
//
//	table.RawSetString("Read", state.NewFunction(func(L *lua.LState) int {
//		if L.GetTop() < 1 {
//			return 0
//		}
//
//		buffer := L.CheckTable(0)
//		buffer.ForEach(func(luaIndex, entry lua.LValue) {
//
//		})
//
//		if err := s.RemoveSource(src); err == nil {
//			L.Push(lua.LString(err.Error()))
//			return 1
//		}
//
//		return 0
//	}))
//}

func (s *Service) UnexportFromLua(state *lua.LState, rootTable *lua.LTable) {
	state.SetField(rootTable, LuaApiKey, lua.LNil)
}

func (s *Service) luaGroupsGet(L *lua.LState) int {
	if L.GetTop() != 0 { // we do not want any args
		return 0
	}

	result := L.NewTable()
	for _, group := range s.Groups() {
		result.Append(lua.LString(group))
	}

	L.Push(result)

	return 1
}

func (s *Service) luaSourceAdd(L *lua.LState) int {
	if L.GetTop() != 1 {
		return 0
	}

	srcURI := L.CheckString(0)
	src := NewSource(srcURI)

	if err := s.AddSource(src); err == nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

func (s *Service) luaAddSourceToGroup(L *lua.LState) int {
	if L.GetTop() != 2 {
		return 0
	}

	srcURI := L.CheckString(0)
	group := L.CheckString(1)

	src := NewSource(srcURI)

	if err := s.AddSourceToGroup(src, group); err == nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

func (s *Service) luaRemoveSource(L *lua.LState) int {
	if L.GetTop() != 1 {
		return 0
	}

	srcURI := L.CheckString(0)
	src := NewSource(srcURI)

	if err := s.RemoveSource(src); err == nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

func (s *Service) luaRemoveSourceFromGroup(L *lua.LState) int {
	if L.GetTop() != 2 {
		return 0
	}

	srcURI := L.CheckString(0)
	group := L.CheckString(1)

	src := NewSource(srcURI)

	if err := s.RemoveSourceFromGroup(src, group); err == nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

func (s *Service) luaRemoveGroup(L *lua.LState) int {
	if L.GetTop() != 1 {
		return 0
	}

	group := L.CheckString(0)

	if err := s.RemoveGroup(group); err == nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	return 0
}

func (s *Service) luaOpen(L *lua.LState) int {
	//// file path/uri is required
	//numArgs := L.GetTop()
	//if numArgs < 1 {
	//	return 0
	//}
	//
	//// make slice of strings
	//args := make([]string, 0)
	//for idx := 0; idx < numArgs; idx++ {
	//	arg := L.CheckString(idx)
	//	args = append(args, arg)
	//}
	//
	//// first arg is path/uri
	//// all other args are optional, which fs groups to look in
	//path := args[0]
	//groups := make([]string, 0)
	//if len(args) > 1 {
	//	groups = args[1:]
	//}
	//
	//// open the file
	//fh, err := s.Open(path, groups...)
	//if err != nil {
	//	L.Push(lua.LString(err.Error()))
	//	return 1
	//}
	//
	//// convert file handle to lua table
	//fht, err := convertFileHandleToLuaTable(fh, state)
	//if err != nil {
	//	L.Push(lua.LString(err.Error()))
	//	return 1
	//}
	//
	//// yield the file handle table
	//L.Push(fht)

	return 0
}
