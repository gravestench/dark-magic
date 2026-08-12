// Package modruntime hosts Dark Magic's serialized Lua execution environment.
package modruntime

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// Module is a versioned engine capability exposed through Lua require.
type Module struct {
	Name   string
	Loader lua.LGFunction
	Help   ModuleHelp
}

// Installer configures the Lua state before boot code runs, for behavior such
// as content-backed require searchers that is not itself a required module.
type Installer struct {
	Name    string
	Install func(*lua.LState) error
}

type request struct {
	run  func(*lua.LState) error
	ctx  context.Context
	stop bool
	done chan error
}

// Runtime owns one Lua state on one goroutine. All access is serialized through
// Run; the state is never exposed outside the duration of a Run callback.
type Runtime struct {
	mu              sync.RWMutex
	started         bool
	modules         []Module
	installers      []Installer
	requests        chan request
	done            chan struct{}
	executionBudget time.Duration

	// activeScope is only read or written by the Lua owner goroutine.
	activeScope *Scope
}

// RegisterInstaller registers VM setup performed on the owner goroutine.
func (r *Runtime) RegisterInstaller(installer Installer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.started {
		return errors.New("modruntime: cannot register an installer while running")
	}
	if installer.Name == "" || installer.Install == nil {
		return errors.New("modruntime: installer name and function are required")
	}
	for _, existing := range r.installers {
		if existing.Name == installer.Name {
			return fmt.Errorf("modruntime: installer %q is already registered", installer.Name)
		}
	}
	r.installers = append(r.installers, installer)
	return nil
}

func (r *Runtime) runScoped(ctx context.Context, scope *Scope, fn func(*lua.LState) error) error {
	return r.runScopedWithBudget(ctx, scope, r.executionBudget, fn)
}

func (r *Runtime) runScopedWithBudget(ctx context.Context, scope *Scope, budget time.Duration, fn func(*lua.LState) error) error {
	return r.runWithBudget(ctx, budget, func(state *lua.LState) error {
		previous := r.activeScope
		r.activeScope = scope
		defer func() { r.activeScope = previous }()
		return fn(state)
	})
}

// RunScoped executes fn on the runtime owner with resources assigned to scope.
// Interactive shells use a dedicated scope so reset and disconnect cannot leak
// handles into scene or server lifetimes.
func (r *Runtime) RunScoped(ctx context.Context, scope *Scope, fn func(*lua.LState) error) error {
	if scope == nil {
		return errors.New("modruntime: nil invocation scope")
	}
	return r.runScoped(ctx, scope, fn)
}

// ModuleNames returns the registered capability names without exposing loaders.
func (r *Runtime) ModuleNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, len(r.modules))
	for index, module := range r.modules {
		names[index] = module.Name
	}
	return names
}

// ModuleHelp returns an immutable snapshot of the registered Lua API metadata.
func (r *Runtime) ModuleHelp() map[string]ModuleHelp {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]ModuleHelp, len(r.modules))
	for _, module := range r.modules {
		result[module.Name] = cloneModuleHelp(module.Help)
	}
	return result
}

func (r *Runtime) requireActiveScope() (*Scope, error) {
	if r.activeScope == nil {
		return nil, errors.New("modruntime: capability requires an active component scope")
	}
	return r.activeScope, nil
}

// New constructs a stopped runtime.
func New() *Runtime { return &Runtime{executionBudget: time.Second} }

// SetExecutionBudget bounds each Lua VM invocation. A zero duration disables
// the runtime deadline but caller cancellation still interrupts Lua execution.
func (r *Runtime) SetExecutionBudget(budget time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.started {
		return errors.New("modruntime: cannot change execution budget while running")
	}
	if budget < 0 {
		return errors.New("modruntime: execution budget cannot be negative")
	}
	r.executionBudget = budget
	return nil
}

// RegisterModule registers a capability before the runtime starts.
func (r *Runtime) RegisterModule(module Module) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.started {
		return errors.New("modruntime: cannot register a module while running")
	}
	if module.Name == "" || module.Loader == nil {
		return errors.New("modruntime: module name and loader are required")
	}
	for _, existing := range r.modules {
		if existing.Name == module.Name {
			return fmt.Errorf("modruntime: module %q is already registered", module.Name)
		}
	}
	module.Help = cloneModuleHelp(module.Help)
	r.modules = append(r.modules, module)
	return nil
}

