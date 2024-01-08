package fileLoader

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/gravestench/mpq"
)

func NewSource(uri string) Source {
	s := Source{
		Path: uri,
	}

	return s
}

type SourceType int

const (
	SourceUnknown SourceType = iota
	SourceDirectory
	SourceArchive
	SourceRepo
)

type Source struct {
	Path string
}

func (s *Source) String() string {
	return s.Path
}

func (s *Source) Filesystem() (fs.FS, error) {
	switch s.Type() {
	case SourceDirectory:
		return s.getDirectoryFilesystem()
	case SourceArchive:
		return s.getArchiveFilesystem()
	case SourceRepo:
		return s.getRepoFilesystem()
	default:
		return nil, fmt.Errorf("getting reader for %q: bad path or incompatible file")
	}
}

func (s *Source) getDirectoryFilesystem() (fs.FS, error) {
	return os.DirFS(s.Path), nil
}

func (s *Source) getArchiveFilesystem() (fs.FS, error) {
	switch ext := strings.ToLower(filepath.Ext(s.Path)); ext {
	case "mpq":
		return mpq.New(s.Path)
	case "zip":
		return nil, fmt.Errorf("not implemented")
	case "gzip", "gz":
		return nil, fmt.Errorf("not implemented")
	case "tar":
		return nil, fmt.Errorf("not implemented")
	case "7z", "7zip":
		return nil, fmt.Errorf("not implemented")
	default:
		return nil, fmt.Errorf("not implemented")
	}
}

func (s *Source) getRepoFilesystem() (fs.FS, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Source) Type() (t SourceType) {
	info, err := os.Stat(s.Path)
	if err != nil && os.IsNotExist(err) {
		return
	}

	if info.IsDir() {
		t = SourceDirectory
		return
	}

	switch ext := strings.ToLower(filepath.Ext(s.Path)); ext {
	case "mpq", "zip", "7z", "gz":
		t = SourceArchive
	}

	return
}
