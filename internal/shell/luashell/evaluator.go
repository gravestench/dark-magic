// Package luashell adapts a serialized Lua runtime to the shared shell core.
package luashell

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/shell"
	glua "github.com/yuin/gopher-lua"
)

var keywords = []string{
	"and", "break", "do", "else", "elseif", "end", "false", "for", "function",
	"goto", "if", "in", "local", "nil", "not", "or", "repeat", "return", "then",
	"true", "until", "while",
}

type Evaluator struct {
	runtime    *modruntime.Runtime
	scope      *modruntime.Scope
	env        *glua.LTable
	modules    []string
	registered []string
	allowed    map[string]struct{}
	aliases    map[string]string
	help       map[string]modruntime.ModuleHelp
	mutable    bool
}

func New(runtime *modruntime.Runtime) (*Evaluator, error) {
	if runtime == nil {
		return nil, fmt.Errorf("lua shell: runtime is required")
	}
	modules := runtime.ModuleNames()
	return newEvaluator(runtime, modules, modules, true)
}

// NewForPolicy limits module discovery/require and shell-global assignment to
// the authority explicitly granted to the session.
func NewForPolicy(runtime *modruntime.Runtime, policy shell.Policy) (*Evaluator, error) {
	if runtime == nil {
		return nil, fmt.Errorf("lua shell: runtime is required")
	}
	registered := make(map[string]struct{})
	for _, module := range runtime.ModuleNames() {
		registered[module] = struct{}{}
	}
	modules := make([]string, 0, len(policy.Capabilities))
	for _, module := range policy.Capabilities {
		if _, ok := registered[module]; ok {
			modules = append(modules, module)
		}
	}
	registeredModules := runtime.ModuleNames()
	return newEvaluator(runtime, registeredModules, modules, policy.Mutable)
}

func newEvaluator(runtime *modruntime.Runtime, registered, modules []string, mutable bool) (*Evaluator, error) {
	allowed := make(map[string]struct{}, len(modules))
	aliases := make(map[string]string, len(modules))
	for _, module := range modules {
		allowed[module] = struct{}{}
		aliases[moduleAlias(module)] = module
	}
	return &Evaluator{
		runtime: runtime, scope: &modruntime.Scope{}, modules: modules, registered: registered,
		allowed: allowed, aliases: aliases, mutable: mutable,
		help: runtime.ModuleHelp(),
	}, nil
}

func (e *Evaluator) Evaluate(ctx context.Context, source string) (shell.Result, error) {
	var result shell.Result
	err := e.runtime.RunScoped(ctx, e.scope, func(state *glua.LState) error {
		environment := e.environment(state)
		restoreModules := e.restrictModules(state)
		defer restoreModules()
		printed := make([]string, 0, 4)
		installShellOutput(state, environment, &printed)
		e.installRequire(state, environment)
		function, expression, err := compile(state, source)
		if err != nil {
			return err
		}
		state.SetFEnv(function, environment)
		base := state.GetTop()
		state.Push(function)
		if err := state.PCall(0, glua.MultRet, nil); err != nil {
			state.SetTop(base)
			return err
		}
		values := make([]string, 0, state.GetTop()-base)
		for index := base + 1; index <= state.GetTop(); index++ {
			values = append(values, formatValue(state.Get(index), 0))
		}
		state.SetTop(base)
		outputs := append(printed, values...)
		result.Kind = "statement"
		if len(printed) > 0 {
			result.Kind = "output"
		}
		if expression && len(values) > 0 {
			result.Kind = "value"
		}
		if len(outputs) > 0 {
			result.Text = strings.Join(outputs, "\n")
		} else {
			result.Text = "ok"
		}
		return nil
	})
	return result, err
}

func (e *Evaluator) restrictModules(state *glua.LState) func() {
	packageTable, ok := state.GetGlobal("package").(*glua.LTable)
	if !ok {
		return func() {}
	}
	preload, _ := packageTable.RawGetString("preload").(*glua.LTable)
	loaded, _ := packageTable.RawGetString("loaded").(*glua.LTable)
	type hiddenModule struct {
		name            string
		preload, loaded glua.LValue
	}
	hidden := make([]hiddenModule, 0)
	for _, name := range e.registered {
		if _, allowed := e.allowed[name]; allowed {
			continue
		}
		entry := hiddenModule{name: name}
		if preload != nil {
			entry.preload = preload.RawGetString(name)
			preload.RawSetString(name, glua.LNil)
		}
		if loaded != nil {
			entry.loaded = loaded.RawGetString(name)
			loaded.RawSetString(name, glua.LNil)
		}
		hidden = append(hidden, entry)
	}
	return func() {
		for _, entry := range hidden {
			if preload != nil {
				preload.RawSetString(entry.name, entry.preload)
			}
			if loaded != nil {
				loaded.RawSetString(entry.name, entry.loaded)
			}
		}
	}
}