// RegisterModuleDefault installs a fallback capability only when a richer host
// composition has not already supplied that exact module name. Required
// capabilities should continue using strict RegisterModule.
func (r *Runtime) RegisterModuleDefault(module Module) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.started {
		return errors.New("modruntime: cannot register a module while running")
	}
	if module.Name == "" || module.Loader == nil {
		return errors.New("modruntime: module name and loader are required")
	}
	for _, existing := range r.modules {
		if existing.Name == module.Name {
			return nil
		}
	}
	module.Help = cloneModuleHelp(module.Help)
	r.modules = append(r.modules, module)
	return nil
}

// RegisterModuleOverride installs a host-specific capability, replacing an
// earlier policy-neutral default with the same name. It is deliberately
// separate from strict registration so ordinary duplicate wiring still fails.
func (r *Runtime) RegisterModuleOverride(module Module) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.started {
		return errors.New("modruntime: cannot register a module while running")
	}
	if module.Name == "" || module.Loader == nil {
		return errors.New("modruntime: module name and loader are required")
	}
	module.Help = cloneModuleHelp(module.Help)
	for index, existing := range r.modules {
		if existing.Name == module.Name {
			r.modules[index] = module
			return nil
		}
	}
	r.modules = append(r.modules, module)
	return nil
}

// Start creates the Lua state on its owner goroutine.
func (r *Runtime) Start(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.started {
		return nil
	}
	requests := make(chan request)
	done := make(chan struct{})
	ready := make(chan error, 1)
	modules := append([]Module(nil), r.modules...)
	installers := append([]Installer(nil), r.installers...)
	go runLoop(requests, done, ready, installers, modules)
	select {
	case err := <-ready:
		if err != nil {
			return err
		}
		r.requests = requests
		r.done = done
		r.started = true
		return nil
	case <-ctx.Done():
		response := make(chan error, 1)
		requests <- request{stop: true, done: response}
		<-response
		<-done
		return fmt.Errorf("modruntime: start: %w", ctx.Err())
	}
}

// Stop closes the Lua state after all active Run calls finish.
func (r *Runtime) Stop(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.started {
		return nil
	}
	response := make(chan error, 1)
	select {
	case r.requests <- request{stop: true, done: response}:
	case <-ctx.Done():
		return fmt.Errorf("modruntime: stop: %w", ctx.Err())
	}
	err := <-response
	<-r.done
	r.started = false
	r.requests = nil
	r.done = nil
	return err
}

// Run executes fn on the Lua owner goroutine.
func (r *Runtime) Run(ctx context.Context, fn func(*lua.LState) error) error {
	return r.runWithBudget(ctx, r.executionBudget, fn)
}

