package content

import (
	"errors"
	"io/fs"
	"testing"
)

type missingMPQFS struct{ wrapped bool }

func (m missingMPQFS) Open(string) (fs.File, error) {
	if m.wrapped {
		return nil, errors.New("getting file stream: file not found")
	}
	return nil, errors.New("file not found")
}

func TestNormalizedMPQTranslatesMissingFile(t *testing.T) {
	source := normalizedFS{FS: missingMPQFS{}, backslash: true}
	_, err := source.Open("data/global/example.dc6")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected fs.ErrNotExist, got %v", err)
	}
}

func TestNormalizedMPQTranslatesWrappedMissingFile(t *testing.T) {
	source := normalizedFS{FS: missingMPQFS{wrapped: true}, backslash: true}
	_, err := source.Open("data/global/example.dc6")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected fs.ErrNotExist, got %v", err)
	}
}
