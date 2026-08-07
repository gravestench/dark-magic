package luashell

import (
	"context"
	"strings"
	"testing"

	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/shell"
	glua "github.com/yuin/gopher-lua"
)

func TestEvaluatorPersistsLocalsFormatsValuesAndCompletesWithoutExecution(t *testing.T) {
	runtime := modruntime.New()
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	evaluator, err := New(runtime)
	if err != nil {
		t.Fatal(err)
	}
	defer evaluator.Close()
	if result, err := evaluator.Evaluate(context.Background(), "answer = 42"); err != nil || result.Text != "ok" {
		t.Fatalf("assignment = %#v, %v", result, err)
	}
	if result, err := evaluator.Evaluate(context.Background(), "answer"); err != nil || result.Text != "42" {
		t.Fatalf("answer = %#v, %v", result, err)
	}
	if result, err := evaluator.Evaluate(context.Background(), `{name="hero", level=2}`); err != nil || !strings.Contains(result.Text, "level=2") {
		t.Fatalf("table = %#v, %v", result, err)
	}
	if result, err := evaluator.Evaluate(context.Background(), `print("visible", 7)`); err != nil || result.Text != "visible\t7" || result.Kind != "output" {
		t.Fatalf("print = %#v, %v", result, err)
	}
	if result, err := evaluator.Evaluate(context.Background(), `printregs()`); err != nil || !strings.Contains(result.Text, "Lua call frames:") {
		t.Fatalf("printregs = %#v, %v", result, err)
	}
	candidates, err := evaluator.Complete(context.Background(), "pri")
	if err != nil || len(candidates) == 0 || candidates[0].Value != "print" {
		t.Fatalf("completion = %#v, %v", candidates, err)
	}
	if err := runtime.Run(context.Background(), func(state *glua.LState) error {
		methods := state.NewTable()
		methods.RawSetString("inspect", state.NewFunction(func(*glua.LState) int { return 0 }))
		metatable := state.NewTable()
		metatable.RawSetString("__index", methods)
		value := state.NewUserData()
		state.SetMetatable(value, metatable)
		state.SetGlobal("hero", value)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	candidates, err = evaluator.Complete(context.Background(), "hero.in")
	if err != nil || len(candidates) != 1 || candidates[0].Value != "hero.inspect" || candidates[0].Detail != "userdata member" {
		t.Fatalf("userdata completion = %#v, %v", candidates, err)
	}
}

func TestEvaluatorEnforcesSessionPolicy(t *testing.T) {
	runtime := modruntime.New()
	if err := runtime.RegisterModule(modruntime.Module{Name: "secret/v1", Loader: func(state *glua.LState) int {
		state.Push(state.NewTable())
		return 1
	}}); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	evaluator, err := NewForPolicy(runtime, shell.Policy{Name: "observer", Mutable: false})
	if err != nil {
		t.Fatal(err)
	}
	defer evaluator.Close()
	if _, err := evaluator.Evaluate(context.Background(), `require("secret/v1")`); err == nil || !strings.Contains(err.Error(), "does not permit") {
		t.Fatalf("require error = %v", err)
	}
	if _, err := evaluator.Evaluate(context.Background(), `answer = 42`); err == nil || !strings.Contains(err.Error(), "read-only") {
		t.Fatalf("assignment error = %v", err)
	}
	if result, err := evaluator.Evaluate(context.Background(), `getmetatable(_G), package, os`); err != nil || !strings.Contains(result.Text, "protected shell environment") || !strings.Contains(result.Text, "nil") {
		t.Fatalf("protected globals = %#v, %v", result, err)
	}
}
