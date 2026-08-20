package luashell

import (
	"context"
	"strings"
	"testing"

	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/shell"
	glua "github.com/yuin/gopher-lua"
)

// startRuntime registers modules in caller order and starts the real serialized
// runtime. Cleanup reports stop failures without hiding the original assertion.
func startRuntime(t *testing.T, modules ...modruntime.Module) *modruntime.Runtime {
	t.Helper()

	runtime := modruntime.New()
	for _, module := range modules {
		if err := runtime.RegisterModule(module); err != nil {
			t.Fatal(err)
		}
	}

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := runtime.Stop(context.Background()); err != nil {
			t.Errorf("stop runtime: %v", err)
		}
	})

	return runtime
}

// unrestrictedEvaluator constructs the all-module mutable shell and closes its
// scope before startRuntime's later cleanup stops the owning runtime.
func unrestrictedEvaluator(t *testing.T, runtime *modruntime.Runtime) *Evaluator {
	t.Helper()

	evaluator, err := New(runtime)
	if err != nil {
		t.Fatal(err)
	}

	registerEvaluatorCleanup(t, evaluator)

	return evaluator
}

// policyEvaluator constructs a shell with the exact policy under test and
// registers deterministic scope cleanup.
func policyEvaluator(t *testing.T, runtime *modruntime.Runtime, policy shell.Policy) *Evaluator {
	t.Helper()

	evaluator, err := NewForPolicy(runtime, policy)
	if err != nil {
		t.Fatal(err)
	}

	registerEvaluatorCleanup(t, evaluator)

	return evaluator
}

// registerEvaluatorCleanup ensures every persistent Lua scope closes before its runtime.
func registerEvaluatorCleanup(t *testing.T, evaluator *Evaluator) {
	t.Helper()

	t.Cleanup(func() {
		if err := evaluator.Close(); err != nil {
			t.Errorf("close evaluator: %v", err)
		}
	})
}

// evaluate executes one source string and fails at the call site, keeping each
// behavioral assertion focused on returned kind or text.
func evaluate(t *testing.T, evaluator *Evaluator, source string) shell.Result {
	t.Helper()

	result, err := evaluator.Evaluate(context.Background(), source)
	if err != nil {
		t.Fatalf("evaluate %q: %v", source, err)
	}

	return result
}

// assertEvaluationError verifies both rejection and its policy-relevant diagnostic.
func assertEvaluationError(t *testing.T, evaluator *Evaluator, source, diagnostic string) {
	t.Helper()

	_, err := evaluator.Evaluate(context.Background(), source)
	if err == nil || !strings.Contains(err.Error(), diagnostic) {
		t.Fatalf("evaluate %q error = %v, want %q", source, err, diagnostic)
	}
}

