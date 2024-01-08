package service_template

import (
	"log/slog"

	"github.com/gravestench/servicemesh"
)

type Service struct {
	logger *slog.Logger

	// these should be the exported integration
	// interface exported by other services
	bar any
	baz any
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	// This init method will be invoked by the servicemesh
	// as soon as the dependency resolution has finished.
	// If the service does not implement servicemesh.HasDependencies,
	// then this method is invoked immediately.
}

func (s *Service) Name() string {
	return "Template"
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

func (s *Service) Foo() {
	// do stuff here
}
