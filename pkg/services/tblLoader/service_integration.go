package tblLoader

import (
	tbl "github.com/gravestench/tbl_text"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/cacheManager"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

var (
	_ servicemesh.Service         = &Service{}
	_ servicemesh.HasLogger       = &Service{}
	_ servicemesh.HasDependencies = &Service{}
	_ configFile.HasDefaultConfig = &Service{}
	_ cacheManager.HasCache       = &Service{}
	_ LoadsTblFiles               = &Service{}
)

type Dependency = LoadsTblFiles

type LoadsTblFiles = interface {
	Load(filepath string) (tbl.TextTable, error)
}
