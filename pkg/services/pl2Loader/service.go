package pl2Loader

import (
	"fmt"
	"image"
	"image/color"
	"log/slog"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/cache"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
)

type Service struct {
	logger *slog.Logger
	mpq    mpqLoader.Dependency
	config configFile.Dependency
	cache  *cache.Cache
}

func (s *Service) Init(mesh servicemesh.Mesh) {

}

func (s *Service) Name() string {
	return "PL2 Loader"
}

func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}

func (s *Service) ExtractPaletteFromPl2(pathPL2 string) (color.Palette, error) {
	paletteStream, err := s.mpq.Load(pathPL2)
	if err != nil {
		return nil, fmt.Errorf("loading pl2", "error", err)
	}

	const (
		numColors    = 256
		numBytesRGBA = numColors * 4
		opaque       = 255
	)

	paletteData := make([]byte, numBytesRGBA)
	numRead, err := paletteStream.Read(paletteData)
	if err != nil {
		return nil, fmt.Errorf("reading from PL2 stream", "error", err)
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
