package modruntime

import (
	"context"
	"fmt"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/presentation/navigation"
	lua "github.com/yuin/gopher-lua"
)

type navigationRequest struct {
	kind string
	id   string
	slot string
}

// Scenes adapts Lua-authored scenes to the engine navigation stack.
type Scenes struct {
	runtime *Runtime
	manager *navigation.Manager

	mu       sync.Mutex
	pending  []navigationRequest
	profiler SceneProfiler
	input    *inputstate.Store
}

type SceneProfiler interface{ CaptureSceneHeap(string) error }

func (s *Scenes) SetProfiler(profiler SceneProfiler) { s.profiler = profiler }

// SetInputStore applies capability-level suppression while nonfocused scenes
// continue simulation beneath transparent overlays.
func (s *Scenes) SetInputStore(input *inputstate.Store) { s.input = input }

// FrameContext labels work performed after Lua scene callbacks, including
// deferred renderer uploads and native drawing, with the focused scene.
func (s *Scenes) FrameContext(ctx context.Context) context.Context {
	name, ok := s.manager.Focused()
	if !ok {
		name = "none"
	}
	return pprof.WithLabels(ctx, pprof.Labels("scene", name))
}

// NewScenes constructs a Lua scene controller.
func NewScenes(runtime *Runtime, manager *navigation.Manager) *Scenes {
	return &Scenes{runtime: runtime, manager: manager}
}

