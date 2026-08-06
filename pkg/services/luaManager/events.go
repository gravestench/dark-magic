package luaManager

import "github.com/gravestench/servicemesh"

func (s *Service) OnServiceInitialized(service servicemesh.Service) {
	s.exportToLuaEnvironment(service)
}
