package fileWatcher

import (
	"github.com/gravestench/runtime"
)

var _ runtime.EventHandlerServiceAdded = &Service{}

func (s *Service) OnServiceAdded(args ...interface{}) {
	if len(args) < 1 {
		return
	}

	service, ok := args[0].(runtime.Service)
	if !ok {
		return
	}

	go s.setupServiceToWatchFiles(service)
}