// registerUserdataFixture installs a raw userdata method table so completion can
// prove it inspects metatables without invoking methods.
func registerUserdataFixture(t *testing.T, runtime *modruntime.Runtime) {
	t.Helper()

	err := runtime.Run(context.Background(), func(state *glua.LState) error {
		methods := state.NewTable()
		methods.RawSetString("inspect", state.NewFunction(func(*glua.LState) int { return 0 }))

		metatable := state.NewTable()
		metatable.RawSetString("__index", methods)

		value := state.NewUserData()
		state.SetMetatable(value, metatable)
		state.SetGlobal("hero", value)

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestEvaluatorPersistsLocalsFormatsValuesAndCompletesWithoutExecution covers
// persistent scope state, bounded formatting, captured output, and raw completion.
func TestEvaluatorPersistsLocalsFormatsValuesAndCompletesWithoutExecution(t *testing.T) {
	runtime := startRuntime(t)
	evaluator := unrestrictedEvaluator(t, runtime)

	if result := evaluate(t, evaluator, "answer = 42"); result.Text != "ok" {
		t.Fatalf("assignment = %#v", result)
	}

	if result := evaluate(t, evaluator, "answer"); result.Text != "42" {
		t.Fatalf("answer = %#v", result)
	}

	if result := evaluate(t, evaluator, `{name="hero", level=2}`); !strings.Contains(result.Text, "level = 2") {
		t.Fatalf("table = %#v", result)
	}

	structured := evaluate(
		t,
		evaluator,
		`local value={nested={answer=42}}; value.self=value; return value`,
	)
	if !strings.Contains(structured.Text, "nested = {\n") ||
		!strings.Contains(structured.Text, "self = <cycle>") {
		t.Fatalf("structured table = %#v", structured)
	}

	printed := evaluate(t, evaluator, `print("visible", 7)`)
	if printed.Text != "visible\t7" || printed.Kind != "output" {
		t.Fatalf("print = %#v", printed)
	}

	multiline := evaluate(t, evaluator, `"heading\n  indented\tvalue"`)
	if multiline.Text != "heading\n  indented\tvalue" {
		t.Fatalf("multiline string = %#v", multiline)
	}

	if result := evaluate(t, evaluator, `printregs()`); !strings.Contains(result.Text, "Lua call frames:") {
		t.Fatalf("printregs = %#v", result)
	}

	candidates, err := evaluator.Complete(context.Background(), "pri")
	if err != nil || len(candidates) == 0 || candidates[0].Value != "print" {
		t.Fatalf("completion = %#v, %v", candidates, err)
	}

	registerUserdataFixture(t, runtime)

	candidates, err = evaluator.Complete(context.Background(), "hero.in")

	invalidUserdataCompletion := err != nil ||
		len(candidates) != 1 ||
		candidates[0].Value != "hero.inspect" ||
		candidates[0].Detail != "userdata member"
	if invalidUserdataCompletion {
		t.Fatalf("userdata completion = %#v, %v", candidates, err)
	}
}

// secretModule exposes metadata and a loader that an observer policy must never discover.
func secretModule() modruntime.Module {
	return modruntime.Module{
		Name: "secret/v1",
		Help: modruntime.ModuleHelp{Summary: "Top secret capability."},
		Loader: func(state *glua.LState) int {
			state.Push(state.NewTable())

			return 1
		},
	}
}

// TestEvaluatorEnforcesSessionPolicy verifies denied require/discovery, read-only
// globals, protected metatables, and hidden unsafe standard libraries.
func TestEvaluatorEnforcesSessionPolicy(t *testing.T) {
	runtime := startRuntime(t, secretModule())
	evaluator := policyEvaluator(t, runtime, shell.Policy{Name: "observer", Mutable: false})

	assertEvaluationError(t, evaluator, `require("secret/v1")`, "does not permit")
	assertEvaluationError(t, evaluator, `engine.modules["secret/v1"]`, "does not permit")
	assertEvaluationError(t, evaluator, `answer = 42`, "read-only")

	if result := evaluate(t, evaluator, `engine.docs()`); strings.Contains(result.Text, "secret") {
		t.Fatalf("policy-filtered docs = %#v", result)
	}

	search := evaluate(t, evaluator, `engine.apropos("secret")`)
	if !strings.Contains(search.Text, "No permitted") {
		t.Fatalf("policy-filtered search = %#v", search)
	}

	candidates, err := evaluator.Complete(context.Background(), "d2legacy.se")
	if err != nil || len(candidates) != 0 {
		t.Fatalf("denied completion = %#v, %v", candidates, err)
	}

	protected := evaluate(t, evaluator, `getmetatable(_G), package, os`)
	if !strings.Contains(protected.Text, "protected shell environment") ||
		!strings.Contains(protected.Text, "nil") {
		t.Fatalf("protected globals = %#v", protected)
	}
}

// TestCompletionRetainsTopLevelVersionedIDsForUnknownDottedBase protects the
// legacy fallback where a dot in an exact module ID is not mistaken for member access.
func TestCompletionRetainsTopLevelVersionedIDsForUnknownDottedBase(t *testing.T) {
	module := modruntime.Module{
		Name: "foo.bar/v1",
		Loader: func(state *glua.LState) int {
			state.Push(state.NewTable())

			return 1
		},
	}
	runtime := startRuntime(t, module)
	evaluator := unrestrictedEvaluator(t, runtime)

	candidates, err := evaluator.Complete(context.Background(), "foo.")
	if err != nil || len(candidates) != 1 || candidates[0].Value != "foo.bar/v1" {
		t.Fatalf("dotted module completion = %#v, %v", candidates, err)
	}
}

// demoModule provides documented, undocumented, and scalar members so root
// discovery tests cover every help fallback and completion source.
func demoModule() modruntime.Module {
	return modruntime.Module{
		Name: "d2legacy.demo/v1",
		Help: modruntime.ModuleHelp{
			Summary: "Demonstrates documented shell APIs.",
			Commands: map[string]modruntime.CommandHelp{
				"greet": {
					Summary: "Greet a named adventurer.",
					Usage:   "d2legacy.demo.greet(name)",
					Parameters: []modruntime.ParameterHelp{{
						Name: "name", Type: "string", Description: "Adventurer name.",
					}},
					Returns: []modruntime.ReturnHelp{{
						Name: "greeting", Type: "string", Description: "Friendly greeting.",
					}},
					Examples: []string{`engine.demo.greet("Deckard")`},
				},
			},
		},
		Loader: loadDemoModule,
	}
}

// loadDemoModule constructs the table shape consumed by discovery and help tests.
func loadDemoModule(state *glua.LState) int {
	module := state.NewTable()
	module.RawSetString("name", glua.LString("demo"))
	module.RawSetString("greet", state.NewFunction(func(current *glua.LState) int {
		current.Push(glua.LString("Hello, " + current.CheckString(1)))

		return 1
	}))
	module.RawSetString("undocumented", state.NewFunction(func(*glua.LState) int { return 0 }))
	state.Push(module)

	return 1
}

// TestEvaluatorExposesDiscoverableD2LegacyRoot verifies lazy cached access,
// capability order, every help form, generated docs, search, and completion details.
func TestEvaluatorExposesDiscoverableD2LegacyRoot(t *testing.T) {
	runtime := startRuntime(t, demoModule())
	policy := shell.Policy{
		Name:         "developer",
		Mutable:      true,
		Capabilities: []string{"d2legacy.demo/v1"},
	}
	evaluator := policyEvaluator(t, runtime, policy)

	// Repeating alias access proves the first lazy lookup caches the module without
	// changing results; exact versioned access must resolve the same table.
	paths := []string{
		`d2legacy.demo.name`,
		`d2legacy.demo.name`,
		`engine.modules["d2legacy.demo/v1"].name`,
	}
	for _, source := range paths {
		if result := evaluate(t, evaluator, source); result.Text != `"demo"` {
			t.Fatalf("%s = %#v", source, result)
		}
	}

	capabilities := evaluate(t, evaluator, `table.concat(engine.capabilities(), ",")`)
	if capabilities.Text != `"d2legacy.demo/v1"` {
		t.Fatalf("capabilities = %#v", capabilities)
	}

	assertHelpContains(t, evaluator, `engine.help()`, "demo")
	assertHelpContains(
		t,
		evaluator,
		`engine.help(d2legacy.demo)`,
		"greet",
		"undocumented",
		"documented shell APIs",
	)
	assertHelpContains(
		t,
		evaluator,
		`engine.help(d2legacy.demo.greet)`,
		"d2legacy.demo.greet(name)",
		"Adventurer name",
	)
	assertHelpContains(t, evaluator, `engine.help("d2legacy.demo.greet")`, "Friendly greeting")
	assertHelpContains(
		t,
		evaluator,
		`engine.help("d2legacy.demo.undocumented")`,
		"Lua command provided by d2legacy.demo/v1",
	)
	assertHelpContains(t, evaluator, `engine.apropos("adventurer")`, "d2legacy.demo.greet")
	assertHelpContains(
		t,
		evaluator,
		`engine.docs()`,
		"# Dark Magic Lua API",
		"d2legacy.demo.greet(name)",
	)

	assertCompletion(t, evaluator, "d2legacy.de", "d2legacy.demo", "member")
	assertCompletion(
		t,
		evaluator,
		"d2legacy.demo.gr",
		"d2legacy.demo.greet",
		"Greet a named adventurer.",
	)
}

// assertHelpContains evaluates a help/search/docs command and requires every
// semantic fragment, keeping long scenario assertions readable.
func assertHelpContains(t *testing.T, evaluator *Evaluator, source string, fragments ...string) {
	t.Helper()

	result := evaluate(t, evaluator, source)
	for _, fragment := range fragments {
		if !strings.Contains(result.Text, fragment) {
			t.Fatalf("%s = %#v, missing %q", source, result, fragment)
		}
	}
}

// assertCompletion verifies the one-candidate value and rationale exposed to adapters.
func assertCompletion(t *testing.T, evaluator *Evaluator, source, value, detail string) {
	t.Helper()

	candidates, err := evaluator.Complete(context.Background(), source)

	invalid := err != nil ||
		len(candidates) != 1 ||
		candidates[0].Value != value ||
		candidates[0].Detail != detail
	if invalid {
		t.Fatalf("completion %q = %#v, %v", source, candidates, err)
	}
}

// engineSettingsModule models a non-mod engine capability with a friendly alias.
func engineSettingsModule() modruntime.Module {
	return modruntime.Module{
		Name: "engine.settings/v1",
		Help: modruntime.ModuleHelp{Summary: "Client settings."},
		Loader: func(state *glua.LState) int {
			module := state.NewTable()
			module.RawSetString("name", glua.LString("settings"))
			state.Push(module)

			return 1
		},
	}
}

// TestEvaluatorExposesEngineCapabilityWithoutVersionedRequire verifies engine
// aliases use the same admitted lazy resolver as d2legacy modules.
func TestEvaluatorExposesEngineCapabilityWithoutVersionedRequire(t *testing.T) {
	runtime := startRuntime(t, engineSettingsModule())
	policy := shell.Policy{
		Name:         "developer",
		Mutable:      true,
		Capabilities: []string{"engine.settings/v1"},
	}
	evaluator := policyEvaluator(t, runtime, policy)

	result := evaluate(t, evaluator, `engine.settings.name`)
	if result.Text != `"settings"` {
		t.Fatalf("engine.settings = %#v", result)
	}

	candidates, err := evaluator.Complete(context.Background(), "engine.set")
	if err != nil || len(candidates) != 1 || candidates[0].Value != "engine.settings" {
		t.Fatalf("settings completion = %#v, %v", candidates, err)
	}
}
