// Package modruntime hosts Dark Magic's serialized Lua execution environment.
package modruntime

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

// Module is a versioned engine capability exposed through Lua require.
type Module struct {
	Name   string
	Loader lua.LGFunction
}

// Installer configures the Lua state before boot code runs, for behavior such
// as content-backed require searchers that is not itself a required module.
type Installer struct {
	Name    string
	Install func(*lua.LState) error
}

type request struct {
	run  func(*lua.LState) error
	stop bool
	done chan error
}

// Runtime owns one Lua state on one goroutine. All access is serialized through
// Run; the state is never exposed outside the duration of a Run callback.
type Runtime struct {
	mu         sync.RWMutex
	started    bool
	modules    []Module
	installers []Installer
	requests   chan request
	done       chan struct{}

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
	return r.Run(ctx, func(state *lua.LState) error {
		previous := r.activeScope
		r.activeScope = scope
		defer func() { r.activeScope = previous }()
		return fn(state)
	})
}

func (r *Runtime) requireActiveScope() (*Scope, error) {
	if r.activeScope == nil {
		return nil, errors.New("modruntime: capability requires an active component scope")
	}
	return r.activeScope, nil
}

// New constructs a stopped runtime.
func New() *Runtime { return &Runtime{} }

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
	r.modules = append(r.modules, module)
	return nil
}

// Start creates the Lua state on its owner goroutine.
func (r *Runtime) Start(ctx context.Context) error {
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
	if fn == nil {
		return errors.New("modruntime: nil invocation")
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if !r.started {
		return errors.New("modruntime: runtime is not started")
	}
	response := make(chan error, 1)
	select {
	case r.requests <- request{run: fn, done: response}:
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
	data, err := fs.ReadFile(source, name)
	if err != nil {
		return fmt.Errorf("modruntime: read %q: %w", name, err)
	}
	return r.Run(ctx, func(state *lua.LState) error {
		function, err := state.Load(bytes.NewReader(data), "@"+name)
		if err != nil {
			return fmt.Errorf("compile %q: %w", name, err)
		}
		state.Push(function)
		if err := state.PCall(0, 0, nil); err != nil {
			return fmt.Errorf("execute %q: %w", name, err)
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
	state := lua.NewState()
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
		current.done <- current.run(state)
	}
}