func (r *Runtime) runWithBudget(ctx context.Context, budget time.Duration, fn func(*lua.LState) error) error {
	if fn == nil {
		return errors.New("modruntime: nil invocation")
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if !r.started {
		return errors.New("modruntime: runtime is not started")
	}
	requestContext := ctx
	cancel := func() {}
	if budget > 0 {
		requestContext, cancel = context.WithTimeout(ctx, budget)
	}
	defer cancel()
	response := make(chan error, 1)
	select {
	case r.requests <- request{run: fn, ctx: requestContext, done: response}:
	case <-ctx.Done():
		return fmt.Errorf("modruntime: invoke: %w", ctx.Err())
	}
	select {
	case err := <-response:
		return err
	case <-ctx.Done():
		return fmt.Errorf("modruntime: invoke: %w", ctx.Err())
	}
}

// Execute loads and runs a Lua source file from a content filesystem.
func (r *Runtime) Execute(ctx context.Context, source fs.FS, name string) error {
	return r.execute(ctx, nil, source, name)
}

// ExecuteScoped loads and runs a Lua source file with resources assigned to
// scope. It is the file-backed counterpart to RunScoped and lets integration
// harnesses keep Lua assertions in Lua files rather than Go string literals.
func (r *Runtime) ExecuteScoped(ctx context.Context, scope *Scope, source fs.FS, name string) error {
	if scope == nil {
		return errors.New("modruntime: nil invocation scope")
	}
	return r.execute(ctx, scope, source, name)
}

func (r *Runtime) execute(ctx context.Context, scope *Scope, source fs.FS, name string) error {
	data, err := fs.ReadFile(source, name)
	if err != nil {
		return fmt.Errorf("modruntime: read %q: %w", name, err)
	}
	run := r.Run
	if scope != nil {
		run = func(ctx context.Context, callback func(*lua.LState) error) error {
			return r.RunScoped(ctx, scope, callback)
		}
	}
	return run(ctx, func(state *lua.LState) error {
		function, err := state.Load(bytes.NewReader(data), "@"+name)
		if err != nil {
			return scriptError(name, "compile", err)
		}
		state.Push(function)
		if err := state.PCall(0, 0, nil); err != nil {
			return scriptError(name, "execute", err)
		}
		return nil
	})
}

// InvalidateModule removes a content-backed module from Lua's require cache so
// the next require observes the current layered VFS. It does not mutate an
// already-running component; transactional replacement remains responsible for
// switching component ownership.
func (r *Runtime) InvalidateModule(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("modruntime: module name is required")
	}
	return r.Run(ctx, func(state *lua.LState) error {
		packageTable, ok := state.GetGlobal("package").(*lua.LTable)
		if !ok {
			return errors.New("modruntime: Lua package table is unavailable")
		}
		loaded, ok := packageTable.RawGetString("loaded").(*lua.LTable)
		if !ok {
			return errors.New("modruntime: Lua package.loaded table is unavailable")
		}
		loaded.RawSetString(name, lua.LNil)
		return nil
	})
}

func runLoop(requests <-chan request, done chan<- struct{}, ready chan<- error, installers []Installer, modules []Module) {
	state := newSandboxedState()
	for _, installer := range installers {
		if err := installer.Install(state); err != nil {
			state.Close()
			ready <- fmt.Errorf("modruntime: install %q: %w", installer.Name, err)
			close(done)
			return
		}
	}
	for _, module := range modules {
		state.PreloadModule(module.Name, module.Loader)
	}
	ready <- nil
	defer close(done)
	defer state.Close()
	for current := range requests {
		if current.stop {
			current.done <- nil
			return
		}
		state.SetContext(current.ctx)
		err := current.run(state)
		state.RemoveContext()
		current.done <- err
	}
}

// newSandboxedState opens only deterministic, in-memory language facilities.
// The default GopherLua state also exposes the host filesystem, environment,
// process clock, ambient random generator, debug hooks, and dynamic source
// loading. Mods reach the outside world exclusively through reviewed engine
// capabilities, so those ambient APIs never enter the global environment.
func newSandboxedState() *lua.LState {
	state := lua.NewState(lua.Options{SkipOpenLibs: true})
	lua.OpenBase(state)
	lua.OpenPackage(state)
	lua.OpenTable(state)
	lua.OpenString(state)
	lua.OpenMath(state)

	for _, name := range []string{"dofile", "load", "loadfile", "loadstring", "print", "_printregs"} {
		state.SetGlobal(name, lua.LNil)
	}
	if mathTable, ok := state.GetGlobal("math").(*lua.LTable); ok {
		mathTable.RawSetString("random", lua.LNil)
		mathTable.RawSetString("randomseed", lua.LNil)
	}
	if packageTable, ok := state.GetGlobal("package").(*lua.LTable); ok {
		packageTable.RawSetString("loadlib", lua.LNil)
		packageTable.RawSetString("path", lua.LString(""))
		packageTable.RawSetString("cpath", lua.LString(""))
		// Preloaded engine capabilities are the only default loader. A reviewed
		// content installer may append its VFS-backed loader afterward.
		loaders := state.NewTable()
		loaders.Append(state.NewFunction(func(state *lua.LState) int {
			name := state.CheckString(1)
			preload := state.GetField(state.GetGlobal("package"), "preload")
			loader := state.GetField(preload, name)
			if loader == lua.LNil {
				state.Push(lua.LString(fmt.Sprintf("\n\tno preloaded module %q", name)))
				return 1
			}
			state.Push(loader)
			return 1
		}))
		packageTable.RawSetString("loaders", loaders)
		state.SetField(state.Get(lua.RegistryIndex), "_LOADERS", loaders)
	}
	return state
}
