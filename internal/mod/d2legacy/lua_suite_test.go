package d2legacy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	lua "github.com/yuin/gopher-lua"
)

// Lua suites are authored beside the d2legacy modules they cover. Each test is
// an ordered list of Lua callbacks and host actions. Callbacks execute inside
// the production authority; host actions run between callbacks so stepping the
// session never re-enters the Lua owner goroutine.
func TestLuaSuites(t *testing.T) {
	source := content.D2Legacy()
	paths, err := discoverLuaSuites(source)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no d2legacy Lua test suites found")
	}
	for _, path := range paths {
		path := path
		t.Run(luaSuiteName(path), func(t *testing.T) {
			t.Parallel()
			runLuaSuite(t, source, path)
		})
	}
}

func discoverLuaSuites(source fs.FS) ([]string, error) {
	var paths []string
	err := fs.WalkDir(source, "lua/d2legacy", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_test.lua") {
			paths = append(paths, path)
		}
		return nil
	})
	sort.Strings(paths)
	return paths, err
}

func luaSuiteName(path string) string {
	name := strings.TrimPrefix(path, "lua/")
	return strings.TrimSuffix(filepath.ToSlash(name), "_test.lua")
}

func runLuaSuite(t *testing.T, source fs.FS, path string) {
	t.Helper()
	fixture := newLuaSuiteFixture(t, source)
	var names []string
	config := luaSuiteConfig{seed: 42, records: fixtureRecords{}}
	err := fixture.authority.Runtime.Run(t.Context(), func(state *lua.LState) error {
		suite, err := loadLuaSuite(state, source, path)
		if err != nil {
			return err
		}
		tests, ok := suite.RawGetString("tests").(*lua.LTable)
		if !ok {
			return fmt.Errorf("%s: suite must return a table containing tests", path)
		}
		tests.ForEach(func(key, _ lua.LValue) { names = append(names, key.String()) })
		sort.Strings(names)
		return config.read(suite)
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Fatalf("%s: suite contains no tests", path)
	}
	validateLuaCoverage(t, path, names, config.covers)
	if !luaTestTierEnabled(config.tier) {
		t.Skipf("%s tier is not enabled", config.tier)
	}
	shuffleLuaTestCases(t, names)
	repeat := luaTestRepeat(t)
	for _, name := range names {
		name := name
		for iteration := 1; iteration <= repeat; iteration++ {
			testName := name
			if repeat > 1 {
				testName = fmt.Sprintf("%s/repeat_%d", name, iteration)
			}
			t.Run(testName, func(t *testing.T) {
				runLuaCase(t, source, path, name, config)
			})
		}
	}
}

func shuffleLuaTestCases(t *testing.T, names []string) {
	t.Helper()
	value := os.Getenv("DARK_MAGIC_LUA_TEST_ORDER_SEED")
	if value == "" {
		return
	}
	seed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		t.Fatalf("DARK_MAGIC_LUA_TEST_ORDER_SEED must be an integer, got %q", value)
	}
	rand.New(rand.NewSource(seed)).Shuffle(len(names), func(left, right int) {
		names[left], names[right] = names[right], names[left]
	})
}

func luaTestTierEnabled(tier string) bool {
	enabled := os.Getenv("DARK_MAGIC_LUA_TEST_TIERS")
	if enabled == "" {
		enabled = "fast,integration"
	}
	for _, candidate := range strings.Split(enabled, ",") {
		if strings.TrimSpace(candidate) == tier {
			return true
		}
	}
	return false
}

func luaTestRepeat(t *testing.T) int {
	t.Helper()
	value := os.Getenv("DARK_MAGIC_LUA_TEST_REPEAT")
	if value == "" {
		return 1
	}
	repeat, err := strconv.Atoi(value)
	if err != nil || repeat < 1 {
		t.Fatalf("DARK_MAGIC_LUA_TEST_REPEAT must be a positive integer, got %q", value)
	}
	return repeat
}

