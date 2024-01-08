package luaManager

import (
	lua "github.com/yuin/gopher-lua"
)

type task struct {
	fn   func(state *lua.LState) error
	done chan error
}

func (s *Service) luaStateAccessSingletonWorker() {
	defer s.wg.Done()
	for {
		for t := range s.tasks {
			t.done <- t.fn(s.state)
		}
	}
}
