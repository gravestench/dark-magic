package loading

import (
	"image"
	"image/color"
	"image/draw"
	"log/slog"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/paths"
	"github.com/gravestench/dark-magic/pkg/services/gui"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
	"github.com/gravestench/dark-magic/pkg/services/screenManager"
)

type recipe interface {
	servicemesh.Service
	servicemesh.HasLogger
	servicemesh.HasDependencies
	screenManager.ModalGameUI
}

type Screen struct {
	logger *slog.Logger

	renderer raylibRenderer.Dependency
	gui      gui.Dependency

	root   raylibRenderer.Renderable
	update func()
}

func (s *Screen) Init(mesh servicemesh.Mesh) {
	s.root = s.renderer.NewRenderable()

	s.initBackground()
	s.initLoadingAnimation()
}

func (s *Screen) initBackground() {
	// Create a new 10x10 image with RGBA color model
	img := image.NewRGBA(image.Rect(0, 0, 800, 600))

	// Fill the image with red color
	bgColor := color.RGBA{0, 0, 0, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	s.root.SetOrigin(0, 0)
	s.root.SetImage(img)
}

func (s *Screen) initLoadingAnimation() {
	anim, err := s.gui.NewAnimationDC6(paths.LoadingScreen, paths.PaletteTransformLoading)
	for err != nil {
		anim, err = s.gui.NewAnimationDC6(paths.LoadingScreen, paths.PaletteTransformLoading)
	}

	w, h := s.renderer.WindowSize()
	centerX := float32(w / 2)
	centerY := float32(h / 2)

	anim.SetPlayMode(gui.PlayYoYo)
	anim.Renderable().SetPosition(centerX, centerY)
}

func (s *Screen) Name() string {
	return "Loading Screen"
}

func (s *Screen) Ready() bool {
	if s.gui == nil {
		return false
	}

	if !s.gui.Ready() {
		return false
	}

	return true
}

func (s *Screen) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Screen) Logger() *slog.Logger {
	return s.logger
}

func (s *Screen) DependenciesResolved() bool {
	if s.renderer == nil {
		return false
	}

	if !s.renderer.IsInit() {
		return false
	}

	if s.gui == nil {
		return false
	}

	if !s.gui.Ready() {
		return false
	}

	return true
}

func (s *Screen) ResolveDependencies(services []servicemesh.Service) {
	for _, service := range services {
		switch candidate := service.(type) {
		case raylibRenderer.Dependency:
			s.renderer = candidate
		case gui.Dependency:
			s.gui = candidate
		}
	}
}

func (s *Screen) Mode() string {
	return "loading"
}

func (s *Screen) Renderable() raylibRenderer.Renderable {
	return s.root
}

func (s *Screen) Update() {
	if s.update == nil {
		return
	}

	s.update()

}

func (s *Screen) OnUpdate(f func()) {

}
