package luaManager

import (
	"fmt"

	"github.com/yuin/gopher-lua"
)

func (s *Service) runScript(script string) error {
	return s.WithState(func(state *lua.LState) error {
		if err := state.DoFile(script); err != nil {
			return fmt.Errorf("executing init script %q: %v", script, err)
		}
		return nil
	})
}
