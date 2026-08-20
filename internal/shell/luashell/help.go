package luashell

import (
	"fmt"
	"sort"
	"strings"

	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	glua "github.com/yuin/gopher-lua"
)

// apropos searches only policy-permitted module, command, type, and method
// metadata, then sorts matches so map iteration cannot reorder shell output.
func (e *Evaluator) apropos(query string) string {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return "Usage: d2legacy.apropos(\"search terms\")"
	}

	var matches []string

	for _, module := range e.modules {
		doc := e.help[module]
		alias := "d2legacy." + moduleAlias(module)

		if strings.Contains(strings.ToLower(module+" "+alias+" "+doc.Summary), query) {
			matches = append(matches, alias+" — "+doc.Summary)
		}

		matches = append(matches, matchingCommandHelp(alias, doc.Commands, query)...)
		matches = append(matches, matchingTypeHelp(doc.Types, query)...)
	}

	sort.Strings(matches)

	if len(matches) == 0 {
		return fmt.Sprintf("No permitted Dark Magic APIs match %q.", query)
	}

	return strings.Join(matches, "\n")
}

// matchingCommandHelp returns command matches without exposing commands from a
// module that the caller's policy omitted.
func matchingCommandHelp(
	alias string,
	commands map[string]modruntime.CommandHelp,
	query string,
) []string {
	var matches []string

	for name, command := range commands {
		haystack := name + " " + command.Usage + " " + command.Summary
		if strings.Contains(strings.ToLower(haystack), query) {
			matches = append(matches, alias+"."+name+" — "+command.Summary)
		}
	}

	return matches
}

// matchingTypeHelp searches type summaries and methods while retaining authored
// usage strings in the displayed matches.
func matchingTypeHelp(types map[string]modruntime.TypeHelp, query string) []string {
	var matches []string

	for typeName, typeDoc := range types {
		if strings.Contains(strings.ToLower(typeName+" "+typeDoc.Summary), query) {
			matches = append(matches, typeName+" — "+typeDoc.Summary)
		}

		for name, method := range typeDoc.Methods {
			haystack := typeName + " " + name + " " + method.Usage + " " + method.Summary
			if strings.Contains(strings.ToLower(haystack), query) {
				matches = append(matches, method.Usage+" — "+method.Summary)
			}
		}
	}

	return matches
}

// helpFor resolves string paths, loaded module members, and userdata metatables
// without granting access to metadata outside the permitted module list.
func (e *Evaluator) helpFor(state *glua.LState, value glua.LValue) string {
	if text, ok := value.(glua.LString); ok {
		return e.helpForPath(state, string(text))
	}

	if help := e.helpForLoadedValue(state, value); help != "" {
		return help
	}

	if userData, ok := value.(*glua.LUserData); ok {
		metatable := state.GetMetatable(userData)

		for _, module := range e.modules {
			for typeName, typeDoc := range e.help[module].Types {
				if state.GetTypeMetatable(typeName) == metatable {
					return formatTypeHelp(typeName, typeDoc)
				}
			}
		}
	}

	return "No help metadata is available for that value. Pass a path such as \"engine.audio.play\"."
}

// helpForPath resolves and loads an admitted module before formatting a module
// or command, ensuring string help follows the same policy as direct access.
func (e *Evaluator) helpForPath(state *glua.LState, path string) string {
	module, command := e.helpPath(path)
	if module == "" {
		return fmt.Sprintf("No permitted Dark Magic API matches %q.", path)
	}

	e.requireModule(state, module)
	value := state.Get(-1)
	state.Pop(1)

	if command != "" {
		return e.formatCommandHelp(module, command)
	}

	return e.formatModuleHelp(module, value)
}

// helpForLoadedValue identifies an already loaded module or one of its direct
// members by identity, avoiding arbitrary execution during help lookup.
func (e *Evaluator) helpForLoadedValue(state *glua.LState, value glua.LValue) string {
	packageTable, _ := state.GetGlobal("package").(*glua.LTable)

	loaded, _ := packageTable.RawGetString("loaded").(*glua.LTable)
	if loaded == nil {
		return ""
	}

	for _, module := range e.modules {
		moduleValue := loaded.RawGetString(module)
		if moduleValue == value {
			return e.formatModuleHelp(module, moduleValue)
		}

		if table, ok := moduleValue.(*glua.LTable); ok {
			matched := matchingTableMember(table, value)
			if matched != "" {
				return e.formatCommandHelp(module, matched)
			}
		}
	}

	return ""
}

// matchingTableMember returns the last authored key whose value matches, which
// preserves the prior table-iteration behavior when aliases share a function.
func matchingTableMember(table *glua.LTable, value glua.LValue) string {
	matched := ""

	table.ForEach(func(key, member glua.LValue) {
		if member == value {
			matched = key.String()
		}
	})

	return matched
}

