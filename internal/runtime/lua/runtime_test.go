package modruntime

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func TestRuntimeExecutionBudgetInterruptsLuaAndRecovers(t *testing.T) {
	runtime := New()
	if err := runtime.SetExecutionBudget(25 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	err := runtime.Execute(context.Background(), fstest.MapFS{"loop.lua": &fstest.MapFile{Data: []byte(`while true do end`)}}, "loop.lua")
	if err == nil || !strings.Contains(err.Error(), "deadline exceeded") {
		t.Fatalf("budget error = %v", err)
	}
	if err := runtime.Execute(context.Background(), fstest.MapFS{"ok.lua": &fstest.MapFile{Data: []byte(`recovered = true`)}}, "ok.lua"); err != nil {
		t.Fatal(err)
	}
}

func TestRuntimeSupportsASeparateBoundedLifecycleBudget(t *testing.T) {
	runtime := New()
	if err := runtime.SetExecutionBudget(5 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())

	installPause := func(state *lua.LState) error {
		state.SetGlobal("pause_for_cold_asset", state.NewFunction(func(*lua.LState) int {
			time.Sleep(25 * time.Millisecond)
			return 0
		}))
		return nil
	}
	if err := runtime.runWithBudget(context.Background(), time.Second, installPause); err != nil {
		t.Fatal(err)
	}
	invoke := func(state *lua.LState) error {
		return state.DoString(`pause_for_cold_asset(); lifecycle_finished = true`)
	}
	if err := runtime.Run(context.Background(), invoke); err == nil || !strings.Contains(err.Error(), "deadline exceeded") {
		t.Fatalf("normal invocation escaped its tight budget: %v", err)
	}
	if err := runtime.runWithBudget(context.Background(), 100*time.Millisecond, invoke); err != nil {
		t.Fatalf("bounded lifecycle invocation failed: %v", err)
	}
}

func TestRuntimeReturnsSourceAwareTraceback(t *testing.T) {
	runtime := New()
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	err := runtime.Execute(context.Background(), fstest.MapFS{"broken.lua": &fstest.MapFile{Data: []byte("local function inner() error('boom') end\ninner()")}}, "broken.lua")
	if err == nil || !strings.Contains(err.Error(), "broken.lua") || !strings.Contains(err.Error(), "stack traceback") {
		t.Fatalf("diagnostic = %v", err)
	}
}

func TestRuntimeExecutesVersionedModuleOnOneOwner(t *testing.T) {
	t.Parallel()

	runtime := New()
	if err := runtime.RegisterModule(Module{Name: "dm.test/v1", Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"answer": func(state *lua.LState) int { state.Push(lua.LNumber(42)); return 1 },
		})
		state.Push(module)
		return 1
	}}); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	source := fstest.MapFS{"boot.lua": &fstest.MapFile{Data: []byte(`local test = require("dm.test/v1"); result = test.answer()`)}}
	if err := runtime.Execute(context.Background(), source, "boot.lua"); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		if got := state.GetGlobal("result"); got.String() != "42" {
			t.Fatalf("result = %s", got)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestRuntimeSerializesConcurrentCalls(t *testing.T) {
	t.Parallel()

	runtime := New()
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	const workers = 32
	var wait sync.WaitGroup
	wait.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wait.Done()
			if err := runtime.Run(context.Background(), func(state *lua.LState) error {
				value := state.GetGlobal("count")
				count, _ := value.(lua.LNumber)
				state.SetGlobal("count", count+1)
				return nil
			}); err != nil {
				t.Errorf("Run: %v", err)
			}
		}()
	}
	wait.Wait()
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		if got := state.GetGlobal("count"); got.String() != "32" {
			t.Fatalf("count = %s", got)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestScopeClosesInReverseAndJoinsErrors(t *testing.T) {
	t.Parallel()

	first := errors.New("first")
	second := errors.New("second")
	var calls []int
	scope := &Scope{}
	for index, failure := range []error{first, nil, second} {
		index, failure := index, failure
		if err := scope.Add(func() error { calls = append(calls, index); return failure }); err != nil {
			t.Fatal(err)
		}
	}
	err := scope.Close()
	if !errors.Is(err, first) || !errors.Is(err, second) {
		t.Fatalf("Close error = %v", err)
	}
	if !reflect.DeepEqual(calls, []int{2, 1, 0}) {
		t.Fatalf("calls = %v", calls)
	}
	if err := scope.Close(); err != nil {
		t.Fatal(err)
	}
	if err := scope.Add(func() error { return nil }); err == nil {
		t.Fatal("expected closed scope to reject resource")
	}
}

func TestHandlesRejectStaleAndMistypedReferences(t *testing.T) {
	t.Parallel()

	var handles Handles
	texture, err := handles.Add("texture", "first")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := handles.Get(Handle{Type: "sound", Slot: texture.Slot, Generation: texture.Generation}); err == nil {
		t.Fatal("expected mistyped handle to fail")
	}
	if err := handles.Release(texture); err != nil {
		t.Fatal(err)
	}
	if _, err := handles.Get(texture); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("stale Get error = %v", err)
	}
	replacement, err := handles.Add("texture", "second")
	if err != nil {
		t.Fatal(err)
	}
	if replacement.Slot != texture.Slot || replacement.Generation == texture.Generation {
		t.Fatalf("replacement = %#v, original = %#v", replacement, texture)
	}
}
