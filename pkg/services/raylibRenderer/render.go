package raylibRenderer

import (
	"image"
	"image/color"
	"sort"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (s *Service) render() {
	for s.cache == nil {
		time.Sleep(time.Second)
	}

	s.renderRecursively(s.rootNode)
}

func (s *Service) renderRecursively(node Renderable) {
	if node.IsEnabled() {
		s.renderNode(node)
	}

	children := node.Children()

	sort.Slice(children, func(i, j int) bool {
		return children[i].ZIndex() < children[j].ZIndex()
	})

	for _, child := range children {
		s.renderRecursively(child)
	}
}

func (s *Service) renderNode(node Renderable) {
	x, y := node.Position()
	tx := node.Texture()

	if node.dirty() {
		px := getAllPixelData(node.Image())
		if len(px) < 4 {
			return
		}

		rl.UpdateTexture(tx, px)
	}

	//rl.DrawTextureEx(
	//	tx,
	//	rl.Vector2{X: float32(x), Y: float32(y)},
	//	node.Rotation(),
	//	node.Scale(),
	//	rl.NewColor(255, 255, 255, uint8(node.Opacity()*255)))

	origin := node.Origin()
	scale := node.Scale()

	// src rect is at 0,0 and dimension of src texture
	srcWidth, srcHeight := float32(tx.Width), float32(tx.Height)
	srcRect := rl.NewRectangle(0, 0, srcWidth, srcHeight)

	// dst rect is at position of node, with scaled dimension of texture
	dstWidth, dstHeight := float32(tx.Width)*scale, float32(tx.Height)*scale
	dstRect := rl.NewRectangle(float32(x), float32(y), dstWidth, dstHeight)

	// node origin uses normalized value, applied to scaled dimension of texture
	// to provide relative offset, regardless of texture dimensions
	originX, originY := dstWidth*origin.X, dstHeight*origin.Y
	dstOrigin := rl.Vector2{X: originX, Y: originY}

	tint := rl.NewColor(255, 255, 255, uint8(node.Opacity()*255))

	rl.DrawTexturePro(tx, srcRect, dstRect, dstOrigin, node.Rotation(), tint)
}

func getAllPixelData(img image.Image) []color.RGBA {
	// Get the dimensions of the image
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Convert the RGBA image to a slice of color.RGBA
	pixels := make([]color.RGBA, width*height)
	index := 0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := img.At(x, y).(color.RGBA)
			pixels[index] = pixel
			index++
		}
	}

	return pixels
}
