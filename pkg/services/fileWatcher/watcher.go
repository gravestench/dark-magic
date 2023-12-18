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
		prevPath := ""
		prevTime := time.Now()
		for {
			select {
			case event, ok := <-s.watcher.Events:
				if !ok {
					return
				}
				// ignore duplicate events (a few editors trigger more than one file
				// event on save, e.g. Sublime).
				now := time.Now()
				if len(prevPath) != 0 && prevPath == event.Name {
					if now.Sub(prevTime) < debounceDuration {
						continue // skip duplicate event.
					}
				}
				prevPath = event.Name
				prevTime = now

				s.logger.Debug("file watcher event", "event", event)

				f, ok := s.activeWatchers[event.Name]
				if !ok {
					s.logger.Warn("unable to locate registered watcher for path %q", event.Name)
					continue
				}
				go func() {
					if err := f(event.Name); err != nil {
						s.logger.Warn("%+v", err)
					}
				}()
			case err, ok := <-s.watcher.Errors:
				if !ok {
					return
				}
				s.logger.Warn("error from file watcher; %+v", err)
			}
		}
	}()
	return nil
}

// AddWatcher watches the given file for changes and invokes f with the file
// path when a change is detected.
func (s *Service) AddWatcher(path string, f func(path string) error) {
	s.logger.Debug("adding watcher for %q", path)

	if err := s.watcher.Add(path); err != nil {
		s.logger.Warn("unable to add watcher", "path", path, "error", err)
		return
	}

	s.activeWatchers[path] = f
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
	s.logger.Debug("closing watcher")

	if err := s.watcher.Close(); err != nil {
		s.logger.Warn("unable to close file watcher; %+v", err)
	}
}
