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
	return nil
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
	authority, err := StartWithConfig(t.Context(), source, config.records, engine, session, Config{
		Seed: config.seed, InitialData: config.initialData, DisableExecutionBudget: config.disableBudget,
	})
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
	if function, ok := action.RawGetString("run").(*lua.LFunction); ok {
		return fixture.authority.Runtime.RunScoped(t.Context(), fixture.scope, func(state *lua.LState) error {
			return state.CallByParam(lua.P{Fn: function, Protect: true})
		})
	}
	if value := action.RawGetString("step"); value != lua.LNil {
		count := int(lua.LVAsNumber(value))
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
		return restoreLuaSuiteCheckpoint(t, fixture)
	}
	if value := action.RawGetString("checkpoint_parity_steps"); value != lua.LNil {
		return compareLuaSuiteCheckpoint(t, fixture, int(lua.LVAsNumber(value)))
	}
	if value := action.RawGetString("engine_update_ms"); value != lua.LNil {
		return fixture.engine.Update(time.Duration(lua.LVAsNumber(value)) * time.Millisecond)
	}
	for field, authority := range map[string]simulation.Authority{
		"submit": simulation.AuthorityPlayer, "submit_system": simulation.AuthoritySystem,
	} {
		command, ok := action.RawGetString(field).(*lua.LTable)
		if !ok {
			continue
		}
		player := command.RawGetString("player").String()
		if authority == simulation.AuthoritySystem && player == "nil" {
			player = "system:lua-test"
		}
		payload, err := luaCommandPayload(command.RawGetString("payload"))
		if err != nil {
			return err
		}
		return fixture.session.Submit(simulation.Command{
			Tick:     uint64(lua.LVAsNumber(command.RawGetString("tick"))),
			Sequence: uint64(lua.LVAsNumber(command.RawGetString("sequence"))),
			Player:   player, Authority: authority,
			Kind:    command.RawGetString("kind").String(),
			Payload: payload,
		})
	}
	return fmt.Errorf("unknown action; want run, step, engine_update_ms, checkpoint_restore, submit, or submit_system")
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
	fixture.engine, fixture.session, fixture.authority, fixture.scope = engine, session, authority, &modruntime.Scope{}
	return nil
}
