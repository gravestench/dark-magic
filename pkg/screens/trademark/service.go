package trademark

import (
	"image"
	"image/color"
	"log/slog"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/paths"
	"github.com/gravestench/dark-magic/pkg/services/dc6Loader"
	"github.com/gravestench/dark-magic/pkg/services/dccLoader"
	"github.com/gravestench/dark-magic/pkg/services/gui"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/pl2Loader"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
	"github.com/gravestench/dark-magic/pkg/services/screenManager"
	"github.com/gravestench/dark-magic/pkg/services/spriteManager"
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
	mpq      mpqLoader.Dependency
	dc6      dc6Loader.Dependency
	dcc      dccLoader.Dependency
	pl2      pl2Loader.Dependency
	gui      gui.Dependency
	sprite   spriteManager.Dependency

	root   raylibRenderer.Renderable
	update func()

	objects struct {
		logo                raylibRenderer.Renderable
		trademarkBackground *gui.TileSprite
	}
}

func (s *Screen) Init(mesh servicemesh.Mesh) {
	s.root = s.renderer.NewRenderable()
	time.Sleep(time.Second)

	s.initBackground()
	s.initLogo()
}

func (s *Screen) initLogo() {
	const (
		logoLeftFG = paths.Diablo2LogoFireLeft
		logoLeftBG = paths.Diablo2LogoBlackLeft

		logoRightFG = paths.Diablo2LogoFireRight
		logoRightBG = paths.Diablo2LogoBlackRight

		paletteAct1 = paths.PaletteTransformAct1
	)

	s.objects.logo = s.renderer.NewRenderable()
	s.objects.logo.SetParent(s.root)
	s.objects.logo.Disable()

	// load the graphic, keep trying until it is loaded.
	anim, err := s.gui.NewAnimationDC6(logoLeftFG, paletteAct1)
	for err != nil {
		anim, err = s.gui.NewAnimationDC6(logoLeftFG, paletteAct1)
	}

	node := anim.Renderable()
	node.SetParent(s.objects.logo)
	node.SetPosition(400, 20)
	node.SetOrigin(1, 0)
	node.SetBlendMode(rl.BlendAdditive)
}

func (s *Screen) initBackground() {
	ts, err := s.gui.NewTileSprite(paths.TrademarkScreen, paths.PaletteTransformSky, 4, 4)
	for err != nil {
		ts, err = s.gui.NewTileSprite(paths.TrademarkScreen, paths.PaletteTransformSky, 4, 4)
	}

	ts.Renderable().SetParent(s.root)

	s.root.SetScale(1024.0 / 800.0)
}

func (s *Screen) Name() string {
	return "Trademark Screen"
}

func (s *Screen) Ready() bool {
	if s.renderer == nil {
		return false
	}

	if s.mpq == nil {
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

	if s.sprite == nil {
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

	if s.sprite == nil {
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
		case gui.Dependency:
			s.gui = candidate
		case spriteManager.Dependency:
			s.sprite = candidate
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

func fillRect(w, h int, rgba color.RGBA) image.Image {
	// Create a new RGBA image with the specified width and height.
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Fill the entire rectangle with the specified color.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, rgba)
		}
	}

	return img
}
