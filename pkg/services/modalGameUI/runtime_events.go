package modalGameUI

import (
	"github.com/gravestench/runtime"
)

func (s *Service) OnServiceAdded(args ...any) {
	if len(args) < 1 {
		return
	}

	service, ok := args[0].(runtime.Service)
	if !ok {
		return
	}

	s.attemptBindService(service)
}