func TestLuaHarnessContract(t *testing.T) {
	t.Run("table encoding distinguishes arrays and objects", func(t *testing.T) {
		state := lua.NewState()
		defer state.Close()
		object := state.NewTable()
		encoded, err := luaValueToGo(object, map[*lua.LTable]bool{})
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := encoded.(map[string]any); !ok {
			t.Fatalf("empty table encoded as %T, want object", encoded)
		}
		array := state.NewTable()
		metatable := state.NewTable()
		metatable.RawSetString("__d2legacy_test_array", lua.LTrue)
		array.Metatable = metatable
		encoded, err = luaValueToGo(array, map[*lua.LTable]bool{})
		if err != nil {
			t.Fatal(err)
		}
		if values, ok := encoded.([]any); !ok || len(values) != 0 {
			t.Fatalf("marked empty array encoded as %#v", encoded)
		}
	})

	t.Run("cyclic input is rejected", func(t *testing.T) {
		state := lua.NewState()
		defer state.Close()
		table := state.NewTable()
		table.RawSetString("self", table)
		if _, err := luaValueToGo(table, map[*lua.LTable]bool{}); err == nil || !strings.Contains(err.Error(), "cyclic") {
			t.Fatalf("cyclic table error = %v", err)
		}
	})

	t.Run("tier selection is explicit", func(t *testing.T) {
		t.Setenv("DARK_MAGIC_LUA_TEST_TIERS", "fast,stress")
		if !luaTestTierEnabled("fast") || !luaTestTierEnabled("stress") || luaTestTierEnabled("real_assets") {
			t.Fatal("tier filter did not honor the configured set")
		}
	})

	t.Run("order randomization is reproducible", func(t *testing.T) {
		t.Setenv("DARK_MAGIC_LUA_TEST_ORDER_SEED", "73")
		left := []string{"a", "b", "c", "d", "e"}
		right := append([]string(nil), left...)
		shuffleLuaTestCases(t, left)
		shuffleLuaTestCases(t, right)
		if strings.Join(left, ",") != strings.Join(right, ",") {
			t.Fatalf("seeded orders differ: %v and %v", left, right)
		}
	})

	t.Run("authority actions reject narrower profiles", func(t *testing.T) {
		state := lua.NewState()
		defer state.Close()
		action := state.NewTable()
		action.RawSetString("step", lua.LNumber(1))
		err := runLuaAction(t, &luaSuiteFixture{config: luaSuiteConfig{profile: "module"}}, action)
		if err == nil || !strings.Contains(err.Error(), "requires the authority profile") {
			t.Fatalf("module step error = %v", err)
		}
	})

	t.Run("actions reject ambiguity and unknown operations", func(t *testing.T) {
		state := lua.NewState()
		defer state.Close()
		fixture := &luaSuiteFixture{config: luaSuiteConfig{profile: "authority"}}
		for name, action := range map[string]*lua.LTable{
			"unknown": state.NewTable(),
			"ambiguous": func() *lua.LTable {
				value := state.NewTable()
				value.RawSetString("step", lua.LNumber(1))
				value.RawSetString("checkpoint_restore", lua.LTrue)
				return value
			}(),
		} {
			t.Run(name, func(t *testing.T) {
				if err := runLuaAction(t, fixture, action); err == nil {
					t.Fatal("invalid action was accepted")
				}
			})
		}
	})

	t.Run("numeric action fields require non-negative integers", func(t *testing.T) {
		state := lua.NewState()
		defer state.Close()
		for _, value := range []lua.LValue{lua.LString("1"), lua.LNumber(-1), lua.LNumber(1.5), lua.LNil} {
			if _, err := luaNonNegativeInteger(value, "value"); err == nil {
				t.Fatalf("accepted invalid integer %s (%s)", value, value.Type())
			}
		}
	})

	t.Run("commands reject malformed fields before submission", func(t *testing.T) {
		state := lua.NewState()
		defer state.Close()
		command := state.NewTable()
		command.RawSetString("tick", lua.LNumber(-1))
		command.RawSetString("sequence", lua.LNumber(0))
		command.RawSetString("kind", lua.LString("example"))
		action := state.NewTable()
		action.RawSetString("submit", command)
		err := runLuaAction(t, &luaSuiteFixture{config: luaSuiteConfig{profile: "authority"}}, action)
		if err == nil || !strings.Contains(err.Error(), "command tick") {
			t.Fatalf("malformed command error = %v", err)
		}
	})

	t.Run("checkpoint restore requires a completed step", func(t *testing.T) {
		fixture := newLuaSuiteFixture(t, content.D2Legacy(), luaSuiteConfig{
			profile: "authority", seed: 42, records: fixtureRecords{},
		})
		state := lua.NewState()
		defer state.Close()
		action := state.NewTable()
		action.RawSetString("checkpoint_restore", lua.LTrue)
		err := runLuaAction(t, fixture, action)
		if err == nil || !strings.Contains(err.Error(), "completed step") {
			t.Fatalf("empty checkpoint error = %v", err)
		}
	})
}

