package tweens

import (
	"time"

	lua "github.com/yuin/gopher-lua"

	"github.com/gravestench/dark-magic/pkg/easing"
	luaService "github.com/gravestench/dark-magic/pkg/services/luaManager"
)

const LuaApiKey = "tweens"

var _ luaService.LuaPlugin = &Service{}

// these methods are automatically invoked
// by the lua service to export stuff into the
// lua environment for use in scripts.

func (s *Service) ExportToLua(state *lua.LState, rootTable *lua.LTable) {
	table := state.NewTable()

	fnMap := map[string]lua.LGFunction{
		"New": s.luaNewTween,
		"Add": s.luaAddTween,
	}

	for key, fn := range fnMap {
		state.SetField(table, key, state.NewFunction(fn))
	}

	state.SetField(rootTable, LuaApiKey, table)
}

func (s *Service) UnexportFromLua(state *lua.LState, rootTable *lua.LTable) {
	state.SetField(rootTable, LuaApiKey, lua.LNil)
}

func (s *Service) luaNewTween(L *lua.LState) int {
	if L.GetTop() != 0 {
		return 0
	}

	tween := s.New()
	tweenTable := s.LuaMakeTween(L, tween)

	L.Push(tweenTable)

	return 1
}

func (s *Service) luaAddTween(L *lua.LState) int {
	if L.GetTop() != 1 {
		return 0
	}

	table := L.CheckTable(1)
	field := L.GetField(table, "_userdata")

	if userData, ok := field.(*lua.LUserData); ok {
		if tween, isTween := userData.Value.(*Tween); isTween {
			s.Add(tween)
		}
	}

	return 0
}

func (s *Service) LuaMakeTween(L *lua.LState, t *Tween) *lua.LTable {
	table := L.NewTable()

	ud := L.NewUserData()
	ud.Value = t

	// this is how we can create and add new tweens inside of lua code
	L.SetField(table, "_userdata", ud)

	for key, fn := range map[string]lua.LGFunction{
		"Start":      s.luaTweenClosureStart(t),
		"Stop":       s.luaTweenClosureStop(t),
		"Play":       s.luaTweenClosurePlay(t),
		"Pause":      s.luaTweenClosurePause(t),
		"Progress":   s.luaTweenClosureProgress(t),
		"Time":       s.luaTweenClosureTime(t),
		"Ease":       s.luaTweenClosureEase(t),
		"OnStart":    s.luaTweenClosureOnStart(t),
		"OnComplete": s.luaTweenClosureOnComplete(t),
		"OnRepeat":   s.luaTweenClosureOnRepeat(t),
		"OnUpdate":   s.luaTweenClosureOnUpdate(t),
		"Delay":      s.luaTweenClosureDelay(t),
		"Repeat":     s.luaTweenClosureRepeat(t),
	} {
		L.SetField(table, key, L.NewFunction(fn))
	}

	return table
}

func (s *Service) luaTweenClosureStart(t *Tween) lua.LGFunction {
	return func(L *lua.LState) int {
		numArgs := L.GetTop()
		if numArgs > 0 {
			return 0
		}

		t.Start()
		return 0
	}
}

func (s *Service) luaTweenClosureStop(t *Tween) lua.LGFunction {
	return func(L *lua.LState) int {
		numArgs := L.GetTop()
		if numArgs > 0 {
			return 0
		}

		t.Stop()
		return 0
	}
}

func (s *Service) luaTweenClosurePlay(t *Tween) lua.LGFunction {
	return func(L *lua.LState) int {
		numArgs := L.GetTop()
		if numArgs > 0 {
			return 0
		}

		t.Play()
		return 0
	}
}

func (s *Service) luaTweenClosurePause(t *Tween) lua.LGFunction {
	return func(L *lua.LState) int {
		numArgs := L.GetTop()
		if numArgs > 0 {
			return 0
		}

		t.Pause()
		return 0
	}
}

func (s *Service) luaTweenClosureProgress(t *Tween) lua.LGFunction {
	return func(L *lua.LState) int {
		numArgs := L.GetTop()
		if numArgs > 0 {
			return 0
		}

		progress := t.Progress()
		L.Push(lua.LNumber(progress))

		return 1
	}
}

func (s *Service) luaTweenClosureTime(t *Tween) lua.LGFunction {
	return func(L *lua.LState) int {
		numArgs := L.GetTop()
		if numArgs < 1 {
			L.Push(lua.LNumber(t.duration))
			return 1
		}

		tweenTime := L.CheckNumber(1)
		t.Time(time.Duration(tweenTime))

		return 0
	}
}

func (s *Service) luaTweenClosureEase(t *Tween) lua.LGFunction {
	return func(L *lua.LState) int {
		numArgs := L.GetTop()
		if numArgs < 1 {
			_, name := getEaseFn(t.easeName, nil)
			L.Push(lua.LString(name))
			return 1
		}

		var args []any
		for i := 1; i < numArgs; i++ {
			args = append(args, L.CheckAny(i))
		}

		ease, ok := args[0].(string)
		if !ok {
			return 0
		}

		if _, found := easing.EaseMap[ease]; found {
			t.Ease(args...)
		}

		return 0
	}
}

func (s *Service) luaTweenClosureOnStart(t *Tween) lua.LGFunction {
	return func(L *lua.LState) int {
		if L.GetTop() < 1 {
			return 0
		}
		fn := L.CheckFunction(1)
		t.OnStart(func() { callLua(L, fn) })
		return 0
	}
}

func (s *Service) luaTweenClosureOnComplete(t *Tween) lua.LGFunction {
	return func(L *lua.LState) int {
		if L.GetTop() < 1 {
			return 0
		}
		fn := L.CheckFunction(1)
		t.OnComplete(func() { callLua(L, fn) })
		return 0
	}
}

func (s *Service) luaTweenClosureOnRepeat(t *Tween) lua.LGFunction {
	return func(L *lua.LState) int {
		if L.GetTop() < 1 {
			return 0
		}
		fn := L.CheckFunction(1)
		t.OnRepeat(func() { callLua(L, fn) })
		return 0
	}
}

func (s *Service) luaTweenClosureOnUpdate(t *Tween) lua.LGFunction {
	return func(L *lua.LState) int {
		numArgs := L.GetTop()
		if numArgs < 1 {
			return 0
		}

		// Get the Lua function from the argument
		fn := L.CheckFunction(1)

		// Create a Go wrapper function
		t.onUpdate = func(input float64) {
			callLua(L, fn, lua.LNumber(input))
		}

		return 0
	}
}

func (s *Service) luaTweenClosureDelay(t *Tween) lua.LGFunction {
	return func(L *lua.LState) int {
		numArgs := L.GetTop()

		if numArgs < 1 {
			L.Push(lua.LNumber(t.delay))
			return 1
		}

		delay := L.CheckNumber(1)
		t.Delay(time.Duration(delay))

		return 0
	}
}

func (s *Service) luaTweenClosureRepeat(t *Tween) lua.LGFunction {
	return func(L *lua.LState) int {
		numArgs := L.GetTop()
		if numArgs < 1 {
			L.Push(lua.LNumber(t.repeatCount))
			return 1
		}

		repeat := L.CheckNumber(1)
		t.Repeat(int(repeat))

		return 0
	}
}

func callLua(state *lua.LState, fn *lua.LFunction, args ...lua.LValue) {
	_ = state.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}, args...)
}
