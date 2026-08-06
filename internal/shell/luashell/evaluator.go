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
	runtime *modruntime.Runtime
	scope   *modruntime.Scope
	env     *glua.LTable
	modules []string
}

func New(runtime *modruntime.Runtime) (*Evaluator, error) {
	if runtime == nil {
		return nil, fmt.Errorf("lua shell: runtime is required")
	}
	return &Evaluator{runtime: runtime, scope: &modruntime.Scope{}, modules: runtime.ModuleNames()}, nil
}

func (e *Evaluator) Evaluate(ctx context.Context, source string) (shell.Result, error) {
	var result shell.Result
	err := e.runtime.RunScoped(ctx, e.scope, func(state *glua.LState) error {
		environment := e.environment(state)
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
		result.Kind = "statement"
		if expression || len(values) > 0 {
			result.Kind = "value"
			result.Text = strings.Join(values, "\t")
		} else {
			result.Text = "ok"
		}
		return nil
	})
	return result, err
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
			if table := rawTable(environment.RawGetString(base), state.GetGlobal(base)); table != nil {
				seen = make(map[string]string)
				table.ForEach(func(key, _ glua.LValue) { seen[base+"."+key.String()] = "member" })
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
	metatable.RawSetString("__index", state.Get(glua.GlobalsIndex))
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

func rawTable(primary, fallback glua.LValue) *glua.LTable {
	if table, ok := primary.(*glua.LTable); ok {
		return table
	}
	table, _ := fallback.(*glua.LTable)
	return table
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
