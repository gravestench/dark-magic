package guiManager

import (
	"image"
	"image/color"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/runtime"
	"github.com/rs/zerolog"

	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
	"github.com/gravestench/dark-magic/pkg/services/spriteManager"
)

type Service struct {
	logger *zerolog.Logger

	sprite   spriteManager.Dependency
	renderer raylibRenderer.Dependency

	inputState struct {
		keys []int
	}

	texture *rl.Texture2D

	root Node
}

func (s *Service) Init(rt runtime.Runtime) {
	// This init method will be invoked by the runtime
	// as soon as the dependency resolution has finished.
	// If the service does not implement runtime.HasDependencies,
	// then this method is invoked immediately.
	//keyboard.Listen(s.updateKeyboardKeys)

	s.root = s.NewTreeNode(800, 600)

	n1 := s.NewNode()

	n1.SetImageFunc(func() image.Image {
		width, height := 100, 100
		img := image.NewRGBA(image.Rect(0, 0, width, height))

		// Fill the image with a white background
		white := color.RGBA{255, 0, 0, 255}
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				img.Set(x, y, white)
			}
		}

		return img
	})

	n2 := s.NewNode()
	n2.SetParent(n1)
	n2.SetPosition(image.Point{2, 2})
	n2.SetUpdateFunc(func() {
		p := n2.Position()
		p.X++
		n2.SetPosition(p)
	})

	n2.SetImageFunc(func() image.Image {
		width, height := 10, 10
		img := image.NewRGBA(image.Rect(0, 0, width, height))

		// Fill the image with a white background
		white := color.RGBA{0, 255, 0, 255}
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				img.Set(x, y, white)
			}
		}

		return img
	})

	ticker := time.NewTicker(time.Second / 60)

	// Create a Goroutine to handle the ticker
	go func() {
		for {
			select {
			case <-ticker.C:
				// Call your function here
				s.Update()
			}
		}
	}()
}

func (s *Service) Name() string {
	return "GUI"
}

// the following methods are boilerplate, but they are used
// by the runtime to enforce a standard logging format.

func (s *Service) BindLogger(logger *zerolog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *zerolog.Logger {
	return s.logger
}

func (s *Service) NewNode() Node {
	n := s.NewTreeNode(0, 0)

	s.root.AddChild(n)

	return n
}

func (s *Service) Nodes() []Node {
	return s.root.Children()
}

func (s *Service) Update() {
	s.root.Update()
}
