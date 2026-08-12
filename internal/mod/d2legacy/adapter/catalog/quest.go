// Package catalog exposes immutable, recovered Diablo II relationships to the
// d2legacy Lua mod. Parsing the source files remains a Go data-boundary job;
// deciding how quests and maps use these relationships belongs to d2legacy.
package catalog

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gravestench/dark-magic/internal/mod/d2legacy/data/recovered"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	lua "github.com/yuin/gopher-lua"
)

type snapshotter interface {
	Snapshot() (recovered.Snapshot, error)
}

type textLookup interface {
	Text(string) (string, error)
}

// QuestModule provides read-only recovered quest and speech records. It does
// not execute quest policy; the Lua mod decides visibility, progression,
// rewards, and presentation from these immutable facts.
func QuestModule(source snapshotter, locale ...textLookup) modruntime.Module {
	var text textLookup
	if len(locale) > 0 {
		text = locale[0]
	}
	return modruntime.Module{
		Name: "d2legacy.quest_catalog/v1",
		Help: modruntime.ModuleHelp{Summary: "Read recovered Diablo II quest and speech relationships.", Commands: map[string]modruntime.CommandHelp{
			"quest":  {Usage: "d2legacy.quest_catalog.quest(id)", Summary: "Return one normalized quest definition or nil."},
			"quests": {Usage: "d2legacy.quest_catalog.quests([act])", Summary: "Return quests in canonical ID order, optionally filtered by zero-based act."},
			"speech": {Usage: "d2legacy.quest_catalog.speech(sound)", Summary: "Return the localization key for a logical sound ID."},
			"dialog": {Usage: "d2legacy.quest_catalog.dialog(sound)", Summary: "Resolve a speech ID into timed localized text."},
		}},
		Loader: func(state *lua.LState) int {
			module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
				"quest": func(state *lua.LState) int {
					snapshot, err := source.Snapshot()
					if err != nil {
						return raise(state, err)
					}
					quest, found := snapshot.QuestsByID[state.CheckInt(1)]
					if !found {
						state.Push(lua.LNil)
						return 1
					}
					state.Push(questTable(state, quest))
					return 1
				},
				"quests": func(state *lua.LState) int {
					snapshot, err := source.Snapshot()
					if err != nil {
						return raise(state, err)
					}
					act := -1
					if state.GetTop() > 0 && state.Get(1) != lua.LNil {
						act = state.CheckInt(1)
					}
					result := state.NewTable()
					for _, quest := range snapshot.Quests {
						if act < 0 || quest.Act == act {
							result.Append(questTable(state, quest))
						}
					}
					state.Push(result)
					return 1
				},
				"speech": func(state *lua.LState) int {
					snapshot, err := source.Snapshot()
					if err != nil {
						return raise(state, err)
					}
					entry, found := snapshot.SpeechByName[strings.ToLower(state.CheckString(1))]
					if !found {
						state.Push(lua.LNil)
						return 1
					}
					state.Push(speechTable(state, entry.Sound, entry.StringKey))
					return 1
				},
				"dialog": func(state *lua.LState) int {
					if text == nil {
						return raise(state, fmt.Errorf("quest catalog: no localization source"))
					}
					snapshot, err := source.Snapshot()
					if err != nil {
						return raise(state, err)
					}
					entry, found := snapshot.SpeechByName[strings.ToLower(state.CheckString(1))]
					if !found {
						state.Push(lua.LNil)
						return 1
					}
					localized, err := text.Text(entry.StringKey)
					if err != nil {
						return raise(state, err)
					}
					rate, body, err := ParseDialogueText(localized)
					if err != nil {
						return raise(state, fmt.Errorf("quest catalog: dialog %q: %w", entry.Sound, err))
					}
					result := speechTable(state, entry.Sound, entry.StringKey)
					result.RawSetString("text", lua.LString(body))
					result.RawSetString("scroll_lines_per_second", lua.LNumber(rate))
					state.Push(result)
					return 1
				},
			})
			module.RawSetString("api", lua.LNumber(1))
			state.Push(module)
			return 1
		},
	}
}

// ParseDialogueText decodes a legacy localized speech payload: the first line
// is a 60 Hz scroll-rate value and the remaining lines are displayed text.
func ParseDialogueText(value string) (float64, string, error) {
	header, body, found := strings.Cut(strings.ReplaceAll(value, "\r\n", "\n"), "\n")
	if !found || strings.TrimSpace(body) == "" {
		return 0, "", fmt.Errorf("localized value has no timing line and body")
	}
	timing, err := strconv.ParseFloat(strings.TrimSpace(header), 64)
	if err != nil || timing < 0 {
		return 0, "", fmt.Errorf("invalid timing value %q", header)
	}
	return timing / 60, body, nil
}

func raise(state *lua.LState, err error) int { state.RaiseError("%v", err); return 0 }

func speechTable(state *lua.LState, sound, key string) *lua.LTable {
	result := state.NewTable()
	result.RawSetString("sound", lua.LString(sound))
	result.RawSetString("string_key", lua.LString(key))
	return result
}

func questTable(state *lua.LState, quest recovered.Quest) *lua.LTable {
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
