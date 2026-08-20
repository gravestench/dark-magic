package common

import (
	"log/slog"
)

// Service owns dependencies shared by small raylib platform adapters. Keeping the logger here avoids giving those
// adapters access to renderer internals merely to emit diagnostics.
type Service struct {
	log *slog.Logger
}

// SetLogger installs the process logger used by common raylib adapters. A nil logger is retained intentionally so each
// adapter can decide whether silence or slog.Default is appropriate for its contract.
func (s *Service) SetLogger(l *slog.Logger) {
	s.log = l
}

// Logger returns the configured logger without manufacturing a fallback, preserving the caller's explicit nil choice.
func (s *Service) Logger() *slog.Logger {
	return s.log
}
