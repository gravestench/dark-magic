package modruntime

import (
	"fmt"
	"sort"
	"strings"
)

// ParameterHelp documents one positional or named Lua command parameter.
type ParameterHelp struct {
	Name, Type, Description string
	Optional                bool
}

// ReturnHelp documents one value returned by a Lua command.
type ReturnHelp struct {
	Name, Type, Description string
}

// CommandHelp is the single source of truth for runtime help, completion, and
// generated API documentation for one Lua function.
type CommandHelp struct {
	Summary    string
	Usage      string
	Parameters []ParameterHelp
	Returns    []ReturnHelp
	Examples   []string
}

// ModuleHelp documents one versioned Lua capability and its public commands.
// Commands omitted from this map still receive discoverable fallback help.
type ModuleHelp struct {
	Summary  string
	Commands map[string]CommandHelp
	Types    map[string]TypeHelp
}

// TypeHelp documents a userdata type returned by a capability.
type TypeHelp struct {
	Summary string
	Methods map[string]CommandHelp
}

func cloneModuleHelp(help ModuleHelp) ModuleHelp {
	clone := ModuleHelp{Summary: help.Summary, Commands: make(map[string]CommandHelp, len(help.Commands)), Types: make(map[string]TypeHelp, len(help.Types))}
	for name, command := range help.Commands {
		command.Parameters = append([]ParameterHelp(nil), command.Parameters...)
		command.Returns = append([]ReturnHelp(nil), command.Returns...)
		command.Examples = append([]string(nil), command.Examples...)
		clone.Commands[name] = command
	}
	for name, typeDoc := range help.Types {
		methods := make(map[string]CommandHelp, len(typeDoc.Methods))
		for method, command := range typeDoc.Methods {
			methods[method] = command
		}
		typeDoc.Methods = methods
		clone.Types[name] = typeDoc
	}
	return clone
}

func documentedModule(summary string, commands map[string]CommandHelp, types ...map[string]TypeHelp) ModuleHelp {
	help := ModuleHelp{Summary: summary, Commands: commands}
	if len(types) > 0 {
		help.Types = types[0]
	}
	return help
}

func commandHelp(usage, summary string) CommandHelp {
	return CommandHelp{Usage: usage, Summary: summary}
}

// Markdown renders deterministic API documentation from the same metadata
// used by runtime help and completion.
func (r *Runtime) Markdown(modules []string) string {
	help := r.ModuleHelp()
	if modules == nil {
		modules = r.ModuleNames()
	}
	modules = append([]string(nil), modules...)
	sort.Strings(modules)
	var output strings.Builder
	output.WriteString("# Dark Magic Lua API\n")
	for _, module := range modules {
		doc, ok := help[module]
		if !ok {
			continue
		}
		fmt.Fprintf(&output, "\n## `%s`\n\n%s\n", module, doc.Summary)
		names := make([]string, 0, len(doc.Commands))
		for name := range doc.Commands {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			command := doc.Commands[name]
			fmt.Fprintf(&output, "\n### `%s`\n\n%s\n", command.Usage, command.Summary)
			if len(command.Parameters) > 0 {
				output.WriteString("\nParameters:\n")
				for _, parameter := range command.Parameters {
					optional := ""
					if parameter.Optional {
						optional = " (optional)"
					}
					fmt.Fprintf(&output, "\n- `%s` (`%s`%s): %s", parameter.Name, parameter.Type, optional, parameter.Description)
				}
				output.WriteByte('\n')
			}
			if len(command.Returns) > 0 {
				output.WriteString("\nReturns:\n")
				for _, result := range command.Returns {
					fmt.Fprintf(&output, "\n- `%s` (`%s`): %s", result.Name, result.Type, result.Description)
				}
				output.WriteByte('\n')
			}
			if len(command.Examples) > 0 {
				output.WriteString("\n```lua\n" + strings.Join(command.Examples, "\n") + "\n```\n")
			}
		}
		typeNames := make([]string, 0, len(doc.Types))
		for name := range doc.Types {
			typeNames = append(typeNames, name)
		}
		sort.Strings(typeNames)
		for _, typeName := range typeNames {
			typeDoc := doc.Types[typeName]
			fmt.Fprintf(&output, "\n### Userdata `%s`\n\n%s\n", typeName, typeDoc.Summary)
			methodNames := make([]string, 0, len(typeDoc.Methods))
			for name := range typeDoc.Methods {
				methodNames = append(methodNames, name)
			}
			sort.Strings(methodNames)
			for _, name := range methodNames {
				method := typeDoc.Methods[name]
				fmt.Fprintf(&output, "\n- `%s` — %s", method.Usage, method.Summary)
			}
			output.WriteByte('\n')
		}
	}
	return output.String()
}
