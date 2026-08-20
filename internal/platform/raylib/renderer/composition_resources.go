package raylibRenderer

import (
	"fmt"
	"image"

	"github.com/gravestench/dark-magic/internal/presentation/render"
)

// animationPlayback joins a backend-neutral player with resolved frame images and their stable texture keys.
type animationPlayback struct {
	player       *render.AnimationPlayer
	frames       []image.Image
	keys         []string
	seekRevision uint64
}

// warmTexture performs the requested owner-thread upload without retaining semantic resource ownership.
func (b *compositionBackend) warmTexture(resource render.Resource) error {
	if resource.TextureKey == "" {
		return fmt.Errorf("warm texture key is empty")
	}

	b.renderer.getTexture(resource.TextureKey, resource.Payload.(image.Image))

	return nil
}

// createResource records semantic ownership without allocating native resources until a node actually uses it.
func (b *compositionBackend) createResource(id render.ResourceID, resource render.Resource) error {
	b.resources[id] = resource
	return nil
}

// updateResource replaces a streaming texture and updates every attached node while retaining compatible allocations.
func (b *compositionBackend) updateResource(id render.ResourceID, replacement render.Resource) error {
	resource, exists := b.resources[id]
	if !exists {
		return fmt.Errorf("resource %v does not exist", id)
	}

	if resource.Kind != render.ResourceTexture || replacement.Kind != render.ResourceTexture {
		return fmt.Errorf("resource %v is not an updateable texture", id)
	}

	previous := resource.Payload.(image.Image).Bounds().Size()
	next := replacement.Payload.(image.Image).Bounds().Size()
	b.resources[id] = replacement

	for nodeID, resourceID := range b.nodeResources {
		if resourceID != id {
			continue
		}

		node := b.nodes[nodeID]
		// Streaming frames retain one native texture unless dimensions change and require a different GPU allocation.
		if previous != next {
			node.ClearTextures()
		}

		node.UpdateImageResource(replacement.Payload.(image.Image), replacement.TextureKey)
	}

	return nil
}

// destroyResource closes a resource-specific palette effect before removing semantic resource ownership.
func (b *compositionBackend) destroyResource(id render.ResourceID) error {
	if _, exists := b.resources[id]; !exists {
		return fmt.Errorf("resource %v does not exist", id)
	}

	if effect := b.paletteEffects[id]; effect != nil {
		effect.close()
		delete(b.paletteEffects, id)
	}

	delete(b.resources, id)

	return nil
}

// attachAnimation resolves all frame resources before publishing playback, preventing partially attached animations.
func (b *compositionBackend) attachAnimation(id render.NodeID, node *node, resource render.Resource) error {
	animation := resource.Payload.(render.AnimationData)
	frames := make([]image.Image, len(animation.Frames))
	keys := make([]string, len(animation.Frames))

	for index, frameID := range animation.Frames {
		frame, exists := b.resources[frameID]
		if !exists {
			return fmt.Errorf("animation %v frame %d is unavailable", resource.ID, index)
		}

		frames[index] = frame.Payload.(image.Image)
		keys[index] = frame.TextureKey
	}

	player := render.NewAnimationPlayer(animation.Durations, animation.Loop)
	b.playbacks[id] = &animationPlayback{player: player, frames: frames, keys: keys}
	node.SetAnimationFrame(frames[0], keys[0], 0)

	return nil
}

// setAnimationFrame keeps animation changes on the node's stable texture-variant path.
func (b *compositionBackend) setAnimationFrame(node *node, frame image.Image, key string, index int) {
	node.SetAnimationFrame(frame, key, index)
}

// drawableImage resolves the initial image for each drawable resource kind without allocating native textures.
func (b *compositionBackend) drawableImage(resource render.Resource) (image.Image, error) {
	switch resource.Kind {
	case render.ResourceTexture:
		return resource.Payload.(image.Image), nil
	case render.ResourceAnimation:
		animation := resource.Payload.(render.AnimationData)

		frame, exists := b.resources[animation.Frames[0]]
		if !exists {
			return nil, fmt.Errorf("animation %v frame is unavailable", resource.ID)
		}

		return frame.Payload.(image.Image), nil
	case render.ResourceRenderTarget:
		target := resource.Payload.(render.RenderTargetData)
		return image.NewRGBA(image.Rect(0, 0, target.Width, target.Height)), nil
	default:
		return nil, fmt.Errorf("resource kind %q is not drawable", resource.Kind)
	}
}
