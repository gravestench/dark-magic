// Package luashell adapts a serialized Lua runtime to the shared shell core.
package luashell

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/shell"
	glua "github.com/yuin/gopher-lua"
)

// Evaluator owns one persistent shell scope and the policy-filtered module
// metadata needed by evaluation, help, and completion.
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

// New exposes every registered runtime module through a mutable developer shell.
func New(runtime *modruntime.Runtime) (*Evaluator, error) {
	if runtime == nil {
		return nil, fmt.Errorf("lua shell: runtime is required")
	}

	modules := runtime.ModuleNames()

	return newEvaluator(runtime, modules, modules, true)
}

// NewForPolicy limits module discovery, require, and shell-global assignment to
// authority explicitly granted to the session. Unknown capabilities remain hidden.
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
		if _, exists := registered[module]; exists {
			modules = append(modules, module)
		}
	}

	registeredModules := runtime.ModuleNames()

	return newEvaluator(runtime, registeredModules, modules, policy.Mutable)
}

// newEvaluator builds lookup indexes around one persistent runtime scope. Module
// slice order is retained because help, capability listing, and resolution expose it.
func newEvaluator(
	runtime *modruntime.Runtime,
	registered []string,
	modules []string,
	mutable bool,
) (*Evaluator, error) {
	allowed := make(map[string]struct{}, len(modules))

	aliases := make(map[string]string, len(modules))
	for _, module := range modules {
		allowed[module] = struct{}{}
		aliases[moduleAlias(module)] = module
	}

	return &Evaluator{
		runtime:    runtime,
		scope:      &modruntime.Scope{},
		modules:    modules,
		registered: registered,
		allowed:    allowed,
		aliases:    aliases,
		help:       runtime.ModuleHelp(),
		mutable:    mutable,
	}, nil
}

// Evaluate executes one submission inside the persistent scope. Module hiding
// is restored before releasing the runtime so another scope never inherits policy.
func (e *Evaluator) Evaluate(ctx context.Context, source string) (shell.Result, error) {
	var result shell.Result

	err := e.runtime.RunScoped(ctx, e.scope, func(state *glua.LState) error {
		evaluated, err := e.evaluateState(state, source)
		if err == nil {
			result = evaluated
		}

		return err
	})

	return result, err
}

// evaluateState installs shell-local policy and output shims, executes the
// expression-or-statement, and removes execution results from the Lua stack.
func (e *Evaluator) evaluateState(state *glua.LState, source string) (shell.Result, error) {
	environment := e.environment(state)

	restoreModules := e.restrictModules(state)
	defer restoreModules()

	printed := make([]string, 0, 4)
	installShellOutput(state, environment, &printed)
	e.installRequire(state, environment)

	function, expression, err := compile(state, source)
	if err != nil {
		return shell.Result{}, err
	}

	state.SetFEnv(function, environment)
	base := state.GetTop()
	state.Push(function)

	if err := state.PCall(0, glua.MultRet, nil); err != nil {
		state.SetTop(base)

		return shell.Result{}, err
	}

	values := make([]string, 0, state.GetTop()-base)
	for index := base + 1; index <= state.GetTop(); index++ {
		values = append(values, formatValue(state.Get(index), 0))
	}

	state.SetTop(base)

	return evaluationResult(printed, values, expression), nil
}

// evaluationResult preserves printed-output-before-return-value ordering and
// classifies the result for richer adapters without requiring them to parse text.
func evaluationResult(printed, values []string, expression bool) shell.Result {
	result := shell.Result{Kind: "statement"}
	if len(printed) > 0 {
		result.Kind = "output"
	}

	if expression && len(values) > 0 {
		result.Kind = "value"
	}

	outputs := append(printed, values...)
	if len(outputs) > 0 {
		result.Text = strings.Join(outputs, "\n")
	} else {
		result.Text = "ok"
	}

	return result
}

// Close releases the persistent runtime scope; Session serializes this against
// evaluation and completion so no Lua work can race scope teardown.
func (e *Evaluator) Close() error {
	return e.scope.Close()
}

// environment lazily constructs the persistent shell global table. Read-only
// policies hide unsafe standard libraries and reject new global assignment.
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
		if !e.mutable && restrictedGlobal(key) {
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

// restrictedGlobal identifies standard libraries withheld from read-only shells
// because they can mutate modules, host files, processes, or runtime internals.
func restrictedGlobal(key string) bool {
	return key == "package" || key == "debug" || key == "io" || key == "os"
}

// compile first treats input as an expression, then falls back to a statement.
// This retains convenient REPL values without requiring users to type return.
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
