package service_template

import (
	"context"
	"log/slog"
)

// Service is a deliberately small example of an explicitly constructed engine
// component. Required dependencies belong in New; optional dependencies can be
// expressed with functional options or narrow setter methods before Start.
type Service struct {
	logger *slog.Logger
	foo    FooDependency
	run    bool
}

// FooDependency is the narrow contract this component needs from a peer.
type FooDependency interface {
	Foo()
}

func New(logger *slog.Logger, foo FooDependency) *Service {
	return &Service{logger: logger, foo: foo}
}

func (s *Service) Start(context.Context) error {
	s.run = true
	return nil
}

func (s *Service) Stop(context.Context) error {
	s.run = false
	return nil
}

func (s *Service) DoWork() {
	if s.run && s.foo != nil {
		s.foo.Foo()
	}
}