type luaSuiteFixture struct {
	engine    *gameecs.Engine
	session   *gamesession.Session
	authority *Authority
	config    luaSuiteConfig
	scope     *modruntime.Scope
}

type luaSuiteConfig struct {
	apiVersion    int
	profile       string
	tier          string
	seed          uint64
	initialData   map[string]any
	records       fixtureRecords
	disableBudget bool
	covers        []string
}

func (config *luaSuiteConfig) read(suite *lua.LTable) error {
	config.apiVersion = int(lua.LVAsNumber(suite.RawGetString("api_version")))
	if config.apiVersion != 1 {
		return fmt.Errorf("suite must use d2legacy.tests/v1 (api_version = %d)", config.apiVersion)
	}
	config.profile = suite.RawGetString("profile").String()
	config.tier = suite.RawGetString("tier").String()
	if config.profile == "" || config.profile == "nil" {
		return fmt.Errorf("suite must declare a runtime profile")
	}
	if config.tier == "" || config.tier == "nil" {
		return fmt.Errorf("suite must declare a test tier")
	}
	if value := suite.RawGetString("seed"); value != lua.LNil {
		config.seed = uint64(lua.LVAsNumber(value))
	}
	if value, ok := suite.RawGetString("initial_data").(*lua.LTable); ok {
		converted, err := luaTableToStringMap(value)
		if err != nil {
			return fmt.Errorf("invalid initial_data: %w", err)
		}
		config.initialData = converted
	}
	if value, ok := suite.RawGetString("records").(*lua.LTable); ok {
		converted, err := luaValueToGo(value, map[*lua.LTable]bool{})
		if err != nil {
			return fmt.Errorf("invalid records: %w", err)
		}
		data, err := json.Marshal(converted)
		if err != nil {
			return fmt.Errorf("encode records: %w", err)
		}
		if err := json.Unmarshal(data, &config.records); err != nil {
			return fmt.Errorf("decode records: %w", err)
		}
	}
	config.disableBudget = suite.RawGetString("disable_execution_budget") == lua.LTrue
	if value, ok := suite.RawGetString("covers").(*lua.LTable); ok {
		value.ForEach(func(_, family lua.LValue) {
			config.covers = append(config.covers, family.String())
		})
	}
	return nil
}

func validateLuaCoverage(t *testing.T, suitePath string, cases, claims []string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "architecture", "d2legacy-test-coverage.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	caseSet := make(map[string]bool, len(cases))
	for _, name := range cases {
		caseSet[name] = true
	}
	claimSet := make(map[string]bool, len(claims))
	for _, family := range claims {
		if claimSet[family] {
			t.Errorf("%s declares duplicate coverage family %q", suitePath, family)
		}
		claimSet[family] = true
	}
	ledgerClaims := map[string]bool{}
	allLedgerFamilies := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Split(strings.TrimSpace(line), "\t")
		if len(fields) != 4 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		allLedgerFamilies[fields[0]] = true
		evidence := strings.SplitN(fields[1], "#", 2)
		if len(evidence) == 2 && evidence[0] == "internal/content/d2legacy/"+suitePath {
			ledgerClaims[fields[0]] = true
			if !caseSet[evidence[1]] {
				t.Errorf("coverage family %q names undiscovered case %q in %s", fields[0], evidence[1], suitePath)
			}
			if !claimSet[fields[0]] {
				t.Errorf("coverage family %q names %s but its executed suite metadata does not claim it", fields[0], suitePath)
			}
		}
	}
	for family := range claimSet {
		if !allLedgerFamilies[family] {
			t.Errorf("%s claims unknown coverage family %q", suitePath, family)
		}
	}
}

