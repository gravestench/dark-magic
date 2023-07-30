package mpq_loader

import (
	"io"

	mpq "github.com/gravestench/mpq/pkg"

	"github.com/gravestench/runtime"
	"github.com/gravestench/runtime/examples/services/config_file"
)

var (
	_ runtime.Service              = &Service{}
	_ runtime.HasLogger            = &Service{}
	_ runtime.HasDependencies      = &Service{}
	_ config_file.HasDefaultConfig = &Service{}
	_ ReadsMpqArchives             = &Service{}
)

type Dependency = ReadsMpqArchives

type ReadsMpqArchives interface {
	Archives() map[string]*mpq.MPQ
	AddArchive(filepath string) error
	RemoveArchive(filepath string) error
	Load(filepath string) (io.Reader, error)
}
