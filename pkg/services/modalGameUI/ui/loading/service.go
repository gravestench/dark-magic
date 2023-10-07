package loading

import (
	"image"
	"image/color"
	"image/draw"
	"time"

	"github.com/gravestench/runtime"
	"github.com/rs/zerolog"

	"github.com/gravestench/dark-magic/pkg/paths"
	"github.com/gravestench/dark-magic/pkg/services/dc6Loader"
	"github.com/gravestench/dark-magic/pkg/services/dccLoader"
	"github.com/gravestench/dark-magic/pkg/services/modalGameUI"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/pl2Loader"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

type recipe interface {
	runtime.Service
	runtime.HasLogger
	runtime.HasDependencies
	modalGameUI.ModalGameUI
}

type Screen struct {
	logger *zerolog.Logger

	renderer raylibRenderer.Dependency
	mpq      mpqLoader.Dependency
	dc6      dc6Loader.Dependency
	dcc      dccLoader.Dependency
	pl2      pl2Loader.Dependency

	root   raylibRenderer.Renderable
	update func()

	progress []bool
}

func (s *Screen) Init(rt runtime.Runtime) {
	s.root = s.renderer.NewRenderable()

	s.progress = make([]bool, 100)

	go func() {
		for {
			s.progress = make([]bool, 100)

			for idx := 0; idx < len(s.progress); idx++ {
				time.Sleep(time.Second / 10)
				s.progress[idx] = true
			}
		}
	}()

	s.initBackground()
	s.initLoadingImage()
}

func (s *Screen) initBackground() {
	// Create a new 10x10 image with RGBA color model
	img := image.NewRGBA(image.Rect(0, 0, 800, 600))

	// Fill the image with red color
	red := color.RGBA{0, 0, 0, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{red}, image.Point{}, draw.Src)

	s.root.SetImage(img)
}

func (s *Screen) initLoadingImage() {
	dc6Image, err := s.dc6.Load(paths.LoadingScreen)
	for err != nil {
		time.Sleep(time.Second)
		dc6Image, err = s.dc6.Load(paths.LoadingScreen)
	}

	palette, err := s.pl2.ExtractPaletteFromPl2(paths.PaletteTransformLoading)
	if err != nil {
		s.logger.Fatal().Msgf("couldn't load the palette transform for the loading screen: %v", err)
	}

	dc6Image.SetPalette(palette)

	loading := s.renderer.NewRenderable()
	d1 := dc6Image.Directions[0]
	frames := d1.Frames
	currentFrame := 0
	loading.SetImage(frames[currentFrame].ToImageRGBA())

	t := time.Now()

	loading.OnUpdate(func() {
		if time.Since(t) < time.Second/24 {
			return
		}

		t = time.Now()

		var total float64
		var current float64

		total = float64((len(s.progress)))
		for _, b := range s.progress {
			if b {
				current += 1
			}
		}

		currentFrame = int(float64(current/total) * float64(len(frames)))
		loading.SetPosition(float32(400-(frames[currentFrame].Width/2)), float32(300-(frames[currentFrame].Height/2)))
		loading.SetImage(frames[currentFrame%len(frames)].ToImageRGBA())
	})
}

func (s *Screen) Name() string {
	return "Loading Screen"
}

func (s *Screen) BindLogger(logger *zerolog.Logger) {
	s.logger = logger
}

func (s *Screen) Logger() *zerolog.Logger {
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

	return true
}

func (s *Screen) ResolveDependencies(rt runtime.Runtime) {
	for _, service := range rt.Services() {
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
	//TODO implement me
	panic("implement me")
}
