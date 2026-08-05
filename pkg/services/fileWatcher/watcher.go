package fileWatcher

import (
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

// InitWatcher initializes a watcher for file changes, such as changes to
// configuration files in the game directory.
func (s *Service) initWatcher() error {
	// debounceDuration specifies the threshold for when to ignore duplicate
	// events triggered for the same file.
	const debounceDuration = 10 * time.Millisecond

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.WithStack(err)
	}

	s.watcher = w

	go func() {
		lastEvent := make(map[string]time.Time)
		for {
			select {
			case event, ok := <-s.watcher.Events:
				if !ok {
					return
				}
				// ignore duplicate events (a few editors trigger more than one file
				// event on save, e.g. Sublime).
				now := time.Now()
				if previous := lastEvent[event.Name]; !previous.IsZero() && now.Sub(previous) < debounceDuration {
					continue
				}
				lastEvent[event.Name] = now

				s.logger.Debug("file watcher event", "event", event)

				s.handleEvent(event)
			case err, ok := <-s.watcher.Errors:
				if !ok {
					return
				}
				s.logger.Warn("error from file watcher", "error", err)
			}
		}
	}()
	return nil
}

func (s *Service) handleEvent(event fsnotify.Event) {
	if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
		return
	}
	s.mux.RLock()
	callback, ok := s.activeWatchers[event.Name]
	s.mux.RUnlock()
	if !ok {
		s.logger.Warn("unable to locate registered watcher", "path", event.Name)
		return
	}
	if err := callback(event.Name); err != nil {
		s.logger.Warn("file watcher callback failed", "path", event.Name, "error", err)
	}
	if event.Op&fsnotify.Rename != 0 {
		if err := s.watcher.Add(event.Name); err != nil {
			s.logger.Debug("waiting for renamed watch target to reappear", "path", event.Name, "error", err)
		}
	}
}

// AddWatcher watches the given file for changes and invokes f with the file
// path when a change is detected.
func (s *Service) AddWatcher(path string, f func(path string) error) {
	s.logger.Debug("adding watcher", "path", path)

	s.mux.Lock()
	if s.activeWatchers == nil {
		s.activeWatchers = make(map[string]FileHandlerFunc)
	}
	s.activeWatchers[path] = f
	s.mux.Unlock()
	if err := s.watcher.Add(path); err != nil {
		s.mux.Lock()
		delete(s.activeWatchers, path)
		s.mux.Unlock()
		s.logger.Warn("unable to add watcher", "path", path, "error", err)
		return
	}
}

// WatchAndLoad watches the given file for changes and invokes f with the file
// path when a change is detected. The given file is loaded once using f when
// calling WatchAndLoad.
func (s *Service) WatchAndLoad(path string, f func(path string) error) {
	// Add watcher for the given file.
	s.AddWatcher(path, f)

	// Load file.
	if err := f(path); err != nil {
		s.logger.Warn("unable to process", "path", path, "error", err)
		return
	}
}

// CloseWatcher closes the watcher for file changes.
func (s *Service) CloseWatcher() {
	s.closeOnce.Do(func() {
		s.logger.Debug("closing watcher")
		if s.watcher != nil {
			if err := s.watcher.Close(); err != nil {
				s.logger.Warn("unable to close file watcher", "error", err)
			}
		}
	})
}
