package lua

import (
	"fmt"
)

func (s *Service) runScript(script string) error {
	if err := s.state.DoFile(script); err != nil {
		return fmt.Errorf("executing init script %q: %v", script, err)
	}

	return nil
}
