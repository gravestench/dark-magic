package modruntime

import (
	"errors"
	"fmt"
	"sync"
)

// ReleaseFunc releases one script-owned resource.
type ReleaseFunc func() error

// Scope owns all resources acquired by one script component or invocation.
//
// A scope is the bridge between disposable Lua code and longer-lived native
// capabilities: callbacks, nodes, handles, and subscriptions register releases
// here so reload or failure cannot strand them in their owning subsystem.
type Scope struct {
	mu       sync.Mutex
	closed   bool
	releases []ReleaseFunc
}

// Add transfers ownership of release to the scope.
func (s *Scope) Add(release ReleaseFunc) error {
	if release == nil {
		return errors.New("modruntime: nil resource release")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return errors.New("modruntime: resource scope is closed")
	}

	s.releases = append(s.releases, release)

	return nil
}

// Close releases resources in reverse acquisition order and joins failures.
func (s *Scope) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}

	s.closed = true
	releases := s.releases
	s.releases = nil
	s.mu.Unlock()

	var errs []error

	for i := len(releases) - 1; i >= 0; i-- {
		if err := releases[i](); err != nil {
			errs = append(errs, fmt.Errorf("modruntime: release resource %d: %w", i, err))
		}
	}

	return errors.Join(errs...)
}
