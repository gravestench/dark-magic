package tweens

import (
	"testing"

	"github.com/yuin/gopher-lua"
)

func TestLuaTweenCallbacksAndRepeat(t *testing.T) {
	state := lua.NewState()
	defer state.Close()
	service := &Service{}
	root := state.NewTable()
	service.ExportToLua(state, root)
	state.SetGlobal("api", root)

	script := `
local count = 0
local tween = api.tweens.New()
tween.OnStart(function() count = count + 1 end)
tween.OnUpdate(function(_) count = count + 1 end)
tween.OnComplete(function() count = count + 1 end)
tween.Repeat(2)
result = { tween = tween, count = function() return count end }
`
	if err := state.DoString(script); err != nil {
		t.Fatal(err)
	}
	result := state.GetGlobal("result")
	tweenTable := state.GetField(result, "tween").(*lua.LTable)
	tween := state.GetField(tweenTable, "_userdata").(*lua.LUserData).Value.(*Tween)
	if tween.repeatCount != 2 {
		t.Fatalf("repeat count = %d, want 2", tween.repeatCount)
	}
	tween.Update(defaultDuration)
	countFn := state.GetField(result, "count")
	if err := state.CallByParam(lua.P{Fn: countFn, NRet: 1, Protect: true}); err != nil {
		t.Fatal(err)
	}
	count := state.Get(-1)
	state.Pop(1)
	if count != lua.LNumber(3) {
		t.Fatalf("callback count = %v, want 3", count)
	}
}
