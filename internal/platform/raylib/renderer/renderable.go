package raylibRenderer

import (
	"fmt"
	"image"
	"math"
	"sort"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/google/uuid"
)

func (s *Service) newNode() *node {
	n := &node{
		renderer:       s,
		uuid:           uuid.New(),
		opacity:        1,
		tint:           rl.White,
		enabled:        true,
		visible:        true,
		transformDirty: true,
		local:          rl.MatrixIdentity(),
		world:          rl.MatrixIdentity(),
		origin:         rl.Vector2{X: 0.5, Y: 0.5},
	}

	n.setParent(s.rootNode)

	return n
}

type node struct {
	renderer              *Service
	uuid                  uuid.UUID
	opacity               float32
	tint                  rl.Color
	blendMode             rl.BlendMode
	image                 image.Image
	enabled               bool
	visible               bool
	origin                rl.Vector2
	textureVariant        string
	textureKey            string
	sharedTexture         bool
	textureKeys           map[string]struct{}
	clip                  *rl.Rectangle
	shader                *rl.Shader
	shaderTexture         *rl.Texture2D
	shaderTextureLocation int32

	world          rl.Matrix
	local          rl.Matrix
	parent         *node
	children       []*node
	childrenSorted bool
	transformDirty bool
	worldRotation  float32

	isDirty bool
}

func (n *node) Shader() *rl.Shader           { return n.shader }
func (n *node) ShaderTexture() *rl.Texture2D { return n.shaderTexture }
func (n *node) ShaderTextureLocation() int32 { return n.shaderTextureLocation }

