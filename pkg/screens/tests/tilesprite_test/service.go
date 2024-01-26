package tilesprite_test

import (
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

var _ recipe = &Screen{}

type Screen struct {
	logger *slog.Logger

	renderer raylibRenderer.Dependency
	gui      gui.Dependency

	root   raylibRenderer.Renderable
	update func()
}

func (s *Screen) Init(mesh servicemesh.Mesh) {
	s.root = s.renderer.NewRenderable()

	s.initTest()
}

func (s *Screen) initTest() {
	ts, err := s.gui.NewTileSprite(paths.TrademarkScreen, paths.PaletteTransformSky, 4, 4)
	for err != nil {
		ts, err = s.gui.NewTileSprite(paths.TrademarkScreen, paths.PaletteTransformSky, 4, 4)
	}

	ts.Renderable().SetParent(s.root)

	s.root.SetScale(1024.0 / 800.0)
}

func (s *Screen) Name() string {
	return "Tilesprite Test Screen"
}

func (s *Screen) Ready() bool {
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
	return "trademark"
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
