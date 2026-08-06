package raylibRenderer

import (
	"image"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (s *Service) render() {
	if s.cache == nil {
		return
	}

	s.renderRecursively(s.rootNode)
}

func (s *Service) renderRecursively(renderable Renderable) {
	if clip := renderable.Clip(); clip != nil {
		rl.BeginScissorMode(int32(clip.X), int32(clip.Y), int32(clip.Width), int32(clip.Height))
		defer rl.EndScissorMode()
	}
	if renderable.IsEnabled() {
		s.renderNode(renderable)
	}

	children := renderable.Children()
	if concrete, ok := renderable.(*node); ok {
		children = concrete.sortedChildren()
	}

	for _, child := range children {
		s.renderRecursively(child)
	}
}

func (s *Service) renderNode(node Renderable) {
	x, y := node.Position()
	tx := node.Texture()

	if node.dirty() {
		img := node.Image()
		if px, ok := contiguousRGBA(img); ok {
			if len(px) < 4 {
				return
			}
			rl.UpdateTexture(tx, px)
		} else {
			px := getAllPixelData(img)
			if len(px) == 0 {
				return
			}

			rl.UpdateTexture(tx, px)
		}
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
	if shader := node.Shader(); shader != nil {
		rl.BeginShaderMode(*shader)
		defer rl.EndShaderMode()
	}

	if node.BlendMode() != rl.BlendAlpha {
		rl.BeginBlendMode(node.BlendMode())
		defer rl.EndBlendMode()
	}
	rl.DrawTexturePro(tx, srcRect, dstRect, dstOrigin, node.Rotation(), tint)
}

// contiguousRGBA exposes an already GPU-ready RGBA surface without allocating
// or performing color-model conversion. Decoded and normalized engine assets
// use this layout; subimages with padded rows safely take the fallback path.
func contiguousRGBA(img image.Image) ([]byte, bool) {
	rgba, ok := img.(*image.RGBA)
	if !ok {
		return nil, false
	}
	bounds := rgba.Bounds()
	size := bounds.Dx() * bounds.Dy() * 4
	if size == 0 || rgba.Stride != bounds.Dx()*4 {
		return nil, false
	}
	start := rgba.PixOffset(bounds.Min.X, bounds.Min.Y)
	if start < 0 || start+size > len(rgba.Pix) {
		return nil, false
	}
	return rgba.Pix[start : start+size], true
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
			pixel := color.RGBAModel.Convert(img.At(bounds.Min.X+x, bounds.Min.Y+y)).(color.RGBA)
			pixels[index] = pixel
			index++
		}
	}

	return pixels
}
