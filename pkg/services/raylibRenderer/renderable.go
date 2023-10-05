package raylibRenderer

import (
	"image"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/google/uuid"
)

func (s *Service) NewRenderable() Renderable {
	n := &node{
		renderer: s,
		uuid:     uuid.New(),
		opacity:  1,
		enabled:  true,
		local:    rl.MatrixIdentity(),
		world:    rl.MatrixIdentity(),
	}

	n.SetParent(s.rootNode)

	return n
}

type node struct {
	renderer  *Service
	uuid      uuid.UUID
	opacity   float32
	blendMode rl.BlendMode
	image     image.Image
	enabled   bool

	world    rl.Matrix
	local    rl.Matrix
	parent   Renderable
	children []Renderable
}

func (n *node) UUID() uuid.UUID {
	return n.uuid
}

func (n *node) ZIndex() float32 {
	return n.local.M14
}

func (n *node) SetZIndex(i float32) {
	n.local.M14 = i
}

func (n *node) Position() (x, y float32) {
	x = n.local.M12
	y = n.local.M13
	//z = float32(matrix.M14)
	return x, y
}

func (n *node) SetPosition(x, y float32) {
	n.local.M12 = x
	n.local.M13 = y
	//matrix.M14 = z
}

func (n *node) Rotation() (degrees float32) {
	degrees = float32(math.Atan2(float64(n.local.M8), float64(n.local.M0))) * (180.0 / math.Pi)
	//degrees = float32(math.Asin(-float64(n.local.M4))) * (180.0 / math.Pi)
	//degrees = float32(math.Atan2(float64(n.local.M9), float64(n.local.M5))) * (180.0 / math.Pi)

	return degrees
}

func (n *node) SetRotation(degrees float32) {
	radians := degrees * (math.Pi / 180.0) // Convert degrees to radians
	rotationMatrix := rl.MatrixRotateZ(radians)
	n.local = rl.MatrixMultiply(rotationMatrix, n.local)
}

func (n *node) Scale() (scale float32) {
	scale = float32(n.local.M0)
	//y := float32(n.matrix.M5)
	//z := float32(n.matrix.M10)

	return scale
}

func (n *node) SetScale(scale float32) {
	n.local.M0 = scale
	n.local.M5 = scale
	n.local.M10 = scale
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

func (n *node) Enable() {
	n.enabled = true
}

func (n *node) Disable() {
	n.enabled = false
}

func (n *node) IsEnabled() bool {
	return n.enabled
}

// SetParent sets the parent of this scene graph node
func (n *node) SetParent(p Renderable) {
	if n == p {
		return
	}

	if n.parent != nil {
		n.parent.(hasChildren).removeChild(n)
	}

	n.parent = p

	if p != nil {
		n.parent.(hasChildren).addChild(n)
	}
}

func (n *node) addChild(m Renderable) {
	if m == nil {
		return
	}

	var exists bool

	for _, child := range n.children {
		if child == m {
			exists = true
		}
	}

	if !exists {
		n.children = append(n.children, m)
	}
}

func (n *node) removeChild(m Renderable) {
	if m == nil {
		return
	}

	for idx := len(n.children) - 1; idx >= 0; idx-- {
		if n.children[idx] != m {
			continue
		}

		n.children = append(n.children[:idx], n.children[idx+1:]...)
	}
}

// Children yields the child nodes for this node
func (n *node) Children() []Renderable {
	return n.children
}

// UpdateWorldMatrix recursively updates the children world matrices with this
// nodes world matrix
func (n *node) UpdateWorldMatrix(parent rl.Matrix) {
	n.world = parent

	for idx := range n.children {
		n.children[idx].(hasMatrix).UpdateWorldMatrix(n.GetWorldMatrix())
	}
}

// GetWorldMatrix applies the local transform to the world matrix and returns it
func (n *node) GetWorldMatrix() rl.Matrix {
	return rl.MatrixMultiply(n.world, n.local)
}

// GetLocalMatrix gets the local matrix
func (n *node) GetLocalMatrix() rl.Matrix {
	return n.local
}
