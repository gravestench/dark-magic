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
	for _, module := range modules {
		allowed[module] = struct{}{}
	}
	return &Evaluator{
		runtime: runtime, scope: &modruntime.Scope{}, modules: modules, registered: registered, allowed: allowed, mutable: mutable,
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
		if _, ok := e.allowed[name]; !ok {
			current.RaiseError("shell policy does not permit module %q", name)
			return 0
		}
		require, ok := current.GetGlobal("require").(*glua.LFunction)
		if !ok {
			current.RaiseError("Lua require is unavailable")
			return 0
		}
		current.Push(require)
		current.Push(glua.LString(name))
		current.Call(1, 1)
		return 1
	}))
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
			if table, detail := rawMemberTable(state, environment.RawGetString(base), state.GetGlobal(base)); table != nil {
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

func (e *Evaluator) Close() error { return e.scope.Close() }

func (e *Evaluator) environment(state *glua.LState) *glua.LTable {
	if e.env != nil {
		return e.env
	}
	e.env = state.NewTable()
	e.env.RawSetString("_G", e.env)
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
		return strconv.Quote(string(typed))
	case *glua.LFunction:
		return "<function>"
	case *glua.LUserData:
		return "<userdata>"
	default:
		return value.String()
	}
}
