package raylibRenderer

import (
	"image"
	"image/color"
	"math"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (s *Service) render() {
	if s.cache == nil {
		return
	}

	// These counters live for the whole renderer lifetime. Remember where this
	// frame starts so the debug overlay can also answer the much more useful
	// question: "how much work did the frame I can see require?"
	start := s.BackendDiagnostics()
	started := time.Now()
	s.frames.Add(1)
	s.renderRecursively(s.rootNode, nil)
	s.recordFrameWork(start)
	s.lastFrameRenderNS.Store(uint64(time.Since(started)))
}

func (s *Service) recordFrameWork(start BackendDiagnostics) {
	s.lastFrameDrawCalls.Store(s.drawCalls.Load() - start.DrawCalls)
	s.lastFrameNodesVisited.Store(s.nodesVisited.Load() - start.NodesVisited)
	s.lastFrameSubtreesCulled.Store(s.subtreesCulled.Load() - start.SubtreesCulled)
	s.lastFrameTextureUpdates.Store(s.textureUpdates.Load() - start.TextureUpdates)
	s.lastFrameUploadNS.Store(s.textureUploadNS.Load() - start.TextureUploadNS)
}

func (s *Service) renderRecursively(renderable *node, inheritedClip *rl.Rectangle) {
	s.nodesVisited.Add(1)
	if !renderable.visible {
		s.subtreesCulled.Add(1)
		return
	}
	clip := intersectClip(inheritedClip, renderable.Clip())
	if renderable.IsEnabled() {
		width, height := s.Resolution()
		if len(renderable.children) == 0 && renderable.outside(float32(width), float32(height)) {
			s.subtreesCulled.Add(1)
			return
		}
		if clip != nil {
			rl.BeginScissorMode(int32(clip.X), int32(clip.Y), int32(clip.Width), int32(clip.Height))
		}
		s.renderNode(renderable)
		if clip != nil {
			rl.EndScissorMode()
		}
	}

	children := renderable.sortedChildren()

	for _, child := range children {
		s.renderRecursively(child, clip)
	}
}

func (n *node) outside(viewWidth, viewHeight float32) bool {
	if n.Rotation() != 0 {
		return false
	}
	x, y := n.Position()
	scaleX, scaleY := n.scaleXY()
	bounds := n.Image().Bounds()
	width := float32(bounds.Dx()) * float32(math.Abs(float64(scaleX)))
	height := float32(bounds.Dy()) * float32(math.Abs(float64(scaleY)))
	origin := n.Origin()
	left, top := x-width*origin.X, y-height*origin.Y
	return left+width <= 0 || top+height <= 0 || left >= viewWidth || top >= viewHeight
}

func intersectClip(parent, own *rl.Rectangle) *rl.Rectangle {
	if parent == nil {
		return own
	}
	if own == nil {
		return parent
	}
	x1 := max(parent.X, own.X)
	y1 := max(parent.Y, own.Y)
	x2 := min(parent.X+parent.Width, own.X+own.Width)
	y2 := min(parent.Y+parent.Height, own.Y+own.Height)
	result := rl.NewRectangle(x1, y1, max(0, x2-x1), max(0, y2-y1))
	return &result
}

func (s *Service) renderNode(node *node) {
	s.drawCalls.Add(1)
	x, y := node.Position()
	tx := node.Texture()

	if node.dirty() {
		s.textureUpdates.Add(1)
		started := time.Now()
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
		s.textureUploadNS.Add(uint64(time.Since(started)))
	}

	//rl.DrawTextureEx(
	//	tx,
	//	rl.Vector2{X: float32(x), Y: float32(y)},
	//	node.Rotation(),
	//	node.Scale(),
	//	rl.NewColor(255, 255, 255, uint8(node.Opacity()*255)))

	origin := node.Origin()
	scaleX, scaleY := node.scaleXY()

	// src rect is at 0,0 and dimension of src texture
	srcWidth, srcHeight := float32(tx.Width), float32(tx.Height)
	srcRect := rl.NewRectangle(0, 0, srcWidth, srcHeight)

	// dst rect is at position of node, with scaled dimension of texture
	dstWidth, dstHeight := float32(tx.Width)*scaleX, float32(tx.Height)*scaleY
	dstRect := rl.NewRectangle(float32(x), float32(y), dstWidth, dstHeight)

	// node origin uses normalized value, applied to scaled dimension of texture
	// to provide relative offset, regardless of texture dimensions
	originX, originY := dstWidth*origin.X, dstHeight*origin.Y
	dstOrigin := rl.Vector2{X: originX, Y: originY}

	tint := rl.NewColor(255, 255, 255, uint8(node.Opacity()*255))
	if shader := node.Shader(); shader != nil {
		rl.BeginShaderMode(*shader)
		if texture := node.ShaderTexture(); texture != nil {
			// Auxiliary sampler registrations live for one Raylib batch only.
			rl.SetShaderValueTexture(*shader, node.ShaderTextureLocation(), *texture)
		}
		defer rl.EndShaderMode()
	}

	if node.BlendMode() != rl.BlendAlpha {
		if node.BlendMode() == rl.BlendCustom {
			// Diablo II UI draw mode 3 uses screen-like blending:
			// GL_ONE, GL_ONE_MINUS_SRC_COLOR, GL_FUNC_ADD.
			rl.SetBlendFactors(rl.One, rl.OneMinusSrcColor, rl.FuncAdd)
		}
		rl.BeginBlendMode(node.BlendMode())
		defer rl.EndBlendMode()
	}
	rl.DrawTexturePro(tx, srcRect, dstRect, dstOrigin, node.Rotation(), tint)
}

// contiguousRGBA exposes an already GPU-ready RGBA surface without allocating
// or performing color-model conversion. Decoded and normalized engine assets
// use this layout; subimages with padded rows safely take the fallback path.
func contiguousRGBA(img image.Image) ([]byte, bool) {
	bounds := img.Bounds()
	size := bounds.Dx() * bounds.Dy() * 4
	stride, pixels, offset := 0, []byte(nil), 0
	switch typed := img.(type) {
	case *image.RGBA:
		stride, pixels, offset = typed.Stride, typed.Pix, typed.PixOffset(bounds.Min.X, bounds.Min.Y)
	case *image.NRGBA:
		stride, pixels, offset = typed.Stride, typed.Pix, typed.PixOffset(bounds.Min.X, bounds.Min.Y)
	default:
		return nil, false
	}
	if size == 0 || stride != bounds.Dx()*4 {
		return nil, false
	}
	if offset < 0 || offset+size > len(pixels) {
		return nil, false
	}
	return pixels[offset : offset+size], true
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