func (e *Evaluator) installRequire(state *glua.LState, environment *glua.LTable) {
	environment.RawSetString("require", state.NewFunction(func(current *glua.LState) int {
		name := current.CheckString(1)
		return e.requireModule(current, name)
	}))
}

func (e *Evaluator) requireModule(state *glua.LState, name string) int {
	if _, ok := e.allowed[name]; !ok {
		state.RaiseError("shell policy does not permit module %q", name)
		return 0
	}
	require, ok := state.GetGlobal("require").(*glua.LFunction)
	if !ok {
		state.RaiseError("Lua require is unavailable")
		return 0
	}
	state.Push(require)
	state.Push(glua.LString(name))
	state.Call(1, 1)
	return 1
}

func (e *Evaluator) installDarkMagicRoot(state *glua.LState, environment *glua.LTable) {
	root := state.NewTable()
	modules := state.NewTable()
	root.RawSetString("modules", modules)
	root.RawSetString("require", state.NewFunction(func(current *glua.LState) int {
		return e.requireModule(current, current.CheckString(1))
	}))
	root.RawSetString("capabilities", state.NewFunction(func(current *glua.LState) int {
		values := current.NewTable()
		for index, module := range e.modules {
			values.RawSetInt(index+1, glua.LString(module))
		}
		current.Push(values)
		return 1
	}))
	aliases := make([]string, 0, len(e.aliases))
	for alias := range e.aliases {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)
	help := "Dark Magic shell root\n  dm.<capability>       lazy friendly module access\n  dm.modules[<id>]      exact versioned module access\n  dm.require(<id>)      policy-checked require\n  dm.capabilities()     permitted module IDs\n  dm.help([value])      help for a module or command"
	if len(aliases) > 0 {
		help += "\nAvailable aliases: " + strings.Join(aliases, ", ")
	}
	root.RawSetString("help", state.NewFunction(func(current *glua.LState) int {
		if current.GetTop() == 0 {
			current.Push(glua.LString(help))
			return 1
		}
		current.Push(glua.LString(e.helpFor(current, current.Get(1))))
		return 1
	}))
	lazyIndex := func(aliasLookup bool) *glua.LFunction {
		return state.NewFunction(func(current *glua.LState) int {
			name := current.CheckString(2)
			if aliasLookup {
				var ok bool
				name, ok = e.aliases[name]
				if !ok {
					current.Push(glua.LNil)
					return 1
				}
			}
			e.requireModule(current, name)
			value := current.Get(-1)
			current.CheckTable(1).RawSetString(current.CheckString(2), value)
			return 1
		})
	}
	rootMeta := state.NewTable()
	rootMeta.RawSetString("__index", lazyIndex(true))
	rootMeta.RawSetString("__metatable", glua.LString("protected Dark Magic root"))
	state.SetMetatable(root, rootMeta)
	moduleMeta := state.NewTable()
	moduleMeta.RawSetString("__index", lazyIndex(false))
	moduleMeta.RawSetString("__metatable", glua.LString("protected Dark Magic modules"))
	state.SetMetatable(modules, moduleMeta)
	environment.RawSetString("dm", root)
	environment.RawSetString("darkmagic", root)
}

func (e *Evaluator) helpFor(state *glua.LState, value glua.LValue) string {
	if text, ok := value.(glua.LString); ok {
		module, command := e.helpPath(string(text))
		if module == "" {
			return fmt.Sprintf("No permitted Dark Magic API matches %q.", text)
		}
		e.requireModule(state, module)
		value = state.Get(-1)
		state.Pop(1)
		if command != "" {
			return e.formatCommandHelp(module, command)
		}
		return e.formatModuleHelp(module, value)
	}
	packageTable, _ := state.GetGlobal("package").(*glua.LTable)
	loaded, _ := packageTable.RawGetString("loaded").(*glua.LTable)
	if loaded != nil {
		for _, module := range e.modules {
			moduleValue := loaded.RawGetString(module)
			if moduleValue == value {
				return e.formatModuleHelp(module, moduleValue)
			}
			if table, ok := moduleValue.(*glua.LTable); ok {
				matched := ""
				table.ForEach(func(key, member glua.LValue) {
					if member == value {
						matched = key.String()
					}
				})
				if matched != "" {
					return e.formatCommandHelp(module, matched)
				}
			}
		}
	}
	return "No help metadata is available for that value. Pass a path such as \"dm.audio.play\"."
}

