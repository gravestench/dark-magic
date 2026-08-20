package raylibRenderer

import (
	"image"
	"image/color"
	"math"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// render draws the visible retained tree and publishes work deltas for the exact frame just produced.
func (s *Service) render() {
	if s.cache == nil {
		return
	}

	// Lifetime counters are monotonic; capture their starting values so diagnostics describe this visible frame.
	start := s.BackendDiagnostics()
	started := time.Now()

	s.frames.Add(1)
	s.renderRecursively(s.rootNode, nil)
	s.recordFrameWork(start)
	s.lastFrameRenderNS.Store(uint64(time.Since(started)))
}

// recordFrameWork converts cumulative counters into last-frame deltas without resetting lock-free diagnostics.
func (s *Service) recordFrameWork(start BackendDiagnostics) {
	s.lastFrameDrawCalls.Store(s.drawCalls.Load() - start.DrawCalls)
	s.lastFrameNodesVisited.Store(s.nodesVisited.Load() - start.NodesVisited)
	s.lastFrameSubtreesCulled.Store(s.subtreesCulled.Load() - start.SubtreesCulled)
	s.lastFrameTextureUpdates.Store(s.textureUpdates.Load() - start.TextureUpdates)
	s.lastFrameUploadNS.Store(s.textureUploadNS.Load() - start.TextureUploadNS)
}

// renderRecursively culls hidden subtrees, intersects inherited clips, and visits children in cached Z order.
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

	for _, child := range renderable.sortedChildren() {
		s.renderRecursively(child, clip)
	}
}

// outside culls only unrotated leaves because rotated bounds and child transforms require a more conservative test.
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

// intersectClip returns the common rectangle while preserving nil as the absence of clipping.
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

// renderNode uploads changed pixels before drawing with node-local shader, blend, transform, and tint state.
func (s *Service) renderNode(node *node) {
	s.drawCalls.Add(1)

	x, y := node.Position()

	texture := node.Texture()
	if !s.uploadDirtyNodeTexture(node, texture) {
		return
	}

	source, destination, destinationOrigin := nodeTextureGeometry(node, texture, x, y)
	tint := node.Tint()
	tint.A = uint8(node.Opacity() * 255)

	drawNodeTexture(node, texture, source, destination, destinationOrigin, tint)
}

// uploadDirtyNodeTexture consumes one dirty flag.
// It preserves the fast contiguous-RGBA path before falling back to color-model conversion.
func (s *Service) uploadDirtyNodeTexture(node *node, texture rl.Texture2D) bool {
	if !node.dirty() {
		return true
	}

	s.textureUpdates.Add(1)

	started := time.Now()

	img := node.Image()
	if pixels, ok := contiguousRGBA(img); ok {
		if len(pixels) < 4 {
			return false
		}

		rl.UpdateTexture(texture, pixels)
	} else {
		pixels := getAllPixelData(img)
		if len(pixels) == 0 {
			return false
		}

		rl.UpdateTexture(texture, pixels)
	}

	s.textureUploadNS.Add(uint64(time.Since(started)))

	return true
}

// nodeTextureGeometry derives Raylib source and destination rectangles without allocating intermediate image data.
func nodeTextureGeometry(
	node *node,
	texture rl.Texture2D,
	x float32,
	y float32,
) (rl.Rectangle, rl.Rectangle, rl.Vector2) {
	sourceWidth, sourceHeight := float32(texture.Width), float32(texture.Height)
	source := rl.NewRectangle(0, 0, sourceWidth, sourceHeight)

	scaleX, scaleY := node.scaleXY()
	destinationWidth := float32(texture.Width) * scaleX
	destinationHeight := float32(texture.Height) * scaleY
	destination := rl.NewRectangle(x, y, destinationWidth, destinationHeight)

	origin := node.Origin()
	destinationOrigin := rl.Vector2{X: destinationWidth * origin.X, Y: destinationHeight * origin.Y}

	return source, destination, destinationOrigin
}

// drawNodeTexture brackets the draw with shader then blend state so deferred cleanup unwinds blend before shader.
func drawNodeTexture(
	node *node,
	texture rl.Texture2D,
	source rl.Rectangle,
	destination rl.Rectangle,
	destinationOrigin rl.Vector2,
	tint rl.Color,
) {
	if shader := node.Shader(); shader != nil {
		rl.BeginShaderMode(*shader)

		if shaderTexture := node.ShaderTexture(); shaderTexture != nil {
			// Auxiliary sampler registration is valid only for the Raylib batch bracketed by this shader mode.
			rl.SetShaderValueTexture(*shader, node.ShaderTextureLocation(), *shaderTexture)
		}

		defer rl.EndShaderMode()
	}

	if node.BlendMode() != rl.BlendAlpha {
		if node.BlendMode() == rl.BlendCustom {
			// Diablo II UI draw mode 3: GL_ONE, GL_ONE_MINUS_SRC_COLOR, GL_FUNC_ADD.
			rl.SetBlendFactors(rl.One, rl.OneMinusSrcColor, rl.FuncAdd)
		}

		rl.BeginBlendMode(node.BlendMode())

		defer rl.EndBlendMode()
	}

	rl.DrawTexturePro(texture, source, destination, destinationOrigin, node.Rotation(), tint)
}

// contiguousRGBA exposes an already GPU-ready surface without allocation or color-model conversion.
// Subimages with padded rows safely take the conversion fallback.
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

// getAllPixelData converts arbitrary image models into a tightly packed RGBA slice in row-major bounds order.
func getAllPixelData(img image.Image) []color.RGBA {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
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
