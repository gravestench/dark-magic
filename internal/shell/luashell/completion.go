package luashell

import (
	"context"
	"sort"
	"strings"

	"github.com/gravestench/dark-magic/internal/shell"
	glua "github.com/yuin/gopher-lua"
)

var keywords = []string{
	"and", "break", "do", "else", "elseif", "end", "false", "for", "function",
	"goto", "if", "in", "local", "nil", "not", "or", "repeat", "return", "then",
	"true", "until", "while",
}

// Complete inspects one persistent scope without executing user source. Runtime
// serialization keeps completion from racing evaluation or scope teardown.
func (e *Evaluator) Complete(ctx context.Context, source string) ([]shell.Candidate, error) {
	var result []shell.Candidate

	err := e.runtime.RunScoped(ctx, e.scope, func(state *glua.LState) error {
		result = e.completionCandidates(state, source)

		return nil
	})

	return result, err
}

// completionCandidates builds the visible namespace, narrows it for dotted
// access, and emits deterministic prefix matches with descriptive details.
func (e *Evaluator) completionCandidates(state *glua.LState, source string) []shell.Candidate {
	environment := e.environment(state)
	token := completionToken(source)
	seen := e.topLevelCandidates(state, environment)

	if dot := strings.LastIndexByte(token, '.'); dot >= 0 {
		base := token[:dot]
		member := token[dot+1:]

		if members, found := e.memberCandidates(state, environment, base); found {
			seen = members
		}

		token = base + "." + member
	}

	result := make([]shell.Candidate, 0, len(seen))
	for value, detail := range seen {
		if strings.HasPrefix(value, token) {
			result = append(result, shell.Candidate{Value: value, Detail: detail})
		}
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Value < result[j].Value })

	return result
}

// topLevelCandidates combines language keywords, permitted modules, persistent
// session globals, and remaining runtime globals without overwriting closer details.
func (e *Evaluator) topLevelCandidates(
	state *glua.LState,
	environment *glua.LTable,
) map[string]string {
	seen := make(map[string]string)
	for _, keyword := range keywords {
		seen[keyword] = "keyword"
	}

	for _, module := range e.modules {
		seen[module] = "module"
	}

	environment.ForEach(func(key, _ glua.LValue) {
		seen[key.String()] = "session"
	})

	if global, ok := state.Get(glua.GlobalsIndex).(*glua.LTable); ok {
		global.ForEach(func(key, _ glua.LValue) {
			if _, exists := seen[key.String()]; !exists {
				seen[key.String()] = "global"
			}
		})
	}

	return seen
}

// memberCandidates chooses the specialized Dark Magic/module schema when
// possible and otherwise inspects a raw table or userdata method table.
func (e *Evaluator) memberCandidates(
	state *glua.LState,
	environment *glua.LTable,
	base string,
) (map[string]string, bool) {
	if base == "dm" || base == "darkmagic" {
		return e.rootMemberCandidates(base), true
	}

	if module, found := e.completionModule(base); found {
		return e.moduleMemberCandidates(environment, base, module), true
	}

	return e.rawMemberCandidates(state, environment, base)
}

// rootMemberCandidates advertises policy-permitted aliases and stable root
// helpers without forcing any module to load.
func (e *Evaluator) rootMemberCandidates(base string) map[string]string {
	seen := make(map[string]string)

	for alias, module := range e.aliases {
		detail := e.help[module].Summary
		if detail == "" {
			detail = "Dark Magic capability"
		}

		seen[base+"."+alias] = detail
	}

	for _, helper := range []string{"help", "apropos", "docs", "capabilities", "require", "modules"} {
		seen[base+"."+helper] = "Dark Magic helper"
	}

	return seen
}

