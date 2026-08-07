package modruntime

import (
	"strings"

	"github.com/gravestench/dark-magic/internal/game/data/recovered"
	lua "github.com/yuin/gopher-lua"
)

type questCatalogSnapshotter interface {
	Snapshot() (recovered.Snapshot, error)
}

// QuestCatalogModule exposes normalized quest hierarchy and speech metadata
// recovered by Riiablo. It deliberately exposes string and sound identifiers;
// localization and audio capabilities remain responsible for resolving assets.
func QuestCatalogModule(catalog questCatalogSnapshotter) Module {
	return Module{Name: "dm.quest_catalog/v1", Help: documentedModule("Query recovered Diablo II quest and speech relationships.", map[string]CommandHelp{
		"quest":  commandHelp("dm.quest_catalog.quest(id)", "Return one normalized quest definition or nil."),
		"quests": commandHelp("dm.quest_catalog.quests([act])", "Return quests in canonical ID order, optionally filtered by zero-based act."),
		"speech": commandHelp("dm.quest_catalog.speech(sound)", "Return the localization key associated with a logical sound ID."),
	}), Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"quest": func(state *lua.LState) int {
				snapshot, err := catalog.Snapshot()
				if err != nil {
					return pushLuaError(state, err)
				}
				quest, found := snapshot.QuestsByID[state.CheckInt(1)]
				if !found {
					state.Push(lua.LNil)
					return 1
				}
				state.Push(questToLua(state, quest))
				return 1
			},
			"quests": func(state *lua.LState) int {
				snapshot, err := catalog.Snapshot()
				if err != nil {
					return pushLuaError(state, err)
				}
				filterAct := -1
				if state.GetTop() > 0 && state.Get(1) != lua.LNil {
					filterAct = state.CheckInt(1)
				}
				result := state.NewTable()
				for _, quest := range snapshot.Quests {
					if filterAct < 0 || quest.Act == filterAct {
						result.Append(questToLua(state, quest))
					}
				}
				state.Push(result)
				return 1
			},
			"speech": func(state *lua.LState) int {
				snapshot, err := catalog.Snapshot()
				if err != nil {
					return pushLuaError(state, err)
				}
				entry, found := snapshot.SpeechByName[strings.ToLower(state.CheckString(1))]
				if !found {
					state.Push(lua.LNil)
					return 1
				}
				result := state.NewTable()
				result.RawSetString("sound", lua.LString(entry.Sound))
				result.RawSetString("string_key", lua.LString(entry.StringKey))
				state.Push(result)
				return 1
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}

func questToLua(state *lua.LState, quest recovered.Quest) *lua.LTable {
	result := state.NewTable()
	result.RawSetString("id", lua.LNumber(quest.ID))
	result.RawSetString("name", lua.LString(quest.Name))
	result.RawSetString("act", lua.LNumber(quest.Act))
	result.RawSetString("order", lua.LNumber(quest.Order))
	result.RawSetString("visible", lua.LBool(quest.Visible))
	result.RawSetString("icon", lua.LString(quest.Icon))
	result.RawSetString("title_string_key", lua.LString(quest.TitleStringKey))
	if quest.PrerequisiteID != nil {
		result.RawSetString("prerequisite_id", lua.LNumber(*quest.PrerequisiteID))
	}
	stages := state.NewTable()
	for _, stage := range quest.Stages {
		entry := state.NewTable()
		entry.RawSetString("index", lua.LNumber(stage.Index))
		entry.RawSetString("string_key", lua.LString(stage.StringKey))
		alternates := state.NewTable()
		for _, key := range stage.Alternates {
			alternates.Append(lua.LString(key))
		}
		entry.RawSetString("alternates", alternates)
		stages.Append(entry)
	}
	result.RawSetString("stages", stages)
	return result
}
