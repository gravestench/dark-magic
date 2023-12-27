package mpqLoader

import (
	"io"

	"github.com/gravestench/mpq"
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

var (
	_ servicemesh.Service         = &Service{}
	_ servicemesh.HasLogger       = &Service{}
	_ servicemesh.HasDependencies = &Service{}
	_ configFile.HasDefaultConfig = &Service{}
	_ ReadsMpqArchives            = &Service{}
)

type Dependency = ReadsMpqArchives

type ReadsMpqArchives interface {
	servicemesh.Service
	RequiredArchivesLoaded() bool
	Archives() map[string]*mpq.MPQ
	AddArchive(filepath string) error
	RemoveArchive(filepath string) error
	Load(filepath string) (io.Reader, error)
}
