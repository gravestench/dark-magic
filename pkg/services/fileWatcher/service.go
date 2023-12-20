package fileWatcher

import (
	"log/slog"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gravestench/servicemesh"
)

type Service struct {
	logger         *slog.Logger
	watcher        *fsnotify.Watcher
	activeWatchers map[string]FileHandlerFunc
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.activeWatchers = make(map[string]func(string) error)

	if err := s.initWatcher(); err != nil {
		s.logger.Error("initializing file watcher", "error", err)
		panic(err)
	}

	for _, service := range mesh.Services() {
		// try to bind existing services
		go s.setupServiceToWatchFiles(service)
		// there is a servicemesh event handler that does this in servicemesh_events.go
	}
}

func (s *Service) Name() string {
	return "File Watcher"
}

func (s *Service) Ready() bool {
	return true
}

// the following methods are boilerplate, but they are used
// by the servicemesh to enforce a standard logging format.

func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}

func (s *Service) setupServiceToWatchFiles(service servicemesh.Service) {
	candidate, ok := service.(NeedsFileWatcher)
	if !ok {
		return
	}

	for !service.Ready() {
		time.Sleep(time.Millisecond * 10)
	}

	dependent, ok := service.(servicemesh.HasDependencies)
	if ok {
		// we want to wait for the other service to resolve its dependencies,
		// like in the case that it depends on the config file service.
		for !dependent.DependenciesResolved() {
			time.Sleep(time.Millisecond * 10)
		}
	}

	if dependent != nil {
		s.logger.Debug("ready to set up file watchers", "dependent", service.Name())
	}

	for path, handler := range candidate.FileHandlers() {
		s.logger.Info("setting up file watchers", "for", service.Name(), "path", path)
		s.AddWatcher(path, handler)
	}
}
