package gui

import (
	"image"
	"time"

	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

type transportMode int

const (
	Stopped transportMode = iota
	Playing
	Paused
)

type PlayMode int

const (
	PlayForward PlayMode = iota
	PlayBackward
	PlayYoYo
)

const LoopForever = -1

type LoopingMode int

const (
	NotLooping LoopingMode = iota
	Looping
	LoopingSpecificFrames
)

type animation interface {
	Renderable() raylibRenderer.Renderable
	Destroy()
	transportControls
	layerManagement
	frameManagement
	timeControls
	loopingControls
}

type directionalAnimation interface {
	animation
	directionControls
}

type transportControls interface {
	Start()
	Stop()
	Play()
	Pause()
	SetPlayMode(mode PlayMode)
}

type layerManagement interface {
	NumLayers() int

	CurrentLayerIndex() int
	SetCurrentLayerIndex(int)

	BoundingBoxOfAllLayers() image.Rectangle
	BoundingBoxOfCurrentLayer() image.Rectangle
}

type frameManagement interface {
	NumFrames() int

	CurrentFrameIndex() int
	SetCurrentFrameIndex(int)

	BoundingBoxOfAllFrames() image.Rectangle
	BoundingBoxOfCurrentFrame() image.Rectangle
}

type timeControls interface {
	FramesPerSecond() float32
	SetFramesPerSecond(float32)

	Duration() time.Duration
	SetDuration(duration time.Duration)

	Advance(elapsed time.Duration)
}

type loopingControls interface {
	LoopingMode() LoopingMode
	SetLoopingMode(mode LoopingMode)

	SubLoopFrames() []int
	SetSubLoopFrames([]int)

	SubLoopRange() (firstFrame, lastFrame int)
	SetSubLoopRange(firstFrame, lastFrame int)

	LoopCount() int
	SetLoopCount(int)
	NumTimesLooped() int
}

type directionControls interface {
	NumDirections() int

	Direction() int
	SetDirection(int)
}
