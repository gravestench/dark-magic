package lua

import (
	ee "github.com/gravestench/eventemitter"
	"github.com/gravestench/runtime"
)

func (s *Service) BindsEvents(emitter *ee.EventEmitter) {
	s.events = emitter
}

func (s *Service) tryToExportToLuaEnvironment(args ...any) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if len(args) < 1 {
		return
	}

	service, ok := args[0].(runtime.Service)
	if !ok {
		return
	}

	luaUser, ok := args[0].(UsesLuaEnvironment)
	if !ok {
		return
	}

	luaUser.ExportToLua(s.state)
	s.logger.Info().Msgf("successfully exported '%s' to lua", service.Name())
}
