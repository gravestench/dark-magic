package loading

import (
	"image"
	"image/color"
	"image/draw"
	"log/slog"
	"time"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/paths"
	"github.com/gravestench/dark-magic/pkg/services/assetLoader"
	"github.com/gravestench/dark-magic/pkg/services/modalGameUI"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

type recipe interface {
	servicemesh.Service
	servicemesh.HasLogger
	servicemesh.HasDependencies
	modalGameUI.ModalGameUI
}

type Screen struct {
	logger *slog.Logger

	renderer raylibRenderer.Dependency
	assets   assetLoader.Dependency

	root   raylibRenderer.Renderable
	update func()
}

func (s *Screen) Init(mesh servicemesh.Mesh) {
	s.root = s.renderer.NewRenderable()

	s.initBackground()
	s.initLoadingImage()
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

func (s *Screen) initLoadingImage() {
	// load the dc6 image
	dc6Image, err := s.assets.LoadDc6(paths.LoadingScreen)
	for err != nil {
		time.Sleep(time.Second)
		dc6Image, err = s.assets.LoadDc6(paths.LoadingScreen)
	}

	// load a palette
	palette, err := s.assets.ExtractPaletteFromPl2(paths.PaletteTransformLoading)
	if err != nil {
		s.logger.Error("couldn't load the palette transform for the loading screen", "error", err)
		panic(err)
	}

	// apply the palette
	dc6Image.SetPalette(palette)

	frames := dc6Image.Directions[0].Frames
	currentFrame := 0

	// get a renderable
	r := s.renderer.NewRenderable()
	r.SetImage(frames[currentFrame].ToImageRGBA())

	t := time.Now()

	w, h := s.renderer.WindowSize()
	centerX := float32(w / 2)
	centerY := float32(h / 2)

	// example update callback for the renderable
	r.OnUpdate(func() {
		if time.Since(t) < time.Second/24 {
			return
		}

		t = time.Now()

		currentFrame = (currentFrame + 1) % len(frames)

		r.SetPosition(centerX, centerY)
		r.SetImage(frames[currentFrame].ToImageRGBA())
	})
}

func (s *Screen) Name() string {
	return "Loading Screen"
}

func (s *Screen) Ready() bool {
	if s.renderer == nil {
		return false
	}

	if s.assets == nil {
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
	if s.assets == nil {
		return false
	}

	if s.renderer == nil {
		return false
	}

	if !s.renderer.IsInit() {
		return false
	}

	return true
}

func (s *Screen) ResolveDependencies(services []servicemesh.Service) {
	for _, service := range services {
		switch candidate := service.(type) {
		case assetLoader.Dependency:
			s.assets = candidate
		case raylibRenderer.Dependency:
			s.renderer = candidate
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
