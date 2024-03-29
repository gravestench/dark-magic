package pl2Loader

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"log/slog"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/cache"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
)

type Service struct {
	logger *slog.Logger
	mpq    mpqLoader.Dependency
	config *configFile.Config
	cache  *cache.Cache
}

func (s *Service) Init(mesh servicemesh.Mesh) {

}

func (s *Service) Name() string {
	return "PL2 Loader"
}

func (s *Service) Ready() bool {
	if s.mpq == nil {
		return false
	}

	if s.config == nil {
		return false
	}

	if !s.mpq.RequiredArchivesLoaded() {
		return false
	}

	return true
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
		return nil, fmt.Errorf("loading pl2: %v", err)
	}

	const (
		numColors          = 256
		numColorComponents = 4
		numBytesRGBA       = numColors * numColorComponents
		opaque             = 255
	)

	paletteData := make([]byte, numBytesRGBA)
	numRead, err := paletteStream.Read(paletteData)
	if err != nil && err != io.EOF {
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
			R: paletteData[(idx*numColorComponents)+0],
			G: paletteData[(idx*numColorComponents)+1],
			B: paletteData[(idx*numColorComponents)+2],
			A: opaque,
		}
	}

	return p, nil
}
