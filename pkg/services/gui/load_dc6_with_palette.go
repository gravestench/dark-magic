package gui

import (
	"fmt"

	dc6 "github.com/gravestench/dc6/pkg"
)

func (s *Service) loadDc6WithPalette(dc6Path, pl2Path string) (*dc6.DC6, error) {
	img, err := s.dc6.Load(dc6Path)
	if err != nil {
		return nil, fmt.Errorf("loading dc6: %v", err)
	}

	pal, err := s.pl2.ExtractPaletteFromPl2(pl2Path)
	if err != nil {
		return nil, fmt.Errorf("loading dc6: %v", err)
	}

	img.SetPalette(pal)

	return img, nil
}
