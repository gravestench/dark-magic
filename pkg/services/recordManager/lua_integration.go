package recordManager

import (
	"sort"

	lua "github.com/yuin/gopher-lua"
)

const LuaAPIKey = "records"

func (s *Service) ExportToLua(state *lua.LState, rootTable *lua.LTable) {
	table := state.NewTable()
	state.SetField(table, "Load", state.NewFunction(s.luaLoadRecords))
	state.SetField(table, "Reload", state.NewFunction(s.luaReloadRecords))
	state.SetField(table, "Loaded", state.NewFunction(s.luaLoadedRecords))
	state.SetField(rootTable, LuaAPIKey, table)
}

func (s *Service) UnexportFromLua(state *lua.LState, rootTable *lua.LTable) {
	state.SetField(rootTable, LuaAPIKey, lua.LNil)
}

func (s *Service) luaLoadRecords(L *lua.LState) int {
	path := L.CheckString(1)
	records, err := s.loadGenericRecords(path)
	return pushRecords(L, records, err)
}

func (s *Service) luaReloadRecords(L *lua.LState) int {
	path := L.CheckString(1)
	records, err := s.reloadGenericRecords(path)
	return pushRecords(L, records, err)
}

func pushRecords(L *lua.LState, records []map[string]string, err error) int {
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	result := L.NewTable()
	for idx, record := range records {
		row := L.NewTable()
		keys := make([]string, 0, len(record))
		for key := range record {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			row.RawSetString(key, lua.LString(record[key]))
		}
		result.RawSetInt(idx+1, row)
	}
	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

func (s *Service) luaLoadedRecords(L *lua.LState) int {
	s.recordMux.RLock()
	paths := make([]string, 0, len(s.recordRegistry))
	for path := range s.recordRegistry {
		paths = append(paths, path)
	}
	s.recordMux.RUnlock()
	sort.Strings(paths)
	result := L.NewTable()
	for idx, path := range paths {
		result.RawSetInt(idx+1, lua.LString(path))
	}
	L.Push(result)
	return 1
}
