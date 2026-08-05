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

type normalizedFS struct {
	fs.FS
	mpq bool
}

func (n normalizedFS) Open(name string) (fs.File, error) {
	name = strings.TrimLeft(name, "/\\")
	if n.mpq {
		name = strings.ReplaceAll(name, "/", "\\")
	} else {
		name = strings.ReplaceAll(name, "\\", "/")
	}
	return n.FS.Open(name)
}

func (s *Source) String() string {
	return s.Path
}

func (s Source) Filesystem() (fs.FS, error) {
	switch s.Type() {
	case SourceDirectory:
		return s.getDirectoryFilesystem()
	case SourceArchive:
		return s.getArchiveFilesystem()
	case SourceRepo:
		return s.getRepoFilesystem()
	default:
		return nil, fmt.Errorf("getting reader for %q: bad path or incompatible file", s.Path)
	}
}

func (s *Source) getDirectoryFilesystem() (fs.FS, error) {
	return normalizedFS{FS: os.DirFS(s.Path)}, nil
}

func (s *Source) getArchiveFilesystem() (fs.FS, error) {
	switch ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(s.Path)), "."); ext {
	case "mpq":
		archive, err := mpq.New(s.Path)
		if err != nil {
			return nil, err
		}
		return normalizedFS{FS: archive, mpq: true}, nil
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

func (s Source) Type() (t SourceType) {
	info, err := os.Stat(s.Path)
	if err != nil && os.IsNotExist(err) {
		return
	}

	if info.IsDir() {
		t = SourceDirectory
		return
	}

	switch ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(s.Path)), "."); ext {
	case "mpq", "zip", "7z", "gz":
		t = SourceArchive
	}

	return
}
