package cacheManager

import (
	"github.com/gravestench/servicemesh"
)

func (s *Service) OnServiceAdded(service servicemesh.Service) {
	s.tryToFlushCacheForService(service)
}
