package luaManager

import "github.com/gravestench/servicemesh"

func (s *Service) OnServiceAdded(service servicemesh.Service) {
	go s.exportToLuaEnvironment(service)
}
