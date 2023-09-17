package config_file

import (
	"sync"

	"github.com/rs/zerolog"

	"github.com/gravestench/runtime"
	"github.com/gravestench/runtime/pkg/events"
)

const (
	defaultConfigDir  = "~/.config"
	defaultConfigFile = "config.json"
)

// Service is a config file manager that marshals to and from json files.
type Service struct {
	log                        *zerolog.Logger
	mux                        sync.Mutex
	configs                    map[string]*Config
	servicesWithDefaultConfigs map[string]HasDefaultConfig
	RootDirectory              string
}

// BindLogger satisfies the runtime.HasLogger interface
func (s *Service) BindLogger(l *zerolog.Logger) {
	s.log = l
}

// Logger satisfies the runtime.HasLogger interface
func (s *Service) Logger() *zerolog.Logger {
	return s.log
}

// Name satisfies the runtime.IsRuntimeService interface
func (s *Service) Name() string {
	return "Config File Manager"
}

// Init satisfies the runtime.IsRuntimeService interface
func (s *Service) Init(rt runtime.Runtime) {
	s.configs = make(map[string]*Config)
	s.servicesWithDefaultConfigs = make(map[string]HasDefaultConfig)

	for _, candidate := range rt.Services() {
		err := s.applyDefaultConfig(candidate)
		if err != nil {
			s.log.Error().Msgf("applying default config for %q: %v", candidate.Name(), err)
		}
	}

	rt.Events().On(events.EventServiceAdded, func(args ...any) {
		if len(args) < 1 {
			return
		}

		if candidate, ok := args[0].(runtime.Service); ok {
			err := s.applyDefaultConfig(candidate)
			if err != nil {
				s.log.Error().Msgf("applying default config for %q: %v", candidate.Name(), err)
			}
		}
	})
}
