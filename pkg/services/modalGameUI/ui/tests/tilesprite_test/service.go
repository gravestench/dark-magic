package tilesprite_test

import (
	"image"
	"log/slog"
	"time"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/paths"
	"github.com/gravestench/dark-magic/pkg/services/dc6Loader"
	"github.com/gravestench/dark-magic/pkg/services/dccLoader"
	"github.com/gravestench/dark-magic/pkg/services/guiManager"
	"github.com/gravestench/dark-magic/pkg/services/modalGameUI"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/pl2Loader"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

type recipe interface {
	servicemesh.Service
	servicemesh.HasLogger
	servicemesh.HasDependencies
	modalGameUI.ModalGameUI
}

var _ recipe = &Screen{}

type Screen struct {
	logger *slog.Logger

	renderer raylibRenderer.Dependency
	mpq      mpqLoader.Dependency
	dc6      dc6Loader.Dependency
	dcc      dccLoader.Dependency
	pl2      pl2Loader.Dependency
	gui      guiManager.Dependency

	root   raylibRenderer.Renderable
	update func()

	testTileSprite *guiManager.TileSprite
}

func (s *Screen) Init(mesh servicemesh.Mesh) {
	s.root = s.renderer.NewRenderable()

	s.initTest()
}

func (s *Screen) initTest() {
	// load the dc6 image
	spritePath := paths.TrademarkScreen
	dc6Image, err := s.dc6.Load(spritePath)
	for err != nil {
		time.Sleep(time.Second)
		dc6Image, err = s.dc6.Load(spritePath)
	}

	// load a palette
	palettePath := paths.PaletteTransformSky
	palette, err := s.pl2.ExtractPaletteFromPl2(palettePath)
	for err != nil {
		palette, err = s.pl2.ExtractPaletteFromPl2(palettePath)
	}

	// apply the palette
	dc6Image.SetPalette(palette)

	var tiles []image.Image
	var tileConfigs []guiManager.TileConfig
	const size = 4

	for idx, frame := range dc6Image.Directions[0].Frames {
		tiles = append(tiles, frame)
		tileConfigs = append(tileConfigs, guiManager.TileConfig{
			Index:    idx,
			Position: image.Point{X: 256 * (idx % size), Y: 256 * (idx / size)},
		})
	}

	s.testTileSprite, err = s.gui.NewTileSprite(guiManager.TileSpriteConfig{
		Parent:      s.root,
		Tiles:       tiles,
		TileConfigs: tileConfigs,
	})

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
	if s.mpq == nil {
		return false
	}

	if !s.mpq.RequiredArchivesLoaded() {
		return false
	}

	if s.dc6 == nil {
		return false
	}

	if s.dcc == nil {
		return false
	}

	if s.pl2 == nil {
		return false
	}

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
		case mpqLoader.Dependency:
			s.mpq = candidate
		case dc6Loader.Dependency:
			s.dc6 = candidate
		case dccLoader.Dependency:
			s.dcc = candidate
		case pl2Loader.Dependency:
			s.pl2 = candidate
		case raylibRenderer.Dependency:
			s.renderer = candidate
		case guiManager.Dependency:
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
