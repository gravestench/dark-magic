package input

import (
	"io"
	"log/slog"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

type testLuaManager struct{ state *lua.LState }

func (m *testLuaManager) Ready() bool                                { return true }
func (m *testLuaManager) WithState(fn func(*lua.LState) error) error { return fn(m.state) }
func (m *testLuaManager) GlobalsExist(...string) bool                { return true }
func (m *testLuaManager) RebuildState()                              {}

func TestLuaKeyPressedSubscriptionDispatches(t *testing.T) {
	state := lua.NewState()
	defer state.Close()
	service := &Service{
		lua:                 &testLuaManager{state: state},
		keyStates:           make(map[int32]InputState),
		keyPressedCallbacks: make(map[int32][]*lua.LFunction),
	}
	service.SetLogger(slog.New(slog.NewTextHandler(io.Discard, nil)))
	api := state.NewTable()
	state.SetGlobal("api", api)
	service.ExportToLua(state, api)
	if err := state.DoString(`
api.input.OnKeyPressed(api.input.Key.Grave, function(key) receivedKey = key end)
`); err != nil {
		t.Fatal(err)
	}
	service.dispatchKeyPressed(96)
	if got := state.GetGlobal("receivedKey"); got != lua.LNumber(96) {
		t.Fatalf("received key = %v, want 96", got)
	}
}
