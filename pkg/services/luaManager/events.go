package luaManager

import (
	"time"

	"github.com/gravestench/servicemesh"
)

func (s *Service) OnServiceAdded(service servicemesh.Service) {
	for s.state == nil {
		time.Sleep(time.Millisecond * 10)
	}

	go s.exportToLuaEnvironment(service)
}
