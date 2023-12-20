package configFile

import (
	"log/slog"
	"sync"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/fileWatcher"
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
	fileWatcher                fileWatcher.Dependency
}

// BindLogger satisfies the runtime.HasLogger interface
func (s *Service) SetLogger(l *slog.Logger) {
	s.log = l
}

// Logger satisfies the runtime.HasLogger interface
func (s *Service) Logger() *slog.Logger {
	return s.log
}

// Name satisfies the runtime.IsRuntimeService interface
func (s *Service) Name() string {
	return "Config File Manager"
}

func (s *Service) Ready() bool {
	if s.fileWatcher == nil {
		return false
	}

	return true
}

// Init satisfies the runtime.IsRuntimeService interface
func (s *Service) Init(mesh servicemesh.Mesh) {
	s.mesh = mesh
	s.configs = make(map[string]*Config)
	s.servicesWithDefaultConfigs = make(map[string]HasDefaultConfig)

	for _, candidate := range mesh.Services() {
		err := s.initConfigForServiceCandidate(candidate)
		if err != nil {
			s.log.Error("applying default config", "for", candidate.Name(), "error", err)
		}
	}
}

func (s *Service) OnServiceAdded(service servicemesh.Service) {
	err := s.initConfigForServiceCandidate(service)
	if err != nil {
		s.log.Error("applying default config", "for", service.Name(), "error", err)
	}
}
