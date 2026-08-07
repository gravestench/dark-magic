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
	if _, err := evaluator.Evaluate(context.Background(), `dm.modules["secret/v1"]`); err == nil || !strings.Contains(err.Error(), "does not permit") {
		t.Fatalf("dm.modules error = %v", err)
	}
	candidates, err := evaluator.Complete(context.Background(), "dm.se")
	if err != nil || len(candidates) != 0 {
		t.Fatalf("denied completion = %#v, %v", candidates, err)
	}
	if _, err := evaluator.Evaluate(context.Background(), `answer = 42`); err == nil || !strings.Contains(err.Error(), "read-only") {
		t.Fatalf("assignment error = %v", err)
	}
	if result, err := evaluator.Evaluate(context.Background(), `getmetatable(_G), package, os`); err != nil || !strings.Contains(result.Text, "protected shell environment") || !strings.Contains(result.Text, "nil") {
		t.Fatalf("protected globals = %#v, %v", result, err)
	}
}

func TestEvaluatorExposesDiscoverableDarkMagicRoot(t *testing.T) {
	runtime := modruntime.New()
	if err := runtime.RegisterModule(modruntime.Module{Name: "dm.demo/v1", Help: modruntime.ModuleHelp{
		Summary: "Demonstrates documented shell APIs.",
		Commands: map[string]modruntime.CommandHelp{"greet": {
			Summary: "Greet a named adventurer.", Usage: "dm.demo.greet(name)",
			Parameters: []modruntime.ParameterHelp{{Name: "name", Type: "string", Description: "Adventurer name."}},
			Returns:    []modruntime.ReturnHelp{{Name: "greeting", Type: "string", Description: "Friendly greeting."}},
			Examples:   []string{`dm.demo.greet("Deckard")`},
		}},
	}, Loader: func(state *glua.LState) int {
		module := state.NewTable()
		module.RawSetString("name", glua.LString("demo"))
		module.RawSetString("greet", state.NewFunction(func(current *glua.LState) int {
			current.Push(glua.LString("Hello, " + current.CheckString(1)))
			return 1
		}))
		module.RawSetString("undocumented", state.NewFunction(func(*glua.LState) int { return 0 }))
		state.Push(module)
		return 1
	}}); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	evaluator, err := NewForPolicy(runtime, shell.Policy{Name: "developer", Mutable: true, Capabilities: []string{"dm.demo/v1"}})
	if err != nil {
		t.Fatal(err)
	}
	defer evaluator.Close()

	for _, source := range []string{`dm.demo.name`, `darkmagic.demo.name`, `dm.modules["dm.demo/v1"].name`} {
		if result, evaluateErr := evaluator.Evaluate(context.Background(), source); evaluateErr != nil || result.Text != `"demo"` {
			t.Fatalf("%s = %#v, %v", source, result, evaluateErr)
		}
	}
	if result, evaluateErr := evaluator.Evaluate(context.Background(), `table.concat(dm.capabilities(), ",")`); evaluateErr != nil || result.Text != `"dm.demo/v1"` {
		t.Fatalf("capabilities = %#v, %v", result, evaluateErr)
	}
	if result, evaluateErr := evaluator.Evaluate(context.Background(), `dm.help()`); evaluateErr != nil || !strings.Contains(result.Text, "demo") {
		t.Fatalf("help = %#v, %v", result, evaluateErr)
	}
	if result, evaluateErr := evaluator.Evaluate(context.Background(), `dm.help(dm.demo)`); evaluateErr != nil || !strings.Contains(result.Text, "greet") || !strings.Contains(result.Text, "undocumented") || !strings.Contains(result.Text, "documented shell APIs") {
		t.Fatalf("module help = %#v, %v", result, evaluateErr)
	}
	if result, evaluateErr := evaluator.Evaluate(context.Background(), `dm.help(dm.demo.greet)`); evaluateErr != nil || !strings.Contains(result.Text, "dm.demo.greet(name)") || !strings.Contains(result.Text, "Adventurer name") {
		t.Fatalf("command help = %#v, %v", result, evaluateErr)
	}
	if result, evaluateErr := evaluator.Evaluate(context.Background(), `dm.help("dm.demo.greet")`); evaluateErr != nil || !strings.Contains(result.Text, "Friendly greeting") {
		t.Fatalf("string command help = %#v, %v", result, evaluateErr)
	}
	if result, evaluateErr := evaluator.Evaluate(context.Background(), `dm.help("dm.demo.undocumented")`); evaluateErr != nil || !strings.Contains(result.Text, "Lua command provided by dm.demo/v1") {
		t.Fatalf("fallback command help = %#v, %v", result, evaluateErr)
	}
	candidates, err := evaluator.Complete(context.Background(), "dm.de")
	if err != nil || len(candidates) != 1 || candidates[0].Value != "dm.demo" || candidates[0].Detail != "Demonstrates documented shell APIs." {
		t.Fatalf("completion = %#v, %v", candidates, err)
	}
	candidates, err = evaluator.Complete(context.Background(), "dm.demo.gr")
	if err != nil || len(candidates) != 1 || candidates[0].Value != "dm.demo.greet" || candidates[0].Detail != "Greet a named adventurer." {
		t.Fatalf("command completion = %#v, %v", candidates, err)
	}
}
