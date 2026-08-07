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
}

func cloneModuleHelp(help ModuleHelp) ModuleHelp {
	clone := ModuleHelp{Summary: help.Summary, Commands: make(map[string]CommandHelp, len(help.Commands))}
	for name, command := range help.Commands {
		command.Parameters = append([]ParameterHelp(nil), command.Parameters...)
		command.Returns = append([]ReturnHelp(nil), command.Returns...)
		command.Examples = append([]string(nil), command.Examples...)
		clone.Commands[name] = command
	}
	return clone
}

func documentedModule(summary string, commands map[string]CommandHelp) ModuleHelp {
	return ModuleHelp{Summary: summary, Commands: commands}
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
	}
	return output.String()
}
