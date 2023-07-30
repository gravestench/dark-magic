package record_manager

import (
	"github.com/gravestench/runtime"
)

var (
	_ runtime.Service         = &Service{}
	_ runtime.HasDependencies = &Service{}
	//_ config_file.HasDefaultConfig = &Service{}
	_ LoadsDiablo2Records = &Service{}
)

type Dependency = LoadsDiablo2Records

type LoadsDiablo2Records interface {
	LoadRecords() error
}
