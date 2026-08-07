package modruntime

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
