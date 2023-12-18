package pl2Loader

import (
	"image/color"

	"github.com/gravestench/pl2"

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
	_ LoadsPl2Files               = &Service{}
)

type Dependency = LoadsPl2Files

type LoadsPl2Files = interface {
	Load(filepath string) (*pl2.PL2, error)
	ExtractPaletteFromPl2(pathPL2 string) (color.Palette, error)
}
