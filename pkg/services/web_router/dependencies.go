package web_router

import (
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/config_file"
)

func (s *Service) DependenciesResolved() bool {
	if s.cfgManager == nil {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(rt runtime.R) {
	for _, other := range rt.Services() {
		if cfg, ok := other.(config_file.Manager); ok {
			s.cfgManager = cfg
		}
	}
}