func luaTableToStringMap(table *lua.LTable) (map[string]any, error) {
	converted, err := luaValueToGo(table, map[*lua.LTable]bool{})
	if err != nil {
		return nil, err
	}
	result, ok := converted.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("want an object, got an array")
	}
	return result, nil
}

func luaValueToGo(value lua.LValue, active map[*lua.LTable]bool) (any, error) {
	switch value := value.(type) {
	case *lua.LNilType:
		return nil, nil
	case lua.LBool:
		return bool(value), nil
	case lua.LNumber:
		return float64(value), nil
	case lua.LString:
		return string(value), nil
	case *lua.LTable:
		if active[value] {
			return nil, fmt.Errorf("cyclic table")
		}
		active[value] = true
		defer delete(active, value)
		length := value.Len()
		array := make([]any, length)
		isArray := length > 0 || luaTableMarkedAsArray(value)
		object := make(map[string]any)
		var conversionErr error
		value.ForEach(func(key, item lua.LValue) {
			if conversionErr != nil {
				return
			}
			converted, err := luaValueToGo(item, active)
			if err != nil {
				conversionErr = err
				return
			}
			if number, ok := key.(lua.LNumber); ok && int(number) >= 1 && int(number) <= length && number == lua.LNumber(int(number)) {
				array[int(number)-1] = converted
				return
			}
			isArray = false
			stringKey, ok := key.(lua.LString)
			if !ok {
				conversionErr = fmt.Errorf("object key %s must be a string", key.String())
				return
			}
			object[string(stringKey)] = converted
		})
		if conversionErr != nil {
			return nil, conversionErr
		}
		if isArray {
			return array, nil
		}
		for index, item := range array {
			object[strconv.Itoa(index+1)] = item
		}
		return object, nil
	default:
		return nil, fmt.Errorf("unsupported Lua value %s", value.Type())
	}
}

func luaTableMarkedAsArray(table *lua.LTable) bool {
	metatable, ok := table.Metatable.(*lua.LTable)
	return ok && metatable.RawGetString("__d2legacy_test_array") == lua.LTrue
}

func newLuaSuiteFixture(t *testing.T, source fs.FS, configs ...luaSuiteConfig) *luaSuiteFixture {
	t.Helper()
	config := luaSuiteConfig{seed: 42, records: fixtureRecords{}}
	if len(configs) > 0 {
		config = configs[0]
	}
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		engine.Close()
		t.Fatal(err)
	}
	authority, err := startLuaSuiteProfile(t.Context(), source, config, engine, session)
	if err != nil {
		session.Close()
		engine.Close()
		t.Fatal(err)
	}
	fixture := &luaSuiteFixture{engine: engine, session: session, authority: authority, config: config, scope: &modruntime.Scope{}}
	t.Cleanup(func() {
		fixture.scope.Close()
		_ = fixture.authority.Stop(context.Background())
		_ = fixture.session.Close()
		_ = fixture.engine.Close()
	})
	return fixture
}

