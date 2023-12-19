package configFile

import (
	"log/slog"
	"sync"

	"github.com/gravestench/servicemesh"
)

const (
	defaultConfigDir  = "~/.config"
	defaultConfigFile = "config.json"
)

// Service is a config file manager that marshals to and from json files.
type Service struct {
	mesh                       servicemesh.Mesh
	log                        *slog.Logger
	mux                        sync.Mutex
	configs                    map[string]*Config
	servicesWithDefaultConfigs map[string]HasDefaultConfig
	RootDirectory              string
}

// SetLogger satisfies the servicemesh.HasLogger interface
func (s *Service) SetLogger(l *slog.Logger) {
	s.log = l
}

// Logger satisfies the servicemesh.HasLogger interface
func (s *Service) Logger() *slog.Logger {
	return s.log
}

// Name satisfies the servicemesh.IsRuntimeService interface
func (s *Service) Name() string {
	return "Config File Manager"
}

func (s *Service) Ready() bool {
	return true
}

// Init satisfies the servicemesh.IsRuntimeService interface
func (s *Service) Init(mesh servicemesh.Mesh) {
	s.mesh = mesh
	s.configs = make(map[string]*Config)
	s.servicesWithDefaultConfigs = make(map[string]HasDefaultConfig)

	for _, candidate := range mesh.Services() {
		err := s.initConfigForServiceCandidate(candidate)
		if err != nil {
			s.log.Error("applying default config for %q: %v", candidate.Name(), err)
		}
	}

	mesh.Events().On(servicemesh.EventServiceAdded, func(args ...any) {
		if len(args) < 1 {
			return
		}

		if candidate, ok := args[0].(servicemesh.Service); ok {
			err := s.initConfigForServiceCandidate(candidate)
			if err != nil {
				s.log.Error("applying default config for %q: %v", candidate.Name(), err)
			}
		}
	})
}
