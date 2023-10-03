package guiManager

import (
	"image"
)

func (s *Service) LayerIndex() int {
	return 1000
}

func (s *Service) Position() (x, y int) {
	return 0, 0
}

func (s *Service) Rotation() (degrees float32) {
	return 0
}

func (s *Service) Scale() (scale float32) {
	return 1
}

func (s *Service) Opacity() (opcaity float64) {
	return 1.0
}

func (s *Service) LayerImage() image.Image {
	return s.root.Image()
}