func (n *node) SetShader(shader *rl.Shader, texture *rl.Texture2D, location int32) {
	n.shader, n.shaderTexture, n.shaderTextureLocation = shader, texture, location
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

func (n *node) UUID() uuid.UUID {
	return n.uuid
}

func (n *node) ZIndex() float32 {
	return n.local.M14
}

func (n *node) SetZIndex(i float32) {
	if n.local.M14 == i {
		return
	}
	n.local.M14 = i
	n.transformDirty = true
	if n.parent != nil {
		n.parent.childrenSorted = false
	}
}

func (n *node) Position() (x, y float32) {
	x = n.world.M12
	y = n.world.M13
	//z = float32(matrix.M14)
	return x, y
}

func (n *node) SetPosition(x, y float32) {
	if n.local.M12 == x && n.local.M13 == y {
		return
	}
	n.local.M12 = x
	n.local.M13 = y
	n.transformDirty = true
}

func (n *node) Rotation() (degrees float32) {
	return n.worldRotation
}

func matrixRotation(matrix rl.Matrix) float32 {
	// Compute scale factors
	scaleX := math.Sqrt(float64(matrix.M0*matrix.M0 + matrix.M1*matrix.M1 + matrix.M2*matrix.M2))
	if scaleX == 0 {
		return 0
	}

	// Normalize matrix components to remove scale
	m0Prime := matrix.M0 / float32(scaleX)
	m1Prime := matrix.M1 / float32(scaleX)

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
	n.transformDirty = true
}

func (n *node) Scale() (scale float32) {
	return n.world.M0
}

func (n *node) scaleXY() (float32, float32) {
	return n.world.M0, n.world.M5
}

func (n *node) SetScale(scale float32) {
	n.setScaleXY(scale, scale)
}

func (n *node) setScaleXY(x, y float32) {
	if n.local.M0 == x && n.local.M5 == y {
		return
	}
	n.local.M0 = x
	n.local.M5 = y
	n.local.M10 = 1
	n.transformDirty = true
}

func (n *node) Opacity() (opacity float32) {
	return n.opacity
}

func (n *node) SetOpacity(opacity float32) {
	n.opacity = opacity
}

func (n *node) Tint() rl.Color { return n.tint }

func (n *node) SetTint(tint rl.Color) { n.tint = tint }

func (n *node) BlendMode() (mode rl.BlendMode) {
	return n.blendMode
}

func (n *node) SetBlendMode(mode rl.BlendMode) {
	n.blendMode = mode
}

func (n *node) Clip() *rl.Rectangle { return n.clip }

func (n *node) SetClip(clip *rl.Rectangle) { n.clip = clip }

func (n *node) Texture() rl.Texture2D {
	key := n.uuid.String() + n.textureVariant
	if n.textureKey != "" {
		key = n.textureKey
	}
	tx, isNew := n.renderer.getTexture(key, n.Image())

	if isNew {
		// LoadTextureFromImage already uploaded the complete image. Consume the
		// dirty flag so renderNode does not immediately upload the same pixels.
		n.dirty()
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
	n.textureVariant = ""
	n.textureKey = ""
	n.sharedTexture = false
}

func (n *node) SetImageResource(image image.Image, key string) {
	n.isDirty = false
	n.image = image
	n.textureVariant = ""
	n.textureKey = key
	n.sharedTexture = key != ""
}

func (n *node) UpdateImageResource(image image.Image, key string) {
	n.isDirty = true
	n.image = image
	n.textureVariant = ""
	n.textureKey = key
	n.sharedTexture = key != ""
}

func (n *node) SetAnimationFrame(frame image.Image, key string, index int) {
	n.image = frame
	n.isDirty = false
	n.textureVariant = fmt.Sprintf("/animation/%d", index)
	n.textureKey = key
	n.sharedTexture = key != ""
	if n.sharedTexture {
		return
	}
	if n.textureKeys == nil {
		n.textureKeys = make(map[string]struct{})
	}
	n.textureKeys[n.uuid.String()+n.textureVariant] = struct{}{}
}

func (n *node) ClearTextures() {
	if n.renderer.cache == nil {
		return
	}
	if !n.sharedTexture {
		n.renderer.cache.Remove(n.uuid.String())
		for key := range n.textureKeys {
			n.renderer.cache.Remove(key)
		}
	}
	n.textureKeys = nil
	n.textureVariant = ""
	n.textureKey = ""
	n.sharedTexture = false
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

func (n *node) setVisible(visible bool) { n.visible = visible }

func (n *node) parentNode() *node {
	if n.parent == nil {
		return n.renderer.rootNode
	}

	return n.parent
}

// setParent updates the private native hierarchy mirrored from retained state.
func (n *node) setParent(p *node) {
	if n == p {
		return
	}

	if n.parent != nil {
		n.parent.removeChild(n)
	}

	n.parent = p
	n.transformDirty = true

	if p != nil {
		n.parent.addChild(n)
	}
}

func (n *node) addChild(m *node) {
	if m == nil {
		return
	}

	n.children = append(n.children, m)
	n.childrenSorted = false
}

func (n *node) removeChild(m *node) {
	if m == nil {
		return
	}

	for idx := len(n.children) - 1; idx >= 0; idx-- {
		if n.children[idx] != m {
			continue
		}

		n.children = append(n.children[:idx], n.children[idx+1:]...)
		n.childrenSorted = false
	}
}

func (n *node) Children() []*node {
	return n.children
}

func (n *node) sortedChildren() []*node {
	if n.childrenSorted {
		return n.children
	}
	sort.SliceStable(n.children, func(i, j int) bool {
		return n.children[i].ZIndex() < n.children[j].ZIndex()
	})
	n.childrenSorted = true
	return n.children
}

// UpdateWorldMatrix recursively updates the children world matrices with this
// nodes world matrix
func (n *node) UpdateWorldMatrix(parent rl.Matrix, parentDirty bool) {
	dirty := parentDirty || n.transformDirty
	if dirty {
		n.world = rl.MatrixMultiply(n.local, parent)
		n.worldRotation = matrixRotation(n.world)
		n.transformDirty = false
	}

	for idx := range n.children {
		n.children[idx].UpdateWorldMatrix(n.world, dirty)
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