func startLuaSuiteProfile(ctx context.Context, source fs.FS, config luaSuiteConfig, engine *gameecs.Engine, session *gamesession.Session) (*Authority, error) {
	if config.profile == "" || config.profile == "authority" {
		return StartWithConfig(ctx, source, config.records, engine, session, Config{
			Seed: config.seed, InitialData: config.initialData, DisableExecutionBudget: config.disableBudget,
		})
	}
	if config.profile != "module" && config.profile != "ecs" {
		return nil, fmt.Errorf("d2legacy Lua test: unsupported runtime profile %q", config.profile)
	}
	streams, err := NewRandomStreams(config.seed)
	if err != nil {
		return nil, err
	}
	runtime := modruntime.New()
	if config.disableBudget {
		if err := runtime.SetExecutionBudget(0); err != nil {
			return nil, err
		}
	}
	if err := ConfigureModuleRuntime(runtime, source, config.records, streams, config.initialData); err != nil {
		return nil, err
	}
	if config.profile == "ecs" {
		if err := ConfigureECSRuntime(runtime, engine); err != nil {
			return nil, err
		}
	}
	if err := runtime.Start(ctx); err != nil {
		return nil, err
	}
	if config.profile == "ecs" {
		if err := runtime.Run(ctx, func(state *lua.LState) error {
			for _, moduleName := range []string{"d2legacy.components.shared", "d2legacy.components.melee"} {
				module, err := requireLuaModule(state, moduleName)
				if err != nil {
					return err
				}
				register, ok := module.RawGetString("register").(*lua.LFunction)
				if !ok {
					return fmt.Errorf("Lua test ECS profile: %s has no register function", moduleName)
				}
				if err := state.CallByParam(lua.P{Fn: register, Protect: true}); err != nil {
					return fmt.Errorf("Lua test ECS profile: register %s: %w", moduleName, err)
				}
			}
			return nil
		}); err != nil {
			_ = runtime.Stop(context.Background())
			return nil, err
		}
	}
	return &Authority{Runtime: runtime, Random: streams}, nil
}

func requireLuaModule(state *lua.LState, name string) (*lua.LTable, error) {
	require, ok := state.GetGlobal("require").(*lua.LFunction)
	if !ok {
		return nil, fmt.Errorf("Lua require function is unavailable")
	}
	if err := state.CallByParam(lua.P{Fn: require, NRet: 1, Protect: true}, lua.LString(name)); err != nil {
		return nil, err
	}
	value := state.Get(-1)
	state.Pop(1)
	module, ok := value.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("Lua module %s returned %s, want table", name, value.Type())
	}
	return module, nil
}

func loadLuaSuite(state *lua.LState, source fs.FS, path string) (*lua.LTable, error) {
	data, err := fs.ReadFile(source, path)
	if err != nil {
		return nil, err
	}
	function, err := state.Load(bytes.NewReader(data), "@"+path)
	if err != nil {
		return nil, fmt.Errorf("compile %s: %w", path, err)
	}
	state.Push(function)
	if err := state.PCall(0, 1, nil); err != nil {
		return nil, fmt.Errorf("execute %s: %w", path, err)
	}
	value := state.Get(-1)
	state.Pop(1)
	suite, ok := value.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("%s: suite must return a table", path)
	}
	return suite, nil
}

