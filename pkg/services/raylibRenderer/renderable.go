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
		origin:   rl.Vector2{X: 0.5, Y: 0.5},
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
	origin    rl.Vector2

	onUpdate func()

	world    rl.Matrix
	local    rl.Matrix
	parent   Renderable
	children []Renderable

	isDirty bool
}

func (n *node) dirty() bool {
	if !n.isDirty {
		return false
	}

	n.isDirty = false

	return true
}

func (n *node) Origin() rl.Vector2 {
	return n.origin
}

func (n *node) SetOrigin(x, y float64) {
	n.origin.X = float32(x)
	n.origin.Y = float32(y)
}

func (n *node) update() {
	if n.onUpdate != nil {
		n.onUpdate()
	}

	for _, child := range n.children {
		child.update()
	}
}

func (n *node) OnUpdate(f func()) {
	n.onUpdate = f
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
	x = n.world.M12
	y = n.world.M13
	//z = float32(matrix.M14)
	return x, y
}

func (n *node) SetPosition(x, y float32) {
	n.local.M12 = x
	n.local.M13 = y
	//matrix.M14 = z
}

func (n *node) Rotation() (degrees float32) {
	// Compute scale factors
	scaleX := math.Sqrt(float64(n.world.M0*n.world.M0 + n.world.M1*n.world.M1 + n.world.M2*n.world.M2))

	// Normalize matrix components to remove scale
	m0Prime := n.world.M0 / float32(scaleX)
	m1Prime := n.world.M1 / float32(scaleX)

	// Get rotation in radians
	theta := math.Atan2(float64(m1Prime), float64(m0Prime))

	// Convert to degrees if necessary
	angleInDegrees := theta * 180.0 / math.Pi

	return float32(angleInDegrees)
}

//func (n *node) SetRotation(degrees float32) {
//	n.rotation = degrees
//	//// TODO :: setting rotation in the matrix isnt working right...
//	//radians := degrees * (math.Pi / 180.0) // Convert degrees to radians
//	//rotationMatrix := rl.MatrixRotateZ(radians)
//	//n.local = rl.MatrixMultiply(rotationMatrix, n.local)
//}

func (n *node) SetRotation(degrees float32) {
	// Extract existing scale factors
	scaleX := math.Sqrt(float64(n.local.M0*n.local.M0 + n.local.M1*n.local.M1 + n.local.M2*n.local.M2))
	scaleY := math.Sqrt(float64(n.local.M4*n.local.M4 + n.local.M5*n.local.M5 + n.local.M6*n.local.M6))
	scaleZ := math.Sqrt(float64(n.local.M8*n.local.M8 + n.local.M9*n.local.M9 + n.local.M10*n.local.M10))

	// Extract translation components
	tx := n.local.M12
	ty := n.local.M13
	tz := n.local.M14

	// Calculate new rotation matrix for Z-axis
	radians := float64(degrees) * math.Pi / 180.0
	cosTheta := math.Cos(radians)
	sinTheta := math.Sin(radians)

	// Set new rotation while maintaining existing scale and translation
	n.local.M0 = float32(cosTheta) * float32(scaleX)
	n.local.M1 = float32(sinTheta) * float32(scaleX)
	n.local.M4 = float32(-sinTheta) * float32(scaleY)
	n.local.M5 = float32(cosTheta) * float32(scaleY)

	// Z-axis values remain unchanged for 2D rotation
	n.local.M8 = n.local.M8 / float32(scaleZ)
	n.local.M9 = n.local.M9 / float32(scaleZ)
	n.local.M10 = float32(scaleZ)

	// Restore translation components
	n.local.M12 = tx
	n.local.M13 = ty
	n.local.M14 = tz
}

func (n *node) Scale() (scale float32) {
	//x := float32(n.world.M0)
	//y := float32(n.matrix.M5)
	//z := float32(n.matrix.M10)
	scale = float32(n.world.M10)

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
		n.renderer.logger.Error("caching texture", "key", key, "error", err)
	}
}

func (n *node) Image() image.Image {
	if n.image == nil {
		n.SetImage(defaultImage(60, 60))
	}

	return n.image
}

func (n *node) SetImage(image image.Image) {
	n.isDirty = true
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
		n.parent.removeChild(n)
	}

	n.parent = p

	if p != nil {
		n.parent.addChild(n)
	}
}

func (n *node) addChild(m Renderable) {
	if m == nil {
		return
	}

	n.children = append(n.children, m)
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
	n.world = rl.MatrixMultiply(n.local, parent)

	for idx := range n.children {
		n.children[idx].UpdateWorldMatrix(n.world)
	}
}

// WorldMatrix applies the local transform to the world matrix and returns it
func (n *node) WorldMatrix() rl.Matrix {
	return n.world
}

// LocalMatrix gets the local matrix
func (n *node) LocalMatrix() rl.Matrix {
	return n.local
}