func (e *Evaluator) helpPath(path string) (string, string) {
	path = strings.TrimSpace(path)
	if _, ok := e.allowed[path]; ok {
		return path, ""
	}
	path = strings.TrimPrefix(path, "darkmagic.")
	path = strings.TrimPrefix(path, "dm.")
	parts := strings.SplitN(path, ".", 2)
	module, ok := e.aliases[parts[0]]
	if !ok {
		return "", ""
	}
	if len(parts) == 2 {
		return module, parts[1]
	}
	return module, ""
}

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
			if _, ok := member.(*glua.LFunction); ok {
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
	fmt.Fprintf(&output, "dm.%s (%s)\n%s", alias, module, summary)
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

func (e *Evaluator) formatCommandHelp(module, name string) string {
	doc := e.help[module].Commands[name]
	path := "dm." + moduleAlias(module) + "." + name
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
	if len(doc.Parameters) > 0 {
		output.WriteString("\n\nParameters:")
		for _, parameter := range doc.Parameters {
			optional := ""
			if parameter.Optional {
				optional = " (optional)"
			}
			fmt.Fprintf(&output, "\n  %s  %s%s  %s", parameter.Name, parameter.Type, optional, parameter.Description)
		}
	}
	if len(doc.Returns) > 0 {
		output.WriteString("\n\nReturns:")
		for _, result := range doc.Returns {
			fmt.Fprintf(&output, "\n  %s  %s  %s", result.Name, result.Type, result.Description)
		}
	}
	if len(doc.Examples) > 0 {
		output.WriteString("\n\nExamples:\n  " + strings.Join(doc.Examples, "\n  "))
	}
	return output.String()
}

func moduleAlias(module string) string {
	module = strings.TrimPrefix(module, "dm.")
	if separator := strings.IndexByte(module, '/'); separator >= 0 {
		module = module[:separator]
	}
	return strings.ReplaceAll(module, ".", "_")
}

func installShellOutput(state *glua.LState, environment *glua.LTable, output *[]string) {
	printFunction := state.NewFunction(func(current *glua.LState) int {
		values := make([]string, current.GetTop())
		for index := 1; index <= current.GetTop(); index++ {
			values[index-1] = current.ToStringMeta(current.Get(index)).String()
		}
		*output = append(*output, strings.Join(values, "\t"))
		return 0
	})
	environment.RawSetString("print", printFunction)

	registerFunction := state.NewFunction(func(current *glua.LState) int {
		lines := []string{"Lua call frames:"}
		for level := 1; ; level++ {
			frame, ok := current.GetStack(level)
			if !ok {
				break
			}
			_, _ = current.GetInfo("nSl", frame, glua.LNil)
			location := frame.Source
			if frame.CurrentLine >= 0 {
				location += fmt.Sprintf(":%d", frame.CurrentLine)
			}
			lines = append(lines, fmt.Sprintf("  [%d] %s %s", level, frame.Name, location))
			for local := 1; ; local++ {
				name, value := current.GetLocal(frame, local)
				if name == "" {
					break
				}
				lines = append(lines, fmt.Sprintf("      %s = %s", name, formatValue(value, 0)))
			}
		}
		*output = append(*output, strings.Join(lines, "\n"))
		return 0
	})
	// GopherLua's private _printregs writes directly to process stderr. Shadow
	// it inside shell scopes and provide a friendly alias that returns the same
	// class of debugging information through the shell transcript.
	environment.RawSetString("_printregs", registerFunction)
	environment.RawSetString("printregs", registerFunction)
}

func (e *Evaluator) Complete(ctx context.Context, source string) ([]shell.Candidate, error) {
	var result []shell.Candidate
	err := e.runtime.RunScoped(ctx, e.scope, func(state *glua.LState) error {
		environment := e.environment(state)
		token := completionToken(source)
		seen := make(map[string]string)
		for _, keyword := range keywords {
			seen[keyword] = "keyword"
		}
		for _, module := range e.modules {
			seen[module] = "module"
		}
		environment.ForEach(func(key, _ glua.LValue) { seen[key.String()] = "session" })
		if global, ok := state.Get(glua.GlobalsIndex).(*glua.LTable); ok {
			global.ForEach(func(key, _ glua.LValue) {
				if _, exists := seen[key.String()]; !exists {
					seen[key.String()] = "global"
				}
			})
		}
		if dot := strings.LastIndexByte(token, '.'); dot >= 0 {
			base, member := token[:dot], token[dot+1:]
			if base == "dm" || base == "darkmagic" {
				seen = make(map[string]string)
				for alias, module := range e.aliases {
					detail := e.help[module].Summary
					if detail == "" {
						detail = "Dark Magic capability"
					}
					seen[base+"."+alias] = detail
				}
				for _, helper := range []string{"help", "capabilities", "require", "modules"} {
					seen[base+"."+helper] = "Dark Magic helper"
				}
				token = base + "." + member
			} else if module, ok := e.completionModule(base); ok {
				seen = make(map[string]string)
				for name, command := range e.help[module].Commands {
					detail := command.Summary
					if detail == "" {
						detail = "Lua command"
					}
					seen[base+"."+name] = detail
				}
				if table := rawPathTable(environment, strings.Split(base, ".")); table != nil {
					table.ForEach(func(key, value glua.LValue) {
						if _, ok := value.(*glua.LFunction); !ok {
							return
						}
						name := key.String()
						if _, exists := seen[base+"."+name]; !exists {
							seen[base+"."+name] = "Lua command"
						}
					})
				}
				token = base + "." + member
			} else if table, detail := rawMemberTable(state, environment.RawGetString(base), state.GetGlobal(base)); table != nil {
				seen = make(map[string]string)
				table.ForEach(func(key, _ glua.LValue) { seen[base+"."+key.String()] = detail })
				token = base + "." + member
			}
		}
		for value, detail := range seen {
			if strings.HasPrefix(value, token) {
				result = append(result, shell.Candidate{Value: value, Detail: detail})
			}
		}
		sort.Slice(result, func(i, j int) bool { return result[i].Value < result[j].Value })
		return nil
	})
	return result, err
}

func (e *Evaluator) completionModule(base string) (string, bool) {
	base = strings.TrimPrefix(base, "darkmagic.")
	base = strings.TrimPrefix(base, "dm.")
	module, ok := e.aliases[base]
	return module, ok
}

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

func (e *Evaluator) Close() error { return e.scope.Close() }

func (e *Evaluator) environment(state *glua.LState) *glua.LTable {
	if e.env != nil {
		return e.env
	}
	e.env = state.NewTable()
	e.env.RawSetString("_G", e.env)
	e.installDarkMagicRoot(state, e.env)
	metatable := state.NewTable()
	global := state.Get(glua.GlobalsIndex).(*glua.LTable)
	metatable.RawSetString("__index", state.NewFunction(func(current *glua.LState) int {
		key := current.CheckString(2)
		if !e.mutable && (key == "package" || key == "debug" || key == "io" || key == "os") {
			current.Push(glua.LNil)
			return 1
		}
		current.Push(global.RawGetString(key))
		return 1
	}))
	metatable.RawSetString("__metatable", glua.LString("protected shell environment"))
	if !e.mutable {
		metatable.RawSetString("__newindex", state.NewFunction(func(current *glua.LState) int {
			current.RaiseError("shell policy is read-only; cannot assign %q", current.Get(2).String())
			return 0
		}))
	}
	state.SetMetatable(e.env, metatable)
	return e.env
}

func compile(state *glua.LState, source string) (*glua.LFunction, bool, error) {
	function, err := state.Load(bytes.NewBufferString("return "+source), "=shell")
	if err == nil {
		return function, true, nil
	}
	function, statementErr := state.Load(bytes.NewBufferString(source), "=shell")
	if statementErr != nil {
		return nil, false, statementErr
	}
	return function, false, nil
}

func completionToken(source string) string {
	index := strings.LastIndexFunc(source, func(current rune) bool {
		return !(current == '_' || current == '.' || current >= 'a' && current <= 'z' || current >= 'A' && current <= 'Z' || current >= '0' && current <= '9')
	})
	return source[index+1:]
}

func rawMemberTable(state *glua.LState, values ...glua.LValue) (*glua.LTable, string) {
	for _, value := range values {
		if table, ok := value.(*glua.LTable); ok {
			return table, "member"
		}
		if _, ok := value.(*glua.LUserData); !ok {
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

func formatValue(value glua.LValue, depth int) string {
	switch typed := value.(type) {
	case *glua.LTable:
		if depth >= 2 {
			return "{…}"
		}
		parts := make([]string, 0, 8)
		typed.ForEach(func(key, item glua.LValue) {
			if len(parts) < 8 {
				parts = append(parts, key.String()+"="+formatValue(item, depth+1))
			}
		})
		sort.Strings(parts)
		return "{" + strings.Join(parts, ", ") + "}"
	case glua.LString:
		if strings.ContainsAny(string(typed), "\r\n\t") {
			return string(typed)
		}
		return strconv.Quote(string(typed))
	case *glua.LFunction:
		return "<function>"
	case *glua.LUserData:
		return "<userdata>"
	default:
		return value.String()
	}
}