func runLuaCase(t *testing.T, source fs.FS, path, name string, config luaSuiteConfig) {
	t.Helper()
	fixture := newLuaSuiteFixture(t, source, config)
	var actions *lua.LTable
	err := fixture.authority.Runtime.Run(t.Context(), func(state *lua.LState) error {
		suite, err := loadLuaSuite(state, source, path)
		if err != nil {
			return err
		}
		tests := suite.RawGetString("tests").(*lua.LTable)
		var ok bool
		actions, ok = tests.RawGetString(name).(*lua.LTable)
		if !ok {
			return fmt.Errorf("%s: test %q must be an action array", path, name)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for index := 1; index <= actions.Len(); index++ {
		action, ok := actions.RawGetInt(index).(*lua.LTable)
		if !ok {
			t.Fatalf("%s %s action %d must be a table", path, name, index)
		}
		if err := runLuaAction(t, fixture, action); err != nil {
			t.Fatalf("%s %s action %d: %v", path, name, index, err)
		}
	}
}

func runLuaAction(t *testing.T, fixture *luaSuiteFixture, action *lua.LTable) error {
	t.Helper()
	known := []string{"run", "step", "checkpoint_restore", "checkpoint_parity_steps", "engine_update_ms", "submit", "submit_system"}
	present := 0
	for _, field := range known {
		if action.RawGetString(field) != lua.LNil {
			present++
		}
	}
	if present != 1 {
		return fmt.Errorf("action must contain exactly one recognized operation, got %d", present)
	}
	if function, ok := action.RawGetString("run").(*lua.LFunction); ok {
		return fixture.authority.Runtime.RunScoped(t.Context(), fixture.scope, func(state *lua.LState) error {
			return state.CallByParam(lua.P{Fn: function, Protect: true})
		})
	}
	if value := action.RawGetString("step"); value != lua.LNil {
		if fixture.config.profile != "authority" {
			return fmt.Errorf("step requires the authority profile, got %q", fixture.config.profile)
		}
		count, err := luaNonNegativeInteger(value, "step count")
		if err != nil {
			return err
		}
		if count < 1 {
			return fmt.Errorf("step count must be positive")
		}
		for range count {
			if err := fixture.session.Step(); err != nil {
				return err
			}
		}
		return nil
	}
	if action.RawGetString("checkpoint_restore") == lua.LTrue {
		if fixture.config.profile != "authority" {
			return fmt.Errorf("checkpoint restore requires the authority profile, got %q", fixture.config.profile)
		}
		return restoreLuaSuiteCheckpoint(t, fixture)
	}
	if value := action.RawGetString("checkpoint_parity_steps"); value != lua.LNil {
		if fixture.config.profile != "authority" {
			return fmt.Errorf("checkpoint parity requires the authority profile, got %q", fixture.config.profile)
		}
		steps, err := luaNonNegativeInteger(value, "checkpoint parity steps")
		if err != nil || steps < 1 {
			return fmt.Errorf("checkpoint parity steps must be a positive integer")
		}
		return compareLuaSuiteCheckpoint(t, fixture, steps)
	}
	if value := action.RawGetString("engine_update_ms"); value != lua.LNil {
		milliseconds, err := luaNonNegativeInteger(value, "engine update milliseconds")
		if err != nil {
			return err
		}
		return fixture.engine.Update(time.Duration(milliseconds) * time.Millisecond)
	}
	for field, authority := range map[string]simulation.Authority{
		"submit": simulation.AuthorityPlayer, "submit_system": simulation.AuthoritySystem,
	} {
		command, ok := action.RawGetString(field).(*lua.LTable)
		if !ok {
			continue
		}
		if fixture.config.profile != "authority" {
			return fmt.Errorf("command submission requires the authority profile, got %q", fixture.config.profile)
		}
		player := command.RawGetString("player").String()
		if authority == simulation.AuthoritySystem && player == "nil" {
			player = "system:lua-test"
		}
		payload, err := luaCommandPayload(command.RawGetString("payload"))
		if err != nil {
			return err
		}
		tick, err := luaNonNegativeInteger(command.RawGetString("tick"), "command tick")
		if err != nil {
			return err
		}
		sequence, err := luaNonNegativeInteger(command.RawGetString("sequence"), "command sequence")
		if err != nil {
			return err
		}
		kind := command.RawGetString("kind")
		if kind.Type() != lua.LTString || kind.String() == "" {
			return fmt.Errorf("command kind must be a non-empty string")
		}
		return fixture.session.Submit(simulation.Command{
			Tick:     uint64(tick),
			Sequence: uint64(sequence),
			Player:   player, Authority: authority,
			Kind:    kind.String(),
			Payload: payload,
		})
	}
	return fmt.Errorf("unknown action; want run, step, engine_update_ms, checkpoint_restore, submit, or submit_system")
}

func luaNonNegativeInteger(value lua.LValue, label string) (int, error) {
	number, ok := value.(lua.LNumber)
	if !ok || number < 0 || number != lua.LNumber(int(number)) {
		return 0, fmt.Errorf("%s must be a non-negative integer", label)
	}
	return int(number), nil
}

func luaCommandPayload(value lua.LValue) ([]byte, error) {
	if value == lua.LNil {
		return []byte("{}"), nil
	}
	if text, ok := value.(lua.LString); ok {
		return []byte(text), nil
	}
	converted, err := luaValueToGo(value, map[*lua.LTable]bool{})
	if err != nil {
		return nil, fmt.Errorf("invalid command payload: %w", err)
	}
	payload, err := json.Marshal(converted)
	if err != nil {
		return nil, fmt.Errorf("encode command payload: %w", err)
	}
	return payload, nil
}

func restoreLuaSuiteCheckpoint(t *testing.T, fixture *luaSuiteFixture) error {
	t.Helper()
	replay, err := fixture.session.Replay()
	if err != nil {
		return err
	}
	if len(replay.Checkpoints) == 0 {
		return fmt.Errorf("checkpoint_restore requires at least one completed step")
	}
	checkpoint := replay.Checkpoints[len(replay.Checkpoints)-1]
	fixture.scope.Close()
	if err := fixture.authority.Stop(t.Context()); err != nil {
		return err
	}
	if err := fixture.session.Close(); err != nil {
		return err
	}
	if err := fixture.engine.Close(); err != nil {
		return err
	}
	engine, err := gameecs.RestoreSnapshot(*checkpoint.Snapshot)
	if err != nil {
		return err
	}
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		engine.Close()
		return err
	}
	authority, err := StartWithConfig(t.Context(), content.D2Legacy(), fixture.config.records, engine, session, Config{
		Seed: fixture.config.seed, InitialData: fixture.config.initialData, Restore: checkpoint.Participants, DisableExecutionBudget: fixture.config.disableBudget,
	})
	if err != nil {
		session.Close()
		engine.Close()
		return err
	}
	fixture.engine, fixture.session, fixture.authority, fixture.scope = engine, session, authority, &modruntime.Scope{}
	return nil
}

func compareLuaSuiteCheckpoint(t *testing.T, fixture *luaSuiteFixture, steps int) error {
	t.Helper()
	if steps < 1 {
		return fmt.Errorf("checkpoint_parity_steps must be positive")
	}
	replay, err := fixture.session.Replay()
	if err != nil {
		return err
	}
	if len(replay.Checkpoints) == 0 {
		return fmt.Errorf("checkpoint parity requires a checkpoint")
	}
	checkpoint := replay.Checkpoints[len(replay.Checkpoints)-1]
	for range steps {
		if err := fixture.session.Step(); err != nil {
			return err
		}
	}
	originalReplay, err := fixture.session.Replay()
	if err != nil {
		return err
	}
	originalChecksum := originalReplay.Checkpoints[len(originalReplay.Checkpoints)-1].Checksum
	fixture.scope.Close()
	if err := fixture.authority.Stop(t.Context()); err != nil {
		return err
	}
	_ = fixture.session.Close()
	_ = fixture.engine.Close()
	engine, err := gameecs.RestoreSnapshot(*checkpoint.Snapshot)
	if err != nil {
		return err
	}
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		engine.Close()
		return err
	}
	authority, err := StartWithConfig(t.Context(), content.D2Legacy(), fixture.config.records, engine, session, Config{
		Seed: fixture.config.seed, InitialData: fixture.config.initialData, Restore: checkpoint.Participants, DisableExecutionBudget: fixture.config.disableBudget,
	})
	if err != nil {
		session.Close()
		engine.Close()
		return err
	}
	fixture.engine, fixture.session, fixture.authority, fixture.scope = engine, session, authority, &modruntime.Scope{}
	for range steps {
		if err := session.Step(); err != nil {
			return err
		}
	}
	restoredReplay, err := session.Replay()
	if err != nil {
		return err
	}
	restoredChecksum := restoredReplay.Checkpoints[len(restoredReplay.Checkpoints)-1].Checksum
	if restoredChecksum != originalChecksum {
		return fmt.Errorf("continued checksum = %s, want %s", restoredChecksum, originalChecksum)
	}
	return nil
}
