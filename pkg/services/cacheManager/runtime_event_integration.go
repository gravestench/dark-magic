package cacheManager

import (
	"github.com/gravestench/runtime"
)

func (s *Service) OnServiceAdded(args ...any) {
	if len(args) < 1 {
		return
	}

	if candidate, ok := args[0].(runtime.Service); ok {
		s.tryToFlushCacheForService(candidate)
	}
}
