package raylibRenderer

import (
	"image"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/google/uuid"
)

// newNode creates a visible local node with identity transforms and immediately attaches it to the scene root.
func (s *Service) newNode() *node {
	node := &node{
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
	node.setParent(s.rootNode)

	return node
}

// node mirrors one retained render node and owns its native texture associations and private hierarchy links.
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

// UUID returns the stable identity used to namespace node-owned texture cache entries.
func (n *node) UUID() uuid.UUID {
	return n.uuid
}

// Opacity returns the independent alpha multiplier applied at draw time.
func (n *node) Opacity() float32 {
	return n.opacity
}

// SetOpacity updates draw-time alpha without changing the RGB tint or texture contents.
func (n *node) SetOpacity(opacity float32) {
	n.opacity = opacity
}

// Tint returns the RGB multiplier whose alpha channel is replaced by Opacity during drawing.
func (n *node) Tint() rl.Color {
	return n.tint
}

// SetTint updates the RGB multiplier without marking texture pixels dirty.
func (n *node) SetTint(tint rl.Color) {
	n.tint = tint
}

// BlendMode returns the native blend bracket selected for this node.
func (n *node) BlendMode() rl.BlendMode {
	return n.blendMode
}

// SetBlendMode changes native draw state without modifying retained texture ownership.
func (n *node) SetBlendMode(mode rl.BlendMode) {
	n.blendMode = mode
}

// Clip returns the node-local clipping rectangle copied from retained state.
func (n *node) Clip() *rl.Rectangle {
	return n.clip
}

// SetClip replaces the clip pointer with backend-owned storage prepared by composition application.
func (n *node) SetClip(clip *rl.Rectangle) {
	n.clip = clip
}
