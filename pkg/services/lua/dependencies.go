package lua

import (
	"github.com/gravestench/runtime"
	"github.com/gravestench/runtime/examples/services/config_file"
)

func (s *Service) DependenciesResolved() bool {
	if s.cfg == nil {
		return false
	}

	if cfg, _ := s.Config(); cfg == nil {
		return false // make sure to wait for the config file to get init'd
	}

	return true
}

func (s *Service) ResolveDependencies(rt runtime.R) {
	for _, service := range rt.Services() {
		if candidate, ok := service.(config_file.Manager); ok {
			s.cfg = candidate
		}
	}
}
