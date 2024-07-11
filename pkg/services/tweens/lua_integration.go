package tweens

import (
	"fmt"
	"strconv"
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

func (s *Service) LuaPluginPreload(state *lua.LState) {
	//TODO implement me
	panic("implement me")
}

func (s *Service) LuaPluginLoadIntoTable(state *lua.LState, rootTable *lua.LTable) {
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

func (s *Service) LuaPluginUnloadFromTable(state *lua.LState, rootTable *lua.LTable) {
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

	tbl := L.CheckTable(1)
	if ud, ok := L.GetField(tbl, "UserData").(*lua.LUserData); ok {
		if tween, ok := ud.Value.(*Tween); ok {
			s.Add(tween)
		}
	}

	return 0
}

func (s *Service) LuaMakeTween(L *lua.LState, t *Tween) *lua.LTable {
	table := L.NewTable()

	for key, fn := range map[string]lua.LGFunction{
		"Start":      s.luaTweenClosureStart(t),
		"Stop":       s.luaTweenClosureStop(t),
		"Play":       s.luaTweenClosurePlay(t),
		"Pause":      s.luaTweenClosurePause(t),
		"Progress":   s.luaTweenClosureProgress(t),
		"Update":     s.luaTweenClosureUpdate(t),
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

	L.SetField(table, "UserData", s.luaTweenClosureUserData(L, t))

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

func (s *Service) luaTweenClosureUpdate(t *Tween) lua.LGFunction {
	return func(L *lua.LState) int {
		numArgs := L.GetTop()
		if numArgs < 1 {
			return 0
		}

		delta := L.CheckNumber(1)
		t.Update(time.Duration(delta))

		return 0
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

		ease := L.CheckString(1)

		var args []lua.LValue
		for i := 2; i < numArgs; i++ {
			args = append(args, L.CheckNumber(i))
		}

		var params []float64

		for _, arg := range args {
			if n, ok := arg.(lua.LNumber); ok {
				if v, err := strconv.ParseFloat(n.String(), 32); err == nil {
					params = append(params, v)
				}
			}
		}

		if _, found := easing.EaseMap[ease]; found {
			t.Ease(ease, params)
		} else if !found {

		}

		return 0
	}
}

func (s *Service) luaTweenClosureOnStart(t *Tween) lua.LGFunction {
	return func(L *lua.LState) int {
		return 0
	}
}

func (s *Service) luaTweenClosureOnComplete(t *Tween) lua.LGFunction {
	return func(L *lua.LState) int {
		return 0
	}
}

func (s *Service) luaTweenClosureOnRepeat(t *Tween) lua.LGFunction {
	return func(L *lua.LState) int {
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
			// Push the Lua function onto the stack
			L.Push(fn)
			// Push the argument onto the stack
			L.Push(lua.LNumber(input))

			// Call the Lua function with 1 argument and 1 return value
			err := L.PCall(1, 0, nil)
			if err != nil {
				fmt.Println("Error calling Lua function:", err)
				return
			}

			// Return 0 if the return value is not a number
			return
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
		t.Delay(time.Duration(repeat))

		return 0
	}
}

func (s *Service) luaTweenClosureUserData(L *lua.LState, t *Tween) *lua.LUserData {
	obj := L.NewUserData()
	obj.Value = t

	return obj
}
