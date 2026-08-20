package raylibRenderer

import (
	"fmt"
	"image"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Shader returns the optional shader whose native mode brackets this node's draw call.
func (n *node) Shader() *rl.Shader {
	return n.shader
}

// ShaderTexture returns the optional auxiliary sampler uploaded within the node shader batch.
func (n *node) ShaderTexture() *rl.Texture2D {
	return n.shaderTexture
}

// ShaderTextureLocation returns the native uniform location paired with ShaderTexture.
func (n *node) ShaderTextureLocation() int32 {
	return n.shaderTextureLocation
}

// SetShader replaces the three shader associations together so drawing never observes a partially updated tuple.
func (n *node) SetShader(shader *rl.Shader, texture *rl.Texture2D, location int32) {
	n.shader, n.shaderTexture, n.shaderTextureLocation = shader, texture, location
}

// dirty consumes one pending pixel update so each changed image uploads at most once.
func (n *node) dirty() bool {
	if !n.isDirty {
		return false
	}

	n.isDirty = false

	return true
}

// Texture resolves the current stable cache key and suppresses a redundant upload after a cache miss created it.
func (n *node) Texture() rl.Texture2D {
	key := n.uuid.String() + n.textureVariant
	if n.textureKey != "" {
		key = n.textureKey
	}

	texture, isNew := n.renderer.getTexture(key, n.Image())
	if isNew {
		// LoadTextureFromImage uploaded the complete image, so renderNode must not upload identical pixels again.
		n.dirty()
	}

	return texture
}

// SetTexture inserts a caller-created native texture under this node's base identity using its current image weight.
func (n *node) SetTexture(texture rl.Texture2D) {
	key := n.uuid.String()
	bounds := n.Image().Bounds()
	weight := bounds.Dx() * bounds.Dy() * 4

	if err := n.renderer.cache.Insert(key, texture, weight); err != nil {
		n.renderer.logger.Error("caching texture", "key", key, "error", err)
	}
}

// Image lazily installs the historical 60x60 placeholder so every enabled node has drawable dimensions.
func (n *node) Image() image.Image {
	if n.image == nil {
		n.SetImage(defaultImage(60, 60))
	}

	return n.image
}

// SetImage assigns caller-owned pixels to this node's private texture namespace and requests one upload.
func (n *node) SetImage(value image.Image) {
	n.isDirty = true
	n.image = value
	n.textureVariant = ""
	n.textureKey = ""
	n.sharedTexture = false
}

// SetImageResource associates immutable pixels with a shared semantic key that the cache owns across nodes.
func (n *node) SetImageResource(value image.Image, key string) {
	n.isDirty = false
	n.image = value
	n.textureVariant = ""
	n.textureKey = key
	n.sharedTexture = key != ""
}

// UpdateImageResource retains a shared allocation but marks its changed pixels for one owner-thread upload.
func (n *node) UpdateImageResource(value image.Image, key string) {
	n.isDirty = true
	n.image = value
	n.textureVariant = ""
	n.textureKey = key
	n.sharedTexture = key != ""
}

// SetAnimationFrame selects a stable per-frame variant when no shared texture key is available.
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

// ClearTextures evicts only node-owned cache entries; shared semantic textures remain under composer ownership.
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
