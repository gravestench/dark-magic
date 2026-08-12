package d2legacy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
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
	for _, name := range names {
		name := name
		t.Run(name, func(t *testing.T) {
			runLuaCase(t, source, path, name, config)
		})
	}
}

type luaSuiteFixture struct {
	engine    *gameecs.Engine
	session   *gamesession.Session
	authority *Authority
	config    luaSuiteConfig
}

type luaSuiteConfig struct {
	seed        uint64
	initialData map[string]any
	records     fixtureRecords
}

func (config *luaSuiteConfig) read(suite *lua.LTable) error {
	if value := suite.RawGetString("seed"); value != lua.LNil {
		config.seed = uint64(lua.LVAsNumber(value))
	}
	for field, destination := range map[string]*map[string]any{
		"initial_data_json": &config.initialData,
	} {
		value := suite.RawGetString(field)
		if value == lua.LNil {
			continue
		}
		if err := json.Unmarshal([]byte(value.String()), destination); err != nil {
			return fmt.Errorf("invalid %s: %w", field, err)
		}
	}
	if value := suite.RawGetString("records_json"); value != lua.LNil {
		if err := json.Unmarshal([]byte(value.String()), &config.records); err != nil {
			return fmt.Errorf("invalid records_json: %w", err)
		}
	}
	return nil
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
		Seed: config.seed, InitialData: config.initialData,
	})
	if err != nil {
		session.Close()
		engine.Close()
		t.Fatal(err)
	}
	fixture := &luaSuiteFixture{engine: engine, session: session, authority: authority, config: config}
	t.Cleanup(func() {
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
		return fixture.authority.Runtime.Run(t.Context(), func(state *lua.LState) error {
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
		return fixture.session.Submit(simulation.Command{
			Tick:     uint64(lua.LVAsNumber(command.RawGetString("tick"))),
			Sequence: uint64(lua.LVAsNumber(command.RawGetString("sequence"))),
			Player:   player, Authority: authority,
			Kind:    command.RawGetString("kind").String(),
			Payload: []byte(command.RawGetString("payload").String()),
		})
	}
	return fmt.Errorf("unknown action; want run, step, checkpoint_restore, submit, or submit_system")
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
		Seed: fixture.config.seed, InitialData: fixture.config.initialData, Restore: checkpoint.Participants,
	})
	if err != nil {
		session.Close()
		engine.Close()
		return err
	}
	fixture.engine, fixture.session, fixture.authority = engine, session, authority
	return nil
}
