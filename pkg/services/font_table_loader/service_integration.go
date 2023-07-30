package font_table_loader

import (
	"github.com/gravestench/font_table"
	"github.com/gravestench/runtime"
)

var (
	_ runtime.Service         = &Service{}
	_ runtime.HasLogger       = &Service{}
	_ runtime.HasDependencies = &Service{}
	_ LoadsFontTableFiles     = &Service{}
)

type Dependency = LoadsFontTableFiles

type LoadsFontTableFiles = interface {
	Load(filepath string) (*font_table.Font, error)
}
