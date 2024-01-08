package common

import (
	"log/slog"
)

type Service struct {
	log *slog.Logger
}

func (s *Service) SetLogger(l *slog.Logger) {
	s.log = l
}

func (s *Service) Logger() *slog.Logger {
	return s.log
}