// Module returns the dm.scene/v1 capability.
func (s *Scenes) Module() Module {
	return Module{Name: "dm.scene/v1", Help: documentedModule("Register Lua scenes and navigate the active scene stack.", map[string]CommandHelp{
		"register":       commandHelp("dm.scene.register(definition)", "Register a Lua-authored scene definition."),
		"replace":        commandHelp("dm.scene.replace(id [, payload])", "Replace the active scene."),
		"push":           commandHelp("dm.scene.push(id [, payload])", "Push a scene above the active scene."),
		"pop":            commandHelp("dm.scene.pop([payload])", "Pop the active scene."),
		"toggle_overlay": commandHelp("dm.scene.toggle_overlay(id, slot)", "Toggle an overlay in the left, right, or full spatial slot."),
	}), Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"register":       s.luaRegister,
			"replace":        s.luaRequest("replace"),
			"push":           s.luaRequest("push"),
			"pop":            s.luaRequest("pop"),
			"toggle_overlay": s.luaToggleOverlay,
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}

// Flush applies deferred navigation requests outside scene callbacks.
func (s *Scenes) Flush(ctx context.Context) error {
	s.mu.Lock()
	pending := s.pending
	s.pending = nil
	s.mu.Unlock()
	for _, request := range pending {
		var err error
		switch request.kind {
		case "replace":
			err = s.manager.Replace(ctx, request.id)
		case "push":
			err = s.manager.Push(ctx, request.id)
		case "pop":
			err = s.manager.Pop(ctx)
		case "toggle_overlay":
			err = s.manager.ToggleOverlay(ctx, request.id, request.slot)
		}
		if err != nil {
			return fmt.Errorf("modruntime: scene %s %q: %w", request.kind, request.id, err)
		}
	}
	return nil
}

func (s *Scenes) luaToggleOverlay(state *lua.LState) int {
	id := state.CheckString(1)
	slot := state.CheckString(2)
	s.mu.Lock()
	s.pending = append(s.pending, navigationRequest{kind: "toggle_overlay", id: id, slot: slot})
	s.mu.Unlock()
	return 0
}

// Update advances active scenes and then applies requested transitions.
func (s *Scenes) Update(ctx context.Context, elapsed time.Duration) error {
	if err := s.manager.Update(ctx, elapsed); err != nil {
		return err
	}
	return s.Flush(ctx)
}

// Render invokes active scene render callbacks from bottom to top.
func (s *Scenes) Render(ctx context.Context) error { return s.manager.Render(ctx) }

// Close destroys active scenes before their owning component unregisters the
// scene definitions.
func (s *Scenes) Close(ctx context.Context) error { return s.manager.Close(ctx) }

func (s *Scenes) luaRegister(state *lua.LState) int {
	scope, err := s.runtime.requireActiveScope()
	if err != nil {
		state.RaiseError("%v", err)
		return 0
	}
	id := state.CheckString(1)
	table := state.CheckTable(2)
	definition, err := parseSceneDefinition(id, table)
	if err != nil {
		state.RaiseError("%v", err)
		return 0
	}
	if err := s.manager.Register(id, func(context.Context) (navigation.Scene, error) {
		return &luaScene{id: id, runtime: s.runtime, definition: definition, scope: &Scope{}, profiler: s.profiler, input: s.input}, nil
	}); err != nil {
		state.RaiseError("registering scene %q: %v", id, err)
		return 0
	}
	if err := scope.Add(func() error { return s.manager.Unregister(id) }); err != nil {
		_ = s.manager.Unregister(id)
		state.RaiseError("owning scene registration %q: %v", id, err)
	}
	return 0
}

func (s *Scenes) luaRequest(kind string) lua.LGFunction {
	return func(state *lua.LState) int {
		id := ""
		if kind != "pop" {
			id = state.CheckString(1)
		}
		s.mu.Lock()
		s.pending = append(s.pending, navigationRequest{kind: kind, id: id})
		s.mu.Unlock()
		return 0
	}
}

type luaSceneDefinition struct {
	table                 *lua.LTable
	create, enter, update *lua.LFunction
	render                *lua.LFunction
	exit, destroy         *lua.LFunction
	blocks                bool
	passesInput           bool
	worldView             string
}

func parseSceneDefinition(id string, table *lua.LTable) (luaSceneDefinition, error) {
	definition := luaSceneDefinition{table: table}
	for name, target := range map[string]**lua.LFunction{
		"create": &definition.create, "enter": &definition.enter, "update": &definition.update, "render": &definition.render,
		"exit": &definition.exit, "destroy": &definition.destroy,
	} {
		value := table.RawGetString(name)
		if value == lua.LNil {
			continue
		}
		function, ok := value.(*lua.LFunction)
		if !ok {
			return luaSceneDefinition{}, fmt.Errorf("scene %q %s must be a function", id, name)
		}
		*target = function
	}
	if blocks := table.RawGetString("blocks_update_below"); blocks != lua.LNil {
		value, ok := blocks.(lua.LBool)
		if !ok {
			return luaSceneDefinition{}, fmt.Errorf("scene %q blocks_update_below must be a boolean", id)
		}
		definition.blocks = bool(value)
	}
	if passthrough := table.RawGetString("passes_input_below"); passthrough != lua.LNil {
		value, ok := passthrough.(lua.LBool)
		if !ok {
			return luaSceneDefinition{}, fmt.Errorf("scene %q passes_input_below must be a boolean", id)
		}
		definition.passesInput = bool(value)
	}
	definition.worldView = "center"
	if worldView := table.RawGetString("world_view"); worldView != lua.LNil {
		value, ok := worldView.(lua.LString)
		if !ok || value != "left" && value != "right" && value != "center" {
			return luaSceneDefinition{}, fmt.Errorf("scene %q world_view must be left, right, or center", id)
		}
		definition.worldView = string(value)
	}
	return definition, nil
}

type luaScene struct {
	id         string
	runtime    *Runtime
	definition luaSceneDefinition
	instance   *lua.LTable
	scope      *Scope
	profiler   SceneProfiler
	input      *inputstate.Store
}

// Scene construction legitimately decodes and composes cold MPQ assets. Keep
// the tight normal invocation budget for update/render callbacks, while giving
// transactional lifecycle work enough bounded time to finish on slower disks.
const sceneLifecycleBudget = 10 * time.Second

func (s *luaScene) Create(ctx context.Context) error {
	return s.callLifecycle(ctx, s.definition.create)
}
func (s *luaScene) Enter(ctx context.Context) error  { return s.callLifecycle(ctx, s.definition.enter) }
func (s *luaScene) Render(ctx context.Context) error { return s.call(ctx, s.definition.render) }
func (s *luaScene) Exit(ctx context.Context) error   { return s.callLifecycle(ctx, s.definition.exit) }
func (s *luaScene) BlocksUpdateBelow() bool          { return s.definition.blocks }
func (s *luaScene) PassesInputBelow() bool           { return s.definition.passesInput }
func (s *luaScene) WorldView() string                { return s.definition.worldView }
func (s *luaScene) Update(ctx context.Context, elapsed time.Duration) error {
	return s.UpdateInputFocused(ctx, elapsed, true, true, s.definition.worldView)
}

func (s *luaScene) UpdateFocused(ctx context.Context, elapsed time.Duration, focused bool) error {
	return s.UpdateInputFocused(ctx, elapsed, focused, focused, s.definition.worldView)
}

func (s *luaScene) UpdateInputFocused(ctx context.Context, elapsed time.Duration, focused, inputAllowed bool, worldView string) error {
	mode := "none"
	if focused {
		mode = "focused"
	} else if inputAllowed {
		mode = "gameplay"
	}
	return s.UpdateRoutedInput(ctx, elapsed, focused, mode, worldView)
}

func (s *luaScene) UpdateRoutedInput(ctx context.Context, elapsed time.Duration, focused bool, mode, worldView string) error {
	if s.definition.update == nil {
		return nil
	}
	var result error
	pprof.Do(ctx, pprof.Labels("scene", s.id), func(ctx context.Context) {
		invoke := func() error {
			return s.runtime.runScoped(ctx, s.scope, func(state *lua.LState) error {
				return state.CallByParam(lua.P{Fn: s.definition.update, NRet: 0, Protect: true}, s.instanceTable(state), lua.LNumber(elapsed.Seconds()), lua.LBool(focused), lua.LBool(mode != "none"), lua.LString(worldView))
			})
		}
		if mode == "none" && s.input != nil {
			result = s.input.Suppress(invoke)
		} else if mode == "pointer" && s.input != nil {
			result = s.input.PointerOnly(invoke)
		} else if mode == "gameplay_pointer" && s.input != nil {
			result = s.input.GameplayAndPointer(invoke)
		} else if mode == "gameplay" && s.input != nil {
			result = s.input.GameplayOnly(invoke)
		} else {
			result = invoke()
		}
	})
	return result
}
func (s *luaScene) Destroy(ctx context.Context) error {
	err := s.callLifecycle(ctx, s.definition.destroy)
	if s.profiler != nil {
		err = errorsJoin(err, s.profiler.CaptureSceneHeap(s.id))
	}
	return errorsJoin(err, s.scope.Close())
}

func (s *luaScene) callLifecycle(ctx context.Context, function *lua.LFunction) error {
	if function == nil {
		return nil
	}
	var result error
	pprof.Do(ctx, pprof.Labels("scene", s.id), func(ctx context.Context) {
		result = s.runtime.runScopedWithBudget(ctx, s.scope, sceneLifecycleBudget, func(state *lua.LState) error {
			return state.CallByParam(lua.P{Fn: function, NRet: 0, Protect: true}, s.instanceTable(state))
		})
	})
	return result
}

func (s *luaScene) call(ctx context.Context, function *lua.LFunction) error {
	if function == nil {
		return nil
	}
	var result error
	pprof.Do(ctx, pprof.Labels("scene", s.id), func(ctx context.Context) {
		result = s.runtime.runScoped(ctx, s.scope, func(state *lua.LState) error {
			return state.CallByParam(lua.P{Fn: function, NRet: 0, Protect: true}, s.instanceTable(state))
		})
	})
	return result
}

// instanceTable gives every navigation entry private mutable Lua state. Scene
// definitions are reusable blueprints; sharing their table meant transactional
// replacement could let the old scene's destroy callback erase or destroy the
// newly-entered scene's fields and render handles.
func (s *luaScene) instanceTable(state *lua.LState) *lua.LTable {
	if s.instance != nil {
		return s.instance
	}
	s.instance = state.NewTable()
	s.definition.table.ForEach(func(key, value lua.LValue) {
		s.instance.RawSet(key, value)
	})
	return s.instance
}

func errorsJoin(errs ...error) error {
	var result error
	for _, err := range errs {
		if err == nil {
			continue
		}
		if result == nil {
			result = err
		} else {
			result = fmt.Errorf("%v; %w", result, err)
		}
	}
	return result
}