// moduleMemberCandidates merges authored help with already loaded Lua functions
// so undocumented commands remain completable without invoking lazy loading.
func (e *Evaluator) moduleMemberCandidates(
	environment *glua.LTable,
	base string,
	module string,
) map[string]string {
	seen := make(map[string]string)

	for name, command := range e.help[module].Commands {
		detail := command.Summary
		if detail == "" {
			detail = "Lua command"
		}

		seen[base+"."+name] = detail
	}

	if table := rawPathTable(environment, strings.Split(base, ".")); table != nil {
		table.ForEach(func(key, value glua.LValue) {
			if _, callable := value.(*glua.LFunction); !callable {
				return
			}

			name := key.String()
			if _, exists := seen[base+"."+name]; !exists {
				seen[base+"."+name] = "Lua command"
			}
		})
	}

	return seen
}

// rawMemberCandidates exposes direct members of a session/global table or
// userdata method table and enriches userdata methods from permitted type help.
func (e *Evaluator) rawMemberCandidates(
	state *glua.LState,
	environment *glua.LTable,
	base string,
) (map[string]string, bool) {
	environmentValue := environment.RawGetString(base)
	globalValue := state.GetGlobal(base)

	table, detail := rawMemberTable(state, environmentValue, globalValue)
	if table == nil {
		return nil, false
	}

	value := environmentValue
	if value == glua.LNil {
		value = globalValue
	}

	seen := make(map[string]string)

	table.ForEach(func(key, _ glua.LValue) {
		name := key.String()
		seen[base+"."+name] = e.memberDetail(state, value, name, detail)
	})

	return seen, true
}

// memberDetail substitutes authored userdata method summaries for the generic
// fallback when the metatable belongs to a permitted documented type.
func (e *Evaluator) memberDetail(
	state *glua.LState,
	value glua.LValue,
	member string,
	fallback string,
) string {
	if _, userData := value.(*glua.LUserData); !userData {
		return fallback
	}

	metatable := state.GetMetatable(value)

	for _, module := range e.modules {
		for typeName, typeDoc := range e.help[module].Types {
			if state.GetTypeMetatable(typeName) == metatable {
				if method, found := typeDoc.Methods[member]; found {
					return method.Summary
				}
			}
		}
	}

	return fallback
}

// completionModule resolves friendly d2legacy aliases while preserving support
// for paths carrying the legacy prefix twice.
func (e *Evaluator) completionModule(base string) (string, bool) {
	base = strings.TrimPrefix(base, "d2legacy.")
	base = strings.TrimPrefix(base, "d2legacy.")

	module, allowed := e.aliases[base]

	return module, allowed
}

// rawPathTable follows only raw table members, avoiding metamethod execution
// during side-effect-free completion discovery.
func rawPathTable(root *glua.LTable, path []string) *glua.LTable {
	var value glua.LValue = root

	for _, segment := range path {
		table, ok := value.(*glua.LTable)
		if !ok {
			return nil
		}

		value = table.RawGetString(segment)
	}

	table, _ := value.(*glua.LTable)

	return table
}

// completionToken extracts the identifier-like suffix shared by Lua, raylib,
// and TUI completion without consuming punctuation before the current token.
func completionToken(source string) string {
	index := strings.LastIndexFunc(source, func(current rune) bool {
		return !completionCharacter(current)
	})

	return source[index+1:]
}

// completionCharacter defines the ASCII member-path grammar intentionally used
// by the shell even though Lua string identifiers can be broader.
func completionCharacter(current rune) bool {
	return current == '_' || current == '.' ||
		current >= 'a' && current <= 'z' ||
		current >= 'A' && current <= 'Z' ||
		current >= '0' && current <= '9'
}

// rawMemberTable finds a physical table or userdata __index table without
// invoking metamethods, which keeps completion inspection side-effect-free.
func rawMemberTable(state *glua.LState, values ...glua.LValue) (*glua.LTable, string) {
	for _, value := range values {
		if table, ok := value.(*glua.LTable); ok {
			return table, "member"
		}

		if _, userData := value.(*glua.LUserData); !userData {
			continue
		}

		metatable, ok := state.GetMetatable(value).(*glua.LTable)
		if !ok {
			continue
		}

		if methods, ok := metatable.RawGetString("__index").(*glua.LTable); ok {
			return methods, "userdata member"
		}
	}

	return nil, ""
}
