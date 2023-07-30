package wav_loader

import (
	"github.com/gravestench/runtime"
)

var (
	_ runtime.Service         = &Service{}
	_ runtime.HasLogger       = &Service{}
	_ runtime.HasDependencies = &Service{}
	_ LoadsWavFiles           = &Service{}
)

type Dependency = LoadsWavFiles

type LoadsWavFiles = interface {
	Load(filepath string) ([]byte, error)
}
