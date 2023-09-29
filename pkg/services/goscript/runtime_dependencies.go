package goscript

import (
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

func (s *Service) DependenciesResolved() bool {
	if s.cfg == nil {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(rt runtime.R) {
	for _, service := range rt.Services() {
		if candidate, ok := service.(configFile.Manager); ok {
			s.cfg = candidate
		}
	}
}
