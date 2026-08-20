package luashell

import (
	"sort"
	"strings"

	glua "github.com/yuin/gopher-lua"
)

// hiddenModule records policy-withheld preload and loaded values so the runtime
// registry can be restored exactly after one scoped shell operation.
type hiddenModule struct {
	name            string
	preload, loaded glua.LValue
}

// restrictModules temporarily removes non-permitted modules from both Lua
// registries. The returned restoration must run before releasing the shared VM.
func (e *Evaluator) restrictModules(state *glua.LState) func() {
	packageTable, ok := state.GetGlobal("package").(*glua.LTable)
	if !ok {
		return func() {}
	}

	preload, _ := packageTable.RawGetString("preload").(*glua.LTable)
	loaded, _ := packageTable.RawGetString("loaded").(*glua.LTable)
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

// installRequire shadows global require with the evaluator's policy check while
// leaving the underlying runtime loader available to permitted module requests.
func (e *Evaluator) installRequire(state *glua.LState, environment *glua.LTable) {
	environment.RawSetString("require", state.NewFunction(func(current *glua.LState) int {
		return e.requireModule(current, current.CheckString(1))
	}))
}

// requireModule enforces capability admission before invoking the runtime's
// original require function and leaves its single result on the Lua stack.
func (e *Evaluator) requireModule(state *glua.LState, name string) int {
	if _, allowed := e.allowed[name]; !allowed {
		state.RaiseError("shell policy does not permit module %q", name)

		return 0
	}

	require, ok := state.GetGlobal("require").(*glua.LFunction)
	if !ok {
		state.RaiseError("Lua require is unavailable")

		return 0
	}

	state.Push(require)
	state.Push(glua.LString(name))
	state.Call(1, 1)

	return 1
}

// installDarkMagicRoot installs the policy-filtered engine and d2legacy roots.
// Both aliases share cached module values so requiring through either is consistent.
func (e *Evaluator) installDarkMagicRoot(state *glua.LState, environment *glua.LTable) {
	root := state.NewTable()
	modules := state.NewTable()
	root.RawSetString("modules", modules)

	root.RawSetString("require", state.NewFunction(func(current *glua.LState) int {
		return e.requireModule(current, current.CheckString(1))
	}))
	root.RawSetString("capabilities", e.capabilitiesFunction(state))
	root.RawSetString("help", e.helpFunction(state, darkMagicRootHelp(e.sortedAliases())))
	root.RawSetString("apropos", state.NewFunction(func(current *glua.LState) int {
		current.Push(glua.LString(e.apropos(current.CheckString(1))))

		return 1
	}))
	root.RawSetString("docs", state.NewFunction(func(current *glua.LState) int {
		current.Push(glua.LString(e.runtime.Markdown(e.modules)))

		return 1
	}))

	rootMeta := state.NewTable()
	rootMeta.RawSetString("__index", e.lazyModuleIndex(state, true))
	rootMeta.RawSetString("__metatable", glua.LString("protected engine/mod root"))
	state.SetMetatable(root, rootMeta)

	moduleMeta := state.NewTable()
	moduleMeta.RawSetString("__index", e.lazyModuleIndex(state, false))
	moduleMeta.RawSetString("__metatable", glua.LString("protected engine/mod modules"))
	state.SetMetatable(modules, moduleMeta)

	// Both roots share the policy-filtered resolver. Module IDs themselves retain
	// the ownership distinction: engine.* capabilities versus d2legacy.* mod APIs.
	environment.RawSetString("engine", root)
	environment.RawSetString("d2legacy", root)
}

// capabilitiesFunction returns module IDs in configured policy order, matching
// the order advertised by session policy and generated documentation.
func (e *Evaluator) capabilitiesFunction(state *glua.LState) *glua.LFunction {
	return state.NewFunction(func(current *glua.LState) int {
		values := current.NewTable()
		for index, module := range e.modules {
			values.RawSetInt(index+1, glua.LString(module))
		}

		current.Push(values)

		return 1
	})
}

// helpFunction supports both root overview and value/path-specific help while
// retaining the Lua command's one-string return convention.
func (e *Evaluator) helpFunction(state *glua.LState, overview string) *glua.LFunction {
	return state.NewFunction(func(current *glua.LState) int {
		if current.GetTop() == 0 {
			current.Push(glua.LString(overview))

			return 1
		}

		current.Push(glua.LString(e.helpFor(current, current.Get(1))))

		return 1
	})
}

// lazyModuleIndex resolves either an alias or exact module ID on first access,
// then caches the value under the caller's original table key.
func (e *Evaluator) lazyModuleIndex(state *glua.LState, aliasLookup bool) *glua.LFunction {
	return state.NewFunction(func(current *glua.LState) int {
		key := current.CheckString(2)
		name := key

		if aliasLookup {
			var found bool

			name, found = e.aliases[name]
			if !found {
				current.Push(glua.LNil)

				return 1
			}
		}

		e.requireModule(current, name)
		value := current.Get(-1)
		current.CheckTable(1).RawSetString(key, value)

		return 1
	})
}

// sortedAliases makes root help deterministic despite the alias lookup map.
func (e *Evaluator) sortedAliases() []string {
	aliases := make([]string, 0, len(e.aliases))
	for alias := range e.aliases {
		aliases = append(aliases, alias)
	}

	sort.Strings(aliases)

	return aliases
}

// darkMagicRootHelp constructs the stable overview text and appends only the
// policy-permitted aliases visible to this evaluator.
func darkMagicRootHelp(aliases []string) string {
	const overview = "Engine/mod shell roots\n" +
		"  engine.<capability>   lazy engine capability access\n" +
		"  d2legacy.<module>           lazy d2legacy module access\n" +
		"  engine.modules[<id>]  exact versioned module access\n" +
		"  engine.require(<id>)  policy-checked require\n" +
		"  engine.capabilities() permitted module IDs\n" +
		"  engine.help([value])  help for a module or command"

	if len(aliases) == 0 {
		return overview
	}

	return overview + "\nAvailable aliases: " + strings.Join(aliases, ", ")
}

// moduleAlias converts an exact versioned module ID into its stable shell member
// name while preserving the existing dot-to-underscore compatibility mapping.
func moduleAlias(module string) string {
	module = strings.TrimPrefix(module, "d2legacy.")
	module = strings.TrimPrefix(module, "engine.")

	if separator := strings.IndexByte(module, '/'); separator >= 0 {
		module = module[:separator]
	}

	return strings.ReplaceAll(module, ".", "_")
}
