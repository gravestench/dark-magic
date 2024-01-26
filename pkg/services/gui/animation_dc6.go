package gui

import (
	"fmt"
	"image"
	"image/color"
	"time"

	dc6 "github.com/gravestench/dc6/pkg"

	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

const (
	defaultFPS = 24
)

func (s *Service) NewAnimationDC6(dc6Path, pl2Path string) (*AnimationDC6, error) {
	img, err := s.loadDc6WithPalette(dc6Path, pl2Path)
	if err != nil {
		return nil, fmt.Errorf("loading dc6: %v", err)
	}

	anim := &AnimationDC6{
		service:       s,
		node:          s.renderer.NewRenderable(),
		DC6:           img,
		frameDuration: time.Second / defaultFPS,
	}

	// put the animation into our map to update it later
	uuidAnim := anim.Renderable().UUID()
	s.animations[uuidAnim] = anim

	// get the current frame
	idxLayer, idxFrame := anim.CurrentLayerIndex(), anim.CurrentFrameIndex()
	frame := anim.DC6.Directions[idxLayer].Frames[idxFrame]

	anim.Renderable().SetImage(frame.ToImageRGBA())

	anim.SetSubLoopRange(0, anim.NumFrames())
	textureFootprint := createEmptyRGBA(image.Rectangle{
		Max: anim.BoundingBoxOfAllLayers().Size(),
	})
	anim.Renderable().SetImage(textureFootprint)

	anim.SetLoopingMode(Looping)
	anim.SetLoopCount(LoopForever)
	anim.Play()

	return anim, nil
}

var _ animation = &AnimationDC6{}
var _ directionalAnimation = &AnimationDC6{}

type AnimationDC6 struct {
	service *Service
	node    raylibRenderer.Renderable
	DC6     *dc6.DC6
	transportMode
	PlayMode
	currentLayer, currentFrame int // indices
	elapsed                    time.Duration
	frameDuration              time.Duration
	subLoopFrames              []int
	subLoopIndex               int
	loopMode                   LoopingMode
	timesToLoop                int // times to loop, -1 is forever
	timesLooped                int // how many loops have happened
}

func (a *AnimationDC6) Destroy() {
	a.Renderable().Destory()
	delete(a.service.animations, a.Renderable().UUID())
}

func (a *AnimationDC6) Renderable() raylibRenderer.Renderable {
	return a.node
}

func (a *AnimationDC6) Start() {
	a.timesLooped = 0
	startFrame, _ := a.SubLoopRange()
	a.SetCurrentFrameIndex(startFrame)
	a.transportMode = Playing
}

func (a *AnimationDC6) Stop() {
	a.transportMode = Stopped
	a.timesLooped = 0
	startFrame, _ := a.SubLoopRange()
	a.SetCurrentFrameIndex(startFrame)
}

func (a *AnimationDC6) Play() {
	a.transportMode = Playing
}

func (a *AnimationDC6) Pause() {
	a.transportMode = Paused
}

func (a *AnimationDC6) SetPlayMode(mode PlayMode) {
	switch mode {
	case PlayForward, PlayBackward, PlayYoYo:
		a.PlayMode = mode
	default:
		a.PlayMode = PlayForward
	}
}

func (a *AnimationDC6) NumLayers() int {
	return len(a.DC6.Directions)
}

func (a *AnimationDC6) clampLayerIndex() {
	a.currentLayer = clamp(a.currentLayer, 0, len(a.DC6.Directions))
}

func (a *AnimationDC6) CurrentLayerIndex() int {
	a.clampLayerIndex()
	return a.currentLayer
}

func (a *AnimationDC6) SetCurrentLayerIndex(i int) {
	a.currentLayer = i
	a.clampLayerIndex()
}

// BoundingBoxOfAllLayers returns the bounding box that encompasses all frames in all layers (directions).
func (a *AnimationDC6) BoundingBoxOfAllLayers() (r image.Rectangle) {
	for layerIndex := range a.DC6.Directions {
		r = combineRectangles(r, a.boundingBoxOfLayer(layerIndex))
	}

	return
}

// BoundingBoxOfCurrentLayer returns the bounding box that encompasses all frames in the current layer (direction).
func (a *AnimationDC6) BoundingBoxOfCurrentLayer() image.Rectangle {
	return a.boundingBoxOfLayer(a.currentLayer)
}

func (a *AnimationDC6) boundingBoxOfLayer(layerIndex int) (r image.Rectangle) {
	if layerIndex < 0 || layerIndex >= a.NumLayers() {
		return
	}

	for frameIndex := range a.DC6.Directions[layerIndex].Frames {
		r = combineRectangles(r, a.boundingBoxOfLayerFrame(layerIndex, frameIndex))
	}

	return
}

func (a *AnimationDC6) boundingBoxOfLayerFrame(layerIndex, frameIndex int) (r image.Rectangle) {
	if layerIndex < 0 || layerIndex >= a.NumLayers() {
		return
	}

	layer := a.DC6.Directions[layerIndex]
	if frameIndex < 0 || frameIndex >= len(layer.Frames) {
		return
	}

	frame := layer.Frames[frameIndex]
	ox, oy := int(frame.OffsetX), int(frame.OffsetY)
	offsetRect := createOffsetRectangle(frame.Bounds(), ox, oy)
	frameRect := a.DC6.Directions[layerIndex].Frames[frameIndex].Bounds()

	return combineRectangles(offsetRect, frameRect)
}

func (a *AnimationDC6) NumDirections() int {
	return a.NumLayers()
}

func (a *AnimationDC6) Direction() int {
	return a.CurrentLayerIndex()
}

func (a *AnimationDC6) SetDirection(i int) {
	a.SetCurrentLayerIndex(i)
}

func (a *AnimationDC6) NumFrames() int {
	layer := a.DC6.Directions[a.CurrentLayerIndex()]
	return len(layer.Frames)
}

func (a *AnimationDC6) clampFrameIndex() {
	idxLayer := a.CurrentLayerIndex()
	a.currentFrame = clamp(a.currentFrame, 0, len(a.DC6.Directions[idxLayer].Frames))
}

func (a *AnimationDC6) CurrentFrameIndex() int {
	a.clampFrameIndex()
	return a.currentFrame
}

func (a *AnimationDC6) SetCurrentFrameIndex(i int) {
	a.currentFrame = i
	a.clampFrameIndex()
}

// BoundingBoxOfAllFrames returns the bounding box that encompasses all frames in the animation.
func (a *AnimationDC6) BoundingBoxOfAllFrames() image.Rectangle {
	currentDirection := a.DC6.Directions[a.CurrentLayerIndex()]

	// Initialize the bounding box with invalid values
	minX, minY := 0, 0
	maxX, maxY := 0, 0

	for _, frame := range currentDirection.Frames {
		frameRect := image.Rect(
			int(frame.OffsetX),
			int(frame.OffsetY),
			int(frame.OffsetX)+int(frame.Width),
			int(frame.OffsetY)+int(frame.Height),
		)

		// Update the bounding box based on the frame's position and dimensions
		minX = min(minX, frameRect.Min.X)
		minY = min(minY, frameRect.Min.Y)
		maxX = max(maxX, frameRect.Max.X)
		maxY = max(maxY, frameRect.Max.Y)
	}

	// Create and return the bounding box as an image.Rectangle
	return image.Rect(minX, minY, maxX, maxY)
}

// BoundingBoxOfCurrentFrame returns the bounding box of the current frame.
func (a *AnimationDC6) BoundingBoxOfCurrentFrame() image.Rectangle {
	currentDirection := a.DC6.Directions[a.CurrentLayerIndex()]

	if a.currentFrame < 0 || a.currentFrame < len(currentDirection.Frames) {
		// Return an empty rectangle if the current frame index is out of bounds
		return image.Rect(0, 0, 0, 0)
	}

	frame := currentDirection.Frames[a.currentFrame]
	return image.Rect(
		int(frame.OffsetX),
		int(frame.OffsetY),
		int(frame.OffsetX)+int(frame.Width),
		int(frame.OffsetY)+int(frame.Height),
	)
}

func (a *AnimationDC6) FramesPerSecond() float32 {
	return float32(time.Second / a.frameDuration)
}

func (a *AnimationDC6) SetFramesPerSecond(f float32) {
	if f <= 0 {
		panic("FPS cannot be <= zero")
	}

	frameDurationSeconds := 1.0 / f
	frameDuration := time.Duration(frameDurationSeconds * float32(time.Second))
	a.frameDuration = frameDuration
}

func (a *AnimationDC6) Duration() time.Duration {
	return a.frameDuration
}

func (a *AnimationDC6) SetDuration(duration time.Duration) {
	if duration <= 0 {
		panic("Duration must be greater than zero")
	}

	a.frameDuration = duration / time.Duration(a.NumFrames())
}

func (a *AnimationDC6) Advance(elapsed time.Duration) {
	defer func() {
		a.node.SetImage(a.DC6.Directions[a.CurrentLayerIndex()].Frames[a.CurrentFrameIndex()].ToImageRGBA())
	}()

	if a.transportMode == Stopped { // If the animation is stopped, do nothing
		return
	}

	a.elapsed += elapsed

	for a.elapsed >= a.frameDuration {
		a.elapsed -= a.frameDuration

		switch a.PlayMode {
		case PlayForward:
			a.advanceFrameForward()
		case PlayBackward:
			a.advanceFrameBackward()
		case PlayYoYo:
			a.advanceFrameYoYo()
		}
	}
}

func (a *AnimationDC6) advanceFrameForward() {
	a.currentFrame = clamp(a.currentFrame+1, 0, a.NumFrames())

	isLastFrame := a.currentFrame >= a.NumFrames()-1
	if isLastFrame {
		a.timesLooped++
	}

	if a.timesToLoop == LoopForever {
		return
	}

	isLastLoop := a.timesLooped >= a.timesToLoop
	if isLastLoop && isLastFrame {
		a.transportMode = Stopped
	}
}

func (a *AnimationDC6) advanceFrameBackward() {
	a.currentFrame = clamp(a.currentFrame-1, 0, a.NumFrames())

	isLastFrame := a.currentFrame == 0
	if isLastFrame {
		a.timesLooped++
	}

	if a.timesToLoop == LoopForever {
		return
	}

	isLastLoop := a.timesLooped >= a.timesToLoop
	if isLastLoop && isLastFrame {
		a.transportMode = Stopped
	}
}

func (a *AnimationDC6) advanceFrameYoYo() {
	direction := 1 // Default direction is forward
	if a.timesLooped%2 != 0 {
		direction = -1 // Reverse direction when timesLooped is odd
	}

	// Calculate the new currentFrame with direction applied
	newFrame := a.currentFrame + direction

	// Check if it's a valid frame, and if not, reverse direction
	if newFrame < 0 || newFrame >= a.NumFrames() {
		direction *= -1
		newFrame = clamp(a.currentFrame+direction, 0, a.NumFrames())
		a.timesLooped++
	}

	a.currentFrame = newFrame

	if a.timesToLoop == LoopForever {
		return
	}

	// Check if it's the last loop and last frame to stop the animation
	isLastLoop := a.timesLooped >= (a.timesToLoop * 2) // forward+backward
	isLastFrame := a.currentFrame == 0

	if isLastLoop && isLastFrame {
		a.transportMode = Stopped
	}
}

func (a *AnimationDC6) LoopingMode() LoopingMode {
	return a.loopMode
}

func (a *AnimationDC6) SetLoopingMode(m LoopingMode) {
	a.loopMode = m
}

func (a *AnimationDC6) LoopCount() int {
	return a.timesToLoop
}

func (a *AnimationDC6) SetLoopCount(i int) {
	if i < LoopForever {
		i = LoopForever
	}

	a.timesToLoop = i
}

func (a *AnimationDC6) NumTimesLooped() int {
	return a.timesLooped
}

func (a *AnimationDC6) SubLoopFrames() []int {
	if len(a.subLoopFrames) < 1 {
		a.subLoopFrames = []int{0}
	}

	return a.subLoopFrames
}

func (a *AnimationDC6) SetSubLoopFrames(indices []int) {
	if len(indices) < 1 {
		indices = []int{0}
	}

	a.subLoopFrames = indices
}

func (a *AnimationDC6) SubLoopRange() (firstFrame, lastFrame int) {
	if len(a.subLoopFrames) < 1 {
		return 0, 0
	}

	firstFrame = a.subLoopFrames[0]
	lastFrame = a.subLoopFrames[len(a.subLoopFrames)-1]

	return
}

func (a *AnimationDC6) SetSubLoopRange(firstFrame, lastFrame int) {
	layer := a.DC6.Directions[a.CurrentLayerIndex()]
	firstFrame = clamp(firstFrame, 0, len(layer.Frames))
	lastFrame = clamp(lastFrame, 0, len(layer.Frames))

	a.subLoopFrames = nil

	if firstFrame == lastFrame {
		a.subLoopFrames = []int{firstFrame}
		return
	}

	for idx := firstFrame; idx < lastFrame; idx++ {
		a.subLoopFrames = append(a.subLoopFrames, idx)
	}
}

func clamp(value, min, max int) int {
	rangeSize := max - min
	return ((value-min)%rangeSize+rangeSize)%rangeSize + min
}

func createEmptyRGBA(rectangle image.Rectangle) *image.RGBA {
	// Create a new RGBA image of the specified rectangle size
	width := rectangle.Max.X - rectangle.Min.X
	height := rectangle.Max.Y - rectangle.Min.Y
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill the image with a transparent color (all zeros)
	transparentColor := color.RGBA{0, 0, 0, 0}
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			rgba.SetRGBA(x, y, transparentColor)
		}
	}

	return rgba
}

func createOffsetRectangle(baseRect image.Rectangle, xOffset, yOffset int) image.Rectangle {
	// Calculate the new rectangle coordinates with the offsets
	minX := baseRect.Min.X + xOffset
	minY := baseRect.Min.Y + yOffset
	maxX := baseRect.Max.X + xOffset
	maxY := baseRect.Max.Y + yOffset

	// Create and return the new image.Rectangle
	return image.Rect(minX, minY, maxX, maxY)
}

func combineRectangles(rect1, rect2 image.Rectangle) image.Rectangle {
	// Find the minimum and maximum coordinates
	minX := min(rect1.Min.X, rect2.Min.X)
	minY := min(rect1.Min.Y, rect2.Min.Y)
	maxX := max(rect1.Max.X, rect2.Max.X)
	maxY := max(rect1.Max.Y, rect2.Max.Y)

	// Create and return the new image.Rectangle
	return image.Rect(minX, minY, maxX, maxY)
}
