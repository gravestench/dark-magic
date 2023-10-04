package raylibRenderer

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const DefaultCameraName = ""

func (s *Service) GetDefaultCamera() *rl.Camera2D {
	return s.GetCamera(DefaultCameraName)
}

func (s *Service) AddCamera(name string) *rl.Camera2D {
	if c, found := s.cameras[name]; found {
		return c
	}

	newCamera := rl.NewCamera2D(rl.Vector2{}, rl.Vector2{}, 0, 1)

	s.cameras[name] = &newCamera

	return &newCamera
}

func (s *Service) GetCamera(name string) *rl.Camera2D {
	if c, found := s.cameras[name]; found {
		return c
	}

	return s.AddCamera(name)
}

func (s *Service) RemoveCamera(name string) {
	delete(s.cameras, name)
}
