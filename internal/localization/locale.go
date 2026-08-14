// Package localization loads layered Diablo string tables.
package localization

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"sync"

	tbl "github.com/gravestench/tbl_text"
)

// Locale lazily composes base, expansion, then patch tables for one language.
type Locale struct {
	source   fs.FS
	language string
	mu       sync.Mutex
	once     sync.Once
	strings  map[string]string
	err      error
}

// New constructs a locale; language is typically "English".
func New(source fs.FS, language string) *Locale {
	return &Locale{source: source, language: language}
}

// Text resolves key, returning the key alongside an error when absent.
func (l *Locale) Text(key string) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.once.Do(l.load)
	if l.err != nil {
		return key, l.err
	}
	value, exists := l.strings[key]
	if !exists {
		return key, fmt.Errorf("localization: key %q is not present", key)
	}
	return value, nil
}

// Invalidate discards the composed locale after VFS package layers change.
func (l *Locale) Invalidate() {
	l.mu.Lock()
	l.once = sync.Once{}
	l.strings = nil
	l.err = nil
	l.mu.Unlock()
}

// GetSupportedLanguages satisfies compatibility UI localization seams.
func (l *Locale) GetSupportedLanguages() []string { return []string{l.language} }

func (l *Locale) load() {
	l.strings = make(map[string]string)
	shimPath := fmt.Sprintf("locales/%s.json", l.language)
	if data, err := fs.ReadFile(l.source, shimPath); err == nil {
		if err := json.Unmarshal(data, &l.strings); err != nil {
			l.err = fmt.Errorf("localization: decode %q: %w", shimPath, err)
			return
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		l.err = fmt.Errorf("localization: read %q: %w", shimPath, err)
		return
	}
	paths := []string{
		fmt.Sprintf("data/local/lng/%s/string.tbl", l.language),
		fmt.Sprintf("data/local/lng/%s/expansionstring.tbl", l.language),
		fmt.Sprintf("data/local/lng/%s/patchstring.tbl", l.language),
	}
	loaded := len(l.strings) > 0
	for _, path := range paths {
		file, err := l.source.Open(path)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			l.err = fmt.Errorf("localization: read %q: %w", path, err)
			return
		}
		var table tbl.TextTable
		if reader, ok := file.(io.ReaderAt); ok {
			info, statErr := file.Stat()
			if statErr == nil {
				table, err = tbl.UnmarshalReaderAt(reader, info.Size())
			} else {
				err = statErr
			}
		} else {
			data, readErr := io.ReadAll(file)
			if readErr != nil {
				err = readErr
			} else {
				table, err = tbl.Unmarshal(data)
			}
		}
		closeErr := file.Close()
		if err != nil {
			l.err = fmt.Errorf("localization: decode %q: %w", path, err)
			return
		}
		if closeErr != nil {
			l.err = fmt.Errorf("localization: close %q: %w", path, closeErr)
			return
		}
		loaded = true
		for key, value := range table {
			l.strings[key] = value
		}
	}
	if !loaded {
		l.err = fmt.Errorf("localization: no %s string tables found", l.language)
	}
}
