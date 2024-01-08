package webRouter

import (
	"github.com/gravestench/servicemesh"
)

var _ servicemesh.EventHandlerServiceAdded = &Service{}

func (s *Service) OnServiceAdded(service servicemesh.Service) {
	s.initRoutesForService(service)
}
