package mpqLoader

import (
	"io"

	"github.com/gravestench/mpq"
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

var (
	_ runtime.Service             = &Service{}
	_ runtime.HasLogger           = &Service{}
	_ runtime.HasDependencies     = &Service{}
	_ configFile.HasDefaultConfig = &Service{}
	_ ReadsMpqArchives            = &Service{}
)

type Dependency = ReadsMpqArchives

type ReadsMpqArchives interface {
	RequiredArchivesLoaded() bool
	Archives() map[string]*mpq.MPQ
	AddArchive(filepath string) error
	RemoveArchive(filepath string) error
	Load(filepath string) (io.Reader, error)
}
