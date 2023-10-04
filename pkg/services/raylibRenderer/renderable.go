package raylibRenderer

import (
	"image"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/google/uuid"
)

func (s *Service) NewRenderable() Renderable {
	n := &node{
		renderer: s,
		uuid:     uuid.New(),
		opacity:  1,
		scale:    1,
	}

	s.renderables[n.uuid] = n

	return n
}

type node struct {
	renderer  *Service
	uuid      uuid.UUID
	zIndex    int
	position  rl.Vector2
	rotation  float32
	scale     float32
	opacity   float32
	blendMode rl.BlendMode
	image     image.Image
}

func (n *node) UUID() uuid.UUID {
	return n.uuid
}

func (n *node) ZIndex() int {
	return n.zIndex
}

func (n *node) SetZIndex(i int) {
	n.zIndex = i
}

func (n *node) Position() (x, y int) {
	return int(n.position.X), int(n.position.Y)
}

func (n *node) SetPosition(x, y int) {
	n.position.X, n.position.Y = float32(x), float32(y)
}

func (n *node) Rotation() (degrees float32) {
	return n.rotation
}

func (n *node) SetRotation(degrees float32) {
	n.rotation = degrees
}

func (n *node) Scale() (scale float32) {
	return n.scale
}

func (n *node) SetScale(scale float32) {
	n.scale = scale
}

func (n *node) Opacity() (opacity float32) {
	return n.opacity
}

func (n *node) SetOpacity(opacity float32) {
	n.opacity = opacity
}

func (n *node) BlendMode() (mode rl.BlendMode) {
	return n.blendMode
}

func (n *node) SetBlendMode(mode rl.BlendMode) {
	n.blendMode = mode
}

func (n *node) Texture() rl.Texture2D {
	tx, isNew := n.renderer.GetTexture(n.uuid, n.Image())

	if isNew {
		rl.UpdateTexture(tx, getAllPixelData(n.Image()))
	}

	return tx
}

func (n *node) SetTexture(tx rl.Texture2D) {
	key := n.uuid.String()

	img := n.Image()
	bounds := img.Bounds()
	numBytes := bounds.Dx() * bounds.Dy() * 4

	if err := n.renderer.cache.Insert(key, tx, numBytes); err != nil {
		n.renderer.logger.Error().Msgf("[%s] caching texture: %v", key, err)
	}
}

func (n *node) Image() image.Image {
	if n.image == nil {
		n.image = defaultImage(60, 60)
	}

	return n.image
}

func (n *node) SetImage(image image.Image) {
	n.image = image
}