// formatTypeHelp renders methods in sorted order so generated help remains stable.
func formatTypeHelp(name string, doc modruntime.TypeHelp) string {
	names := make([]string, 0, len(doc.Methods))
	for method := range doc.Methods {
		names = append(names, method)
	}

	sort.Strings(names)

	var output strings.Builder

	fmt.Fprintf(&output, "%s\n%s", name, doc.Summary)

	if len(names) > 0 {
		output.WriteString("\n\nMethods:")

		for _, name := range names {
			method := doc.Methods[name]
			fmt.Fprintf(&output, "\n  %-24s %s", method.Usage, method.Summary)
		}
	}

	return output.String()
}

// helpPath maps exact IDs and friendly d2legacy aliases to permitted modules.
// Two legacy prefix removals preserve compatibility with redundantly qualified paths.
func (e *Evaluator) helpPath(path string) (string, string) {
	path = strings.TrimSpace(path)
	if _, allowed := e.allowed[path]; allowed {
		return path, ""
	}

	path = strings.TrimPrefix(path, "d2legacy.")
	path = strings.TrimPrefix(path, "d2legacy.")
	parts := strings.SplitN(path, ".", 2)

	module, allowed := e.aliases[parts[0]]
	if !allowed {
		return "", ""
	}

	if len(parts) == 2 {
		return module, parts[1]
	}

	return module, ""
}

// formatModuleHelp merges authored command metadata with runtime-provided Lua
// functions so undocumented commands remain discoverable rather than invisible.
func (e *Evaluator) formatModuleHelp(module string, value glua.LValue) string {
	doc := e.help[module]
	alias := moduleAlias(module)

	summary := doc.Summary
	if summary == "" {
		summary = "Dark Magic Lua capability."
	}

	commands := make(map[string]struct{}, len(doc.Commands))
	for name := range doc.Commands {
		commands[name] = struct{}{}
	}

	if table, ok := value.(*glua.LTable); ok {
		table.ForEach(func(key, member glua.LValue) {
			if _, callable := member.(*glua.LFunction); callable {
				commands[key.String()] = struct{}{}
			}
		})
	}

	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}

	sort.Strings(names)

	var output strings.Builder

	fmt.Fprintf(&output, "d2legacy.%s (%s)\n%s", alias, module, summary)

	if len(names) > 0 {
		output.WriteString("\n\nCommands:")

		for _, name := range names {
			command := doc.Commands[name]

			description := command.Summary
			if description == "" {
				description = "Lua command provided by " + module + "."
			}

			fmt.Fprintf(&output, "\n  %-24s %s", name, description)
		}
	}

	return output.String()
}

// formatCommandHelp fills missing metadata with stable fallbacks while retaining
// parameter, return, and example order authored by the module.
func (e *Evaluator) formatCommandHelp(module, name string) string {
	doc := e.help[module].Commands[name]
	path := "d2legacy." + moduleAlias(module) + "." + name

	usage := doc.Usage
	if usage == "" {
		usage = path + "(...)"
	}

	summary := doc.Summary
	if summary == "" {
		summary = "Lua command provided by " + module + "."
	}

	var output strings.Builder

	fmt.Fprintf(&output, "%s\n\n%s", usage, summary)
	appendParameterHelp(&output, doc.Parameters)
	appendReturnHelp(&output, doc.Returns)

	if len(doc.Examples) > 0 {
		output.WriteString("\n\nExamples:\n  " + strings.Join(doc.Examples, "\n  "))
	}

	return output.String()
}

// appendParameterHelp preserves declaration order and marks optional inputs next
// to their type, matching the established command-help text schema.
func appendParameterHelp(output *strings.Builder, parameters []modruntime.ParameterHelp) {
	if len(parameters) == 0 {
		return
	}

	output.WriteString("\n\nParameters:")

	for _, parameter := range parameters {
		optional := ""
		if parameter.Optional {
			optional = " (optional)"
		}

		fmt.Fprintf(
			output,
			"\n  %s  %s%s  %s",
			parameter.Name,
			parameter.Type,
			optional,
			parameter.Description,
		)
	}
}

// appendReturnHelp preserves authored return ordering because positional Lua
// results derive meaning from their order.
func appendReturnHelp(output *strings.Builder, results []modruntime.ReturnHelp) {
	if len(results) == 0 {
		return
	}

	output.WriteString("\n\nReturns:")

	for _, result := range results {
		fmt.Fprintf(output, "\n  %s  %s  %s", result.Name, result.Type, result.Description)
	}
}
