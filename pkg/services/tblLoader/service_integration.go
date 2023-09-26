package tblLoader

import (
	tbl "github.com/gravestench/tbl_text"

	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/cacheManager"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

var (
	_ runtime.Service             = &Service{}
	_ runtime.HasLogger           = &Service{}
	_ runtime.HasDependencies     = &Service{}
	_ configFile.HasDefaultConfig = &Service{}
	_ cacheManager.HasCache       = &Service{}
	_ LoadsTblFiles               = &Service{}
)

type Dependency = LoadsTblFiles

type LoadsTblFiles = interface {
	Load(filepath string) (tbl.TextTable, error)
}
