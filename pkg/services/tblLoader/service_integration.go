package tblLoader

import (
	tbl "github.com/gravestench/tbl_text"

	"github.com/gravestench/runtime"
)

var (
	_ runtime.Service         = &Service{}
	_ runtime.HasLogger       = &Service{}
	_ runtime.HasDependencies = &Service{}
	_ LoadsTblFiles           = &Service{}
)

type Dependency = LoadsTblFiles

type LoadsTblFiles = interface {
	Load(filepath string) (tbl.TextTable, error)
}
