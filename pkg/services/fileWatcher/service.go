package fileWatcher

import (
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gravestench/runtime"
	"github.com/rs/zerolog"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

type Service struct {
	logger         *zerolog.Logger
	cfg            configFile.Dependency
	watcher        *fsnotify.Watcher
	activeWatchers map[string]FileHandlerFunc
}

func (s *Service) Init(rt runtime.Runtime) {
	s.activeWatchers = make(map[string]func(string) error)

	if err := s.initWatcher(); err != nil {
		s.logger.Fatal().Msgf("initializing file watcher: %v", err)
	}

	for _, service := range rt.Services() {
		// try to bind existing services
		go s.setupServiceToWatchFiles(service)
		// there is a runtime event handler that does this in runtime_events.go
	}

}

func (s *Service) Name() string {
	return "File Watcher"
}

// the following methods are boilerplate, but they are used
// by the runtime to enforce a standard logging format.

func (s *Service) BindLogger(logger *zerolog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *zerolog.Logger {
	return s.logger
}

func (s *Service) setupServiceToWatchFiles(service runtime.Service) {
	candidate, ok := service.(NeedsFileWatcher)
	if !ok {
		return
	}

	dependent, ok := service.(runtime.HasDependencies)
	if ok {
		// we wan tto wait for the other service to resolve its dependencies,
		// like in the case that it depends on the config file service.
		for !dependent.DependenciesResolved() {
			time.Sleep(time.Millisecond * 10)
		}
	}

	if dependent != nil {
		s.logger.Debug().Msgf("dependent service %q ready to set up file watchers", service.Name())
	}

	s.logger.Info().Msgf("setting up service %q to watch for file changes", service.Name())

	for path, handler := range candidate.FileHandlers() {
		s.AddWatcher(path, handler)
	}
}
