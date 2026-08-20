// Package localization loads layered Diablo string tables.
package localization

import (
	"fmt"
	"io/fs"
	"sync"
)

// Locale owns the lazy, invalidatable localization cache for one language.
// The mutex keeps sync.Once replacement and cache access in one synchronization domain.
type Locale struct {
	source   fs.FS
	language string
	mu       sync.Mutex
	once     sync.Once
	strings  map[string]string
	sources  map[string]string
	err      error
}

// New constructs a lazy locale without reading its source, so loading failures surface on the first lookup.
// Diablo installations normally use "English", but alternate language directory names remain supported.
func New(source fs.FS, language string) *Locale {
	return &Locale{source: source, language: language}
}

// Text resolves key while preserving it as the fallback value on lookup or loading errors.
// Callers that do not need source attribution can therefore retain their existing display fallback.
func (l *Locale) Text(key string) (string, error) {
	value, _, err := l.Resolve(key)

	return value, err
}

// Resolve returns localized text and the path of the highest-precedence layer that defined it.
// Lookup shares a lock with invalidation so a reset cannot race an in-progress sync.Once load.
func (l *Locale) Resolve(key string) (string, string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.once.Do(l.load)

	if l.err != nil {
		return key, "", l.err
	}

	value, exists := l.strings[key]
	if !exists {
		return key, "", fmt.Errorf("localization: key %q is not present", key)
	}

	return value, l.sources[key], nil
}

// Invalidate discards the composed locale so the next lookup observes changed VFS package layers.
// Resetting sync.Once under the lookup mutex preserves its rule against concurrent use and replacement.
func (l *Locale) Invalidate() {
	l.mu.Lock()

	l.once = sync.Once{}
	l.strings = nil
	l.sources = nil
	l.err = nil

	l.mu.Unlock()
}

// GetSupportedLanguages exposes the configured language through compatibility UI localization seams.
// A new slice prevents callers from mutating the locale's stored language state.
func (l *Locale) GetSupportedLanguages() []string { return []string{l.language} }

// load composes the locale exactly once per cache generation while Resolve holds the cache mutex.
// Publishing maps and their error together prevents readers from observing a partially composed result.
func (l *Locale) load() {
	composer := newLocaleComposer(l.source, l.language)
	l.strings, l.sources, l.err = composer.compose()
}
