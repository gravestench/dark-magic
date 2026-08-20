// Package filewatch provides deterministic development-time directory polling.
package filewatch

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type fingerprint struct {
	size     int64
	modified time.Time
}

// Handler receives one changed path at a time in deterministic lexical order.
type Handler func(context.Context, string) error

// Watcher polls one host directory for development reload. It observes files
// only; the handler decides how virtual content generations change.
type Watcher struct {
	root     string
	interval time.Duration
	handler  Handler

	mu     sync.Mutex
	known  map[string]fingerprint
	cancel context.CancelFunc
	done   chan struct{}
	err    error
}

// New creates a stopped watcher; no goroutine exists until Start.
func New(root string, interval time.Duration, handler Handler) *Watcher {
	if interval <= 0 {
		interval = 250 * time.Millisecond
	}

	return &Watcher{root: root, interval: interval, handler: handler}
}

// Start captures the baseline before polling so existing files are not reported
// as edits. Repeated starts are idempotent.
func (w *Watcher) Start(ctx context.Context) error {
	if w.root == "" || w.handler == nil {
		return errors.New("filewatch: root and handler are required")
	}

	w.mu.Lock()
	if w.cancel != nil {
		w.mu.Unlock()
		return nil
	}

	known, err := w.snapshot()
	if err != nil {
		w.mu.Unlock()
		return err
	}

	runContext, cancel := context.WithCancel(ctx)
	w.known, w.cancel, w.done, w.err = known, cancel, make(chan struct{}), nil
	done := w.done
	w.mu.Unlock()

	go w.poll(runContext, done)

	return nil
}

// poll owns the ticker for one Start call and records only the latest scan result for Stop.
// A later successful scan intentionally clears an earlier transient error.
func (w *Watcher) poll(ctx context.Context, done chan struct{}) {
	defer close(done)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.recordScanError(w.Scan(ctx))
		}
	}
}

// recordScanError serializes the terminal status observed by Stop without holding the lock during handlers.
func (w *Watcher) recordScanError(err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.err = err
}

// Stop cancels polling, waits within ctx, and returns the last scan error.
func (w *Watcher) Stop(ctx context.Context) error {
	w.mu.Lock()
	cancel, done := w.cancel, w.done
	w.cancel = nil
	w.done = nil
	w.mu.Unlock()

	if cancel == nil {
		return nil
	}

	cancel()

	select {
	case <-done:
	case <-ctx.Done():
		return fmt.Errorf("filewatch: stop: %w", ctx.Err())
	}

	w.mu.Lock()
	err := w.err
	w.mu.Unlock()

	return err
}

// Scan performs one deterministic comparison and invokes the handler in
// lexical path order (filepath.WalkDir order).
func (w *Watcher) Scan(ctx context.Context) error {
	current, err := w.snapshot()
	if err != nil {
		return err
	}

	previous := w.replaceKnown(current)
	changed := changedPaths(previous, current)

	return w.notifyChanges(ctx, changed)
}

// replaceKnown commits the observed filesystem generation before calling user code.
// Consequently a handler failure is reported but not retried as a synthetic file change on the next scan.
func (w *Watcher) replaceKnown(current map[string]fingerprint) map[string]fingerprint {
	w.mu.Lock()
	defer w.mu.Unlock()

	previous := w.known
	w.known = current

	return previous
}

// changedPaths reports creations, content-metadata changes, and removals in deterministic lexical order.
func changedPaths(previous, current map[string]fingerprint) []string {
	var changed []string

	for name, value := range current {
		if old, exists := previous[name]; !exists || old != value {
			changed = append(changed, name)
		}
	}

	for name := range previous {
		if _, exists := current[name]; !exists {
			changed = append(changed, name)
		}
	}

	sort.Strings(changed)

	return changed
}

// notifyChanges attempts every changed path and joins failures so one bad file cannot hide later edits.
func (w *Watcher) notifyChanges(ctx context.Context, changed []string) error {
	var errs []error

	for _, name := range changed {
		if err := w.handler(ctx, filepath.ToSlash(name)); err != nil {
			errs = append(errs, fmt.Errorf("filewatch: handle %q: %w", name, err))
		}
	}

	return errors.Join(errs...)
}

// snapshot captures only size and modification time because polling must remain cheap for development trees.
func (w *Watcher) snapshot() (map[string]fingerprint, error) {
	result := make(map[string]fingerprint)

	err := filepath.WalkDir(w.root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		relative, err := filepath.Rel(w.root, name)
		if err != nil {
			return err
		}

		result[relative] = fingerprint{size: info.Size(), modified: info.ModTime()}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("filewatch: scan %q: %w", w.root, err)
	}

	return result, nil
}
