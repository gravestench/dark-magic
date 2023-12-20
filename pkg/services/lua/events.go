package lua

import (
	"time"

	"github.com/gravestench/servicemesh"
)

func (s *Service) OnServiceAdded(service servicemesh.Service) {
	s.tryToExportToLuaEnvironment(service)
}

func (s *Service) tryToExportToLuaEnvironment(service servicemesh.Service) {
	s.mux.Lock()
	defer s.mux.Unlock()

	for s.state == nil {
		time.Sleep(time.Millisecond * 10)
	}

	luaUser, ok := service.(UsesLuaEnvironment)
	if !ok {
		return
	}

	for !service.Ready() {
		s.logger.Warn("waiting for service to become ready", "target", service.Name())
		time.Sleep(time.Second)
	}

	if candidate, ok := service.(servicemesh.HasDependencies); ok {
		for !candidate.DependenciesResolved() {
			s.logger.Warn("waiting for service to resolve dependencies", "target", service.Name())
			time.Sleep(time.Second)
		}
	}

	if _, exists := s.boundServices[service.Name()]; exists {
		return
	}

	luaUser.ExportToLua(s.state)
	s.logger.Info("exporting to lua", "exported", service.Name())

	if s.boundServices == nil {
		s.boundServices = make(map[string]any)
	}

	s.boundServices[service.Name()] = service
}
