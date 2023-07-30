package lua

import (
	lua "github.com/yuin/gopher-lua"
)

func (s *Service) bindLoggerToLuaEnvironment() {
	fnPrint := s.state.NewFunction(func(L *lua.LState) int {
		numArgs := L.GetTop()

		args := L.CheckAny(numArgs)

		s.logger.Info().Msgf(args.String())

		return 0
	})

	s.state.SetGlobal("print", fnPrint)
}

func (s *Service) LuaState() *lua.LState {
	return s.state
}
