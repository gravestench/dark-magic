package lua

import (
	"github.com/gravestench/servicemesh"
)

func (s *Service) tryToExportToLuaEnvironment(args ...any) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if len(args) < 1 {
		return
	}

	service, ok := args[0].(servicemesh.Service)
	if !ok {
		return
	}

	luaUser, ok := args[0].(UsesLuaEnvironment)
	if !ok {
		return
	}

	if candidate, ok := service.(servicemesh.HasDependencies); ok {
		if !candidate.DependenciesResolved() {
			return
		}
	}

	if _, exists := s.boundServices[service.Name()]; exists {
		return
	}

	go luaUser.ExportToLua(s.state)
	s.boundServices[service.Name()] = service
	s.logger.Info("successfully exported to lua", "exported", service.Name())
}
