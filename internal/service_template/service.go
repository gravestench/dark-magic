package service_template

import (
	"github.com/gravestench/runtime"
	"github.com/rs/zerolog"
)

type Service struct {
	logger *zerolog.Logger

	// these should be the exported integration
	// interface exported by other services
	bar any
	baz any
}

func (s *Service) Init(rt runtime.Runtime) {
	// This init method will be invoked by the runtime
	// as soon as the dependency resolution has finished.
	// If the service does not implement runtime.HasDependencies,
	// then this method is invoked immediately.
}

func (s *Service) Name() string {
	return "Template"
}

// the following methods are boilerplate, but they are used
// by the runtime to enforce a standard logging format.

func (s *Service) BindLogger(logger *zerolog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *zerolog.Logger {
	return s.logger
}

func (s *Service) Foo() {
	// do stuff here
}
