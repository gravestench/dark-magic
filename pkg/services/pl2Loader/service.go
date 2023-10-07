package pl2Loader

import (
	"fmt"
	"image"
	"image/color"

	"github.com/gravestench/runtime"
	"github.com/rs/zerolog"

	"github.com/gravestench/dark-magic/pkg/cache"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
)

type Service struct {
	logger *zerolog.Logger
	mpq    mpqLoader.Dependency
	config configFile.Dependency
	cache  *cache.Cache
}

func (s *Service) Init(rt runtime.Runtime) {

}

func (s *Service) Name() string {
	return "PL2 Loader"
}

func (s *Service) BindLogger(logger *zerolog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *zerolog.Logger {
	return s.logger
}

func (s *Service) ExtractPaletteFromPl2(pathPL2 string) (color.Palette, error) {
	paletteStream, err := s.mpq.Load(pathPL2)
	if err != nil {
		return nil, fmt.Errorf("loading pl2: %v", err)
	}

	const (
		numColors    = 256
		numBytesRGBA = numColors * 4
		opaque       = 255
	)

	paletteData := make([]byte, numBytesRGBA)
	numRead, err := paletteStream.Read(paletteData)
	if err != nil {
		return nil, fmt.Errorf("reading from PL2 stream: %v", err)
	} else if numRead != numBytesRGBA {
		return nil, fmt.Errorf("couldn't read all palette bytes")
	}

	p := make(color.Palette, numColors)
	for idx := range p {
		if idx == 0 {
			p[idx] = image.Transparent

			continue
		}

		p[idx] = color.RGBA{
			R: paletteData[(idx*4)+0],
			G: paletteData[(idx*4)+1],
			B: paletteData[(idx*4)+2],
			A: opaque,
		}
	}

	return p, nil
}
