package fileWatcher

import (
	"github.com/gravestench/servicemesh"
)

var _ servicemesh.EventHandlerServiceAdded = &Service{}

func (s *Service) OnServiceAdded(service servicemesh.Service) {
	go s.setupServiceToWatchFiles(service)
}
